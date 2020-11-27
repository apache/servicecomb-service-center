/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/metrics"
	nf "github.com/apache/servicecomb-service-center/pkg/notify"
	"github.com/apache/servicecomb-service-center/pkg/signal"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/command"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	"github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	snf "github.com/apache/servicecomb-service-center/server/syncernotify"
	"net"
	"os"
	"time"
)

const defaultCollectPeriod = 30 * time.Second

var server ServiceCenterServer

func Run() {
	if err := command.ParseConfig(os.Args); err != nil {
		log.Fatal(err.Error(), err)
	}

	server.Run()
}

type endpoint struct {
	Host string
	Port string
}

type ServiceCenterServer struct {
	REST endpoint
	GRPC endpoint

	apiService          *APIServer
	notifyService       *nf.Service
	syncerNotifyService *snf.Service
}

func (s *ServiceCenterServer) Run() {
	s.initialize()

	s.startServices()

	signal.RegisterListener()

	s.waitForQuit()
}

func (s *ServiceCenterServer) waitForQuit() {
	err := <-s.apiService.Err()
	if err != nil {
		log.Errorf(err, "service center catch errors")
	}

	s.Stop()
}

func (s *ServiceCenterServer) initialize() {
	s.initEndpoints()
	// Metrics
	s.initMetrics()
	// SSL
	s.initSSL()
	// Datasource
	s.initDatasource()
	s.apiService = GetAPIServer()
	s.notifyService = notify.Center()
	s.syncerNotifyService = snf.GetSyncerNotifyCenter()
}

func (s *ServiceCenterServer) initEndpoints() {
	s.REST.Host = config.GetString("server.host", "", config.WithStandby("httpaddr"))
	s.REST.Port = config.GetString("server.port", "", config.WithStandby("httpport"))
	s.GRPC.Host = config.GetString("server.rpc.host", "", config.WithStandby("rpcaddr"))
	s.GRPC.Port = config.GetString("server.rpc.port", "", config.WithStandby("rpcport"))
}

func (s *ServiceCenterServer) initDatasource() {
	// init datasource
	kind := datasource.ImplName(config.GetString("registry.kind", "", config.WithStandby("registry_plugin")))
	if err := datasource.Init(datasource.Options{
		PluginImplName:    kind,
		InstanceTTL:       config.GetRegistry().InstanceTTL,
		SchemaEditable:    config.GetRegistry().SchemaEditable,
		CompactInterval:   config.GetRegistry().CompactInterval,
		CompactIndexDelta: config.GetRegistry().CompactIndexDelta,
	}); err != nil {
		log.Fatalf(err, "init datasource failed")
	}
}

func (s *ServiceCenterServer) initMetrics() {
	interval := config.GetDuration("metrics.interval", defaultCollectPeriod, config.WithENV("METRICS_INTERVAL"))
	if interval <= time.Second {
		interval = defaultCollectPeriod
	}
	var instance string
	if len(s.REST.Host) > 0 {
		instance = net.JoinHostPort(s.REST.Host, s.REST.Port)
	} else {
		if len(s.GRPC.Host) > 0 {
			instance = net.JoinHostPort(s.GRPC.Host, s.GRPC.Port)
		} else {
			log.Fatal("init metrics InstanceName failed", nil)
		}
	}

	if err := metrics.Init(metrics.Options{
		Interval:     interval,
		InstanceName: instance,
		SysMetrics: []string{
			"process_resident_memory_bytes",
			"process_cpu_seconds_total",
			"go_threads",
			"go_goroutines",
		},
	}); err != nil {
		log.Fatal("init metrics failed", err)
	}
}

func (s *ServiceCenterServer) initSSL() {
	if !config.GetSSL().SslEnabled {
		return
	}
	options := tlsconf.Options{
		Dir:              config.GetString("ssl.dir", "", config.WithENV("SSL_ROOT")),
		MinVersion:       config.GetString("ssl.minVersion", "TLSv1.2", config.WithStandby("ssl_min_version")),
		ClientMinVersion: config.GetString("ssl.client.minVersion", "", config.WithStandby("ssl_client_min_version")),
		VerifyPeer:       config.GetInt("ssl.verifyClient", 1, config.WithStandby("ssl_verify_client")) != 0,
		Ciphers:          config.GetString("ssl.ciphers", "", config.WithStandby("ssl_ciphers")),
		ClientCiphers:    config.GetString("ssl.client.ciphers", "", config.WithStandby("ssl_client_ciphers")),
	}
	if options.ClientMinVersion == "" {
		options.ClientMinVersion = options.MinVersion
	}
	if options.ClientCiphers == "" {
		options.ClientCiphers = options.Ciphers
	}
	if err := tlsconf.Init(options); err != nil {
		log.Fatal("init ssl failed", err)
	}
}

func (s *ServiceCenterServer) startServices() {
	// notifications
	s.notifyService.Start()

	// notify syncer
	syncerEnabled := config.GetBool("syncer.enabled", false)
	if syncerEnabled {
		s.syncerNotifyService.Start()
	}

	// load server plugins
	plugin.LoadPlugins()
	rbac.Init()
	if err := gov.Init(); err != nil {
		log.Fatal("init gov failed", err)
	}
	// check version
	if config.GetRegistry().SelfRegister {
		if err := datasource.Instance().UpgradeVersion(context.Background()); err != nil {
			os.Exit(1)
		}
	}
	// api service
	s.startAPIService()
}

func (s *ServiceCenterServer) startAPIService() {
	core.Instance.HostName = util.HostName()
	s.apiService.AddListener(REST, s.REST.Host, s.REST.Port)
	s.apiService.AddListener(RPC, s.REST.Host, s.GRPC.Port)
	s.apiService.Start()
}

func (s *ServiceCenterServer) Stop() {
	if s.apiService != nil {
		s.apiService.Stop()
	}

	if s.notifyService != nil {
		s.notifyService.Stop()
	}

	if s.syncerNotifyService != nil {
		s.syncerNotifyService.Stop()
	}

	gopool.CloseAndWait()

	log.Warnf("service center stopped")
	log.Sync()
}

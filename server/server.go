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
	"crypto/tls"
	"os"

	"github.com/apache/servicecomb-service-center/datasource"
	nf "github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/signal"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/command"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	"github.com/apache/servicecomb-service-center/server/service/grc"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/foundation/gopool"
)

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
	Endpoint    endpoint
	APIServer   *APIServer
	eventCenter *nf.BusService
}

func (s *ServiceCenterServer) Run() {
	s.initialize()

	s.startServices()

	signal.RegisterListener()

	s.waitForQuit()
}

func (s *ServiceCenterServer) waitForQuit() {
	err := <-s.APIServer.Err()
	if err != nil {
		log.Error("service center catch errors", err)
	}

	s.Stop()
}

func (s *ServiceCenterServer) initialize() {
	s.initEndpoints()
	// SSL
	s.initSSL()
	// Datasource
	s.initDatasource()
	s.APIServer = GetAPIServer()
	s.eventCenter = event.Center()
}

func (s *ServiceCenterServer) initEndpoints() {
	s.Endpoint.Host = config.GetString("server.host", "", config.WithStandby("httpaddr"))
	s.Endpoint.Port = config.GetString("server.port", "", config.WithStandby("httpport"))
}

func (s *ServiceCenterServer) initDatasource() {
	// init datasource
	kind := config.GetString("registry.kind", "", config.WithStandby("registry_plugin"))
	tlsConfig, err := getDatasourceTLSConfig()
	if err != nil {
		log.Fatal("get datasource tlsConfig failed", err)
	}
	if err := datasource.Init(datasource.Options{
		Kind:       kind,
		Logger:     log.Logger,
		SslEnabled: config.GetSSL().SslEnabled,
		TLSConfig:  tlsConfig,
		ConnectedFunc: func() {
			err := alarm.Clear(alarm.IDBackendConnectionRefuse)
			if err != nil {
				log.Error("", err)
			}
		},
		ErrorFunc: func(err error) {
			if err == nil {
				return
			}
			err = alarm.Raise(alarm.IDBackendConnectionRefuse, alarm.AdditionalContext("%v", err))
			if err != nil {
				log.Error("", err)
			}
		},
		EnableCache: config.GetRegistry().EnableCache,
		InstanceTTL: config.GetRegistry().InstanceTTL,
	}); err != nil {
		log.Fatal("init datasource failed", err)
	}
}

func getDatasourceTLSConfig() (*tls.Config, error) {
	if config.GetSSL().SslEnabled {
		return tlsconf.ClientConfig()
	}
	return nil, nil
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
	s.eventCenter.Start()

	// load server plugins
	plugin.LoadPlugins()
	rbac.Init()
	if err := grc.Init(); err != nil {
		log.Fatal("init gov failed", err)
	}
	// check version
	if config.GetRegistry().SelfRegister {
		if err := datasource.GetSCManager().UpgradeVersion(context.Background()); err != nil {
			os.Exit(1)
		}
	}
	// api service
	s.startAPIService()
}

func (s *ServiceCenterServer) startAPIService() {
	s.APIServer.Listen(s.Endpoint.Host, s.Endpoint.Port)
	s.APIServer.Start()
}

func (s *ServiceCenterServer) Stop() {
	if s.APIServer != nil {
		s.APIServer.Stop()
	}

	if s.eventCenter != nil {
		s.eventCenter.Stop()
	}

	gopool.CloseAndWait()

	log.Warn("service center stopped")
	log.Flush()
}

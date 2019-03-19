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

import _ "github.com/apache/servicecomb-service-center/server/service/event"
import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	nf "github.com/apache/servicecomb-service-center/pkg/notify"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/mux"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
	"os"
	"time"
)

const buildin = "buildin"

type ServiceCenterServer struct {
	apiService    *APIServer
	notifyService *nf.NotifyService
	cacheService  *backend.KvStore
	goroutine     *gopool.Pool
}

func (s *ServiceCenterServer) Run() {
	s.initialize()

	s.startServices()

	s.waitForQuit()
}

func (s *ServiceCenterServer) waitForQuit() {
	err := <-s.apiService.Err()
	if err != nil {
		log.Errorf(err, "service center catch errors")
	}

	s.Stop()
}

func (s *ServiceCenterServer) needUpgrade() bool {
	err := LoadServerVersion()
	if err != nil {
		log.Errorf(err, "check version failed, can not load the system config")
		return false
	}

	update := !serviceUtil.VersionMatchRule(core.ServerInfo.Version,
		fmt.Sprintf("%s+", version.Ver().Version))
	if !update && version.Ver().Version != core.ServerInfo.Version {
		log.Warnf(
			"there is a higher version '%s' in cluster, now running '%s' version may be incompatible",
			core.ServerInfo.Version, version.Ver().Version)
	}

	return update
}

func (s *ServiceCenterServer) loadOrUpgradeServerVersion() {
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		log.Errorf(err, "wait for server ready failed")
		os.Exit(1)
	}
	if s.needUpgrade() {
		core.ServerInfo.Version = version.Ver().Version

		if err := UpgradeServerVersion(); err != nil {
			log.Errorf(err, "upgrade server version failed")
			os.Exit(1)
		}
	}
	lock.Unlock()
}

func (s *ServiceCenterServer) compactBackendService() {
	delta := core.ServerInfo.Config.CompactIndexDelta
	if delta <= 0 || len(core.ServerInfo.Config.CompactInterval) == 0 {
		return
	}
	interval, err := time.ParseDuration(core.ServerInfo.Config.CompactInterval)
	if err != nil {
		log.Errorf(err, "invalid compact interval %s, reset to default interval 12h", core.ServerInfo.Config.CompactInterval)
		interval = 12 * time.Hour
	}
	s.goroutine.Do(func(ctx context.Context) {
		log.Infof("enabled the automatic compact mechanism, compact once every %s, reserve %d",
			core.ServerInfo.Config.CompactInterval, delta)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				lock, err := mux.Try(mux.GLOBAL_LOCK)
				if lock == nil {
					log.Errorf(err, "can not compact backend by this service center instance now")
					continue
				}

				backend.Registry().Compact(ctx, delta)

				lock.Unlock()
			}
		}
	})
}

func (s *ServiceCenterServer) initialize() {
	s.cacheService = backend.Store()
	s.apiService = GetAPIServer()
	s.notifyService = notify.NotifyCenter()
	s.goroutine = gopool.New(context.Background())
}

func (s *ServiceCenterServer) startServices() {
	// notifications
	s.notifyService.Start()

	// load server plugins
	plugin.LoadPlugins()

	// check version
	if core.ServerInfo.Config.SelfRegister {
		s.loadOrUpgradeServerVersion()
	}

	// cache mechanism
	s.cacheService.Run()
	<-s.cacheService.Ready()

	// compact backend automatically
	if buildin != beego.AppConfig.DefaultString("registry_plugin", buildin) {
		s.compactBackendService()
	}

	// api service
	s.startApiService()
}

func (s *ServiceCenterServer) startApiService() {
	restIp := beego.AppConfig.String("httpaddr")
	restPort := beego.AppConfig.String("httpport")
	rpcIp := beego.AppConfig.DefaultString("rpcaddr", "")
	rpcPort := beego.AppConfig.DefaultString("rpcport", "")

	host, err := os.Hostname()
	if err != nil {
		host = restIp
		log.Errorf(err, "parse hostname failed")
	}
	core.Instance.HostName = host
	s.apiService.AddListener(REST, restIp, restPort)
	s.apiService.AddListener(RPC, rpcIp, rpcPort)
	s.apiService.Start()
}

func (s *ServiceCenterServer) Stop() {
	if s.apiService != nil {
		s.apiService.Stop()
	}

	if s.notifyService != nil {
		s.notifyService.Stop()
	}

	if s.cacheService != nil {
		s.cacheService.Stop()
	}

	s.goroutine.Close(true)

	gopool.CloseAndWait()

	backend.Registry().Close()

	log.Warnf("service center stopped")
	log.Sync()
}

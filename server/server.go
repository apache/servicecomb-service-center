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
	"fmt"
	"os"
	"time"

	//plugin
	_ "github.com/apache/servicecomb-service-center/server/service/event"

	nf "github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/mux"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
)

const buildin = "buildin"

var server ServiceCenterServer

func Run() {
	server.Run()
}

type ServiceCenterServer struct {
	apiService   *APIServer
	eventCenter  *nf.BusService
	cacheService *backend.KvStore
	goroutine    *gopool.Pool
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
	lock, err := mux.Lock(mux.GlobalLock)

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
	err = lock.Unlock()
	if err != nil {
		log.Error("", err)
	}
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
				lock, err := mux.Try(mux.GlobalLock)
				if err != nil {
					log.Errorf(err, "can not compact backend by this service center instance now")
					continue
				}

				err = backend.Registry().Compact(ctx, delta)
				if err != nil {
					log.Error("", err)
				}

				if err := lock.Unlock(); err != nil {
					log.Error("", err)
				}
			}
		}
	})
}

func (s *ServiceCenterServer) initialize() {
	s.cacheService = backend.Store()
	s.apiService = GetAPIServer()
	s.eventCenter = event.Center()
	s.goroutine = gopool.New(context.Background())
}

func (s *ServiceCenterServer) startServices() {
	// notifications
	s.eventCenter.Start()

	// load server plugins
	plugin.LoadPlugins()
	rbac.Init()
	// check version
	if core.ServerInfo.Config.SelfRegister {
		s.loadOrUpgradeServerVersion()
	}

	// cache mechanism
	s.cacheService.Run()
	<-s.cacheService.Ready()

	if buildin != beego.AppConfig.DefaultString("registry_plugin", buildin) {
		// compact backend automatically
		s.compactBackendService()
	}

	// api service
	s.startAPIService()
}

func (s *ServiceCenterServer) startAPIService() {
	restIP := beego.AppConfig.String("httpaddr")
	restPort := beego.AppConfig.String("httpport")
	rpcIP := beego.AppConfig.DefaultString("rpcaddr", "")
	rpcPort := beego.AppConfig.DefaultString("rpcport", "")

	host, err := os.Hostname()
	if err != nil {
		host = restIP
		log.Errorf(err, "parse hostname failed")
	}
	core.Instance.HostName = host
	s.apiService.AddListener(REST, restIP, restPort)
	s.apiService.AddListener(RPC, rpcIP, rpcPort)
	s.apiService.Start()
}

func (s *ServiceCenterServer) Stop() {
	if s.apiService != nil {
		s.apiService.Stop()
	}

	if s.eventCenter != nil {
		s.eventCenter.Stop()
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

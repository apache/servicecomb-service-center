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

import _ "github.com/apache/incubator-servicecomb-service-center/server/service/event"
import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	st "github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	"github.com/apache/incubator-servicecomb-service-center/server/mux"
	nf "github.com/apache/incubator-servicecomb-service-center/server/service/notification"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
	"os"
	"strings"
	"time"
)

var (
	server *ServiceCenterServer
)

func init() {
	server = &ServiceCenterServer{
		store:         st.Store(),
		notifyService: nf.GetNotifyService(),
		apiServer:     GetAPIServer(),
	}
}

type ServiceCenterServer struct {
	apiServer     *APIServer
	notifyService *nf.NotifyService
	store         *st.KvStore
}

func (s *ServiceCenterServer) Run() {
	s.initialize()

	s.startNotifyService()

	s.startApiServer()

	s.waitForQuit()
}

func (s *ServiceCenterServer) waitForQuit() {
	var err error
	select {
	case err = <-s.apiServer.Err():
	case err = <-s.notifyService.Err():
	}
	if err != nil {
		util.Logger().Errorf(err, "service center catch errors")
	}

	s.Stop()

	util.Logger().Warn("service center quit", nil)
}

func (s *ServiceCenterServer) needUpgrade() bool {
	if core.ServerInfo.Version == "0" {
		err := core.LoadServerInformation()
		if err != nil {
			util.Logger().Errorf(err, "check version failed, can not load the system config")
			return false
		}
	}
	return !serviceUtil.VersionMatchRule(core.ServerInfo.Version,
		fmt.Sprintf("%s+", version.Ver().Version))
}

func (s *ServiceCenterServer) initialize() {
	// check version
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	defer lock.Unlock()
	if err != nil {
		util.Logger().Errorf(err, "wait for server ready failed")
		os.Exit(1)
	}
	if s.needUpgrade() {
		core.UpgradeServerVersion()
	}

	// cache mechanism
	s.store.Run()
	<-s.store.Ready()

	// auto compact backend
	s.autoCompactBackend()
}

func (s *ServiceCenterServer) autoCompactBackend() {
	delta := core.ServerInfo.Config.CompactIndexDelta
	if delta <= 0 || len(core.ServerInfo.Config.CompactInterval) == 0 {
		return
	}
	interval, err := time.ParseDuration(core.ServerInfo.Config.CompactInterval)
	if err != nil {
		util.Logger().Errorf(err, "invalid compact interval %s, reset to default interval 12h", core.ServerInfo.Config.CompactInterval)
		interval = 12 * time.Hour
	}
	util.Go(func(stopCh <-chan struct{}) {
		util.Logger().Infof("start the automatic compact mechanism, compact once every %s",
			core.ServerInfo.Config.CompactInterval)
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(interval):
				lock, err := mux.Try(mux.GLOBAL_LOCK)
				if lock == nil {
					util.Logger().Warnf(err, "can not compact backend by this service center instance now")
					continue
				}

				backend.Registry().Compact(context.Background(), delta)

				lock.Unlock()
			}
		}
	})
}

func (s *ServiceCenterServer) startNotifyService() {
	s.notifyService.Config = nf.NotifyServiceConfig{
		AddTimeout:    30 * time.Second,
		NotifyTimeout: 30 * time.Second,
		MaxQueue:      100,
	}
	s.notifyService.Start()
}

func (s *ServiceCenterServer) startApiServer() {
	restIp := beego.AppConfig.String("httpaddr")
	restPort := beego.AppConfig.String("httpport")
	rpcIp := beego.AppConfig.DefaultString("rpcaddr", "")
	rpcPort := beego.AppConfig.DefaultString("rpcport", "")
	cmpName := core.ServerInfo.Config.LoggerName
	hostName := fmt.Sprintf("%s_%s", cmpName, strings.Replace(util.GetLocalIP(), ".", "_", -1))

	s.apiServer.HostName = hostName
	s.addEndpoint(REST, restIp, restPort)
	s.addEndpoint(RPC, rpcIp, rpcPort)
	s.apiServer.Start()
}

func (s *ServiceCenterServer) addEndpoint(t APIType, ip, port string) {
	if s.apiServer.Endpoints == nil {
		s.apiServer.Endpoints = map[APIType]string{}
	}
	if len(ip) == 0 {
		return
	}
	address := util.StringJoin([]string{ip, port}, ":")
	if core.ServerInfo.Config.SslEnabled {
		address += "?sslEnabled=true"
	}
	s.apiServer.Endpoints[t] = fmt.Sprintf("%s://%s", t, address)
}

func (s *ServiceCenterServer) Stop() {
	if s.apiServer != nil {
		s.apiServer.Stop()
	}

	if s.notifyService != nil {
		s.notifyService.Stop()
	}

	if s.store != nil {
		s.store.Stop()
	}

	util.GoCloseAndWait()

	backend.Registry().Close()
}

func Run() {
	server.Run()
}

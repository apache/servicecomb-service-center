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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/mux"
	nf "github.com/apache/incubator-servicecomb-service-center/server/service/notification"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	server *ServiceCenterServer
)

func init() {
	server = &ServiceCenterServer{
		store:         backend.Store(),
		notifyService: nf.GetNotifyService(),
		apiServer:     GetAPIServer(),
		goroutine:     util.NewGo(context.Background()),
	}
}

type ServiceCenterServer struct {
	apiServer     *APIServer
	notifyService *nf.NotifyService
	store         *backend.KvStore
	goroutine     *util.GoRoutine
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

	util.Logger().Debugf("service center stopped")
}

func (s *ServiceCenterServer) LoadServerInformation() error {
	resp, err := backend.Registry().Do(context.Background(),
		registry.GET, registry.WithStrKey(core.GetServerInfoKey()))
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}

	err = json.Unmarshal(resp.Kvs[0].Value, core.ServerInfo)
	if err != nil {
		util.Logger().Errorf(err, "load system config failed, maybe incompatible")
		return nil
	}
	return nil
}

func (s *ServiceCenterServer) UpgradeServerVersion() error {
	core.ServerInfo.Version = version.Ver().Version

	bytes, err := json.Marshal(core.ServerInfo)
	if err != nil {
		return err
	}
	_, err = backend.Registry().Do(context.Background(),
		registry.PUT, registry.WithStrKey(core.GetServerInfoKey()), registry.WithValue(bytes))
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceCenterServer) needUpgrade() bool {
	if core.ServerInfo.Version == "0" {
		err := s.LoadServerInformation()
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
		s.UpgradeServerVersion()
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
	s.goroutine.Do(func(ctx context.Context) {
		util.Logger().Infof("enabled the automatic compact mechanism, compact once every %s, reserve %d",
			core.ServerInfo.Config.CompactInterval, delta)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				lock, err := mux.Try(mux.GLOBAL_LOCK)
				if lock == nil {
					util.Logger().Warnf(err, "can not compact backend by this service center instance now")
					continue
				}

				backend.Registry().Compact(ctx, delta)

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

	s.apiServer.HostName = fmt.Sprintf("%s_%s", core.ServerInfo.Config.LoggerName,
		strings.Replace(restIp, ".", "_", -1))
	s.addEndpoint(REST, restIp, restPort)
	s.addEndpoint(RPC, rpcIp, rpcPort)
	s.apiServer.Start()
}

func (s *ServiceCenterServer) addEndpoint(t APIType, ip, port string) {
	if s.apiServer.Endpoints == nil {
		s.apiServer.Listeners = map[APIType]string{}
		s.apiServer.Endpoints = map[APIType]string{}
	}
	if len(ip) == 0 {
		return
	}
	address := fmt.Sprintf("%s://%s/", t, net.JoinHostPort(url.PathEscape(ip), port))
	if core.ServerInfo.Config.SslEnabled {
		address += "?sslEnabled=true"
	}
	s.apiServer.Listeners[t] = net.JoinHostPort(ip, port)
	s.apiServer.Endpoints[t] = address
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

	s.goroutine.Close(true)
}

func Run() {
	server.Run()
}

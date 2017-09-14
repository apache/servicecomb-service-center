//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package server

import _ "github.com/ServiceComb/service-center/server/service/event"
import _ "github.com/ServiceComb/service-center/server/rest/handlers"
import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/common"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	"github.com/ServiceComb/service-center/server/core/registry"
	st "github.com/ServiceComb/service-center/server/core/registry/store"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/version"
	"github.com/astaxie/beego"
	"os"
	"runtime"
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
	util.Logger().Infof("service center have running simultaneously with %d CPU cores", runtime.GOMAXPROCS(0))

	s.waitForReady()

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
	if core.GetSystemConfig() == nil {
		err := core.LoadSystemConfig()
		if err != nil {
			util.Logger().Errorf(err, "check version failed, can not load the system config")
			return false
		}
	}
	return !serviceUtil.VersionMatchRule(core.GetSystemConfig().Version,
		fmt.Sprintf("%s+", version.Ver().Version))
}

func (s *ServiceCenterServer) waitForReady() {
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		util.Logger().Errorf(err, "wait for server ready failed")
		os.Exit(1)
	}
	if s.needUpgrade() {
		core.UpgradeSystemConfig()
	}
	lock.Unlock()

	s.store.Run()
	<-s.store.Ready()
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
	cmpName := beego.AppConfig.String("ComponentName")
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
	if common.GetServerSSLConfig().SSLEnabled {
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

	registry.GetRegisterCenter().Close()
}

func Run() {
	server.Run()
}

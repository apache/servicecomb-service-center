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
	"github.com/apache/servicecomb-service-center/datasource"
	//plugin
	_ "github.com/apache/servicecomb-service-center/server/service/event"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
)
import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	nf "github.com/apache/servicecomb-service-center/pkg/notify"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/astaxie/beego"
	"os"
)

var server ServiceCenterServer

func Run() {
	server.Run()
}

type ServiceCenterServer struct {
	apiService    *APIServer
	notifyService *nf.Service
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

func (s *ServiceCenterServer) initialize() {
	s.apiService = GetAPIServer()
	s.notifyService = notify.GetNotifyCenter()
}

func (s *ServiceCenterServer) startServices() {
	// notifications
	s.notifyService.Start()

	// load server plugins
	plugin.LoadPlugins()
	rbac.Init()
	// check version
	if core.ServerInfo.Config.SelfRegister {
		if err := datasource.Instance().UpgradeVersion(context.Background()); err != nil {
			os.Exit(1)
		}
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

	if s.notifyService != nil {
		s.notifyService.Stop()
	}

	gopool.CloseAndWait()

	log.Warnf("service center stopped")
	log.Sync()
}

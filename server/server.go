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

import (
	"fmt"
	"github.com/astaxie/beego"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)
import (
	"context"
	"github.com/ServiceComb/service-center/pkg/common"
	"github.com/ServiceComb/service-center/server/api"
	"github.com/ServiceComb/service-center/server/core/registry"
	st "github.com/ServiceComb/service-center/server/core/registry/store"
	rs "github.com/ServiceComb/service-center/server/rest"
	"github.com/ServiceComb/service-center/server/service"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	"github.com/ServiceComb/service-center/util"
	"runtime"
)

var (
	apiServer     *api.APIServer
	notifyService *nf.NotifyService
	store         *st.KvStore
	exit          chan struct{}
)

const CLEAN_UP_TIMEOUT = 5

func init() {
	util.LOGGER.Infof("service center have running simultaneously with %d CPU cores", runtime.GOMAXPROCS(0))

	exit = make(chan struct{})

	store = st.Store()

	notifyService = nf.GetNotifyService()

	apiServer = api.GetAPIServer()

	rs.ServiceAPI, rs.InstanceAPI, rs.GovernServiceAPI = service.AssembleResources()

	go handleSignal()
}

func Run() {
	go startStoreDeamon()

	go startNotifyService()

	go startApiServer()

	waitForQuit()
}

func handleSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	signal.Ignore(syscall.SIGQUIT) // when uses jstack to dump stack

	s := <-sc
	util.LOGGER.Warnf(nil, "Caught signal '%v', now service center quit...", s)

	if apiServer != nil {
		apiServer.Stop()
	}

	if store != nil {
		store.Stop()
	}

	if notifyService != nil {
		notifyService.Stop()
	}

	registry.GetRegisterCenter().Close()

	util.GoCloseAndWait()

	close(exit)
}

func waitForQuit() {
	var err error
	select {
	case err = <-apiServer.Err():
	case err = <-notifyService.Err():
	}
	if err != nil {
		util.LOGGER.Errorf(err, "service center catch errors, %s", err.Error())
	}
	util.LOGGER.Warnf(nil, "waiting for %ds to clean up resources...", CLEAN_UP_TIMEOUT)
	select {
	case <-exit:
	case <-time.After(CLEAN_UP_TIMEOUT * time.Second):
	}
	util.LOGGER.Warn("service center quit", nil)
}

func autoCompact() {
	compactTicker := time.NewTicker(time.Minute * 5)
	defer compactTicker.Stop()
	for t := range compactTicker.C {
		util.LOGGER.Debug(fmt.Sprintf("Compact at %s", t))
		registry.GetRegisterCenter().CompactCluster(context.TODO())
	}
}

func startStoreDeamon() {
	store.Run()
}

func startNotifyService() {
	notifyService.Config = nf.NotifyServiceConfig{
		AddTimeout:    30 * time.Second,
		NotifyTimeout: 30 * time.Second,
		MaxQueue:      100,
	}
	notifyService.Start()
}

func startApiServer() {
	sslMode := common.GetServerSSLConfig().SSLEnabled
	verifyClient := common.GetServerSSLConfig().VerifyClient
	restIp := beego.AppConfig.String("httpaddr")
	restPort := beego.AppConfig.String("httpport")
	grpcIp := beego.AppConfig.DefaultString("grpcaddr", "")
	grpcPort := beego.AppConfig.DefaultString("grpcport", "")
	cmpName := beego.AppConfig.String("ComponentName")
	hostName := fmt.Sprintf("%s_%s", cmpName, strings.Replace(util.GetLocalIP(), ".", "_", -1))
	util.LOGGER.Warnf(nil, "Local listen address: %s:%s, host: %s.", restIp, restPort, hostName)

	eps := map[api.APIType]string{}
	if len(restIp) > 0 && len(restPort) > 0 {
		eps[api.REST] = strings.Join([]string{restIp, restPort}, ":")
	}
	if len(grpcIp) > 0 && len(grpcPort) > 0 {
		eps[api.GRPC] = strings.Join([]string{grpcIp, grpcPort}, ":")
	}
	apiServer.Config = api.APIServerConfig{
		HostName:     hostName,
		Endpoints:    eps,
		SSL:          sslMode,
		VerifyClient: verifyClient,
	}
	apiServer.Start()
}

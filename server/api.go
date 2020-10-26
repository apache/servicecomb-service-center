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
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/grace"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/config"
	rs "github.com/apache/servicecomb-service-center/server/rest"
	"github.com/apache/servicecomb-service-center/server/service"
	"net"
	"strconv"
	"time"
)

var apiServer *APIServer

func init() {
	InitAPI()

	apiServer = &APIServer{
		isClose:   true,
		err:       make(chan error, 1),
		goroutine: gopool.New(context.Background()),
	}
}

func InitAPI() {
	core.ServiceAPI, core.InstanceAPI = service.AssembleResources()
}

type APIType int64

func (t APIType) String() string {
	switch t {
	case RPC:
		return "grpc" // support grpc
	case REST:
		return "rest"
	default:
		return "SCHEME" + strconv.Itoa(int(t))
	}
}

type APIServer struct {
	Listeners map[APIType]string

	restSrv   *rest.Server
	isClose   bool
	forked    bool
	err       chan error
	goroutine *gopool.Pool
}

const (
	RPC  APIType = 0
	REST APIType = 1
)

func (s *APIServer) Err() <-chan error {
	return s.err
}

func (s *APIServer) graceDone() {
	grace.Before(s.MarkForked)
	grace.After(s.Stop)
	if err := grace.Done(); err != nil {
		log.Errorf(err, "server reload failed")
	}
}

func (s *APIServer) MarkForked() {
	s.forked = true
}

func (s *APIServer) AddListener(t APIType, ip, port string) {
	if s.Listeners == nil {
		s.Listeners = map[APIType]string{}
	}
	if len(ip) == 0 {
		return
	}
	s.Listeners[t] = net.JoinHostPort(ip, port)
}

func (s *APIServer) populateEndpoint(t APIType, ipPort string) {
	if len(ipPort) == 0 {
		return
	}
	address := fmt.Sprintf("%s://%s/", t, ipPort)
	if config.ServerInfo.Config.SslEnabled {
		address += "?sslEnabled=true"
	}
	core.Instance.Endpoints = append(core.Instance.Endpoints, address)
}

func (s *APIServer) startRESTServer() (err error) {
	addr, ok := s.Listeners[REST]
	if !ok {
		return
	}
	s.restSrv, err = rs.NewServer(addr)
	if err != nil {
		return
	}
	log.Infof("listen address: %s://%s", REST, s.restSrv.Listener.Addr().String())

	s.populateEndpoint(REST, s.restSrv.Listener.Addr().String())

	s.goroutine.Do(func(_ context.Context) {
		err := s.restSrv.Serve()
		if s.isClose {
			return
		}
		log.Errorf(err, "error to start REST API server %s", addr)
		s.err <- err
	})
	return
}

func (s *APIServer) Start() {
	if !s.isClose {
		return
	}
	s.isClose = false

	core.Instance.Endpoints = nil

	err := s.startRESTServer()
	if err != nil {
		s.err <- err
		return
	}

	s.graceDone()

	defer log.Info("api server is ready")

	if !config.ServerInfo.Config.SelfRegister {
		log.Warnf("self register disabled")
		return
	}

	// 自注册
	s.selfRegister()
}

func (s *APIServer) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	if !s.forked && config.ServerInfo.Config.SelfRegister {
		s.selfUnregister()
	}

	if s.restSrv != nil {
		s.restSrv.Shutdown()
	}

	close(s.err)

	s.goroutine.Close(true)

	log.Info("api server stopped")
}

func (s *APIServer) selfRegister() {
	err := datasource.Instance().SelfRegister(context.Background())
	if err != nil {
		s.err <- err
		return
	}
}

func (s *APIServer) selfUnregister() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := datasource.Instance().SelfUnregister(ctx); err != nil {
		log.Error("stop registry engine failed", err)
	}
}

func GetAPIServer() *APIServer {
	return apiServer
}

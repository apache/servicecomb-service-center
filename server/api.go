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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/grace"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	rs "github.com/apache/incubator-servicecomb-service-center/server/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/rpc"
	"github.com/apache/incubator-servicecomb-service-center/server/service"
	"golang.org/x/net/context"
	"net"
	"strconv"
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
	rpcSrv    *rpc.Server
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
	if core.ServerInfo.Config.SslEnabled {
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

func (s *APIServer) startRPCServer() (err error) {
	addr, ok := s.Listeners[RPC]
	if !ok {
		return
	}

	s.rpcSrv, err = rpc.NewServer(addr)
	if err != nil {
		return
	}
	log.Infof("listen address: %s://%s", RPC, s.rpcSrv.Listener.Addr().String())

	s.populateEndpoint(RPC, s.rpcSrv.Listener.Addr().String())

	s.goroutine.Do(func(_ context.Context) {
		err := s.rpcSrv.Serve()
		if s.isClose {
			return
		}
		log.Errorf(err, "error to start RPC API server %s", addr)
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

	err = s.startRPCServer()
	if err != nil {
		s.err <- err
		return
	}

	s.graceDone()

	defer log.Info("api server is ready")

	if !core.ServerInfo.Config.SelfRegister {
		log.Warnf("self register disabled")
		return
	}

	// 自注册
	err = backend.RegistryEngine().Start()
	if err != nil {
		s.err <- err
		return
	}
}

func (s *APIServer) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	if !s.forked && core.ServerInfo.Config.SelfRegister {
		backend.RegistryEngine().Stop()
	}

	if s.restSrv != nil {
		s.restSrv.Shutdown()
	}

	if s.rpcSrv != nil {
		s.rpcSrv.GracefulStop()
	}

	close(s.err)

	s.goroutine.Close(true)

	log.Info("api server stopped.")
}

func GetAPIServer() *APIServer {
	return apiServer
}

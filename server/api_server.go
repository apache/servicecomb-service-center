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
	"net"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/grace"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/metrics"
	rs "github.com/apache/servicecomb-service-center/server/rest"
	"github.com/apache/servicecomb-service-center/server/service/registry"
	"github.com/go-chassis/foundation/gopool"
)

var apiServer *APIServer

func init() {
	apiServer = &APIServer{
		isClose:   true,
		err:       make(chan error, 1),
		goroutine: gopool.New(gopool.Configure().Workers(5)),
	}
}

type APIServer struct {
	HostPort   string
	HTTPServer *rest.Server

	isClose   bool
	forked    bool
	err       chan error
	goroutine *gopool.Pool
}

func (s *APIServer) Err() <-chan error {
	return s.err
}

func (s *APIServer) graceDone() {
	grace.Before(s.MarkForked)
	grace.After(s.Stop)
	if err := grace.Done(); err != nil {
		log.Error("server reload failed", err)
	}
}

func (s *APIServer) MarkForked() {
	s.forked = true
}

func (s *APIServer) Listen(ip, port string) {
	if len(ip) == 0 {
		return
	}
	s.HostPort = net.JoinHostPort(ip, port)
}

func (s *APIServer) populateEndpoint(ipPort string) {
	if len(ipPort) == 0 {
		return
	}
	address := fmt.Sprintf("rest://%s/", ipPort)
	if config.GetSSL().SslEnabled {
		address += "?sslEnabled=true"
	}
	core.Instance.Endpoints = append(core.Instance.Endpoints, address)
}

func (s *APIServer) serve() (err error) {
	s.HTTPServer, err = rs.NewServer(s.HostPort)
	if err != nil {
		return
	}
	log.Info(fmt.Sprintf("listen address: rest://%s", s.HTTPServer.Listener.Addr().String()))

	s.populateEndpoint(s.HTTPServer.Listener.Addr().String())

	s.goroutine.Do(func(_ context.Context) {
		err := s.HTTPServer.Serve()
		if s.isClose {
			return
		}
		log.Error(fmt.Sprintf("error to serve %s", s.HostPort), err)
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

	err := s.serve()
	if err != nil {
		s.err <- err
		return
	}

	s.graceDone()

	defer log.Info("api server is ready")

	if !config.GetRegistry().SelfRegister {
		log.Warn("self register disabled")
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

	if !s.forked && config.GetRegistry().SelfRegister {
		s.selfUnregister()
	}

	if s.HTTPServer != nil {
		s.HTTPServer.Shutdown()
	}

	close(s.err)

	s.goroutine.Close(true)

	log.Info("api server stopped")
}

func (s *APIServer) selfRegister() {
	err := registry.SelfRegister(context.Background())
	if err != nil {
		s.err <- err
		return
	}
	// report the metrics
	metrics.ReportScInstance()
}

func (s *APIServer) selfUnregister() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := registry.SelfUnregister(ctx); err != nil {
		log.Error("stop registry engine failed", err)
	}
}

func GetAPIServer() *APIServer {
	return apiServer
}

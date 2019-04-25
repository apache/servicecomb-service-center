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
package syncer

import (
	"context"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/datacenter"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	"github.com/apache/servicecomb-service-center/syncer/notify"
	"github.com/apache/servicecomb-service-center/syncer/peer"
	"github.com/apache/servicecomb-service-center/syncer/pkg/events"
	"github.com/apache/servicecomb-service-center/syncer/pkg/syssig"
	"github.com/apache/servicecomb-service-center/syncer/pkg/ticker"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
)

func tickHandler(ctx context.Context) {
	events.Dispatch(events.NewContextEvent(notify.EventTicker, ctx))
}

type Server struct {
	conf   *config.Config
	ending chan struct{}

	tick   *ticker.TaskTicker
	store  datacenter.Store
	agent  *peer.Agent
	broker *grpc.Broker
}

func NewServer(conf *config.Config) *Server {
	return &Server{
		conf:   conf,
		ending: make(chan struct{}),
	}
}

func (s *Server) Run(ctx context.Context) {
	s.initPlugin()

	if err := s.initialization(); err != nil {
		log.Errorf(err, "syncer server initialization faild: %s")
		return
	}

	s.eventListen()

	s.startServers(ctx)

	s.waitQuit(ctx)
}

func (s *Server) Stop() {
	if s.tick != nil {
		s.tick.Stop()
	}

	if s.agent != nil {
		s.agent.DeregisterEventHandler(s)
		s.agent.Shutdown()
	}

	if s.broker != nil {
		s.broker.Stop()
	}

	if s.store != nil {
		s.store.Stop()
	}

	events.Clean()
	gopool.CloseAndWait()
}

func (s *Server) initPlugin() {
	plugins.SetPluginConfig(plugins.PluginStorage.String(), s.conf.StoragePlugin)
	plugins.SetPluginConfig(plugins.PluginRepository.String(), s.conf.RepositoryPlugin)
	plugins.LoadPlugins()
}

func (s *Server) eventListen() {
	s.agent.RegisterEventHandler(s)
	events.AddListener(notify.EventTicker, s.store)
	events.AddListener(notify.EventDiscovery, s)
	events.AddListener(notify.EventPullByPeer, s.store)
}

func (s *Server) initialization() (err error) {
	s.tick = ticker.NewTaskTicker(s.conf.TickerInterval, tickHandler)

	s.store, err = datacenter.NewStore(strings.Split(s.conf.DCAddr, ","))
	if err != nil {
		return err
	}

	s.agent, err = peer.Create(s.conf.Config, createLogFile(s.conf.LogFile))
	if err != nil {
		return err
	}

	s.broker = grpc.NewBroker(s.conf.RPCAddr, s.store)
	return nil
}

func (s *Server) startServers(ctx context.Context) {
	s.agent.Start(ctx)

	if s.conf.JoinAddr != "" {
		_, err := s.agent.Join([]string{s.conf.JoinAddr}, false)
		if err != nil {
			log.Errorf(err, "Syncer join peer cluster failed")
		}
	}

	s.broker.Run()

	gopool.Go(s.tick.Start)
}

func (s *Server) waitQuit(ctx context.Context) {
	err := syssig.AddSignalsHandler(func() {
		s.Stop()
	}, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	if err != nil {
		log.Errorf(err, "Syncer add signals handler failed")
		return
	}
	syssig.Run(ctx)
}

func createLogFile(logFile string) (fw io.Writer) {
	fw = os.Stderr
	if logFile == "" {
		return
	}

	f, err := utils.OpenFile(logFile)
	if err != nil {
		log.Errorf(err, "Syncer open log file %s failed", logFile)
		return
	}
	return f
}

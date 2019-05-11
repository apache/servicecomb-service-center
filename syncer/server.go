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

// tickHandler Timed task handler
func tickHandler(ctx context.Context) {
	events.Dispatch(events.NewContextEvent(notify.EventTicker, ctx))
}

// Server struct for syncer
type Server struct {
	// Stores the syncer configuration
	conf *config.Config

	//Set the task ticker
	tick *ticker.TaskTicker

	// store is the datacenter service
	store datacenter.Store

	// Wraps the serf agents
	agent *peer.Agent

	//Wraps the grpc server and client
	broker *grpc.Broker
}

// NewServer new server with Config
func NewServer(conf *config.Config) *Server {
	return &Server{
		conf: conf,
	}
}

// Run syncer Server
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

// Stop Syncer Server
func (s *Server) Stop() {
	if s.tick != nil {
		s.tick.Stop()
	}

	if s.agent != nil {
		// removes the serf eventHandler
		s.agent.DeregisterEventHandler(s)
		//Leave from Serf
		s.agent.Leave()
		// closes this serf agent
		s.agent.Shutdown()
	}

	if s.broker != nil {
		s.broker.Stop()
	}

	if s.store != nil {
		s.store.Stop()
	}

	// clears the event listeners
	events.Clean()

	// Closes all goroutines in the pool
	gopool.CloseAndWait()
}

// initPlugin Initialize the plugin and load the external plugin according to the configuration
func (s *Server) initPlugin() {
	plugins.SetPluginConfig(plugins.PluginStorage.String(), s.conf.StoragePlugin)
	plugins.SetPluginConfig(plugins.PluginRepository.String(), s.conf.RepositoryPlugin)
	plugins.LoadPlugins()
}

// eventListen Start internal event listener
func (s *Server) eventListen() {
	// Register self as an event handler for serf
	s.agent.RegisterEventHandler(s)

	events.AddListener(notify.EventTicker, s.store)
	events.AddListener(notify.EventDiscovery, s)
	events.AddListener(notify.EventPullByPeer, s.store)
}

// initialization Initialize the starter of the syncer
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

// startServers Start all internal services
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

// waitQuit Waiting for system quit signal
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

// createLogFile create log file
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

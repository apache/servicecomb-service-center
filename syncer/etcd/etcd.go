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

package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.etcd.io/etcd/server/v3/etcdserver/api/v3client"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/pkg/errors"
)

// Server etcd server
type Server struct {
	conf    *embed.Config
	etcd    *embed.Etcd
	running *utils.AtomicBool

	readyCh chan struct{}
	stopCh  chan struct{}
}

// NewServer new etcd server with options
func NewServer(ops ...Option) (*Server, error) {
	conf, err := toEtcdConfig(ops...)
	if err != nil {
		return nil, errors.Wrap(err, "options to etcd config failed")
	}
	return &Server{
		conf:    conf,
		running: utils.NewAtomicBool(false),
		readyCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}, nil
}

// Start etcd server
func (s *Server) Start(ctx context.Context) {
	s.running.DoToReverse(false, func() {
		etcd, err := embed.StartEtcd(s.conf)
		if err != nil {
			log.Error("etcd: start server failed", err)
			close(s.stopCh)
			return
		}
		s.etcd = etcd
		go s.waitNotify(ctx)
	})
}

// AddOptions add some options when server not running
func (s *Server) AddOptions(ops ...Option) error {
	if s.running.Bool() {
		return errors.New("etcd server was running")
	}
	return mergeConfig(s.conf, toConfig(ops...))
}

// Ready Returns a channel that will be closed when etcd is ready
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Stopped Returns a channel that will be closed when etcd is stopped
func (s *Server) Stopped() <-chan struct{} {
	return s.stopCh
}

// Storage returns etcd storage
func (s *Server) Storage() *clientv3.Client {
	return v3client.New(s.etcd.Server)
}

// IsLeader Check leader
func (s *Server) IsLeader() bool {
	if s.etcd == nil || s.etcd.Server == nil {
		return false
	}
	return s.etcd.Server.Leader() == s.etcd.Server.ID()
}

// Stop etcd server
func (s *Server) Stop() {
	s.running.DoToReverse(true, func() {
		if s.etcd != nil {
			log.Info("etcd: begin shutdown")
			s.etcd.Close()
			close(s.stopCh)
		}
		log.Info("etcd: shutdown complete")
	})
}

func (s *Server) waitNotify(ctx context.Context) {
	select {
	// Be returns when the server is readied
	case <-s.etcd.Server.ReadyNotify():
		log.Info("etcd: start server success")
		close(s.readyCh)

	// Be returns when the server is stopped
	case <-s.etcd.Server.StopNotify():
		log.Warn("etcd: server stopped, quitting")
		s.Stop()

	// Returns an error when running goroutine fails in the etcd startup process
	case err := <-s.etcd.Err():
		log.Error("etcd: server happened error, quitting", err)
		s.Stop()

	case <-ctx.Done():
		log.Warn("etcd: cancel server by context")
		s.Stop()
	}
}

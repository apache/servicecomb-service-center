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

package grpc

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCHandler interface {
	Discovery() *pb.SyncData
}
type PullHandle func() *pb.SyncData

// Server struct
type Server struct {
	lsn     net.Listener
	addr    string
	handler GRPCHandler
	readyCh chan struct{}
	stopCh  chan struct{}
	tlsConf *tls.Config
}

// NewServer new grpc server
func NewServer(addr string, handler GRPCHandler, tlsConf *tls.Config) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		readyCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
		tlsConf: tlsConf,
	}
}

// Provide consumers with an interface to pull data
func (s *Server) Pull(ctx context.Context, in *pb.PullRequest) (*pb.SyncData, error) {
	return s.handler.Discovery(), nil
}

// Stop grpc server
func (s *Server) Stop() {
	if s.lsn != nil {
		s.lsn.Close()
		s.lsn = nil
	}
}

// Start grpc server
func (s *Server) Start(ctx context.Context) {
	lsn, err := net.Listen("tcp", s.addr)
	if err == nil {
		var svc *grpc.Server
		if s.tlsConf != nil {
			svc = grpc.NewServer(grpc.Creds(credentials.NewTLS(s.tlsConf)))
		} else {
			svc = grpc.NewServer()
		}

		pb.RegisterSyncServer(svc, s)
		s.lsn = lsn
		gopool.Go(func(ctx context.Context) {
			err = svc.Serve(s.lsn)
		})
	}

	if err != nil {
		log.Error("start grpc failed", err)
		close(s.stopCh)
		return
	}
	log.Info("start grpc success")
	close(s.readyCh)
}

// Ready Returns a channel that will be closed when grpc is ready
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Error Returns a channel that will be closed a grpc is stopped
func (s *Server) Stopped() <-chan struct{} {
	return s.stopCh
}

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
	"math"
	"net"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server struct
type Server struct {
	server   *grpc.Server
	listener net.Listener
	running  *utils.AtomicBool

	readyCh chan struct{}
	stopCh  chan struct{}
}

// NewServer new grpc server with options
func NewServer(ops ...Option) (*Server, error) {
	conf := toGRPCConfig(ops...)
	var srv *grpc.Server
	if conf.tlsConfig != nil {
		srv = grpc.NewServer(grpc.Creds(credentials.NewTLS(conf.tlsConfig)), grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	} else {
		srv = grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	}

	rpc.RegisterGRpcServer(srv)

	ls, err := net.Listen("tcp", conf.addr)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc: listen failed, addr = %s", conf.addr)
	}

	return &Server{
		server:   srv,
		listener: ls,
		running:  utils.NewAtomicBool(false),
		readyCh:  make(chan struct{}),
		stopCh:   make(chan struct{}),
	}, nil
}

// Start grpc server
func (s *Server) Start(ctx context.Context) {
	s.running.DoToReverse(false, func() {
		go func() {
			err := s.server.Serve(s.listener)
			if err != nil {
				log.Error("grpc: start server failed", err)
				s.Stop()
			}
		}()
		close(s.readyCh)
		go s.wait(ctx)
	})
}

// Ready Returns a channel that will be closed when etcd is ready
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Stopped Returns a channel that will be closed when etcd is stopped
func (s *Server) Stopped() <-chan struct{} {
	return s.stopCh
}

// Stop etcd server
func (s *Server) Stop() {
	s.running.DoToReverse(true, func() {
		if s.server != nil {
			log.Info("grpc: begin shutdown")
			s.server.Stop()
			close(s.stopCh)
		}
		log.Info("grpc: shutdown complete")
	})
}

func (s *Server) wait(ctx context.Context) {
	select {
	case <-s.stopCh:
		log.Warn("grpc: server stopped, exited")
	case <-ctx.Done():
		log.Warn("grpc: cancel server by context")
		s.Stop()
	}
}

// InjectClient inject grpc client to proto module
func InjectClient(injection func(conn *grpc.ClientConn), ops ...Option) error {
	conf := toGRPCConfig(ops...)

	var conn *grpc.ClientConn
	var err error

	if conf.tlsConfig != nil {
		conn, err = grpc.Dial(conf.addr, grpc.WithTransportCredentials(credentials.NewTLS(conf.tlsConfig)),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)))
	} else {
		conn, err = grpc.Dial(conf.addr, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)))
	}

	if err != nil {
		return errors.Wrapf(err, "grpc: create grpc client conn failed, addr = %s", conf.addr)
	}

	injection(conn)
	return nil
}

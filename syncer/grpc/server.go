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
	"net"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
)

type GRPCHandler interface {
	GetData() *pb.SyncData
}
type PullHandle func() *pb.SyncData

// Server struct
type Server struct {
	lsn     net.Listener
	addr    string
	handler GRPCHandler
}

// NewServer new grpc server
func NewServer(addr string, handler GRPCHandler) *Server {
	return &Server{addr: addr, handler: handler}
}

// Provide consumers with an interface to pull data
func (s *Server) Pull(ctx context.Context, in *pb.PullRequest) (*pb.SyncData, error) {
	return s.handler.GetData(), nil
}

// Stop grpc server
func (s *Server) Stop() {
	if s.lsn == nil {
		return
	}
	s.lsn.Close()
}

// Run grpc server
func (s *Server) Run() (err error) {
	s.lsn, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	svc := grpc.NewServer()
	pb.RegisterSyncServer(svc, s)
	gopool.Go(func(ctx context.Context) {
		svc.Serve(s.lsn)
	})
	return nil
}

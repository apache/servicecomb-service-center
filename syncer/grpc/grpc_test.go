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
	"errors"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/rpc"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/stretchr/testify/assert"
	ggrpc "google.golang.org/grpc"
)

type testServer struct{}

func (t *testServer) DeclareDataLength(ctx context.Context, request *pb.DeclareRequest) (*pb.DeclareResponse, error) {
	return &pb.DeclareResponse{}, nil
}

func (t *testServer) IncrementPull(ctx context.Context, request *pb.IncrementPullRequest) (*pb.SyncData, error) {
	return &pb.SyncData{}, nil
}

func (t *testServer) Pull(context.Context, *pb.PullRequest) (*pb.SyncData, error) {
	return &pb.SyncData{}, nil
}

func TestGRPCServer(t *testing.T) {
	syncSvr := &testServer{}
	addr := "127.0.0.1:9099"
	svr, err := NewServer(
		WithAddr(addr),
		WithTLSConfig(nil),
	)
	assert.Nil(t, err)

	rpc.RegisterService(func(s *ggrpc.Server) {
		pb.RegisterSyncServer(s, syncSvr)
	})

	err = startServer(context.Background(), svr)
	assert.Nil(t, err)

	err = InjectClient(func(conn *ggrpc.ClientConn) {}, WithAddr(addr))
	assert.Nil(t, err)

	svr.Stop()
}

func startServer(ctx context.Context, svr *Server) (err error) {
	svr.Start(ctx)
	select {
	case <-svr.Ready():
	case <-svr.Stopped():
		err = errors.New("start grpc server failed")
	case <-time.After(time.Second * 3):
		err = errors.New("start grpc server timeout")
	}
	return
}

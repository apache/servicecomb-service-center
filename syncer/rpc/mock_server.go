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

package rpc

import (
	"context"
	"net"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/go-chassis/foundation/gopool"
	"google.golang.org/grpc"
)

const wait = 100 * time.Millisecond

func MockServer(address string, srv v1sync.EventServiceServer) *grpc.Server {
	server := grpc.NewServer()
	v1sync.RegisterEventServiceServer(server, srv)

	gopool.Go(func(context.Context) {
		lis, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal("new tcp connection failed", err)
		}

		err = server.Serve(lis)
		if err != nil {
			log.Fatal("grpc server serve failed", err)
		}
	})

	time.Sleep(wait)

	return server
}

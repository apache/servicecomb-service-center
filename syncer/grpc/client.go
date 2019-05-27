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
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
)

var (
	clients = make(map[string]*Client)
	lock    sync.RWMutex
)

// Client struct
type Client struct {
	addr string
	conn *grpc.ClientConn
	cli  pb.SyncClient
}

func Pull(ctx context.Context, addr string) (*pb.SyncData, error) {
	cli := getClient(addr)

	data, err := cli.cli.Pull(ctx, &pb.PullRequest{})
	if err != nil {
		log.Errorf(err, "Pull from grpc failed, going to close the client")
		closeClient(addr)
	}
	return data, err
}

func closeClient(addr string) {
	lock.RLock()
	cli, ok := clients[addr]
	lock.RUnlock()
	if ok {

		cli.conn.Close()
		lock.Lock()
		delete(clients, addr)
		lock.Unlock()
		log.Infof("Close grpc connection to %s", addr)
	}
}

// GetClient Get the client from the client caches with addr
func getClient(addr string) *Client {
	lock.RLock()
	cli, ok := clients[addr]
	lock.RUnlock()
	if !ok {
		log.Infof("Create new grpc connection to %s", addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil
		}
		cli = &Client{conn: conn, cli: pb.NewSyncClient(conn), addr: addr}
		lock.Lock()
		clients[addr] = cli
		lock.Unlock()
	}
	return cli
}

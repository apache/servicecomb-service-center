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

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	ggrpc "google.golang.org/grpc"
)

var (
	clients sync.Map
)

// Client struct
type Client struct {
	addr string
	conn *ggrpc.ClientConn
	cli  pb.SyncClient
}

// NewSyncClient Get the client from the client caches with addr
func NewSyncClient(addr string, tlsConf *tls.Config) (cli *Client) {
	val, ok := clients.Load(addr)
	if ok {
		cli = val.(*Client)
	} else {
		err := grpc.InjectClient(func(conn *ggrpc.ClientConn) {
			cli = &Client{
				addr: addr,
				conn: conn,
				cli:  pb.NewSyncClient(conn),
			}
			clients.Store(addr, cli)
		}, grpc.WithAddr(addr), grpc.WithTLSConfig(tlsConf))
		if err != nil {
			log.Error("", err)
		}
	}
	return
}

// Pull implement the interface of sync server
func (c *Client) Pull(ctx context.Context, addr string) (*pb.SyncData, error) {
	data, err := c.cli.Pull(ctx, &pb.PullRequest{Addr: addr})
	if err != nil {
		log.Error("Pull from grpc client failed, going to close the client", err)
		closeClient(c.addr)
	}
	return data, err
}

func (c *Client) IncrementPull(ctx context.Context, req *pb.IncrementPullRequest) (*pb.SyncData, error) {
	data, err := c.cli.IncrementPull(ctx, req)
	if err != nil {
		log.Error("Pull from grpc client failed, going to close the client", err)
		closeClient(c.addr)
	}
	return data, err
}

func (c *Client) DeclareDataLength(ctx context.Context, addr string) (*pb.DeclareResponse, error) {
	res, err := c.cli.DeclareDataLength(ctx, &pb.DeclareRequest{Addr: addr})
	if err != nil {
		log.Error("Get SyncDataLength from grpc client failed, going to close the client", err)
		closeClient(c.addr)
	}
	return res, err
}

func closeClient(addr string) {
	val, ok := clients.Load(addr)
	if ok {
		cli := val.(*Client)
		cli.conn.Close()
		clients.Delete(addr)
		log.Info(fmt.Sprintf("Close grpc client connection to %s", addr))
	}
}

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
	"fmt"
	"log"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"golang.org/x/net/websocket"
)

const (
	apiWatcherURL = "/v4/%s/registry/microservices/%s/watcher"
)

func (c *Client) Watch(ctx context.Context, domain, project, selfServiceID string, callback func(*pb.WatchInstanceResponse)) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	conn, err := c.WebsocketDial(ctx, fmt.Sprintf(apiWatcherURL, project, selfServiceID), headers)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	for {
		var message string
		err := websocket.JSON.Receive(conn, &message)
		if err != nil {
			log.Println(err)
			break
		}
		callback(message)
	}
	return pb.NewError(pb.ErrInternal, err.Error())
}

// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (c *LBClient) WebsocketDial(ctx context.Context, api string, headers http.Header) (conn *websocket.Conn, err error) {
	dialer := &websocket.Dialer{TLSClientConfig: c.TLS}
	var errs []string
	for i := 0; i < c.Retries; i++ {
		var addr *url.URL
		addr, err = url.Parse(c.Next())
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s]: %s", addr, err.Error()))
			continue
		}
		if addr.Scheme == "https" {
			addr.Scheme = "wss"
		} else {
			addr.Scheme = "ws"
		}
		conn, _, err = dialer.Dial(addr.String()+api, headers)
		if err == nil {
			break
		}
		errs = append(errs, fmt.Sprintf("[%s]: %s", addr, err.Error()))
	}
	if err != nil {
		err = errors.New(util.StringJoin(errs, ", "))
	}
	return
}

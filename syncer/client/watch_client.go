/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/gorilla/websocket"
)

const watchInstanceURL = "v4/syncer/watch"
const wsScheme = "ws"

var (
	TenantHeader      = "X-Domain-Name"
	HeaderContentType = "Content-Type"
	HeaderUserAgent   = "User-Agent"
)

type WatchClient struct {
	addr     string
	wsDialer *websocket.Dialer
	conn     *websocket.Conn
	ready    bool
	mux      sync.RWMutex
}

// NewWatchClient Get the client from the client caches with addr
func NewWatchClient(addr string) (cli *WatchClient) {
	cli = &WatchClient{
		addr:     addr,
		conn:     new(websocket.Conn),
		wsDialer: new(websocket.Dialer),
	}

	return cli
}

func (c *WatchClient) GetDefaultHeaders() http.Header {
	headers := http.Header{
		HeaderContentType: []string{"application/json"},
		HeaderUserAgent:   []string{"syncer-sc-client/1.0.0"},
		TenantHeader:      []string{"defaut"},
	}
	return headers
}

func (c *WatchClient) WebsocketDial() error {
	hosts := strings.Split(c.addr, "//")
	wsHost := ""

	if len(hosts) == 1 {
		wsHost = hosts[0]
	} else {
		wsHost = hosts[1]
	}

	u := url.URL{
		Scheme: wsScheme,
		Host:   wsHost,
		Path:   watchInstanceURL,
	}

	conn, _, err := c.wsDialer.Dial(u.String(), c.GetDefaultHeaders())

	if err != nil {
		return fmt.Errorf("watching instance dial catch an exception:%s", err.Error())
	}

	c.conn = conn

	log.Info("watching instance dial is success connected ")

	return nil
}

func (c *WatchClient) WatchInstances(callback func(*dump.WatchInstanceChangedEvent)) error {
	c.mux.RLock()
	connIsReady := c.ready
	c.mux.RUnlock()

	if connIsReady {
		return nil
	}

	err := c.WebsocketDial()
	if err != nil {
		return err
	}

	c.mux.Lock()
	c.ready = true
	c.mux.Unlock()

	go func() {
		for {
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Error("err occurred", err)
				err = alarm.Raise(alarm.IDWebsocketOfScSyncerLost, alarm.AdditionalContext("%v", err))
				if err != nil {
					log.Error("alarm error", err)
				}
				break
			}
			if messageType == websocket.TextMessage {
				var response dump.WatchInstanceChangedEvent
				err := json.Unmarshal(message, &response)
				if err != nil {
					break
				}
				callback(&response)

			}
		}

		log.Debugf("close conn:%s", c.conn.RemoteAddr())
		err := c.conn.Close()

		c.mux.Lock()
		c.ready = false
		c.mux.Unlock()

		if err != nil {
			log.Error("close conn error:%s", err)
		}

	}()

	return err
}

func (c *WatchClient) WatchInstanceHeartbeat(callback func(*dump.WatchInstanceChangedEvent)) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-context.Background().Done():
				return
			case <-ticker.C:
				err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))

				if err != nil {
					log.Errorf(err, "fail to send ping to service center, try to conn again")

					c.mux.RLock()
					connIsReady := c.ready
					c.mux.RUnlock()

					if !connIsReady {
						err = c.WatchInstances(callback)
						if err != nil {
							log.Error("", err)
						}
					}
				}
			}
		}
	}()
}

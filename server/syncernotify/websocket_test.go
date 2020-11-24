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
package syncernotify_test

// initialize
import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/server/core"
	. "github.com/apache/servicecomb-service-center/server/syncernotify"
	"github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var closeCh = make(chan struct{})

type watcherConn struct {
}

func init() {
	testing.Init()
	core.Initialize()
}

func (h *watcherConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	conn, _ := upgrader.Upgrade(w, r, nil)
	for {
		conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		<-closeCh
		conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
		conn.Close()
		return
	}
}

func TestDoWebSocketWatch(t *testing.T) {
	s := httptest.NewServer(&watcherConn{})

	conn, _, _ := websocket.DefaultDialer.Dial(
		strings.Replace(s.URL, "http://", "ws://", 1), nil)
	//fmt.Print(conn)
	ws := NewWebSocket(context.Background(), conn)
	ws.Init()

	t.Run("start syncer center", func(t *testing.T) {

		GetSyncerNotifyCenter().Start()

		go DoWebSocketWatch(context.Background(), conn)

		go ws.HandleWatchWebSocketControlMessage()

	})

	t.Run("handle websocket job", func(t *testing.T) {
		// 1. add event to channel, handle event
		instanceEvent := &dump.WatchInstanceChangedEvent{
			Action: "CREATE",
			Service: &dump.Microservice{
				KV: &dump.KV{
					Key:         "/cse-sr/ms/files/default/default/4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
					Rev:         12,
					ClusterName: "clustername",
				},
				Value: &discovery.MicroService{
					ServiceId:   "4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
					AppId:       "default",
					ServiceName: "TEST01",
					Version:     "0.0.1",
					Level:       "BACK",
					Schemas: []string{
						"servicecenter.grpc.api.ServiceCtrl",
						"servicecenter.grpc.api.ServiceInstanceCtrl",
					},
				},
			},
			Instance: &dump.Instance{
				KV: &dump.KV{
					Key:         "/cse-sr/inst/files/default/default/4042a6a3e5a2893698ae363ea99a69eb63fc51cd/7a6be9f861a811e9b3f6fa163eca30e0",
					Rev:         21,
					ClusterName: "clustername",
				},
				Value: &discovery.MicroServiceInstance{
					InstanceId: "8e0fe4b961a811e981a6fa163e86b81a",
					ServiceId:  "4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
					Endpoints: []string{
						"rest://192.168.88.109:30100/",
					},
					HostName: "sunlisen",
				},
			},
			Revision: 10,
		}

		GetSyncerNotifyCenter().AddEvent(instanceEvent)

		<-time.After(time.Second)

		// 2. handle unknown things
		ws.HandleWatchWebSocketJob(nil)

		// 3. handle error
		fakeErr := errors.New("fake error")
		ws.HandleWatchWebSocketJob(fakeErr)

		// 4. heartbeat
		err := ws.Heartbeat(websocket.PingMessage)
		assert.NoError(t, err)
		err = ws.Heartbeat(websocket.PongMessage)
		assert.NoError(t, err)

		closeCh <- struct{}{}

		<-time.After(time.Second)

		err = ws.Heartbeat(websocket.PingMessage)
		assert.Error(t, err)
		err = ws.Heartbeat(websocket.PongMessage)
		assert.Error(t, err)

	})

	Instance().Stop()
	GetSyncerNotifyCenter().Stop()
}

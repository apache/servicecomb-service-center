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
package ws_test

// initialize
import (
	"context"
	"errors"

	_ "github.com/apache/servicecomb-service-center/test"

	wss "github.com/apache/servicecomb-service-center/server/connection/ws"
	"github.com/apache/servicecomb-service-center/server/core"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
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

func TestDoWebSocketListAndWatch(t *testing.T) {
	s := httptest.NewServer(&watcherConn{})

	conn, _, _ := websocket.DefaultDialer.Dial(
		strings.Replace(s.URL, "http://", "ws://", 1), nil)

	wss.SendEstablishError(conn, errors.New("error"))

	w := NewInstanceEventListWatcher("g", "s", func() (results []*discovery.WatchInstanceResponse, rev int64) {
		results = append(results, &discovery.WatchInstanceResponse{
			Response: discovery.CreateResponse(discovery.ResponseSuccess, "ok"),
			Action:   string(discovery.EVT_CREATE),
			Key:      &discovery.MicroServiceKey{},
			Instance: &discovery.MicroServiceInstance{},
		})
		return
	})

	ws := wss.New(context.Background(), conn, w)
	err := ws.Init()
	if err != nil {
		t.Fatalf("TestPublisher_Run")
	}

	event.Center().Start()

	go func() {
		wss.ListAndWatch(context.Background(), "", nil, conn)

		w2 := NewInstanceEventListWatcher("g", "s", func() (results []*discovery.WatchInstanceResponse, rev int64) {
			return
		})
		ws2 := wss.New(context.Background(), conn, w2)
		err := ws2.Init()
		if err != nil {
			t.Fatalf("TestPublisher_Run")
		}
	}()

	go ws.HandleControlMessage()

	w.OnMessage(nil)
	w.OnMessage(&event.InstanceEvent{})

	Center().Publish(NewInstanceEvent("g", "s", 1, &discovery.WatchInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "ok"),
		Action:   string(discovery.EVT_CREATE),
		Key:      &discovery.MicroServiceKey{},
		Instance: &discovery.MicroServiceInstance{},
	}))

	<-time.After(time.Second)

	ws.HandleEvent(nil)

	ws.Heartbeat(websocket.PingMessage)
	ws.Heartbeat(websocket.PongMessage)

	ws.HandleEvent(time.Now())

	closeCh <- struct{}{}

	<-time.After(time.Second)

	ws.Heartbeat(websocket.PingMessage)
	ws.Heartbeat(websocket.PongMessage)

	w.OnMessage(nil)

	wss.Instance().Stop()
}

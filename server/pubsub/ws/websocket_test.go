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

import (
	_ "github.com/apache/servicecomb-service-center/test"

	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/pubsub/ws"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

var closeCh = make(chan struct{})

func init() {
	testing.Init()
	core.Init()
}

type watcherConn struct {
	MockServer *httptest.Server
	ClientConn websocket.Conn
	ServerConn websocket.Conn
}

func (h *watcherConn) Test() {
	h.MockServer = httptest.NewServer(h)
	h.ClientConn, _, _ = websocket.DefaultDialer.Dial(
		strings.Replace(h.MockServer.URL, "http://", "ws://", 1), nil)
	// wait server is ready
	<-time.After(time.Second)
}

func (h *watcherConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	h.ServerConn, _ = upgrader.Upgrade(w, r, nil)
	for {
		//h.ServerConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		//h.ServerConn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
		_, _, err := h.ServerConn.ReadMessage()
		if err != nil {
			return
		}
		<-closeCh
		h.ServerConn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
		h.ServerConn.Close()
		return
	}
	h.MockServer.Close()
}

func NewTest() *watcherConn {
	ts := &watcherConn{}
	ts.Test()
	return ts
}

func TestNewWebSocket(t *testing.T) {
	mock := NewTest()
	t.Run("should return not nil when new", func(t *testing.T) {
		assert.NotNil(t, ws.NewWebSocket("", "", mock.ServerConn))
	})
}

func TestWebSocket_NeedCheck(t *testing.T) {
	mock := NewTest()
	conn := mock.ServerConn
	options := ws.ToOptions()
	webSocket := &ws.WebSocket{
		Options:       options,
		DomainProject: "default",
		ConsumerID:    "",
		Conn:          conn,
	}

	t.Run("should not check when new", func(t *testing.T) {
		webSocket.HealthInterval = time.Second
		webSocket.Init()
		assert.Nil(t, webSocket.NeedCheck())
	})

	t.Run("should check when check time up", func(t *testing.T) {
		webSocket.HealthInterval = time.Microsecond
		webSocket.Init()
		<-time.After(time.Microsecond)
		assert.NotNil(t, webSocket.NeedCheck())
	})
	t.Run("should not check when busy", func(t *testing.T) {
		webSocket.HealthInterval = time.Microsecond
		webSocket.Init()
		<-time.After(time.Microsecond)
		assert.NotNil(t, webSocket.NeedCheck())
		assert.Nil(t, webSocket.NeedCheck())
	})
}

func TestWebSocket_Idle(t *testing.T) {
	mock := NewTest()
	webSocket := ws.NewWebSocket("", "", mock.ServerConn)

	t.Run("should idle when new", func(t *testing.T) {
		select {
		case <-webSocket.Idle():
		default:
			assert.Fail(t, "not idle")
		}
	})
	t.Run("should idle when setIdle", func(t *testing.T) {
		select {
		case <-webSocket.Idle():
			assert.Fail(t, "idle")
		default:
			webSocket.SetIdle()
			select {
			case <-webSocket.Idle():
			default:
				assert.Fail(t, "not idle")
			}
		}
	})
	t.Run("should idle when checkHealth", func(t *testing.T) {
		_ = webSocket.CheckHealth(context.Background())
		select {
		case <-webSocket.Idle():
		default:
			assert.Fail(t, "not idle")
		}
	})
}

func TestWebSocket_CheckHealth(t *testing.T) {
	mock := NewTest()
	event.Center().Start()

	t.Run("should do nothing when recv PING", func(t *testing.T) {
		ws := ws.NewWebSocket("", "", mock.ServerConn)
		mock.ClientConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		<-time.After(time.Second)
		assert.Nil(t, ws.CheckHealth(context.Background()))
	})
	t.Run("should return err when consumer not exist", func(t *testing.T) {
		ws := ws.NewWebSocket("", "", mock.ServerConn)
		assert.Equal(t, "service does not exist", ws.CheckHealth(context.Background()).Error())
	})
}

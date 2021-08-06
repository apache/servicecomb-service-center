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

package syncernotify

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/gorilla/websocket"
)

type WebSocket struct {
	ctx    context.Context
	conn   *websocket.Conn
	err    error
	free   chan struct{}
	closed chan struct{}
}

func DoWebSocketWatch(ctx context.Context, conn *websocket.Conn) {
	log.Debug("begin do websocket watch")

	socket := NewWebSocket(ctx, conn)

	process(socket)
}

func NewWebSocket(ctx context.Context, conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		ctx:  ctx,
		conn: conn,
	}
}

func process(socket *WebSocket) {
	socket.Init()

	socket.HandleWatchWebSocketControlMessage()

	socket.Stop()
}

func (wh *WebSocket) Init() {
	wh.free = make(chan struct{}, 1)
	wh.closed = make(chan struct{})
	wh.SetReady()
	remoteAddr := wh.conn.RemoteAddr().String()

	Instance().Accept(wh)
	log.Debug(fmt.Sprintf("start watching instance status, watcher[%s]", remoteAddr))
}

func (wh *WebSocket) Pick() interface{} {
	select {
	case <-wh.Ready():
		if wh.Err() != nil {
			return wh.Err()
		}

		select {
		case e := <-GetSyncerNotifyCenter().instEventCh:
			return e
		default:
			// reset if idle
			wh.SetReady()
		}
	default:
	}
	return nil
}

func (wh *WebSocket) HandleWatchWebSocketJob(o interface{}) {
	defer wh.SetReady()

	var (
		message    []byte
		remoteAddr = wh.conn.RemoteAddr().String()
	)

	switch o := o.(type) {
	// error will be set in HandleWatchWebSocketControlMessage
	case error:
		log.Error(fmt.Sprintf("watcher[%s] catch an err,", remoteAddr), o)

		message = util.StringToBytesWithNoCopy(fmt.Sprintf("watcher catch an err: %s", o.Error()))
	// InstanceChangedEvent will be set in OnEvent
	case *pb.WatchInstanceChangedEvent:
		resp := o

		log.Info(fmt.Sprintf("event[%s] is coming in, watcher[%s]", resp.Action, remoteAddr))

		data, err := json.Marshal(resp)
		if err != nil {
			log.Error(fmt.Sprintf("watcher[%s] marshal response failed", remoteAddr), err)
			message = util.StringToBytesWithNoCopy(fmt.Sprintf("marshal output file error, %s", err.Error()))
			break
		}
		message = data
	default:
		log.Error(fmt.Sprintf("watcher[%s] unknown input %v", remoteAddr, o), nil)
		return
	}

	select {
	case <-wh.closed:
		return
	default:
	}

	err := wh.WriteMessage(message)

	if err != nil {
		log.Error(fmt.Sprintf("watcher[%s] catch an err", remoteAddr), err)
	}
}

func (wh *WebSocket) ReadTimeout() time.Duration {
	return ReadTimeout
}

func (wh *WebSocket) SendTimeout() time.Duration {
	return SendTimeout
}

func (wh *WebSocket) Heartbeat(messageType int) error {
	err := wh.conn.WriteControl(messageType, []byte{}, time.Now().Add(wh.SendTimeout()))
	if err != nil {
		messageTypeName := "Ping"
		if messageType == websocket.PongMessage {
			messageTypeName = "Pong"
		}
		log.Error(fmt.Sprintf("fail to send '%s' to watcher[%s]", messageTypeName, wh.conn.RemoteAddr()), err)
		return err
	}
	return nil
}

func (wh *WebSocket) HandleWatchWebSocketControlMessage() {
	remoteAddr := wh.conn.RemoteAddr().String()
	// PING
	wh.conn.SetPingHandler(func(message string) error {
		defer func() {
			err := wh.conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout()))
			if err != nil {
				log.Error("", err)
			}
		}()
		log.Debug(fmt.Sprintf("received 'Ping' message '%s' from watcher[%s]", message, remoteAddr))
		return wh.Heartbeat(websocket.PongMessage)
	})
	// PONG
	wh.conn.SetPongHandler(func(message string) error {
		defer func() {
			err := wh.conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout()))
			if err != nil {
				log.Error("", err)
			}
		}()
		log.Debug(fmt.Sprintf("received 'Pong' message '%s' from watcher[%s]", message, remoteAddr))
		return nil
	})
	// CLOSE
	wh.conn.SetCloseHandler(func(code int, text string) error {
		log.Info(fmt.Sprintf("watcher[%s] active closed, code: %d, message: '%s'", remoteAddr, code, text))
		return wh.sendClose(code, text)
	})

	wh.conn.SetReadLimit(ReadMaxBody)
	err := wh.conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout()))
	if err != nil {
		log.Error("", err)
	}
	for {
		_, _, err := wh.conn.ReadMessage()
		if err != nil {
			// client close or conn broken
			wh.SetError(err)
			return
		}
	}
}

func (wh *WebSocket) sendClose(code int, text string) error {
	remoteAddr := wh.conn.RemoteAddr().String()
	var message []byte
	if code != websocket.CloseNoStatusReceived {
		message = websocket.FormatCloseMessage(code, text)
	}
	err := wh.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(wh.SendTimeout()))
	if err != nil {
		log.Error(fmt.Sprintf("watcher[%s] catch an err", remoteAddr), err)
		return err
	}
	return nil
}

func (wh *WebSocket) WriteMessage(message []byte) error {
	err := wh.conn.SetWriteDeadline(time.Now().Add(wh.SendTimeout()))
	if err != nil {
		return err
	}
	return wh.conn.WriteMessage(websocket.TextMessage, message)
}

func (wh *WebSocket) Ready() <-chan struct{} {
	return wh.free
}

func (wh *WebSocket) SetReady() {
	select {
	case wh.free <- struct{}{}:
	default:
	}
}

func (wh *WebSocket) Stop() {
	close(wh.closed)
}

func (wh *WebSocket) SetError(e error) {
	wh.err = e
}

func (wh *WebSocket) Err() error {
	return wh.err
}

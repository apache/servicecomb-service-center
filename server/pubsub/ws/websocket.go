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

package ws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
)

const Websocket = "Websocket"

var errServiceNotExist = errors.New("service does not exist")

type WebSocket struct {
	Options
	Conn          *websocket.Conn
	RemoteAddr    string
	DomainProject string
	ConsumerID    string

	ticker   *time.Ticker
	needPing bool
	idleCh   chan struct{}
}

func (wh *WebSocket) Init() {
	wh.RemoteAddr = wh.Conn.RemoteAddr().String()
	wh.ticker = time.NewTicker(wh.HealthInterval)
	wh.needPing = true
	wh.idleCh = make(chan struct{}, 1)

	wh.registerMessageHandler()

	wh.SetIdle()

	log.Debug(fmt.Sprintf("start watching instance status, subscriber[%s], consumer: %s",
		wh.RemoteAddr, wh.ConsumerID))
}

func (wh *WebSocket) registerMessageHandler() {
	remoteAddr := wh.RemoteAddr
	// PING
	wh.Conn.SetPingHandler(func(message string) error {
		defer func() {
			err := wh.Conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout))
			if err != nil {
				log.Error("", err)
			}
		}()
		if wh.needPing {
			log.Info(fmt.Sprintf("received 'Ping' message '%s' from subscriber[%s], no longer send 'Ping' to it, consumer: %s",
				message, remoteAddr, wh.ConsumerID))
		}
		wh.needPing = false
		return wh.WritePingPong(websocket.PongMessage)
	})
	// PONG
	wh.Conn.SetPongHandler(func(message string) error {
		defer func() {
			err := wh.Conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout))
			if err != nil {
				log.Error("", err)
			}
		}()
		log.Debug(fmt.Sprintf("received 'Pong' message '%s' from subscriber[%s], consumer: %s",
			message, remoteAddr, wh.ConsumerID))
		return nil
	})
	// CLOSE
	wh.Conn.SetCloseHandler(func(code int, text string) error {
		log.Info(fmt.Sprintf("subscriber[%s] active closed, code: %d, message: '%s', consumer: %s",
			remoteAddr, code, text, wh.ConsumerID))
		return wh.sendClose(code, text)
	})
}

func (wh *WebSocket) ReadMessage() error {
	wh.Conn.SetReadLimit(ReadMaxBody)
	err := wh.Conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout))
	if err != nil {
		log.Error("", err)
	}
	for {
		_, _, err := wh.Conn.ReadMessage()
		if err != nil {
			return err
		}
	}
}

func (wh *WebSocket) sendClose(code int, text string) error {
	remoteAddr := wh.Conn.RemoteAddr().String()
	var message []byte
	if code != websocket.CloseNoStatusReceived {
		message = websocket.FormatCloseMessage(code, text)
	}
	err := wh.Conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(wh.SendTimeout))
	if err != nil {
		log.Error(fmt.Sprintf("subscriber[%s] catch an err, consumer: %s",
			remoteAddr, wh.ConsumerID), err)
		return err
	}
	return nil
}

// NeedCheck will be called by checker
func (wh *WebSocket) NeedCheck() interface{} {
	select {
	case <-wh.Idle():
		select {
		case t := <-wh.ticker.C:
			return t
		default:
			// reset if idleCh
			wh.SetIdle()
		}
	default:
	}
	return nil
}

// CheckHealth will be called if NeedCheck() returns not nil
func (wh *WebSocket) CheckHealth(ctx context.Context) error {
	defer wh.SetIdle()

	if !wh.needPing {
		return nil
	}

	ctx = util.SetDomainProjectString(ctx, wh.DomainProject)

	if exist, err := datasource.GetMetadataManager().ExistServiceByID(ctx, &pb.GetExistenceByIDRequest{
		ServiceId: wh.ConsumerID,
	}); err != nil || !exist.Exist {
		return errServiceNotExist
	}

	remoteAddr := wh.Conn.RemoteAddr().String()
	if err := wh.WritePingPong(websocket.PingMessage); err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("send 'Ping' message to subscriber[%s], consumer: %s",
		remoteAddr, wh.ConsumerID))
	return nil
}

func (wh *WebSocket) WritePingPong(messageType int) error {
	return wh.Conn.WriteControl(messageType, []byte{}, time.Now().Add(wh.SendTimeout))
}

func (wh *WebSocket) WriteTextMessage(message []byte) error {
	err := wh.Conn.SetWriteDeadline(time.Now().Add(wh.SendTimeout))
	if err != nil {
		return err
	}
	err = wh.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Error(fmt.Sprintf("subscriber[%s] catch an err, msg size: %d",
			wh.Conn.RemoteAddr().String(), len(message)), err)
	}
	return err
}

func (wh *WebSocket) Idle() <-chan struct{} {
	return wh.idleCh
}

func (wh *WebSocket) SetIdle() {
	select {
	case wh.idleCh <- struct{}{}:
	default:
	}
}

func NewWebSocket(domainProject, serviceID string, conn *websocket.Conn) *WebSocket {
	ws := &WebSocket{
		Options:       ToOptions(),
		DomainProject: domainProject,
		ConsumerID:    serviceID,
		Conn:          conn,
	}
	ws.Init()
	return ws
}

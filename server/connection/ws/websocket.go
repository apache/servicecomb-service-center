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
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/connection"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/notify"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
	"time"
)

const Websocket = "Websocket"

type WebSocket struct {
	ctx    context.Context
	ticker *time.Ticker
	conn   *websocket.Conn
	// watcher subscribe the notification service event
	watcher         *notify.InstanceEventListWatcher
	needPingWatcher bool
	free            chan struct{}
	closed          chan struct{}
}

func (wh *WebSocket) Init() error {
	wh.ticker = time.NewTicker(connection.HeartbeatInterval)
	wh.needPingWatcher = true
	wh.free = make(chan struct{}, 1)
	wh.closed = make(chan struct{})

	wh.SetReady()

	remoteAddr := wh.conn.RemoteAddr().String()

	// put in notification service queue
	if err := notify.Center().AddSubscriber(wh.watcher); err != nil {
		err = fmt.Errorf("establish[%s] websocket watch failed: notify service error, %s",
			remoteAddr, err.Error())
		log.Errorf(nil, err.Error())

		werr := wh.conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error()))
		if werr != nil {
			log.Errorf(werr, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
		}
		return err
	}

	// put in publisher queue
	Instance().Accept(wh)

	log.Debugf("start watching instance status, watcher[%s], subject: %s, group: %s",
		remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
	return nil
}

func (wh *WebSocket) ReadTimeout() time.Duration {
	return connection.ReadTimeout
}

func (wh *WebSocket) SendTimeout() time.Duration {
	return connection.SendTimeout
}

func (wh *WebSocket) Heartbeat(messageType int) error {
	err := wh.conn.WriteControl(messageType, []byte{}, time.Now().Add(wh.SendTimeout()))
	if err != nil {
		messageTypeName := "Ping"
		if messageType == websocket.PongMessage {
			messageTypeName = "Pong"
		}
		log.Errorf(err, "fail to send '%s' to watcher[%s], subject: %s, group: %s",
			messageTypeName, wh.conn.RemoteAddr(), wh.watcher.Subject(), wh.watcher.Group())
		//wh.watcher.SetError(err)
		return err
	}
	return nil
}

func (wh *WebSocket) HandleControlMessage() {
	remoteAddr := wh.conn.RemoteAddr().String()
	// PING
	wh.conn.SetPingHandler(func(message string) error {
		defer func() {
			err := wh.conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout()))
			if err != nil {
				log.Error("", err)
			}
		}()
		if wh.needPingWatcher {
			log.Infof("received 'Ping' message '%s' from watcher[%s], no longer send 'Ping' to it, subject: %s, group: %s",
				message, remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		}
		wh.needPingWatcher = false
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
		log.Debugf("received 'Pong' message '%s' from watcher[%s], subject: %s, group: %s",
			message, remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return nil
	})
	// CLOSE
	wh.conn.SetCloseHandler(func(code int, text string) error {
		log.Infof("watcher[%s] active closed, code: %d, message: '%s', subject: %s, group: %s",
			remoteAddr, code, text, wh.watcher.Subject(), wh.watcher.Group())
		return wh.sendClose(code, text)
	})

	wh.conn.SetReadLimit(connection.ReadMaxBody)
	err := wh.conn.SetReadDeadline(time.Now().Add(wh.ReadTimeout()))
	if err != nil {
		log.Error("", err)
	}
	for {
		_, _, err := wh.conn.ReadMessage()
		if err != nil {
			// client close or conn broken
			wh.watcher.SetError(err)
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
		log.Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return err
	}
	return nil
}

// Pick will be called by publisher
func (wh *WebSocket) Pick() interface{} {
	select {
	case <-wh.Ready():
		if wh.watcher.Err() != nil {
			return wh.watcher.Err()
		}

		select {
		case t := <-wh.ticker.C:
			return t
		case j := <-wh.watcher.Job:
			if j == nil {
				return fmt.Errorf("server shutdown")
			}
			return j
		default:
			// reset if idle
			wh.SetReady()
		}
	default:
	}

	return nil
}

// HandleEvent will be called if Pick() returns not nil
func (wh *WebSocket) HandleEvent(o interface{}) {
	defer wh.SetReady()

	var (
		message    []byte
		remoteAddr = wh.conn.RemoteAddr().String()
	)

	switch o := o.(type) {
	case error:
		log.Errorf(o, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())

		message = util.StringToBytesWithNoCopy(fmt.Sprintf("watcher catch an err: %s", o.Error()))
	case time.Time:
		if exist, err := datasource.Instance().ExistServiceByID(wh.ctx, &pb.GetExistenceByIDRequest{
			ServiceId: wh.watcher.Group(),
		}); err != nil || !exist.Exist {
			message = util.StringToBytesWithNoCopy("Service does not exit.")
			break
		}

		if !wh.needPingWatcher {
			return
		}

		if err := wh.Heartbeat(websocket.PingMessage); err != nil {
			log.Errorf(err, "send 'Ping' message to watcher[%s] failed, subject: %s, group: %s",
				remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
			return
		}

		log.Debugf("send 'Ping' message to watcher[%s], subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return
	case *notify.InstanceEvent:
		resp := o.Response

		providerFlag := fmt.Sprintf("%s/%s/%s", resp.Key.AppId, resp.Key.ServiceName, resp.Key.Version)
		if resp.Action != string(pb.EVT_EXPIRE) {
			providerFlag = fmt.Sprintf("%s/%s(%s)", resp.Instance.ServiceId, resp.Instance.InstanceId, providerFlag)
		}
		log.Infof("event[%s] is coming in, watcher[%s] watch %s, subject: %s, group: %s",
			resp.Action, remoteAddr, providerFlag, wh.watcher.Subject(), wh.watcher.Group())

		resp.Response = nil
		data, err := json.Marshal(resp)
		if err != nil {
			log.Errorf(err, "watcher[%s] watch %s, subject: %s, group: %s",
				remoteAddr, providerFlag, o, wh.watcher.Subject(), wh.watcher.Group())
			message = util.StringToBytesWithNoCopy(fmt.Sprintf("marshal output file error, %s", err.Error()))
			break
		}
		message = data
	default:
		log.Errorf(nil, "watcher[%s] unknown input %v, subject: %s, group: %s",
			remoteAddr, o, wh.watcher.Subject(), wh.watcher.Group())
		return
	}

	select {
	case <-wh.closed:
		return
	default:
	}

	err := wh.WriteMessage(message)
	if evt, ok := o.(*notify.InstanceEvent); ok {
		metrics.ReportPublishCompleted(evt, err)
	}
	if err != nil {
		log.Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
	}
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

func ListAndWatch(ctx context.Context, serviceID string, f func() ([]*pb.WatchInstanceResponse, int64), conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	socket := New(ctx, conn, notify.NewInstanceEventListWatcher(serviceID, domainProject, f))

	metrics.ReportSubscriber(domain, Websocket, 1)
	process(socket)
	metrics.ReportSubscriber(domain, Websocket, -1)
}

func process(socket *WebSocket) {
	if err := socket.Init(); err != nil {
		return
	}

	socket.HandleControlMessage()

	socket.Stop()
}

func SendEstablishError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	log.Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		log.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

func New(ctx context.Context, conn *websocket.Conn, watcher *notify.InstanceEventListWatcher) *WebSocket {
	return &WebSocket{
		ctx:     ctx,
		conn:    conn,
		watcher: watcher,
	}
}

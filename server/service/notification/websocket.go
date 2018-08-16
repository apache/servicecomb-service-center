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
package notification

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"time"
)

type WebSocket struct {
	ctx    context.Context
	ticker *time.Ticker
	conn   *websocket.Conn
	// watcher subscribe the notification service event
	watcher         *ListWatcher
	needPingWatcher bool
	free            chan struct{}
	closed          chan struct{}
}

func (wh *WebSocket) Init() error {
	wh.ticker = time.NewTicker(DEFAULT_HEARTBEAT_INTERVAL)
	wh.needPingWatcher = true
	wh.free = make(chan struct{}, 1)
	wh.closed = make(chan struct{})

	wh.SetReady()

	remoteAddr := wh.conn.RemoteAddr().String()

	// put in notification service queue
	if err := GetNotifyService().AddSubscriber(wh.watcher); err != nil {
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
	publisher.Accept(wh)

	log.Debugf("start watching instance status, watcher[%s], subject: %s, group: %s",
		remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
	return nil
}

func (wh *WebSocket) Timeout() time.Duration {
	return DEFAULT_SEND_TIMEOUT
}

func (wh *WebSocket) heartbeat(messageType int) error {
	err := wh.conn.WriteControl(messageType, []byte{}, time.Now().Add(wh.Timeout()))
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

func (wh *WebSocket) HandleWatchWebSocketControlMessage() {
	remoteAddr := wh.conn.RemoteAddr().String()
	// PING
	wh.conn.SetPingHandler(func(message string) error {
		if wh.needPingWatcher {
			log.Infof("received 'Ping' message '%s' from watcher[%s], no longer send 'Ping' to it, subject: %s, group: %s",
				message, remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		}
		wh.needPingWatcher = false
		return wh.heartbeat(websocket.PongMessage)
	})
	// PONG
	wh.conn.SetPongHandler(func(message string) error {
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
	err := wh.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(wh.Timeout()))
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

// HandleWatchWebSocketJob will be called if Pick() returns not nil
func (wh *WebSocket) HandleWatchWebSocketJob(o interface{}) {
	defer wh.SetReady()

	remoteAddr := wh.conn.RemoteAddr().String()
	var message []byte

	switch o.(type) {
	case error:
		err := o.(error)
		log.Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())

		message = util.StringToBytesWithNoCopy(fmt.Sprintf("watcher catch an err: %s", err.Error()))
	case time.Time:
		domainProject := util.ParseDomainProject(wh.ctx)
		if !serviceUtil.ServiceExist(wh.ctx, domainProject, wh.watcher.Group()) {
			message = util.StringToBytesWithNoCopy("Service does not exit.")
			break
		}

		if !wh.needPingWatcher {
			return
		}

		log.Debugf("send 'Ping' message to watcher[%s], subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		wh.heartbeat(websocket.PingMessage)
		return
	case *WatchJob:
		job := o.(*WatchJob)
		resp := job.Response

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

	err := wh.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
	}
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

func DoWebSocketListAndWatch(ctx context.Context, serviceId string, f func() ([]*pb.WatchInstanceResponse, int64), conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	socket := &WebSocket{
		ctx:     ctx,
		conn:    conn,
		watcher: NewListWatcher(serviceId, apt.GetInstanceRootKey(domainProject)+"/", f),
	}
	process(socket)
}

func process(socket *WebSocket) {
	if err := socket.Init(); err != nil {
		return
	}

	socket.HandleWatchWebSocketControlMessage()

	socket.Stop()
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	log.Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		log.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

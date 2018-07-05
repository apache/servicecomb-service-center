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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"time"
)

type WebSocket struct {
	ctx             context.Context
	ticker          *time.Ticker
	conn            *websocket.Conn
	watcher         *ListWatcher
	needPingWatcher bool
	closed          chan struct{}
}

func (wh *WebSocket) Init() error {
	wh.ticker = time.NewTicker(wh.Timeout())

	remoteAddr := wh.conn.RemoteAddr().String()

	// subscribe backend
	if err := GetNotifyService().AddSubscriber(wh.watcher); err != nil {
		err = fmt.Errorf("establish[%s] websocket watch failed: notify service error, %s",
			remoteAddr, err.Error())
		util.Logger().Errorf(nil, err.Error())

		err = wh.conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error()))
		if err != nil {
			util.Logger().Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
		}
		return err
	}

	// publish events
	publisher.Accept(wh)

	util.Logger().Debugf("start watching instance status, watcher[%s], subject: %s, group: %s",
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
		util.Logger().Errorf(err, "fail to send '%s' to watcher[%s], subject: %s, group: %s",
			messageTypeName, wh.conn.RemoteAddr(), wh.watcher.Subject(), wh.watcher.Group())
		//wh.watcher.SetError(err)
		return err
	}
	return nil
}

func (wh *WebSocket) HandleWatchWebSocketControlMessage() {
	defer close(wh.closed)

	remoteAddr := wh.conn.RemoteAddr().String()
	// PING
	wh.conn.SetPingHandler(func(message string) error {
		if wh.needPingWatcher {
			util.Logger().Infof("received 'Ping' message '%s' from watcher[%s], no longer send 'Ping' to it, subject: %s, group: %s",
				message, remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		}
		wh.needPingWatcher = false
		return wh.heartbeat(websocket.PongMessage)
	})
	// PONG
	wh.conn.SetPongHandler(func(message string) error {
		util.Logger().Debugf("received 'Pong' message '%s' from watcher[%s], subject: %s, group: %s",
			message, remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return nil
	})
	// CLOSE
	wh.conn.SetCloseHandler(func(code int, text string) error {
		util.Logger().Warnf(nil, "watcher[%s] active closed, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return wh.sendClose(code, text)
	})

	for {
		_, _, err := wh.conn.ReadMessage()
		if err != nil {
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
		util.Logger().Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		return err
	}
	return nil
}

func (wh *WebSocket) Pick() interface{} {
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
		return nil
	}
}

func (wh *WebSocket) HandleWatchWebSocketJob(o interface{}) {
	remoteAddr := wh.conn.RemoteAddr().String()
	var message []byte

	switch o.(type) {
	case error:
		err := o.(error)
		util.Logger().Warnf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())

		message = util.StringToBytesWithNoCopy(fmt.Sprintf("watcher catch an err: %s", err.Error()))
	case time.Time:
		domainProject := util.ParseDomainProject(wh.ctx)
		if !serviceUtil.ServiceExist(wh.ctx, domainProject, wh.watcher.Group()) {
			err := fmt.Errorf("Service does not exit.")
			wh.watcher.SetError(err)
			message = util.StringToBytesWithNoCopy(err.Error())
			break
		}

		if !wh.needPingWatcher {
			return
		}

		util.Logger().Debugf("send 'Ping' message to watcher[%s], subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
		err := wh.heartbeat(websocket.PingMessage)
		if err != nil {
			wh.watcher.SetError(err)
		}
		return
	case *WatchJob:
		job := o.(*WatchJob)
		resp := job.Response

		providerFlag := fmt.Sprintf("%s/%s/%s", resp.Key.AppId, resp.Key.ServiceName, resp.Key.Version)
		if resp.Action != string(pb.EVT_EXPIRE) {
			providerFlag = fmt.Sprintf("%s/%s(%s)", resp.Instance.ServiceId, resp.Instance.InstanceId, providerFlag)
		}
		util.Logger().Infof("event[%s] is coming in, watcher[%s] watch %s, subject: %s, group: %s",
			resp.Action, remoteAddr, providerFlag, wh.watcher.Subject(), wh.watcher.Group())

		resp.Response = nil
		data, err := json.Marshal(resp)
		if err != nil {
			wh.watcher.SetError(err)
			message = util.StringToBytesWithNoCopy(fmt.Sprintf("marshal output file error, %s", err.Error()))
			break
		}
		message = data
	default:
		wh.watcher.SetError(fmt.Errorf("unknown input %v", o))
		return
	}

	select {
	case <-wh.closed:
		return
	default:
	}

	err := wh.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		util.Logger().Errorf(err, "watcher[%s] catch an err, subject: %s, group: %s",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Group())
	}
}

func DoWebSocketListAndWatch(ctx context.Context, serviceId string, f func() ([]*pb.WatchInstanceResponse, int64), conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	socket := &WebSocket{
		ctx:             ctx,
		conn:            conn,
		watcher:         NewListWatcher(serviceId, apt.GetInstanceRootKey(domainProject)+"/", f),
		needPingWatcher: true,
		closed:          make(chan struct{}),
	}
	process(socket)
}

func process(socket *WebSocket) {
	if err := socket.Init(); err != nil {
		return
	}
	socket.HandleWatchWebSocketControlMessage()
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	util.Logger().Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		util.Logger().Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

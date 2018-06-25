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

type WebSocketHandler struct {
	ctx             context.Context
	conn            *websocket.Conn
	watcher         *ListWatcher
	needPingWatcher bool
	closed          chan struct{}
	goroutine       *util.GoRoutine
}

func (wh *WebSocketHandler) Init() error {
	remoteAddr := wh.conn.RemoteAddr().String()
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
	util.Logger().Debugf("start watching instance status, watcher[%s] %s %s",
		remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
	return nil
}

func (wh *WebSocketHandler) Timeout() time.Duration {
	return GetNotifyService().Config.NotifyTimeout
}

func (wh *WebSocketHandler) websocketHeartbeat(messageType int) error {
	err := wh.conn.WriteControl(messageType, []byte{}, time.Now().Add(wh.Timeout()))
	if err != nil {
		messageTypeName := "Ping"
		if messageType == websocket.PongMessage {
			messageTypeName = "Pong"
		}
		util.Logger().Errorf(err, "fail to send '%s' to watcher[%s] %s %s", messageTypeName,
			wh.conn.RemoteAddr(), wh.watcher.Subject(), wh.watcher.Id())
		//wh.watcher.SetError(err)
		return err
	}
	return nil
}

func (wh *WebSocketHandler) HandleWatchWebSocketControlMessage(ctx context.Context) {
	defer close(wh.closed)

	remoteAddr := wh.conn.RemoteAddr().String()
	// PING
	wh.conn.SetPingHandler(func(message string) error {
		if wh.needPingWatcher {
			util.Logger().Infof("received 'Ping' message '%s' from watcher[%s] %s %s, no longer send 'Ping' to it",
				message, remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
		}
		wh.needPingWatcher = false
		return wh.websocketHeartbeat(websocket.PongMessage)
	})
	// PONG
	wh.conn.SetPongHandler(func(message string) error {
		util.Logger().Debugf("received 'Pong' message %s from watcher[%s] %s %s",
			message, remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
		return nil
	})
	// CLOSE
	wh.conn.SetCloseHandler(func(code int, text string) error {
		util.Logger().Warnf(nil, "watcher[%s] %s %s active closed", remoteAddr,
			wh.watcher.Subject(), wh.watcher.Id())
		return wh.Close(code, text)
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := wh.conn.ReadMessage()
			if err != nil {
				wh.watcher.SetError(err)
				return
			}
		}
	}
}

func (wh *WebSocketHandler) HandleWatchWebSocketJob() {
	wh.goroutine.Do(wh.HandleWatchWebSocketControlMessage)

	remoteAddr := wh.conn.RemoteAddr().String()
	for {
		timer := time.NewTimer(wh.Timeout())
		select {
		case <-wh.closed:
			timer.Stop()
			return
		case <-wh.ctx.Done():
			timer.Stop()

			util.Logger().Warnf(nil, "handle timed out, watcher[%s] %s %s", remoteAddr,
				wh.watcher.Subject(), wh.watcher.Id())
			return
		case <-timer.C:
			if wh.watcher.Err() != nil {
				return
			}

			domainProject := util.ParseDomainProject(wh.ctx)
			if !serviceUtil.ServiceExist(context.Background(), domainProject, wh.watcher.Id()) {
				err := fmt.Errorf("Service does not exit.")
				util.Logger().Warnf(err, "watcher[%s] %s %s exit", remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				err = wh.conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error()))
				if err != nil {
					util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: write message error",
						remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				}
				return
			}

			if !wh.needPingWatcher {
				continue
			}

			util.Logger().Debugf("send 'Ping' message to watcher[%s] %s %s", remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
			err := wh.websocketHeartbeat(websocket.PingMessage)
			if err != nil {
				return
			}
		case job := <-wh.watcher.Job:
			timer.Stop()

			if wh.watcher.Err() != nil {
				return
			}

			if job == nil {
				err := wh.conn.WriteMessage(websocket.TextMessage,
					util.StringToBytesWithNoCopy("watcher catch an err: server shutdown"))
				if err != nil {
					util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: write message error",
						remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
					return
				}
				util.Logger().Warnf(nil, "watcher[%s] %s %s catch an err: server shutdown",
					remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				return
			}

			resp := job.Response

			providerFlag := fmt.Sprintf("%s/%s/%s", resp.Key.AppId, resp.Key.ServiceName, resp.Key.Version)
			if resp.Action != string(pb.EVT_EXPIRE) {
				providerFlag = fmt.Sprintf("%s/%s(%s)", resp.Instance.ServiceId, resp.Instance.InstanceId, providerFlag)
			}
			util.Logger().Infof("event[%s] is coming in, watcher[%s] %s %s, providers' info %s",
				resp.Action, remoteAddr, wh.watcher.Subject(), wh.watcher.Id(), providerFlag)

			resp.Response = nil
			data, err := json.Marshal(resp)
			if err != nil {
				util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: marshal output file error",
					remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				message := fmt.Sprintf("marshal output file error, %s", err.Error())
				err = wh.conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(message))
				if err != nil {
					util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: write message error",
						remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				}
				return
			}
			err = wh.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: write message error",
					remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
				return
			}
		}
	}
}

func (wh *WebSocketHandler) Close(code int, text string) error {
	defer wh.goroutine.Close(true)

	remoteAddr := wh.conn.RemoteAddr().String()
	var message []byte
	if code != websocket.CloseNoStatusReceived {
		message = websocket.FormatCloseMessage(code, text)
	}
	err := wh.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(wh.Timeout()))
	if err != nil {
		util.Logger().Errorf(err, "watcher[%s] %s %s catch an err: write 'Close' message error",
			remoteAddr, wh.watcher.Subject(), wh.watcher.Id())
		return err
	}
	return nil
}

func DoWebSocketListAndWatch(ctx context.Context, serviceId string, f func() ([]*pb.WatchInstanceResponse, int64), conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	handler := &WebSocketHandler{
		ctx:             ctx,
		conn:            conn,
		watcher:         NewListWatcher(serviceId, apt.GetInstanceRootKey(domainProject)+"/", f),
		needPingWatcher: true,
		closed:          make(chan struct{}),
		goroutine:       util.NewGo(context.Background()),
	}
	processHandler(handler)
}

func processHandler(handler *WebSocketHandler) {
	if err := handler.Init(); err != nil {
		return
	}
	handler.HandleWatchWebSocketJob()
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	util.Logger().Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		util.Logger().Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

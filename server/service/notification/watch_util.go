//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"time"
)

func HandleWatchJob(watcher *ListWatcher, stream pb.ServiceInstanceCtrl_WatchServer, timeout time.Duration) (err error) {
	for {
		select {
		case <-time.After(timeout):
		// TODO grpc 长连接心跳？
		case job := <-watcher.Job:
			if job == nil {
				err = errors.New("channel is closed")
				util.Logger().Errorf(err, "watcher %s %s caught an exception",
					watcher.Subject(), watcher.Id())
				return
			}
			resp := job.(*WatchJob).Response
			util.Logger().Infof("event is coming in, watcher %s %s",
				watcher.Subject(), watcher.Id())

			err = stream.Send(resp)
			if err != nil {
				util.Logger().Errorf(err, "send message error, watcher %s %s",
					watcher.Subject(), watcher.Id())
				watcher.SetError(err)
				return
			}
		}
	}
}

type WebSocketHandler struct {
	ctx             context.Context
	conn            *websocket.Conn
	watcher         *ListWatcher
	needPingWatcher bool
	closed          chan struct{}
}

func (wh *WebSocketHandler) Init() error {
	remoteAddr := wh.conn.RemoteAddr().String()
	if err := GetNotifyService().AddSubscriber(wh.watcher); err != nil {
		err = fmt.Errorf("establish[%s] websocket watch failed: notify service error, %s.",
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

func (wh *WebSocketHandler) HandleWatchWebSocketControlMessage() {
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
		_, _, err := wh.conn.ReadMessage()
		if err != nil {
			wh.watcher.SetError(err)
			return
		}
	}
}

func (wh *WebSocketHandler) HandleWatchWebSocketJob() {
	remoteAddr := wh.conn.RemoteAddr().String()

	for {
		select {
		case <-wh.closed:
			return
		case <-wh.ctx.Done():
			util.Logger().Warnf(nil, "handle timed out, watcher[%s] %s %s", remoteAddr,
				wh.watcher.Subject(), wh.watcher.Id())
			return
		case <-time.After(wh.Timeout()):
			if wh.watcher.Err() != nil {
				return
			}

			tenant := util.ParseTenantProject(wh.ctx)
			if !ms.ServiceExist(context.Background(), tenant, wh.watcher.Id()) {
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

			resp := job.(*WatchJob).Response

			providerFlag := fmt.Sprintf("%s/%s/%s", resp.Key.AppId, resp.Key.ServiceName, resp.Key.Version)
			if resp.Action != string(pb.EVT_EXPIRE) {
				providerFlag = fmt.Sprintf("%s/%s(%s)", resp.Instance.ServiceId, resp.Instance.InstanceId, providerFlag)
			}
			util.Logger().Warnf(nil, "event[%s] is coming in, watcher[%s] %s %s, providers' info %s",
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
	remoteAddr := wh.conn.RemoteAddr().String()
	message := []byte{}
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

func DoWebSocketWatch(ctx context.Context, serviceId string, conn *websocket.Conn) {
	tenant := util.ParseTenantProject(ctx)
	handler := &WebSocketHandler{
		ctx:             ctx,
		conn:            conn,
		watcher:         NewInstanceWatcher(serviceId, apt.GetInstanceRootKey(tenant)+"/"),
		needPingWatcher: true,
		closed:          make(chan struct{}),
	}
	processHandler(handler)
}

func DoWebSocketListAndWatch(ctx context.Context, serviceId string, f func() ([]*pb.WatchInstanceResponse, int64), conn *websocket.Conn) {
	tenant := util.ParseTenantProject(ctx)
	handler := &WebSocketHandler{
		ctx:             ctx,
		conn:            conn,
		watcher:         NewInstanceListWatcher(serviceId, apt.GetInstanceRootKey(tenant)+"/", f),
		needPingWatcher: true,
		closed:          make(chan struct{}),
	}
	processHandler(handler)
}

func processHandler(handler *WebSocketHandler) {
	if err := handler.Init(); err != nil {
		return
	}
	go handler.HandleWatchWebSocketControlMessage()
	handler.HandleWatchWebSocketJob()
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	util.Logger().Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		util.Logger().Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

func PublishInstanceEvent(tenant string, action pb.EventType, serviceKey *pb.MicroServiceKey, instance *pb.MicroServiceInstance, rev int64, subscribers []string) {
	response := &pb.WatchInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Watch instance successfully."),
		Action:   string(action),
		Key:      serviceKey,
		Instance: instance,
	}
	for _, consumerId := range subscribers {
		job := NewWatchJob(INSTANCE, consumerId, apt.GetInstanceRootKey(tenant)+"/", rev, response)
		util.Logger().Debugf("publish event to notify service, %v", job)

		// TODO add超时怎么处理？
		GetNotifyService().AddJob(job)
	}
}

func NewInstanceWatcher(selfServiceId, instanceRoot string) *ListWatcher {
	return NewWatcher(INSTANCE, selfServiceId, instanceRoot)
}

func NewInstanceListWatcher(selfServiceId, instanceRoot string, listFunc func() (results []*pb.WatchInstanceResponse, rev int64)) *ListWatcher {
	return NewListWatcher(INSTANCE, selfServiceId, instanceRoot, listFunc)
}

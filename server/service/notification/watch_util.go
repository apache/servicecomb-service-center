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
	"github.com/gorilla/websocket"
	apt "github.com/servicecomb/service-center/server/core"
	pb "github.com/servicecomb/service-center/server/core/proto"
	"github.com/servicecomb/service-center/util"
	"golang.org/x/net/context"
	"time"
)

func WatchJobHandler(watcher *Watcher, stream pb.ServiceInstanceCtrl_WatchServer, timeout time.Duration) error {
	for {
		select {
		case <-time.After(timeout):
		// TODO grpc 长连接心跳？
		case job := <-watcher.Job:
			resp := job.(*WatchJob).Response
			util.LOGGER.Infof("event is coming in, watcher %s %s",
				watcher.GetSubject(), watcher.GetId())

			err := stream.Send(resp)
			if err != nil {
				util.LOGGER.Errorf(err, "send message error, watcher %s %s",
					watcher.GetSubject(), watcher.GetId())
				watcher.SetError(err)
				return err
			}
		}
	}
}

func WatchWebSocketJobHandler(conn *websocket.Conn, watcher *Watcher, timeout time.Duration) {
	conn.SetPongHandler(func(message string) error {
		util.LOGGER.Debugf("receive heartbeat message %s from watcher %s %s",
			message, watcher.GetSubject(), watcher.GetId())
		return nil
	})
	for {
		select {
		case <-time.After(timeout):
			util.LOGGER.Debugf("send heartbeat to watcher %s %s", watcher.GetSubject(), watcher.GetId())
			err := conn.WriteControl(websocket.PingMessage, []byte("heartbeat"), time.Now().Add(timeout))
			if err != nil {
				util.LOGGER.Errorf(err, "watcher missing heartbeat, watcher %s %s",
					watcher.GetSubject(), watcher.GetId())
				watcher.SetError(err)
				return
			}
		case job := <-watcher.Job:
			if job == nil {
				err := conn.WriteMessage(websocket.TextMessage,
					[]byte("watch catch a err: watcher quit for server shutdown"))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
						watcher.GetSubject(), watcher.GetId())
				}
				return
			}
			resp := job.(*WatchJob).Response
			util.LOGGER.Warnf(nil, "event is coming in, watcher %s %s, providers' info %s %s",
				watcher.GetSubject(), watcher.GetId(), resp.Instance.ServiceId, resp.Instance.InstanceId)

			resp.Response = nil
			data, err := json.Marshal(resp)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: marshal output file error, watcher %s %s",
					watcher.GetSubject(), watcher.GetId())
				watcher.SetError(err)

				message := fmt.Sprintf("marshal output file error, %s", err.Error())
				err = conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
						watcher.GetSubject(), watcher.GetId())
				}
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
					watcher.GetSubject(), watcher.GetId())
				watcher.SetError(err)
				return
			}
		}
	}
}

func WatchPreOpera(in *pb.WatchInstanceRequest, ctx context.Context, server *NotifyService) (error, *Watcher) {
	if in == nil || len(in.SelfServiceId) == 0 {
		return errors.New("request format invalid"), nil
	}
	tenant := util.ParaseTenant(ctx)

	watcher := &Watcher{
		BaseNotifier: BaseNotifier{
			Id:      in.SelfServiceId,
			Subject: apt.GetInstanceRootKey(tenant) + "/",
			Server:  server,
		},
		Job: make(chan NotifyJob),
	}
	return nil, watcher
}

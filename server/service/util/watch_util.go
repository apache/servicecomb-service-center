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
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/websocket"
	apt "github.com/servicecomb/service-center/server/core"
	pb "github.com/servicecomb/service-center/server/core/proto"
	"github.com/servicecomb/service-center/server/core/registry"
	"github.com/servicecomb/service-center/server/service/microservice"
	nf "github.com/servicecomb/service-center/server/service/notification"
	"github.com/servicecomb/service-center/util"
	"golang.org/x/net/context"
	"strings"
	"time"
)

func WatchJobHandler(watcher *nf.ListWatcher, stream pb.ServiceInstanceCtrl_WatchServer, timeout time.Duration) (err error) {
	for {
		select {
		case <-time.After(timeout):
		// TODO grpc 长连接心跳？
		case job := <-watcher.Job:
			if job == nil {
				err = errors.New("channel is closed")
				util.LOGGER.Errorf(err, "watcher %s %s caught an exception",
					watcher.Subject(), watcher.Id())
				return
			}
			resp := job.(*nf.WatchJob).Response
			util.LOGGER.Infof("event is coming in, watcher %s %s",
				watcher.Subject(), watcher.Id())

			err = stream.Send(resp)
			if err != nil {
				util.LOGGER.Errorf(err, "send message error, watcher %s %s",
					watcher.Subject(), watcher.Id())
				watcher.SetError(err)
				return
			}
		}
	}
}

func WatchWebSocketJobHandler(conn *websocket.Conn, watcher *nf.ListWatcher, timeout time.Duration) {
	conn.SetPongHandler(func(message string) error {
		util.LOGGER.Debugf("receive heartbeat message %s from watcher %s %s",
			message, watcher.Subject(), watcher.Id())
		return nil
	})
	for {
		select {
		case <-time.After(timeout):
			util.LOGGER.Debugf("send heartbeat to watcher %s %s", watcher.Subject(), watcher.Id())
			err := conn.WriteControl(websocket.PingMessage, []byte("heartbeat"), time.Now().Add(timeout))
			if err != nil {
				util.LOGGER.Errorf(err, "watcher missing heartbeat, watcher %s %s",
					watcher.Subject(), watcher.Id())
				watcher.SetError(err)
				return
			}
		case job := <-watcher.Job:
			if job == nil {
				err := conn.WriteMessage(websocket.TextMessage,
					[]byte("watch catch a err: watcher quit for server shutdown"))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
						watcher.Subject(), watcher.Id())
				}
				return
			}
			resp := job.(*nf.WatchJob).Response
			util.LOGGER.Warnf(nil, "event is coming in, watcher %s %s, providers' info %s %s",
				watcher.Subject(), watcher.Id(), resp.Instance.ServiceId, resp.Instance.InstanceId)

			resp.Response = nil
			data, err := json.Marshal(resp)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: marshal output file error, watcher %s %s",
					watcher.Subject(), watcher.Id())
				watcher.SetError(err)

				message := fmt.Sprintf("marshal output file error, %s", err.Error())
				err = conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
						watcher.Subject(), watcher.Id())
				}
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher %s %s",
					watcher.Subject(), watcher.Id())
				watcher.SetError(err)
				return
			}
		}
	}
}

func DoWebSocketWatch(service *nf.NotifyService, watcher *nf.ListWatcher, conn *websocket.Conn) {
	if err := service.AddNotifier(watcher); err != nil {
		err = fmt.Errorf("establish web socket watch failed: notify service error, %s.", err.Error())
		util.LOGGER.Errorf(nil, err.Error())

		err = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		if err != nil {
			util.LOGGER.Errorf(err, "establish web socket watch failed: write message failed.")
		}
		return
	}
	util.LOGGER.Infof("start watching instance status, watcher %s %s", watcher.Subject(), watcher.Id())
	WatchWebSocketJobHandler(conn, watcher, service.Config.NotifyTimeout)
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	util.LOGGER.Errorf(err, "establish web socket watch failed.")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
		util.LOGGER.Errorf(err, "establish web socket watch failed: write message failed.")
	}
}

func QueryAllProvidersIntances(ctx context.Context, selfServiceId string) (results []*pb.WatchInstanceResponse, rev int64) {
	results = []*pb.WatchInstanceResponse{}

	tenant := util.ParaseTenantProject(ctx)

	key := apt.GenerateConsumerDependencyKey(tenant, selfServiceId, "")
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		KeyOnly:    true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get %s providers id set failed.", selfServiceId)
		return
	}

	rev = resp.Revision

	for _, depsKv := range resp.Kvs {
		providerDepsKey := string(depsKv.Key)
		providerId := providerDepsKey[strings.LastIndex(providerDepsKey, "/")+1:]

		service, err := microservice.GetById(tenant, providerId, rev)
		if service == nil {
			return
		}
		util.LOGGER.Debugf("query provider service %v with revision %d.", service, rev)

		kvs, err := queryServiceInstancesKvs(ctx, providerId, rev)
		if err != nil {
			return
		}

		util.LOGGER.Debugf("query provider service %s instances[%d] with revision %d.", providerId, len(kvs), rev)
		for _, kv := range kvs {
			util.LOGGER.Debugf("start unmarshal service instance file with revision %d: %s",
				rev, string(kv.Key))
			instance := &pb.MicroServiceInstance{}
			err := json.Unmarshal(kv.Value, instance)
			if err != nil {
				util.LOGGER.Errorf(err, "unmarshal instance of service %s with revision %d failed.",
					providerId, rev)
				return
			}
			results = append(results, &pb.WatchInstanceResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "list instance successfully"),
				Action:   string(pb.EVT_CREATE),
				Key: &pb.MicroServiceKey{
					AppId:       service.AppId,
					ServiceName: service.ServiceName,
					Version:     service.Version,
				},
				Instance: instance,
			})
		}
	}
	return
}

func queryServiceInstancesKvs(ctx context.Context, serviceId string, rev int64) ([]*mvccpb.KeyValue, error) {
	tenant := util.ParaseTenantProject(ctx)
	key := apt.GenerateInstanceKey(tenant, serviceId, "")
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		WithRev:    rev,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "query instance of service %s with revision %d from etcd failed.",
			serviceId, rev)
		return nil, err
	}
	return resp.Kvs, nil
}

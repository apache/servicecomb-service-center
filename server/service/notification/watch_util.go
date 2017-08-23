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
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"strings"
	"time"
)

func WatchJobHandler(watcher *ListWatcher, stream pb.ServiceInstanceCtrl_WatchServer, timeout time.Duration) (err error) {
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
			resp := job.(*WatchJob).Response
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

func websocketHeartbeat(conn *websocket.Conn, messageType int, watcher *ListWatcher, timeout time.Duration) error {
	err := conn.WriteControl(messageType, util.StringToBytesWithNoCopy("heartbeat"), time.Now().Add(timeout))
	if err != nil {
		messageTypeName := "Ping"
		if messageType == websocket.PongMessage {
			messageTypeName = "Pong"
		}
		util.LOGGER.Errorf(err, "fail to send '%s' to watcher[%s] %s %s",
			messageTypeName, conn.RemoteAddr(), watcher.Subject(), watcher.Id())
		watcher.SetError(err)
		return err
	}
	return nil
}

func WatchWebSocketJobHandler(conn *websocket.Conn, watcher *ListWatcher, timeout time.Duration) {
	remoteAddr := conn.RemoteAddr().String()
	conn.SetPongHandler(func(message string) error {
		util.LOGGER.Debugf("receive heartbeat feedback message %s from watcher[%s] %s %s",
			message, remoteAddr, watcher.Subject(), watcher.Id())
		return nil
	})
	conn.SetPingHandler(func(message string) error {
		util.LOGGER.Debugf("receive heartbeat message %s from watcher[%s] %s %s, now give it a reply",
			message, remoteAddr, watcher.Subject(), watcher.Id())
		return websocketHeartbeat(conn, websocket.PongMessage, watcher, timeout)
	})
	for {
		select {
		case <-time.After(timeout):
			util.LOGGER.Debugf("send heartbeat to watcher[%s] %s %s", remoteAddr, watcher.Subject(), watcher.Id())
			err := websocketHeartbeat(conn, websocket.PingMessage, watcher, timeout)
			if err != nil {
				return
			}
		case job := <-watcher.Job:
			if job == nil {
				err := conn.WriteMessage(websocket.TextMessage,
					util.StringToBytesWithNoCopy("watch catch a err: watcher quit for server shutdown"))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher[%s] %s %s",
						remoteAddr, watcher.Subject(), watcher.Id())
					return
				}
				util.LOGGER.Warnf(nil, "watch catch a err: server shutdown, watcher[%s] %s %s",
					remoteAddr, watcher.Subject(), watcher.Id())
				return
			}
			resp := job.(*WatchJob).Response
			util.LOGGER.Warnf(nil, "event[%s] is coming in, watcher[%s] %s %s, providers' info %s %s",
				resp.Action, remoteAddr, watcher.Subject(), watcher.Id(), resp.Instance.ServiceId, resp.Instance.InstanceId)

			resp.Response = nil
			data, err := json.Marshal(resp)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: marshal output file error, watcher[%s] %s %s",
					remoteAddr, watcher.Subject(), watcher.Id())
				watcher.SetError(err)

				message := fmt.Sprintf("marshal output file error, %s", err.Error())
				err = conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(message))
				if err != nil {
					util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher[%s] %s %s",
						remoteAddr, watcher.Subject(), watcher.Id())
				}
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				util.LOGGER.Errorf(err, "watch catch a err: write message error, watcher[%s] %s %s",
					remoteAddr, watcher.Subject(), watcher.Id())
				watcher.SetError(err)
				return
			}
		}
	}
}

func DoWebSocketWatch(service *NotifyService, watcher *ListWatcher, conn *websocket.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	if err := service.AddSubscriber(watcher); err != nil {
		err = fmt.Errorf("establish[%s] websocket watch failed: notify service error, %s.",
			remoteAddr, err.Error())
		util.LOGGER.Errorf(nil, err.Error())

		err = conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error()))
		if err != nil {
			util.LOGGER.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
		}
		return
	}
	util.LOGGER.Debugf("start watching instance status, watcher[%s] %s %s",
		remoteAddr, watcher.Subject(), watcher.Id())
	WatchWebSocketJobHandler(conn, watcher, service.Config.NotifyTimeout)
}

func EstablishWebSocketError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	util.LOGGER.Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		util.LOGGER.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}

func QueryAllProvidersIntances(ctx context.Context, selfServiceId string) (results []*pb.WatchInstanceResponse, rev int64) {
	results = []*pb.WatchInstanceResponse{}

	tenant := util.ParseTenantProject(ctx)

	rev = store.Revision()

	key := apt.GenerateConsumerDependencyKey(tenant, selfServiceId, "")
	resp, err := store.Store().Dependency().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithRev:    rev,
		WithPrefix: true,
		KeyOnly:    true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get %s providers id set failed.", selfServiceId)
		return
	}

	for _, depsKv := range resp.Kvs {
		providerDepsKey := util.BytesToStringWithNoCopy(depsKv.Key)
		providerId := providerDepsKey[strings.LastIndex(providerDepsKey, "/")+1:]

		service, err := ms.GetService(ctx, tenant, providerId, rev)
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
				rev, util.BytesToStringWithNoCopy(kv.Key))
			instance := &pb.MicroServiceInstance{}
			err := json.Unmarshal(kv.Value, instance)
			if err != nil {
				util.LOGGER.Errorf(err, "unmarshal instance of service %s with revision %d failed.",
					providerId, rev)
				return
			}
			results = append(results, &pb.WatchInstanceResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "List instance successfully."),
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
	tenant := util.ParseTenantProject(ctx)
	key := apt.GenerateInstanceKey(tenant, serviceId, "")
	resp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
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

func PublishInstanceEvent(service *NotifyService, tenant string, action pb.EventType, serviceKey *pb.MicroServiceKey, instance *pb.MicroServiceInstance, rev int64, subscribers []string) {
	response := &pb.WatchInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Watch instance successfully."),
		Action:   string(action),
		Key:      serviceKey,
		Instance: instance,
	}
	for _, consumerId := range subscribers {
		job := NewWatchJob(INSTANCE, consumerId, apt.GetInstanceRootKey(tenant)+"/", rev, response)
		util.LOGGER.Debugf("publish event to notify service, %v", job)

		// TODO add超时怎么处理？
		service.AddJob(job)
	}
}

func NewInstanceWatcher(selfServiceId, instanceRoot string) *ListWatcher {
	return NewWatcher(INSTANCE, selfServiceId, instanceRoot)
}

func NewInstanceListWatcher(selfServiceId, instanceRoot string, listFunc func() (results []*pb.WatchInstanceResponse, rev int64)) *ListWatcher {
	return NewListWatcher(INSTANCE, selfServiceId, instanceRoot, listFunc)
}

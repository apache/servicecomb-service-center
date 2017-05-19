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
package proto

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/websocket"
	"strings"
	"golang.org/x/net/context"
)

const (
	EVT_CREATE string = "CREATE"
	EVT_UPDATE string = "UPDATE"
	EVT_DELETE string = "DELETE"
	MS_UP      string = "UP"
	MS_DOWN    string = "DOWN"

	MSI_UP           string = "UP"
	MSI_DOWN         string = "DOWN"
	MSI_STARTING     string = "STARTING"
	MSI_OUTOFSERVICE string = "OUTOFSERVICE"

	CHECK_BY_HEARTBEAT string = "push"
	CHECK_BY_PLATFORM  string = "pull"

	EXISTENCE_MS     string = "microservice"
	EXISTENCE_SCHEMA string = "schema"

	PROP_ALLOW_CROSS_APP = "allowCrossApp"
)

type SerivceInstanceCtrlServerEx interface {
	ServiceInstanceCtrlServer

	WebSocketWatch(ctx context.Context, in *WatchInstanceRequest, conn *websocket.Conn)
}

type GovernServiceCtrlServerEx interface {
	GovernServiceCtrlServer
}

func CreateResponse(code Response_Code, message string) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}
	return resp
}

func EventToResponse(evt *mvccpb.Event) (keys []string, action string, data []byte) {
	keys = strings.Split(string(evt.Kv.Key), "/")
	switch {
	case evt.Type == mvccpb.PUT && evt.Kv.Version == 1:
		action = EVT_CREATE
		data = evt.Kv.Value
		return
	case evt.Type == mvccpb.PUT:
		action = EVT_UPDATE
		data = evt.Kv.Value
		return
	case evt.Type == mvccpb.DELETE:
		action = EVT_DELETE
		if evt.PrevKv == nil {
			// TODO 内嵌无法获取
			return
		}
		data = evt.PrevKv.Value
		return
	}
	return
}

func GetInfoFromInstChangedEvent(evt *mvccpb.Event)  (serviceId, instanceId, tenantProject, action string, data []byte) {
	keys, action, data := EventToResponse(evt)
	if len(keys) < 7 {
		return
	}
	serviceId = keys[len(keys)-2]
	instanceId = keys[len(keys)-1]
	tenantProject = strings.Join([]string{keys[len(keys) - 4 ], keys[len(keys)-3]}, "/")
	keys, action, data = EventToResponse(evt)
	return
}

func GetInfoFromTenantChangeEvent(evt *mvccpb.Event) (tenant string, action string, data []byte) {
	keys, action, data := EventToResponse(evt)
	if len(keys) < 3 {
		return
	}
	tenant = keys[len(keys)-1]
	return
}

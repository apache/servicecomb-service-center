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
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"strings"
)

type EventType string

const (
	EVT_CREATE EventType = "CREATE"
	EVT_UPDATE EventType = "UPDATE"
	EVT_DELETE EventType = "DELETE"
	EVT_ERROR  EventType = "ERROR"
	MS_UP      string    = "UP"
	MS_DOWN    string    = "DOWN"

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
	WebSocketListAndWatch(ctx context.Context, in *WatchInstanceRequest, conn *websocket.Conn)
	CluterHealth(ctx context.Context) (*GetInstancesResponse, error)
}

type GovernServiceCtrlServerEx interface {
	GovernServiceCtrlServer
}

type MicroServiceDependency struct {
	Dependency []*MicroServiceKey
}

type SystemConfig struct {
	Version string `json:"version"`
}

func CreateResponse(code Response_Code, message string) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}
	return resp
}

func KvToResponse(kv *mvccpb.KeyValue) (keys []string, data []byte) {
	keys = strings.Split(string(kv.Key), "/")
	data = kv.Value
	return
}

func GetInfoFromInstKV(kv *mvccpb.KeyValue) (serviceId, instanceId, tenantProject string, data []byte) {
	keys, data := KvToResponse(kv)
	if len(keys) < 4 {
		return
	}
	l := len(keys)
	serviceId = keys[l-2]
	instanceId = keys[l-1]
	tenantProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(kv *mvccpb.KeyValue) (tenant string, data []byte) {
	keys, data := KvToResponse(kv)
	if len(keys) < 1 {
		return
	}
	tenant = keys[len(keys)-1]
	return
}

func GetInfoFromRuleKV(kv *mvccpb.KeyValue) (serviceId, ruleId, tenantProject string, data []byte) {
	keys, data := KvToResponse(kv)
	if len(keys) < 4 {
		return
	}
	l := len(keys)
	serviceId = keys[l-2]
	ruleId = keys[l-1]
	tenantProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func TransferToMicroServiceKeys(in []*DependencyMircroService, tenant string) []*MicroServiceKey {
	rst := []*MicroServiceKey{}
	for _, value := range in {
		rst = append(rst, &MicroServiceKey{
			Tenant:      tenant,
			AppId:       value.AppId,
			ServiceName: value.ServiceName,
			Version:     value.Version,
		})
	}
	return rst
}

func ToMicroServiceKey(tenant string, in *MicroService) *MicroServiceKey {
	return &MicroServiceKey{
		Tenant:      tenant,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Version:     in.Version,
	}
}

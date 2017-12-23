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
package proto

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"strings"
)

type EventType string

const (
	EVT_INIT   EventType = "INIT"
	EVT_CREATE EventType = "CREATE"
	EVT_UPDATE EventType = "UPDATE"
	EVT_DELETE EventType = "DELETE"
	EVT_EXPIRE EventType = "EXPIRE"
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

	Response_SUCCESS int32 = 0

	ENV_DEV    string = "development"
	ENV_TEST   string = "testing"
	ENV_ACCEPT string = "acceptance"
	ENV_PROD   string = "production"

	REGISTERBY_SDK      string = "SDK"
	REGISTERBY_PLATFORM string = "PLATFORM"
	REGISTERBY_SIDECAR  string = "SIDECAR"
	REGISTERBY_UNKNOWM  string = "UNKNOWN"

	FRAMEWORK_UNKNOWN string = "UNKNOWN"
)

type SerivceInstanceCtrlServerEx interface {
	ServiceInstanceCtrlServer

	WebSocketWatch(ctx context.Context, in *WatchInstanceRequest, conn *websocket.Conn)
	WebSocketListAndWatch(ctx context.Context, in *WatchInstanceRequest, conn *websocket.Conn)
	ClusterHealth(ctx context.Context) (*GetInstancesResponse, error)
}

type GovernServiceCtrlServerEx interface {
	GovernServiceCtrlServer
}

type MicroServiceDependency struct {
	Dependency []*MicroServiceKey
}

type ServerConfig struct {
	MaxHeaderBytes int64 `json:"maxHeaderBytes"`
	MaxBodyBytes   int64 `json:"maxBodyBytes"`

	ReadHeaderTimeout string `json:"readHeaderTimeout"`
	ReadTimeout       string `json:"readTimeout"`
	IdleTimeout       string `json:"idleTimeout"`
	WriteTimeout      string `json:"writeTimeout"`

	LimitTTLUnit     string `json:"limitTTLUnit"`
	LimitConnections int64  `json:"limitConnections"`
	LimitIPLookup    string `json:"limitIPLookup"`

	SslEnabled    bool   `json:"sslEnabled,string"`
	SslMinVersion string `json:"sslMinVersion"`
	SslVerifyPeer bool   `json:"sslVerifyPeer,string"`
	SslCiphers    string `json:"sslCiphers"`

	AutoSyncInterval  string `json:"autoSyncInterval"`
	CompactIndexDelta int64  `json:"compactIndexDelta"`

	LoggerName     string `json:"-"`
	LogRotateSize  int64  `json:"logRotateSize"`
	LogBackupCount int64  `json:"logBackupCount"`
	LogFilePath    string `json:"-"`
	LogLevel       string `json:"-"`
	LogFormat      string `json:"-"`
	LogSys         bool   `json:"-"`

	PluginsDir string `json:"-"`
}

func (c *ServerConfig) LogPrint() {

}

type ServerInformation struct {
	Version string        `json:"version"`
	Config  *ServerConfig `json:"-"`
}

func CreateResponse(code int32, message string) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}
	return resp
}

func KvToResponse(kv *mvccpb.KeyValue) (keys []string, data []byte) {
	keys = strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
	data = kv.Value
	return
}

func GetInfoFromSvcKV(kv *mvccpb.KeyValue) (serviceId, domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromInstKV(kv *mvccpb.KeyValue) (serviceId, instanceId, domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	instanceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(kv *mvccpb.KeyValue) (domain string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 1 {
		return
	}
	domain = keys[l-1]
	return
}

func GetInfoFromProjectKV(kv *mvccpb.KeyValue) (domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-2], keys[l-1])
	return
}

func GetInfoFromRuleKV(kv *mvccpb.KeyValue) (serviceId, ruleId, domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	ruleId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromTagKV(kv *mvccpb.KeyValue) (serviceId, domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 3 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromSvcIndexKV(kv *mvccpb.KeyValue) (key *MicroServiceKey, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 6 {
		return
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}, data
}

func GetInfoFromSchemaSummaryKV(kv *mvccpb.KeyValue) (schemaId string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 1 {
		return
	}

	return keys[l-1], data
}

func GetInfoFromSchemaKV(kv *mvccpb.KeyValue) (schemaId string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 1 {
		return
	}

	return keys[l-1], data
}

func GetInfoFromDependencyQueueKV(kv *mvccpb.KeyValue) (consumerId, domainProject string, data []byte) {
	keys, data := KvToResponse(kv)
	l := len(keys)
	if l < 3 {
		return
	}

	return keys[l-2], keys[l-3], data
}

func DependenciesToKeys(in []*DependencyKey, domainProject string) []*MicroServiceKey {
	rst := []*MicroServiceKey{}
	for _, value := range in {
		rst = append(rst, &MicroServiceKey{
			Tenant:      domainProject,
			Environment: value.Environment,
			AppId:       value.AppId,
			ServiceName: value.ServiceName,
			Version:     value.Version,
		})
	}
	return rst
}

func MicroServiceToKey(domainProject string, in *MicroService) *MicroServiceKey {
	return &MicroServiceKey{
		Tenant:      domainProject,
		Environment: in.Environment,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Version:     in.Version,
	}
}

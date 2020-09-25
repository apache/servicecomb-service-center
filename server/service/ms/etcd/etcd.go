// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"strconv"
	"time"
)

// TODO: define error with names here

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
}

type DataSource struct{}

func NewDataSource() *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("microservice mgt data source enable etcd mode")

	inst := &DataSource{}
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	return nil
}

// RegisterService() implement:
// 1. capsule request to etcd kv format
// 2. invoke registry interface to store service information to etcd cluster
// Attention: parameters validation && response check must be checked outside the method
func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*registry.PluginResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceBody := request.Service
	serviceFlag := util.StringJoin([]string{
		serviceBody.Environment, serviceBody.AppId, serviceBody.ServiceName, serviceBody.Version}, "/")

	serviceUtil.SetServiceDefaultValue(serviceBody)

	domainProject := util.ParseDomainProject(ctx)

	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: serviceBody.Environment,
		AppId:       serviceBody.AppId,
		ServiceName: serviceBody.ServiceName,
		Alias:       serviceBody.Alias,
		Version:     serviceBody.Version,
	}

	index := apt.GenerateServiceIndexKey(serviceKey)

	// 产生全局service id
	requestServiceID := serviceBody.ServiceId
	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, index)
		serviceBody.ServiceId = plugin.Plugins().UUID().GetServiceID(ctx)
	}
	serviceBody.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	serviceBody.ModTimestamp = serviceBody.Timestamp

	data, err := json.Marshal(serviceBody)
	if err != nil {
		log.Errorf(err, "create micro-serviceBody[%s] failed, json marshal serviceBody failed, operator: %s",
			serviceFlag, remoteIP)
		return nil, err
	}

	key := apt.GenerateServiceKey(domainProject, serviceBody.ServiceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(serviceKey))

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithKey(keyBytes), registry.WithValue(data)),
		registry.OpPut(registry.WithKey(indexBytes), registry.WithStrValue(serviceBody.ServiceId)),
	}
	uniqueCmpOpts := []registry.CompareOp{
		registry.OpCmp(registry.CmpVer(indexBytes), registry.CmpEqual, 0),
		registry.OpCmp(registry.CmpVer(keyBytes), registry.CmpEqual, 0),
	}
	failOpts := []registry.PluginOp{
		registry.OpGet(registry.WithKey(indexBytes)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, registry.OpPut(registry.WithKey(aliasBytes), registry.WithStrValue(serviceBody.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			registry.OpCmp(registry.CmpVer(aliasBytes), registry.CmpEqual, 0))
		failOpts = append(failOpts, registry.OpGet(registry.WithKey(aliasBytes)))
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)

	return resp, err
}

func (ds *DataSource) GetService(ctx context.Context, service *pb.GetServiceRequest) {
	panic("implement me")
}

func (ds *DataSource) UpdateService(ctx context.Context, service *pb.UpdateServicePropsRequest) {
	panic("implement me")
}

func (ds *DataSource) UnregisterService(ctx context.Context, service *pb.DeleteServiceRequest) {
	panic("implement me")
}

func (ds *DataSource) RegisterInstance() {
	panic("implement me")
}

func (ds *DataSource) SearchInstance() {
	panic("implement me")
}

func (ds *DataSource) UpdateInstance() {
	panic("implement me")
}

func (ds *DataSource) UnRegisterInstance() {
	panic("implement me")
}

func (ds *DataSource) AddSchema() {
	panic("implement me")
}

func (ds *DataSource) GetSchema() {
	panic("implement me")
}

func (ds *DataSource) UpdateSchema() {
	panic("implement me")
}

func (ds *DataSource) DeleteSchema() {
	panic("implement me")
}

func (ds *DataSource) AddTag() {
	panic("implement me")
}

func (ds *DataSource) GetTag() {
	panic("implement me")
}

func (ds *DataSource) UpdateTag() {
	panic("implement me")
}

func (ds *DataSource) DeleteTag() {
	panic("implement me")
}

func (ds *DataSource) AddRule() {
	panic("implement me")
}

func (ds *DataSource) GetRule() {
	panic("implement me")
}

func (ds *DataSource) UpdateRule() {
	panic("implement me")
}

func (ds *DataSource) DeleteRule() {
	panic("implement me")
}

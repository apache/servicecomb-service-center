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

package resource

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	pb "github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	Instance = "instance"
)

func NewInstance(e *v1sync.Event) Resource {
	i := &instance{
		event: e,
	}
	i.manager = new(metadataManage)
	return i
}

type instance struct {
	event *v1sync.Event

	createInput *pb.RegisterInstanceRequest
	updateInput *pb.MicroServiceInstance
	deleteInput *pb.UnregisterInstanceRequest

	serviceID  string
	instanceID string

	service *pb.MicroService
	cur     *pb.MicroServiceInstance

	manager metadataManager
}

func (i *instance) loadInput() error {
	i.createInput = new(pb.RegisterInstanceRequest)
	cre := newInputParam(i.createInput, func() {
		i.serviceID = i.createInput.Instance.ServiceId
		i.instanceID = i.createInput.Instance.InstanceId
	})

	i.updateInput = new(pb.MicroServiceInstance)
	upd := newInputParam(i.updateInput, func() {
		i.serviceID = i.updateInput.ServiceId
		i.instanceID = i.updateInput.InstanceId
	})

	i.deleteInput = new(pb.UnregisterInstanceRequest)
	del := newInputParam(i.deleteInput, func() {
		i.serviceID = i.deleteInput.ServiceId
		i.instanceID = i.deleteInput.InstanceId
	})

	return newInputLoader(
		i.event,
		cre,
		upd,
		del,
	).loadInput()
}

type metadataManage struct {
}

func (m *metadataManage) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	return datasource.GetMetadataManager().RegisterService(ctx, request)
}

func (m *metadataManage) GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.MicroService, error) {
	return datasource.GetMetadataManager().GetService(ctx, in)
}

func (m *metadataManage) PutServiceProperties(ctx context.Context, request *pb.UpdateServicePropsRequest) error {
	return datasource.GetMetadataManager().PutServiceProperties(ctx, request)
}

func (m *metadataManage) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) error {
	return datasource.GetMetadataManager().UnregisterService(ctx, request)
}

func (m *metadataManage) RegisterInstance(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	return datasource.GetMetadataManager().RegisterInstance(ctx, in)
}

func (m *metadataManage) SendHeartbeat(ctx context.Context, in *pb.HeartbeatRequest) error {
	return datasource.GetMetadataManager().SendHeartbeat(ctx, in)
}

func (m *metadataManage) GetInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	return datasource.GetMetadataManager().GetInstance(ctx, in)
}

func (m *metadataManage) PutInstance(ctx context.Context, in *pb.RegisterInstanceRequest) error {
	return datasource.GetMetadataManager().PutInstance(ctx, in)
}

func (m *metadataManage) UnregisterInstance(ctx context.Context, in *pb.UnregisterInstanceRequest) error {
	return datasource.GetMetadataManager().UnregisterInstance(ctx, in)
}

func (m *metadataManage) RegisterEnvironment(ctx context.Context, in *ev.CreateEnvironmentRequest) (*ev.CreateEnvironmentResponse, error) {
	return datasource.GetMetadataManager().RegisterEnvironment(ctx, in)
}

func (m *metadataManage) GetEnvironment(ctx context.Context, in *ev.GetEnvironmentRequest) (*ev.Environment, error) {
	return datasource.GetMetadataManager().GetEnvironment(ctx, in)
}

func (m *metadataManage) UpdateEnvironment(ctx context.Context, in *ev.UpdateEnvironmentRequest) error {
	return datasource.GetMetadataManager().UpdateEnvironment(ctx, in)
}

func (m *metadataManage) UnregisterEnvironment(ctx context.Context, in *ev.DeleteEnvironmentRequest) error {
	return datasource.GetMetadataManager().UnregisterEnvironment(ctx, in)
}

type metadataManager interface {
	serviceManager
	instanceManager
}

type instanceManager interface {
	RegisterInstance(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error)
	SendHeartbeat(ctx context.Context, in *pb.HeartbeatRequest) error
	GetInstance(ctx context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error)
	PutInstance(ctx context.Context, in *pb.RegisterInstanceRequest) error
	UnregisterInstance(ctx context.Context, in *pb.UnregisterInstanceRequest) error
}

func (i *instance) LoadCurrentResource(ctx context.Context) *Result {
	err := i.loadInput()
	if err != nil {
		return FailResult(err)
	}

	serviceID := i.serviceID
	service, err := i.manager.GetService(ctx,
		&pb.GetServiceRequest{
			ServiceId: serviceID,
		})
	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists) {
			log.Warn(fmt.Sprintf("instance service not exist, %s", i.event.Flag()))
			return MicroNonExistResult()
		}
		return FailResult(err)
	}
	i.service = service

	inst, err := i.manager.GetInstance(ctx,
		&pb.GetOneInstanceRequest{
			ProviderServiceId:  serviceID,
			ProviderInstanceId: i.instanceID,
		})
	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrInstanceNotExists) {
			return nil
		}

		return FailResult(err)
	}

	i.cur = inst.Instance
	return nil
}

func (i *instance) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil: i.cur != nil,
		event:     i.event,
		updateTime: func() (int64, error) {
			return formatUpdateTimeSecond(i.cur.ModTimestamp)
		},
		resourceID: "",
	}
	c.tombstoneLoader = c
	return c.needOperate(ctx)
}

func (i *instance) FailHandle(ctx context.Context, code int32) (*v1sync.Event, error) {
	if code != MicroNonExist {
		return nil, nil
	}

	err := i.loadInput()
	if err != nil {
		return nil, err
	}

	ctx = util.SetDomainProject(ctx,
		i.event.Opts[string(util.CtxDomain)],
		i.event.Opts[string(util.CtxProject)])

	serviceID := i.serviceID
	_, err = i.manager.GetService(ctx,
		&pb.GetServiceRequest{
			ServiceId: serviceID,
		})

	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists) {
			log.Warn(fmt.Sprintf("service not exist %s, %s", serviceID, i.event.Flag()))
			return nil, nil
		}
		return nil, err
	}
	return i.event, nil
}

func (i *instance) CanDrop() bool {
	return false
}

func (i *instance) Operate(ctx context.Context) *Result {
	return newOperator(i).operate(ctx, i.event.Action)
}

func (i *instance) CreateHandle(ctx context.Context) error {
	_, err := i.manager.RegisterInstance(ctx, i.createInput)
	return err
}

func (i *instance) UpdateHandle(ctx context.Context) error {
	return i.manager.PutInstance(ctx,
		&pb.RegisterInstanceRequest{
			Instance: i.updateInput,
		})
}

func (i *instance) DeleteHandle(ctx context.Context) error {
	return i.manager.UnregisterInstance(ctx, i.deleteInput)
}

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

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	Microservice = "service"
)

func NewMicroservice(e *v1sync.Event) Resource {
	m := &microservice{
		event: e,
	}
	m.manager = new(metadataManage)
	return m
}

type microservice struct {
	event *v1sync.Event

	createInput *pb.CreateServiceRequest
	updateInput *pb.UpdateServicePropsRequest
	deleteInput *pb.DeleteServiceRequest

	serviceID string

	cur *pb.MicroService

	manager serviceManager

	defaultFailHandler
}

type serviceManager interface {
	RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error)
	GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.MicroService, error)
	PutServiceProperties(ctx context.Context, request *pb.UpdateServicePropsRequest) error
	UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) error
}

func (m *microservice) loadInput() error {
	m.createInput = new(pb.CreateServiceRequest)
	cre := newInputParam(m.createInput, func() {
		m.serviceID = m.createInput.Service.ServiceId
	})

	m.updateInput = new(pb.UpdateServicePropsRequest)
	upd := newInputParam(m.updateInput, func() {
		m.serviceID = m.updateInput.ServiceId
	})

	m.deleteInput = new(pb.DeleteServiceRequest)
	del := newInputParam(m.deleteInput, func() {
		m.serviceID = m.deleteInput.ServiceId
	})

	return newInputLoader(
		m.event,
		cre,
		upd,
		del,
	).loadInput()
}

func (m *microservice) LoadCurrentResource(ctx context.Context) *Result {
	err := m.loadInput()
	if err != nil {
		return FailResult(err)
	}

	cur, err := m.manager.GetService(ctx, &pb.GetServiceRequest{
		ServiceId: m.serviceID,
	})
	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists) {
			return nil
		}

		return FailResult(err)
	}
	m.cur = cur
	return nil
}

func (m *microservice) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil: m.cur != nil,
		event:     m.event,
		updateTime: func() (int64, error) {
			return formatUpdateTimeSecond(m.cur.ModTimestamp)
		},
		resourceID: m.serviceID,
	}
	c.tombstoneLoader = c
	return c.needOperate(ctx)
}

func (m *microservice) CreateHandle(ctx context.Context) error {
	_, err := m.manager.RegisterService(ctx, m.createInput)
	return err
}

func (m *microservice) UpdateHandle(ctx context.Context) error {
	return m.manager.PutServiceProperties(ctx, m.updateInput)
}

func (m *microservice) DeleteHandle(ctx context.Context) error {
	return m.manager.UnregisterService(ctx, m.deleteInput)
}

func (m *microservice) Operate(ctx context.Context) *Result {
	return newOperator(m).operate(ctx, m.event.Action)
}

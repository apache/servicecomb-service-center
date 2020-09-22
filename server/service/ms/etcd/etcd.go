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
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
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
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	return nil
}

func (ds *DataSource) RegisterService(ctx context.Context, service *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if service == nil || service.Service == nil {
		log.Errorf(nil, "create micro-service failed: request body is empty")
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Request body is empty"),
		}, nil
	}
	return nil, nil
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

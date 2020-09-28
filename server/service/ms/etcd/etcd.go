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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/backend"
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
// 2. invoke etcd client to store data
// 3. check etcd-client response && construct createServiceResponse
func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	// start to store service to etcd
	serviceBody := request.Service

	// construct data to invoke etcd client
	opts, uniqueCmpOpts, failOpts, err := CapRegisterData(ctx, request)
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	resp, err := backend.Registry().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)

	return NewRegisterServiceResp(ctx, serviceBody, resp, err)
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	services, err := GetAllServiceUtil(ctx, domainProject)
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	singleService, err := GetSingleService(ctx, domainProject, request.ServiceId)

	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed, get service file failed", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if singleService == nil {
		log.Errorf(nil, "get micro-service[%s] failed, service does not exist", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service successfully."),
		Service:  singleService,
	}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, service *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {
	panic("implement me")
}

func (ds *DataSource) UnregisterService(ctx context.Context, service *pb.DeleteServiceRequest) (
	*pb.DeleteServiceResponse, error) {
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

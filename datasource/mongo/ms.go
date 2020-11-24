/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
)

func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	return &pb.CreateServiceResponse{}, nil
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	return &pb.GetServicesResponse{}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	return &pb.GetServiceResponse{}, nil
}

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	return &pb.GetServiceDetailResponse{}, nil
}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	return &pb.GetServicesInfoResponse{}, nil
}

func (ds *DataSource) GetApplications(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	return &pb.GetAppsResponse{}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {
	return &pb.GetExistenceByIDResponse{}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	return &pb.GetExistenceResponse{}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	return &pb.UpdateServicePropsResponse{}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	return &pb.DeleteServiceResponse{}, nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool, serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {

	}
}

// Instance management
func (ds *DataSource) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	return &pb.RegisterInstanceResponse{}, nil
}

// GetInstances returns instances under the current domain
func (ds *DataSource) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	return &pb.GetOneInstanceResponse{}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	return &pb.GetInstancesResponse{}, nil
}

// GetProviderInstances returns instances under the specified domain
func (ds *DataSource) GetProviderInstances(ctx context.Context, request *pb.GetProviderInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	return nil, "", nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	return nil, "", nil
}

// FindInstances returns instances under the specified domain
func (ds *DataSource) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	return &pb.FindInstancesResponse{}, nil
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	return &pb.UpdateInstanceStatusResponse{}, nil
}

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	return &pb.UpdateInstancePropsResponse{}, nil
}

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	return &pb.UnregisterInstanceResponse{}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return heartbeat.Instance().Heartbeat(ctx, request)
}

func (ds *DataSource) HeartbeatSet(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	return &pb.HeartbeatSetResponse{}, nil
}

func (ds *DataSource) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	return &pb.BatchFindInstancesResponse{}, nil
}

func (ds *DataSource) GetAllInstances(ctx context.Context, request *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error) {
	return &pb.GetAllInstancesResponse{}, nil
}

// Schema management
func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	return &pb.ModifySchemasResponse{}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	return &pb.ModifySchemaResponse{}, nil
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	return &pb.GetExistenceResponse{}, nil
}

func (ds *DataSource) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	return &pb.GetSchemaResponse{}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	return &pb.GetAllSchemaResponse{}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	return &pb.DeleteSchemaResponse{}, nil
}

// Tag management
func (ds *DataSource) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	return &pb.AddServiceTagsResponse{}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	return &pb.GetServiceTagsResponse{}, nil
}

func (ds *DataSource) UpdateTag(ctx context.Context, request *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	return &pb.UpdateServiceTagResponse{}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	return &pb.DeleteServiceTagsResponse{}, nil
}

// White/black list management
func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	return &pb.AddServiceRulesResponse{}, nil
}

func (ds *DataSource) GetRules(ctx context.Context, request *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	return &pb.GetServiceRulesResponse{}, nil
}

func (ds *DataSource) UpdateRule(ctx context.Context, request *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	return &pb.UpdateServiceRuleResponse{}, nil
}

func (ds *DataSource) DeleteRule(ctx context.Context, request *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	return &pb.DeleteServiceRulesResponse{}, nil
}

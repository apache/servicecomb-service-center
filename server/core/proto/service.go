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
	"context"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"google.golang.org/grpc"
)

type ServiceCtrlServer interface {
	Exist(context.Context, *registry.GetExistenceRequest) (*registry.GetExistenceResponse, error)
	Create(context.Context, *registry.CreateServiceRequest) (*registry.CreateServiceResponse, error)
	Delete(context.Context, *registry.DeleteServiceRequest) (*registry.DeleteServiceResponse, error)
	GetOne(context.Context, *registry.GetServiceRequest) (*registry.GetServiceResponse, error)
	GetServices(context.Context, *registry.GetServicesRequest) (*registry.GetServicesResponse, error)
	UpdateProperties(context.Context, *registry.UpdateServicePropsRequest) (*registry.UpdateServicePropsResponse, error)
	AddRule(context.Context, *registry.AddServiceRulesRequest) (*registry.AddServiceRulesResponse, error)
	GetRule(context.Context, *registry.GetServiceRulesRequest) (*registry.GetServiceRulesResponse, error)
	UpdateRule(context.Context, *registry.UpdateServiceRuleRequest) (*registry.UpdateServiceRuleResponse, error)
	DeleteRule(context.Context, *registry.DeleteServiceRulesRequest) (*registry.DeleteServiceRulesResponse, error)
	AddTags(context.Context, *registry.AddServiceTagsRequest) (*registry.AddServiceTagsResponse, error)
	GetTags(context.Context, *registry.GetServiceTagsRequest) (*registry.GetServiceTagsResponse, error)
	UpdateTag(context.Context, *registry.UpdateServiceTagRequest) (*registry.UpdateServiceTagResponse, error)
	DeleteTags(context.Context, *registry.DeleteServiceTagsRequest) (*registry.DeleteServiceTagsResponse, error)
	GetSchemaInfo(context.Context, *registry.GetSchemaRequest) (*registry.GetSchemaResponse, error)
	GetAllSchemaInfo(context.Context, *registry.GetAllSchemaRequest) (*registry.GetAllSchemaResponse, error)
	DeleteSchema(context.Context, *registry.DeleteSchemaRequest) (*registry.DeleteSchemaResponse, error)
	ModifySchema(context.Context, *registry.ModifySchemaRequest) (*registry.ModifySchemaResponse, error)
	ModifySchemas(context.Context, *registry.ModifySchemasRequest) (*registry.ModifySchemasResponse, error)
	AddDependenciesForMicroServices(context.Context, *registry.AddDependenciesRequest) (*registry.AddDependenciesResponse, error)
	CreateDependenciesForMicroServices(context.Context, *registry.CreateDependenciesRequest) (*registry.CreateDependenciesResponse, error)
	GetProviderDependencies(context.Context, *registry.GetDependenciesRequest) (*registry.GetProDependenciesResponse, error)
	GetConsumerDependencies(context.Context, *registry.GetDependenciesRequest) (*registry.GetConDependenciesResponse, error)
	DeleteServices(context.Context, *registry.DelServicesRequest) (*registry.DelServicesResponse, error)
}
type ServiceInstanceCtrlServer interface {
	Register(context.Context, *registry.RegisterInstanceRequest) (*registry.RegisterInstanceResponse, error)
	Unregister(context.Context, *registry.UnregisterInstanceRequest) (*registry.UnregisterInstanceResponse, error)
	Heartbeat(context.Context, *registry.HeartbeatRequest) (*registry.HeartbeatResponse, error)
	Find(context.Context, *registry.FindInstancesRequest) (*registry.FindInstancesResponse, error)
	GetInstances(context.Context, *registry.GetInstancesRequest) (*registry.GetInstancesResponse, error)
	GetOneInstance(context.Context, *registry.GetOneInstanceRequest) (*registry.GetOneInstanceResponse, error)
	UpdateStatus(context.Context, *registry.UpdateInstanceStatusRequest) (*registry.UpdateInstanceStatusResponse, error)
	UpdateInstanceProperties(context.Context, *registry.UpdateInstancePropsRequest) (*registry.UpdateInstancePropsResponse, error)
	Watch(*registry.WatchInstanceRequest, ServiceInstanceCtrlWatchServer) error
	HeartbeatSet(context.Context, *registry.HeartbeatSetRequest) (*registry.HeartbeatSetResponse, error)
}
type ServiceInstanceCtrlWatchServer interface {
	Send(*registry.WatchInstanceResponse) error
	grpc.ServerStream
}
type GovernServiceCtrlServer interface {
	GetServiceDetail(context.Context, *registry.GetServiceRequest) (*registry.GetServiceDetailResponse, error)
	GetServicesInfo(context.Context, *registry.GetServicesInfoRequest) (*registry.GetServicesInfoResponse, error)
	GetApplications(context.Context, *registry.GetAppsRequest) (*registry.GetAppsResponse, error)
}

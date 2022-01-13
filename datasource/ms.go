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

package datasource

import (
	"context"
	"errors"

	pb "github.com/go-chassis/cari/discovery"
)

const (
	ExistTypeMicroservice             = "microservice"
	ExistTypeSchema                   = "schema"
	DefaultLeaseRenewalInterval int32 = 30
	DefaultLeaseRetryTimes      int32 = 3
)

var (
	ErrServiceNotExists     = errors.New("service does not exist")
	ErrInstanceNotExists    = errors.New("instance does not exist")
	ErrUndefinedSchemaID    = errors.New("non-existent schemaID can't be added request")
	ErrModifySchemaNotAllow = errors.New("schema already exist, can not be changed request")
)

// Attention: request validation must be finished before the following interface being invoked!!!
// MetadataManager contains the CRUD of cache metadata
type MetadataManager interface {
	// Microservice management
	RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error)
	ListService(ctx context.Context, request *pb.GetServicesRequest) (*pb.GetServicesResponse, error)
	GetService(ctx context.Context, request *pb.GetServiceRequest) (*pb.MicroService, error)

	GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (*pb.ServiceDetail, error)
	ListServiceDetail(ctx context.Context, request *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error)
	GetOverview(ctx context.Context, request *pb.GetServicesRequest) (*pb.Statistics, error)
	ListApp(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error)

	ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error)
	ExistService(ctx context.Context, request *pb.GetExistenceRequest) (string, error)
	PutServiceProperties(ctx context.Context, request *pb.UpdateServicePropsRequest) error
	UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) error
	CountService(ctx context.Context, request *pb.GetServiceCountRequest) (*pb.GetServiceCountResponse, error)

	// Instance management
	RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error)
	ExistInstance(ctx context.Context, request *pb.MicroServiceInstanceKey) (*pb.GetExistenceByIDResponse, error)
	// GetInstance returns instances under the specified service
	GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error)
	ListInstance(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error)
	// FindInstances returns instances under the specified domain
	FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error)
	PutInstance(ctx context.Context, request *pb.RegisterInstanceRequest) error
	PutInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) error
	PutInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) error
	UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) error
	SendHeartbeat(ctx context.Context, request *pb.HeartbeatRequest) error
	SendManyHeartbeat(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error)
	// ListManyInstances returns instances under the specified domain
	ListManyInstances(ctx context.Context, request *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error)
	CountInstance(ctx context.Context, request *pb.GetServiceCountRequest) (*pb.GetServiceCountResponse, error)

	// Schema management
	ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error)
	ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error)
	ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error)
	GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error)
	GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error)
	DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) error

	// Tag management
	PutManyTags(ctx context.Context, request *pb.AddServiceTagsRequest) error
	ListTag(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error)
	PutTag(ctx context.Context, request *pb.UpdateServiceTagRequest) error
	DeleteManyTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) error

	// RetireService retire the 'RetirePlan.Reserve' latest versions for each of service,
	// delete other versions which doesn't register any instances.
	RetireService(ctx context.Context, plan *RetirePlan) error
}

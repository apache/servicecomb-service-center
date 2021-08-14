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

package ms

import (
	"context"

	pb "github.com/apache/servicecomb-service-center/pkg/registry"
)

// Attention: request validation must be finished before the following interface being invoked!!!
type DataSource interface {
	RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error)
	GetServices(ctx context.Context, request *pb.GetServicesRequest) (*pb.GetServicesResponse, error)
	GetService(ctx context.Context, request *pb.GetServiceRequest) (*pb.GetServiceResponse, error)
	ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error)
	UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error)
	UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error)
	GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
		serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context)

	RegisterInstance(ctx context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error)
	SearchInstance()
	UpdateInstance()
	UnRegisterInstance()

	ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error)
	ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error)
	ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error)
	GetSchema()
	DeleteSchema()

	AddTag()
	GetTag()
	UpdateTag()
	DeleteTag()

	AddRule()
	GetRule()
	UpdateRule()
	DeleteRule()
}

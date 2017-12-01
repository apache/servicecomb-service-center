//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package service

import (
	"github.com/ServiceComb/service-center/pkg/rpc"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"google.golang.org/grpc"
)

var (
	serviceService  pb.ServiceCtrlServer
	instanceService pb.SerivceInstanceCtrlServerEx
)

func init() {
	instanceService = &InstanceService{}
	serviceService = &MicroServiceService{
		instanceService: instanceService,
	}
	rpc.RegisterService(RegisterGrpcServices)
}

func RegisterGrpcServices(s *grpc.Server) {
	pb.RegisterServiceCtrlServer(s, serviceService)
	pb.RegisterServiceInstanceCtrlServer(s, instanceService)
}

func AssembleResources() (pb.ServiceCtrlServer, pb.SerivceInstanceCtrlServerEx) {
	return serviceService, instanceService
}

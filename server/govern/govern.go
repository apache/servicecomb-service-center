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
package govern

import (
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/rpc"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"google.golang.org/grpc"
)

func init() {
	registerGRPC()

	registerREST()
}

func registerGRPC() {
	rpc.RegisterService(func(s *grpc.Server) {
		pb.RegisterGovernServiceCtrlServer(s, GovernServiceAPI)
	})
}

func registerREST() {
	roa.RegisterServent(&GovernServiceControllerV3{})
	roa.RegisterServent(&GovernServiceControllerV4{})
}

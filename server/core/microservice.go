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
package core

import (
	"github.com/ServiceComb/service-center/pkg/util"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/version"
	"golang.org/x/net/context"
)

var Service *pb.MicroService
var Instance *pb.MicroServiceInstance

const (
	REGISTRY_TENANT  = "default"
	REGISTRY_PROJECT = "default"

	registry_app_id               = "default"
	registry_service_name         = "SERVICECENTER"
	registry_default_instance_env = "production"

	REGISTRY_DEFAULT_LEASE_RENEWALINTERVAL int32 = 30
	REGISTRY_DEFAULT_LEASE_RETRYTIMES      int32 = 3

	IS_SC_SELF = "sc_self"
)

func init() {
	Service = &pb.MicroService{
		AppId:       registry_app_id,
		ServiceName: registry_service_name,
		Version:     version.Ver().Version,
		Status:      pb.MS_UP,
		Level:       "BACK",
		Schemas: []string{
			"servicecenter.grpc.api.ServiceCtrl",
			"servicecenter.grpc.api.ServiceInstanceCtrl",
		},
		Properties: map[string]string{
			pb.PROP_ALLOW_CROSS_APP: "true",
		},
	}

	Instance = &pb.MicroServiceInstance{
		Environment: registry_default_instance_env,
		Status:      pb.MSI_UP,
		HealthCheck: &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: REGISTRY_DEFAULT_LEASE_RENEWALINTERVAL,
			Times:    REGISTRY_DEFAULT_LEASE_RETRYTIMES,
		},
	}
}

func AddDefaultContextValue(ctx context.Context) context.Context {
	ctx = util.NewContext(ctx, "tenant", REGISTRY_TENANT)
	ctx = util.NewContext(ctx, "project", REGISTRY_PROJECT)
	ctx = util.NewContext(ctx, IS_SC_SELF, true)
	return ctx
}

func ISSCSelf(ctx context.Context) bool {
	if ctx.Value(IS_SC_SELF) != nil && ctx.Value(IS_SC_SELF).(bool) {
		return true
	}
	return false
}
func GetExistenceRequest() *pb.GetExistenceRequest {
	return &pb.GetExistenceRequest{
		Type:        pb.EXISTENCE_MS,
		AppId:       registry_app_id,
		ServiceName: registry_service_name,
		Version:     version.Ver().Version,
	}
}

func GetServiceRequest(serviceId string) *pb.GetServiceRequest {
	return &pb.GetServiceRequest{
		ServiceId: serviceId,
	}
}

func CreateServiceRequest() *pb.CreateServiceRequest {
	return &pb.CreateServiceRequest{
		Service: Service,
	}
}

func RegisterInstanceRequest(hostName string, endpoints []string) *pb.RegisterInstanceRequest {
	Instance.HostName = hostName
	Instance.Endpoints = endpoints
	return &pb.RegisterInstanceRequest{
		Instance: Instance,
	}
}

func UnregisterInstanceRequest() *pb.UnregisterInstanceRequest {
	return &pb.UnregisterInstanceRequest{
		ServiceId:  Instance.ServiceId,
		InstanceId: Instance.InstanceId,
	}
}

func HeartbeatRequest() *pb.HeartbeatRequest {
	return &pb.HeartbeatRequest{
		ServiceId:  Instance.ServiceId,
		InstanceId: Instance.InstanceId,
	}
}

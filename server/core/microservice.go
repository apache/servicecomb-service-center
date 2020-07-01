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

package core

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"os"
	"strings"
)

var (
	ServiceAPI         pb.ServiceCtrlServer
	InstanceAPI        pb.ServiceInstanceCtrlServerEx
	Service            *pb.MicroService
	Instance           *pb.MicroServiceInstance
	sharedServiceNames map[string]struct{}
)

const (
	RegistryDomain        = "default"
	RegistryProject       = "default"
	RegistryDomainProject = "default/default"

	RegistryAppID        = "default"
	RegistryServiceName  = "SERVICECENTER"
	RegistryServiceAlias = "SERVICECENTER"

	RegistryDefaultLeaseRenewalinterval int32 = 30
	RegistryDefaultLeaseRetrytimes      int32 = 3

	CtxScSelf     = "_sc_self"
	CtxScRegistry = "_registryOnly"
)

func init() {
	prepareSelfRegistration()

	SetSharedMode()
}

func prepareSelfRegistration() {
	Service = &pb.MicroService{
		Environment: pb.ENV_PROD,
		AppId:       RegistryAppID,
		ServiceName: RegistryServiceName,
		Alias:       RegistryServiceAlias,
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
	if beego.BConfig.RunMode == "dev" {
		Service.Environment = pb.ENV_DEV
	}

	Instance = &pb.MicroServiceInstance{
		Status: pb.MSI_UP,
		HealthCheck: &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: RegistryDefaultLeaseRenewalinterval,
			Times:    RegistryDefaultLeaseRetrytimes,
		},
	}
}

func AddDefaultContextValue(ctx context.Context) context.Context {
	return util.SetContext(util.SetContext(util.SetDomainProject(ctx,
		RegistryDomain, RegistryProject),
		CtxScSelf, true),
		CtxScRegistry, "1")
}

func IsDefaultDomainProject(domainProject string) bool {
	return domainProject == RegistryDomainProject
}

func SetSharedMode() {
	sharedServiceNames = make(map[string]struct{})
	for _, s := range strings.Split(os.Getenv("CSE_SHARED_SERVICES"), ",") {
		if len(s) > 0 {
			sharedServiceNames[s] = struct{}{}
		}
	}
	sharedServiceNames[Service.ServiceName] = struct{}{}
}

func IsShared(key *pb.MicroServiceKey) bool {
	if !IsDefaultDomainProject(key.Tenant) {
		return false
	}
	if key.AppId != RegistryAppID {
		return false
	}
	_, ok := sharedServiceNames[key.ServiceName]
	if !ok {
		_, ok = sharedServiceNames[key.Alias]
	}
	return ok
}

func IsSCInstance(ctx context.Context) bool {
	b, _ := ctx.Value(CtxScSelf).(bool)
	return b
}

func GetExistenceRequest() *pb.GetExistenceRequest {
	return &pb.GetExistenceRequest{
		Type:        pb.EXISTENCE_MS,
		Environment: Service.Environment,
		AppId:       Service.AppId,
		ServiceName: Service.ServiceName,
		Version:     Service.Version,
	}
}

func GetServiceRequest(serviceID string) *pb.GetServiceRequest {
	return &pb.GetServiceRequest{
		ServiceId: serviceID,
	}
}

func CreateServiceRequest() *pb.CreateServiceRequest {
	return &pb.CreateServiceRequest{
		Service: Service,
	}
}

func RegisterInstanceRequest() *pb.RegisterInstanceRequest {
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

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
	"github.com/apache/servicecomb-service-center/pkg/proto"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
	"github.com/go-chassis/cari/discovery"
	"strings"
)

var (
	ServiceAPI         proto.ServiceCtrlServer
	InstanceAPI        proto.ServiceInstanceCtrlServerEx
	Service            *discovery.MicroService
	Instance           *discovery.MicroServiceInstance
	globalServiceNames map[string]struct{}
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
}

func prepareSelfRegistration() {
	Service = &discovery.MicroService{
		Environment: discovery.ENV_PROD,
		AppId:       RegistryAppID,
		ServiceName: RegistryServiceName,
		Alias:       RegistryServiceAlias,
		Version:     version.Ver().Version,
		Status:      discovery.MS_UP,
		Level:       "BACK",
		Schemas: []string{
			"servicecenter.grpc.api.ServiceCtrl",
			"servicecenter.grpc.api.ServiceInstanceCtrl",
		},
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "true",
		},
	}
	if beego.BConfig.RunMode == "dev" {
		Service.Environment = discovery.ENV_DEV
	}

	Instance = &discovery.MicroServiceInstance{
		Status: discovery.MSI_UP,
		HealthCheck: &discovery.HealthCheck{
			Mode:     discovery.CHECK_BY_HEARTBEAT,
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

func RegisterGlobalServices() {
	globalServiceNames = make(map[string]struct{})
	for _, s := range strings.Split(config.GetRegistry().GlobalVisible, ",") {
		if len(s) > 0 {
			globalServiceNames[s] = struct{}{}
		}
	}
	globalServiceNames[Service.ServiceName] = struct{}{}
}

func IsGlobal(key *discovery.MicroServiceKey) bool {
	if !IsDefaultDomainProject(key.Tenant) {
		return false
	}
	if key.AppId != RegistryAppID {
		return false
	}
	_, ok := globalServiceNames[key.ServiceName]
	if !ok {
		_, ok = globalServiceNames[key.Alias]
	}
	return ok
}

func IsSCInstance(ctx context.Context) bool {
	b, _ := ctx.Value(CtxScSelf).(bool)
	return b
}

func GetExistenceRequest() *discovery.GetExistenceRequest {
	return &discovery.GetExistenceRequest{
		Type:        discovery.ExistenceMicroservice,
		Environment: Service.Environment,
		AppId:       Service.AppId,
		ServiceName: Service.ServiceName,
		Version:     Service.Version,
	}
}

func GetServiceRequest(serviceID string) *discovery.GetServiceRequest {
	return &discovery.GetServiceRequest{
		ServiceId: serviceID,
	}
}

func CreateServiceRequest() *discovery.CreateServiceRequest {
	return &discovery.CreateServiceRequest{
		Service: Service,
	}
}

func RegisterInstanceRequest() *discovery.RegisterInstanceRequest {
	return &discovery.RegisterInstanceRequest{
		Instance: Instance,
	}
}

func UnregisterInstanceRequest() *discovery.UnregisterInstanceRequest {
	return &discovery.UnregisterInstanceRequest{
		ServiceId:  Instance.ServiceId,
		InstanceId: Instance.InstanceId,
	}
}

func HeartbeatRequest() *discovery.HeartbeatRequest {
	return &discovery.HeartbeatRequest{
		ServiceId:  Instance.ServiceId,
		InstanceId: Instance.InstanceId,
	}
}

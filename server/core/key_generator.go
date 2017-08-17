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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"strings"
)

const (
	REGISTRY_ROOT_KEY        = "cse-sr"
	REGISTRY_SERVICE_KEY     = "ms"
	REGISTRY_INSTANCE_KEY    = "inst"
	REGISTRY_FILE            = "files"
	REGISTRY_INDEX           = "indexes"
	REGISTRY_RULE_KEY        = "rules"
	REGISTRY_TENANT_KEY      = "tenant"
	REGISTRY_ALIAS_KEY       = "alias"
	REGISTRY_TAG_KEY         = "tags"
	REGISTRY_SCHEMA_KEY      = "schemas"
	REGISTRY_LEASE_KEY       = "leases"
	REGISTRY_DEPENDENCY_KEY  = "deps"
	REGISTRY_ENDPOINTS_INDEX = "epsindex"
)

func GetRootKey() string {
	return strings.Join([]string{
		"",
		REGISTRY_ROOT_KEY,
	}, "/")
}

func GetDomainProjectRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		tenant,
	}, "/")
}

func GetServiceRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_FILE,
		tenant,
	}, "/")
}

func GetServiceIndexRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_INDEX,
		tenant,
	}, "/")
}

func GetServiceAliasRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_ALIAS_KEY,
		tenant,
	}, "/")
}

func GetServiceRuleRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_SERVICE_KEY,
		REGISTRY_RULE_KEY,
	}, "/")
}

func GetServiceRuleIndexRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_RULE_KEY,
		REGISTRY_INDEX,
	}, "/")
}

func GetServiceTagRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_SERVICE_KEY,
		REGISTRY_TAG_KEY,
	}, "/")
}

func GetServiceSchemaRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_SERVICE_KEY,
		REGISTRY_SCHEMA_KEY,
	}, "/")
}

func GetInstanceIndexRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_INDEX,
		tenant,
	}, "/")
}

func GetInstanceRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_FILE,
		tenant,
	}, "/")
}

func GetInstanceLeaseRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_LEASE_KEY,
		tenant,
	}, "/")
}

func GetInstancesEndpointsIndexRootKey(tenant string) string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_ENDPOINTS_INDEX,
		tenant,
	}, "/")
}

func GenerateServiceKey(tenant string, serviceId string) string {
	return strings.Join([]string{
		GetServiceRootKey(tenant),
		serviceId,
	}, "/")
}

func GenerateRuleIndexKey(tenant string, serviceId string, attr string, pattern string) string {
	return strings.Join([]string{
		GetServiceRuleIndexRootKey(tenant),
		serviceId,
		attr,
		pattern,
	}, "/")
}

func GenerateServiceIndexKey(key *pb.MicroServiceKey) string {
	appId := key.AppId
	if len(strings.TrimSpace(appId)) == 0 {
		key.AppId = "default"
	}
	stage := key.Stage
	if len(strings.TrimSpace(stage)) == 0 {
		key.Stage = "dev"
	}
	return strings.Join([]string{
		GetServiceIndexRootKey(key.Tenant),
		key.AppId,
		key.Stage,
		key.ServiceName,
		key.Version,
	}, "/")
}

func GenerateServiceAliasKey(key *pb.MicroServiceKey) string {
	appId := key.AppId
	if len(strings.TrimSpace(appId)) == 0 {
		key.AppId = "default"
	}
	stage := key.Stage
	if len(strings.TrimSpace(stage)) == 0 {
		key.Stage = "dev"
	}
	return strings.Join([]string{
		GetServiceAliasRootKey(key.Tenant),
		key.AppId,
		key.Stage,
		key.Alias,
		key.Version,
	}, "/")
}

func GenerateServiceRuleKey(tenant string, serviceId string, ruleId string) string {
	return strings.Join([]string{
		GetServiceRuleRootKey(tenant),
		serviceId,
		ruleId,
	}, "/")
}

func GenerateServiceTagKey(tenant string, serviceId string) string {
	return strings.Join([]string{
		GetServiceTagRootKey(tenant),
		serviceId,
	}, "/")
}

func GenerateServiceSchemaKey(tenant string, serviceId string, schemaId string) string {
	return strings.Join([]string{
		GetServiceSchemaRootKey(tenant),
		serviceId,
		schemaId,
	}, "/")
}

func GenerateInstanceIndexKey(tenant string, instanceId string) string {
	return strings.Join([]string{
		GetInstanceIndexRootKey(tenant),
		instanceId,
	}, "/")
}

func GenerateInstanceKey(tenant string, serviceId string, instanceId string) string {
	return strings.Join([]string{
		GetInstanceRootKey(tenant),
		serviceId,
		instanceId,
	}, "/")
}

func GenerateInstanceLeaseKey(tenant string, serviceId string, instanceId string) string {
	return strings.Join([]string{
		GetInstanceLeaseRootKey(tenant),
		serviceId,
		instanceId,
	}, "/")
}

func generateServiceDependencyRuleKey(serviceType string, tenant string, in *pb.MicroServiceKey) string {
	if in.ServiceName == "*" {
		return strings.Join([]string{
			GenerateServiceDependencyRuleRootKey(tenant),
			serviceType,
			in.ServiceName,
		}, "/")
	}
	appId := in.AppId
	if len(strings.TrimSpace(appId)) == 0 {
		appId = "default"
	}
	stage := in.Stage
	if len(strings.TrimSpace(stage)) == 0 {
		stage = "dev"
	}
	return strings.Join([]string{
		GenerateServiceDependencyRuleRootKey(tenant),
		serviceType,
		appId,
		stage,
		in.ServiceName,
		in.Version,
	}, "/")
}

func GenerateConsumerDependencyRuleKey(tenant string, in *pb.MicroServiceKey) string {
	return generateServiceDependencyRuleKey("c", tenant, in)
}

func GenerateProviderDependencyRuleKey(tenant string, in *pb.MicroServiceKey) string {
	return generateServiceDependencyRuleKey("p", tenant, in)
}

func GenerateServiceDependencyRuleRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPENDENCY_KEY,
		"rule",
	}, "/")
}

func GenerateConsumerDependencyKey(tenant string, consumerId string, providerId string) string {
	return GenerateServiceDependencyKey("c", tenant, consumerId, providerId)
}

func GenerateServiceDependencyKey(serviceType string, tenant string, serviceId1 string, serviceId2 string) string {
	return strings.Join([]string{
		GetServiceDependencyRootKey(tenant),
		serviceType,
		serviceId1,
		serviceId2,
	}, "/")
}

func GenerateProviderDependencyKey(tenant string, providerId string, consumerId string) string {
	return GenerateServiceDependencyKey("p", tenant, providerId, consumerId)
}

func GetServiceDependencyRootKey(tenant string) string {
	return strings.Join([]string{
		GetDomainProjectRootKey(tenant),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPENDENCY_KEY,
	}, "/")
}

func GetDomainRootKey() string {
	return strings.Join([]string{
		GetRootKey(),
		REGISTRY_TENANT_KEY,
	}, "/")
}

func GenerateDomainKey(tenant string) string {
	return strings.Join([]string{
		GetDomainRootKey(),
		tenant,
	}, "/")
}

func GenerateInsEpsIndex(tenant string, serviceId string, inEpsIndex string) string {
	return strings.Join([]string{
		GetInstancesEndpointsIndexRootKey(tenant),
		serviceId,
		inEpsIndex,
	}, "/")
}

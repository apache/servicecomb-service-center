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
	"strings"
)

const (
	REGISTRY_ROOT_KEY           = "cse-sr"
	REGISTRY_SYS_KEY            = "sys"
	REGISTRY_SERVICE_KEY        = "ms"
	REGISTRY_INSTANCE_KEY       = "inst"
	REGISTRY_FILE               = "files"
	REGISTRY_INDEX              = "indexes"
	REGISTRY_RULE_KEY           = "rules"
	REGISTRY_RULE_INDEX_KEY     = "rule-indexes"
	REGISTRY_DOMAIN_KEY         = "domains"
	REGISTRY_PROJECT_KEY        = "projects"
	REGISTRY_ALIAS_KEY          = "alias"
	REGISTRY_TAG_KEY            = "tags"
	REGISTRY_SCHEMA_KEY         = "schemas"
	REGISTRY_SCHEMA_SUMMARY_KEY = "schema-sum"
	REGISTRY_LEASE_KEY          = "leases"
	REGISTRY_DEPENDENCY_KEY     = "deps"
	REGISTRY_DEPS_RULE_KEY      = "dep-rules"
	REGISTRY_METRICS_KEY        = "metrics"
	ENDPOINTS_ROOT_KEY          = "endpoints"
)

func GetRootKey() string {
	return util.StringJoin([]string{
		"",
		REGISTRY_ROOT_KEY,
	}, "/")
}

func GetServiceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_FILE,
		domainProject,
	}, "/")
}

func GetServiceIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_INDEX,
		domainProject,
	}, "/")
}

func GetServiceAliasRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_ALIAS_KEY,
		domainProject,
	}, "/")
}

func GetServiceRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_RULE_KEY,
		domainProject,
	}, "/")
}

func GetServiceRuleIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_RULE_INDEX_KEY,
		domainProject,
	}, "/")
}

func GetServiceTagRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_TAG_KEY,
		domainProject,
	}, "/")
}

func GetServiceSchemaRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_SCHEMA_KEY,
		domainProject,
	}, "/")
}

func GetInstanceIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_INDEX,
		domainProject,
	}, "/")
}

func GetInstanceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_FILE,
		domainProject,
	}, "/")
}

func GetInstanceLeaseRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_LEASE_KEY,
		domainProject,
	}, "/")
}

func GenerateServiceKey(domainProject string, serviceId string) string {
	return util.StringJoin([]string{
		GetServiceRootKey(domainProject),
		serviceId,
	}, "/")
}

func GenerateRuleIndexKey(domainProject string, serviceId string, attr string, pattern string) string {
	return util.StringJoin([]string{
		GetServiceRuleIndexRootKey(domainProject),
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
	return util.StringJoin([]string{
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
	return util.StringJoin([]string{
		GetServiceAliasRootKey(key.Tenant),
		key.AppId,
		key.Stage,
		key.Alias,
		key.Version,
	}, "/")
}

func GenerateServiceRuleKey(domainProject string, serviceId string, ruleId string) string {
	return util.StringJoin([]string{
		GetServiceRuleRootKey(domainProject),
		serviceId,
		ruleId,
	}, "/")
}

func GenerateServiceTagKey(domainProject string, serviceId string) string {
	return util.StringJoin([]string{
		GetServiceTagRootKey(domainProject),
		serviceId,
	}, "/")
}

func GenerateServiceSchemaKey(domainProject string, serviceId string, schemaId string) string {
	return util.StringJoin([]string{
		GetServiceSchemaRootKey(domainProject),
		serviceId,
		schemaId,
	}, "/")
}

func GenerateServiceSchemaSummaryKey(domainProject string, serviceId string, schemaId string) string {
	return util.StringJoin([]string{
		GetServiceSchemaSummaryRootKey(domainProject),
		serviceId,
		schemaId,
	}, "/")
}

func GetServiceSchemaSummaryRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_SCHEMA_SUMMARY_KEY,
		domainProject,
	}, "/")
}

func GenerateInstanceIndexKey(domainProject string, instanceId string) string {
	return util.StringJoin([]string{
		GetInstanceIndexRootKey(domainProject),
		instanceId,
	}, "/")
}

func GenerateInstanceKey(domainProject string, serviceId string, instanceId string) string {
	return util.StringJoin([]string{
		GetInstanceRootKey(domainProject),
		serviceId,
		instanceId,
	}, "/")
}

func GenerateInstanceLeaseKey(domainProject string, serviceId string, instanceId string) string {
	return util.StringJoin([]string{
		GetInstanceLeaseRootKey(domainProject),
		serviceId,
		instanceId,
	}, "/")
}

func GenerateServiceDependencyRuleKey(serviceType string, domainProject string, in *pb.MicroServiceKey) string {
	if in.ServiceName == "*" {
		return util.StringJoin([]string{
			GetServiceDependencyRuleRootKey(domainProject),
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
	return util.StringJoin([]string{
		GetServiceDependencyRuleRootKey(domainProject),
		serviceType,
		appId,
		stage,
		in.ServiceName,
		in.Version,
	}, "/")
}

func GenerateConsumerDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey("c", domainProject, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey("p", domainProject, in)
}

func GetServiceDependencyRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPS_RULE_KEY,
		domainProject,
	}, "/")
}

func GenerateConsumerDependencyKey(domainProject string, consumerId string, providerId string) string {
	return GenerateServiceDependencyKey("c", domainProject, consumerId, providerId)
}

func GenerateServiceDependencyKey(serviceType string, domainProject string, serviceId1 string, serviceId2 string) string {
	return util.StringJoin([]string{
		GetServiceDependencyRootKey(domainProject),
		serviceType,
		serviceId1,
		serviceId2,
	}, "/")
}

func GenerateProviderDependencyKey(domainProject string, providerId string, consumerId string) string {
	return GenerateServiceDependencyKey("p", domainProject, providerId, consumerId)
}

func GetServiceDependencyRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPENDENCY_KEY,
		domainProject,
	}, "/")
}

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_DOMAIN_KEY,
	}, "/")
}

func GenerateDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, "/")
}

func GetSystemKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SYS_KEY,
	}, "/")
}

func GetMetricsRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_METRICS_KEY,
	}, "/")
}

func GenerateMetricsKey(name, utc, domain string) string {
	return util.StringJoin([]string{
		GetMetricsRootKey(),
		name,
		utc,
		domain,
	}, "/")
}

func GetProjectRootKey(domain string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_PROJECT_KEY,
		domain,
	}, "/")
}

func GenerateProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, "/")
}

func GenerateEndpointsRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		ENDPOINTS_ROOT_KEY,
		domainProject,
	}, "/")
}

func GetEndpointsIndexKey(domainProject string, region string, availableZone string, endpoints string) string {
	return util.StringJoin([]string{
		GenerateEndpointsRootKey(domainProject),
		region,
		availableZone,
		endpoints,
	}, "/")
}

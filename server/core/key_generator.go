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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
)

const (
	SPLIT                       = "/"
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
	REGISTRY_DEPS_QUEUE_KEY     = "dep-queue"
	REGISTRY_METRICS_KEY        = "metrics"
	DEPS_QUEUE_UUID             = "0"
	DEPS_CONSUMER               = "c"
	DEPS_PROVIDER               = "p"
)

func GetRootKey() string {
	return SPLIT + REGISTRY_ROOT_KEY
}

func GetServiceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_FILE,
		domainProject,
	}, SPLIT)
}

func GetServiceIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_INDEX,
		domainProject,
	}, SPLIT)
}

func GetServiceAliasRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_ALIAS_KEY,
		domainProject,
	}, SPLIT)
}

func GetServiceAppKey(domainProject, env, appId string) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(domainProject),
		env,
		appId,
	}, SPLIT)
}

func GetServiceRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_RULE_KEY,
		domainProject,
	}, SPLIT)
}

func GetServiceRuleIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_RULE_INDEX_KEY,
		domainProject,
	}, SPLIT)
}

func GetServiceTagRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_TAG_KEY,
		domainProject,
	}, SPLIT)
}

func GetServiceSchemaRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_SCHEMA_KEY,
		domainProject,
	}, SPLIT)
}

func GetInstanceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_FILE,
		domainProject,
	}, SPLIT)
}

func GetInstanceLeaseRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_INSTANCE_KEY,
		REGISTRY_LEASE_KEY,
		domainProject,
	}, SPLIT)
}

func GenerateServiceKey(domainProject string, serviceId string) string {
	return util.StringJoin([]string{
		GetServiceRootKey(domainProject),
		serviceId,
	}, SPLIT)
}

func GenerateRuleIndexKey(domainProject string, serviceId string, attr string, pattern string) string {
	return util.StringJoin([]string{
		GetServiceRuleIndexRootKey(domainProject),
		serviceId,
		attr,
		pattern,
	}, SPLIT)
}

func GenerateServiceIndexKey(key *pb.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.ServiceName,
		key.Version,
	}, SPLIT)
}

func GenerateServiceAliasKey(key *pb.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceAliasRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.Alias,
		key.Version,
	}, SPLIT)
}

func GenerateServiceRuleKey(domainProject string, serviceId string, ruleId string) string {
	return util.StringJoin([]string{
		GetServiceRuleRootKey(domainProject),
		serviceId,
		ruleId,
	}, SPLIT)
}

func GenerateServiceTagKey(domainProject string, serviceId string) string {
	return util.StringJoin([]string{
		GetServiceTagRootKey(domainProject),
		serviceId,
	}, SPLIT)
}

func GenerateServiceSchemaKey(domainProject string, serviceId string, schemaId string) string {
	return util.StringJoin([]string{
		GetServiceSchemaRootKey(domainProject),
		serviceId,
		schemaId,
	}, SPLIT)
}

func GenerateServiceSchemaSummaryKey(domainProject string, serviceId string, schemaId string) string {
	return util.StringJoin([]string{
		GetServiceSchemaSummaryRootKey(domainProject),
		serviceId,
		schemaId,
	}, SPLIT)
}

func GetServiceSchemaSummaryRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_SCHEMA_SUMMARY_KEY,
		domainProject,
	}, SPLIT)
}

func GenerateInstanceKey(domainProject string, serviceId string, instanceId string) string {
	return util.StringJoin([]string{
		GetInstanceRootKey(domainProject),
		serviceId,
		instanceId,
	}, SPLIT)
}

func GenerateInstanceLeaseKey(domainProject string, serviceId string, instanceId string) string {
	return util.StringJoin([]string{
		GetInstanceLeaseRootKey(domainProject),
		serviceId,
		instanceId,
	}, SPLIT)
}

func GenerateServiceDependencyRuleKey(serviceType string, domainProject string, in *pb.MicroServiceKey) string {
	if in == nil {
		return util.StringJoin([]string{
			GetServiceDependencyRuleRootKey(domainProject),
			serviceType,
		}, SPLIT)
	}
	if in.ServiceName == "*" {
		return util.StringJoin([]string{
			GetServiceDependencyRuleRootKey(domainProject),
			serviceType,
			in.Environment,
			in.ServiceName,
		}, SPLIT)
	}
	return util.StringJoin([]string{
		GetServiceDependencyRuleRootKey(domainProject),
		serviceType,
		in.Environment,
		in.AppId,
		in.ServiceName,
		in.Version,
	}, SPLIT)
}

func GenerateConsumerDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey(DEPS_CONSUMER, domainProject, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey(DEPS_PROVIDER, domainProject, in)
}

func GetServiceDependencyRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPS_RULE_KEY,
		domainProject,
	}, SPLIT)
}

func GetServiceDependencyQueueRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPS_QUEUE_KEY,
		domainProject,
	}, SPLIT)
}

func GenerateConsumerDependencyQueueKey(domainProject, consumerId, uuid string) string {
	return util.StringJoin([]string{
		GetServiceDependencyQueueRootKey(domainProject),
		consumerId,
		uuid,
	}, SPLIT)
}

func GetServiceDependencyRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SERVICE_KEY,
		REGISTRY_DEPENDENCY_KEY,
		domainProject,
	}, SPLIT)
}

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_DOMAIN_KEY,
	}, SPLIT)
}

func GenerateDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, SPLIT)
}

func GetServerInfoKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_SYS_KEY,
	}, SPLIT)
}

func GetMetricsRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_METRICS_KEY,
	}, SPLIT)
}

func GenerateMetricsKey(name, utc, domain string) string {
	return util.StringJoin([]string{
		GetMetricsRootKey(),
		name,
		utc,
		domain,
	}, SPLIT)
}

func GetProjectRootKey(domain string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		REGISTRY_PROJECT_KEY,
		domain,
	}, SPLIT)
}

func GenerateProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, SPLIT)
}

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

package path

import (
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	SPLIT                    = "/"
	RegistryRootKey          = "cse-sr"
	RegistrySysKey           = "sys"
	RegistryServiceKey       = "ms"
	RegistryInstanceKey      = "inst"
	RegistryFile             = "files"
	RegistryIndex            = "indexes"
	RegistryRuleKey          = "rules"
	RegistryRuleIndexKey     = "rule-indexes"
	RegistryDomainKey        = "domains"
	RegistryProjectKey       = "projects"
	RegistryAliasKey         = "alias"
	RegistryTagKey           = "tags"
	RegistrySchemaKey        = "schemas"
	RegistrySchemaSummaryKey = "schema-sum"
	RegistryLeaseKey         = "leases"
	RegistryDependencyKey    = "deps"
	RegistryDepsRuleKey      = "dep-rules"
	RegistryDepsQueueKey     = "dep-queue"
	RegistryMetricsKey       = "metrics"
	DepsQueueUUID            = "0"
	DepsConsumer             = "c"
	DepsProvider             = "p"
)

func GetRootKey() string {
	return SPLIT + RegistryRootKey
}

func GenerateETCDAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}

func GetProjectRootKey(domain string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryProjectKey,
		domain,
	}, SPLIT)
}

func GenerateETCDProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, SPLIT)
}

func GenerateETCDDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, SPLIT)
}

func GetServiceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryFile,
		domainProject,
	}, SPLIT)
}

func GenerateServiceKey(domainProject string, serviceID string) string {
	return util.StringJoin([]string{
		GetServiceRootKey(domainProject),
		serviceID,
	}, SPLIT)
}

func GetServiceIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryIndex,
		domainProject,
	}, SPLIT)
}

func GetServiceAliasRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryAliasKey,
		domainProject,
	}, SPLIT)
}

func GetServiceAppKey(domainProject, env, appID string) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(domainProject),
		env,
		appID,
	}, SPLIT)
}

func GetServiceRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryRuleKey,
		domainProject,
	}, SPLIT)
}

func GetServiceRuleIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryRuleIndexKey,
		domainProject,
	}, SPLIT)
}

func GetServiceTagRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryTagKey,
		domainProject,
	}, SPLIT)
}

func GetServiceSchemaRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistrySchemaKey,
		domainProject,
	}, SPLIT)
}

func GetInstanceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryInstanceKey,
		RegistryFile,
		domainProject,
	}, SPLIT)
}

func GetInstanceLeaseRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryInstanceKey,
		RegistryLeaseKey,
		domainProject,
	}, SPLIT)
}

func GenerateRuleIndexKey(domainProject string, serviceID string, attr string, pattern string) string {
	return util.StringJoin([]string{
		GetServiceRuleIndexRootKey(domainProject),
		serviceID,
		attr,
		pattern,
	}, SPLIT)
}

func GenerateServiceIndexKey(key *registry.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.ServiceName,
		key.Version,
	}, SPLIT)
}

func GenerateServiceAliasKey(key *registry.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceAliasRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.Alias,
		key.Version,
	}, SPLIT)
}

func GenerateServiceRuleKey(domainProject string, serviceID string, ruleID string) string {
	return util.StringJoin([]string{
		GetServiceRuleRootKey(domainProject),
		serviceID,
		ruleID,
	}, SPLIT)
}

func GenerateServiceTagKey(domainProject string, serviceID string) string {
	return util.StringJoin([]string{
		GetServiceTagRootKey(domainProject),
		serviceID,
	}, SPLIT)
}

func GenerateServiceSchemaKey(domainProject string, serviceID string, schemaID string) string {
	return util.StringJoin([]string{
		GetServiceSchemaRootKey(domainProject),
		serviceID,
		schemaID,
	}, SPLIT)
}

func GenerateServiceSchemaSummaryKey(domainProject string, serviceID string, schemaID string) string {
	return util.StringJoin([]string{
		GetServiceSchemaSummaryRootKey(domainProject),
		serviceID,
		schemaID,
	}, SPLIT)
}

func GetServiceSchemaSummaryRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistrySchemaSummaryKey,
		domainProject,
	}, SPLIT)
}

func GenerateInstanceKey(domainProject string, serviceID string, instanceID string) string {
	return util.StringJoin([]string{
		GetInstanceRootKey(domainProject),
		serviceID,
		instanceID,
	}, SPLIT)
}

func GenerateInstanceLeaseKey(domainProject string, serviceID string, instanceID string) string {
	return util.StringJoin([]string{
		GetInstanceLeaseRootKey(domainProject),
		serviceID,
		instanceID,
	}, SPLIT)
}

func GenerateServiceDependencyRuleKey(serviceType string, domainProject string, in *registry.MicroServiceKey) string {
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

func GenerateConsumerDependencyRuleKey(domainProject string, in *registry.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey(DepsConsumer, domainProject, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, in *registry.MicroServiceKey) string {
	return GenerateServiceDependencyRuleKey(DepsProvider, domainProject, in)
}

func GetServiceDependencyRuleRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryDepsRuleKey,
		domainProject,
	}, SPLIT)
}

func GetServiceDependencyQueueRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryDepsQueueKey,
		domainProject,
	}, SPLIT)
}

func GenerateConsumerDependencyQueueKey(domainProject, consumerID, uuid string) string {
	return util.StringJoin([]string{
		GetServiceDependencyQueueRootKey(domainProject),
		consumerID,
		uuid,
	}, SPLIT)
}

func GetServiceDependencyRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryDependencyKey,
		domainProject,
	}, SPLIT)
}

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryDomainKey,
	}, SPLIT)
}

func GenerateDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, SPLIT)
}
func GenerateAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}
func GenerateRBACSecretKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		"rbac/secret",
	}, SPLIT)
}
func GetServerInfoKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistrySysKey,
	}, SPLIT)
}

func GetMetricsRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryMetricsKey,
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

func GenerateProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, SPLIT)
}

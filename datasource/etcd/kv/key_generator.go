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

package kv

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"strings"
)

const (
	RegistryRootKey      = "cse-sr"
	RegistryProjectKey   = "projects"
	SPLIT                = "/"
	RegistryDomainKey    = "domains"
	RegistryServiceKey   = "ms"
	RegistryIndex        = "indexes"
	RegistryDepsQueueKey = "dep-queue"
	RegistryInstanceKey  = "inst"
	RegistryFile         = "files"
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

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryDomainKey,
	}, SPLIT)
}

func GenerateETCDDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
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

func GenerateServiceIndexKey(key *registry.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.ServiceName,
		key.Version,
	}, SPLIT)
}

func ToResponse(key []byte) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(key), SPLIT)
}

func GetInfoFromSvcIndexKV(key []byte) *registry.MicroServiceKey {
	keys := ToResponse(key)
	l := len(keys)
	if l < 6 {
		return nil
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &registry.MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func GetServiceAppKey(domainProject, env, appID string) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(domainProject),
		env,
		appID,
	}, SPLIT)
}

func GenerateConsumerDependencyQueueKey(domainProject, consumerID, uuid string) string {
	return util.StringJoin([]string{
		GetServiceDependencyQueueRootKey(domainProject),
		consumerID,
		uuid,
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

func GetInstanceRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryInstanceKey,
		RegistryFile,
		domainProject,
	}, SPLIT)
}

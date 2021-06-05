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
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

func ToResponse(key []byte) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(key), SPLIT)
}

func GetInfoFromSvcKV(key []byte) (serviceID, domainProject string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceID = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromInstKV(key []byte) (serviceID, instanceID, domainProject string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceID = keys[l-2]
	instanceID = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(key []byte) (domain string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 2 {
		return
	}
	domain = keys[l-1]
	return
}

func GetInfoFromProjectKV(key []byte) (domain, project string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 2 {
		return
	}
	return keys[l-2], keys[l-1]
}

func GetInfoFromRuleKV(key []byte) (serviceID, ruleID, domainProject string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceID = keys[l-2]
	ruleID = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromTagKV(key []byte) (serviceID, domainProject string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 3 {
		return
	}
	serviceID = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromSvcIndexKV(key []byte) *discovery.MicroServiceKey {
	keys := ToResponse(key)
	l := len(keys)
	if l < 6 {
		return nil
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &discovery.MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func GetInfoFromSvcAliasKV(key []byte) *discovery.MicroServiceKey {
	return GetInfoFromSvcIndexKV(key)
}

func GetInfoFromSchemaSummaryKV(key []byte) (domainProject, serviceID, schemaID string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return domainProject, keys[l-2], keys[l-1]
}

func GetInfoFromSchemaKV(key []byte) (domainProject, serviceID, schemaID string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return domainProject, keys[l-2], keys[l-1]
}

func GetInfoFromDependencyQueueKV(key []byte) (consumerID, domainProject, uuid string) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	consumerID = keys[l-2]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	uuid = keys[l-1]
	return
}

func GetInfoFromDependencyRuleKV(key []byte) (t string, _ *discovery.MicroServiceKey) {
	keys := ToResponse(key)
	l := len(keys)
	if l < 5 {
		return "", nil
	}
	if keys[l-1] == "*" {
		return keys[l-3], &discovery.MicroServiceKey{
			Tenant:      fmt.Sprintf("%s/%s", keys[l-5], keys[l-4]),
			Environment: keys[l-2],
			ServiceName: keys[l-1],
		}
	}

	return keys[l-5], &discovery.MicroServiceKey{
		Tenant:      fmt.Sprintf("%s/%s", keys[l-7], keys[l-6]),
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

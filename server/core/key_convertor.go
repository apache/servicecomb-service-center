// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"strings"
)

func KvToResponse(key []byte) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(key), SPLIT)
}

func GetInfoFromSvcKV(key []byte) (serviceId, domainProject string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromInstKV(key []byte) (serviceId, instanceId, domainProject string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	instanceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(key []byte) (domain string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 2 {
		return
	}
	domain = keys[l-1]
	return
}

func GetInfoFromProjectKV(key []byte) (domainProject string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 2 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-2], keys[l-1])
	return
}

func GetInfoFromRuleKV(key []byte) (serviceId, ruleId, domainProject string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	ruleId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromTagKV(key []byte) (serviceId, domainProject string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 3 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromSvcIndexKV(key []byte) *pb.MicroServiceKey {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 6 {
		return nil
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func GetInfoFromSvcAliasKV(key []byte) *pb.MicroServiceKey {
	return GetInfoFromSvcIndexKV(key)
}

func GetInfoFromSchemaSummaryKV(key []byte) (domainProject, serviceId, schemaId string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return domainProject, keys[l-2], keys[l-1]
}

func GetInfoFromSchemaKV(key []byte) (domainProject, serviceId, schemaId string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return domainProject, keys[l-2], keys[l-1]
}

func GetInfoFromDependencyQueueKV(key []byte) (consumerId, domainProject, uuid string) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 4 {
		return
	}
	consumerId = keys[l-2]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	uuid = keys[l-1]
	return
}

func GetInfoFromDependencyRuleKV(key []byte) (t string, _ *pb.MicroServiceKey) {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 5 {
		return "", nil
	}
	if keys[l-1] == "*" {
		return keys[l-3], &pb.MicroServiceKey{
			Tenant:      fmt.Sprintf("%s/%s", keys[l-5], keys[l-4]),
			Environment: keys[l-2],
			ServiceName: keys[l-1],
		}
	}

	return keys[l-5], &pb.MicroServiceKey{
		Tenant:      fmt.Sprintf("%s/%s", keys[l-7], keys[l-6]),
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func FromDomainProject(domainProject string) (domain, project string) {
	if i := strings.Index(domainProject, "/"); i >= 0 {
		return domainProject[:i], domainProject[i+1:]
	}
	return domainProject, ""
}

func ToDomainProject(domain, project string) (domainProject string) {
	domainProject = domain + "/" + project
	return
}

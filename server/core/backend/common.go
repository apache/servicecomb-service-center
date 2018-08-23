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

package backend

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"strings"
	"time"
)

const (
	leaseProfTimeFmt           = "15:04:05.000"
	eventBlockSize             = 1000
	deferCheckWindow           = 2 * time.Second // instance DELETE event will be delay.
	selfPreservationPercentage = 0.8
)

var (
	DOMAIN           discovery.Type
	PROJECT          discovery.Type
	SERVICE          discovery.Type
	SERVICE_INDEX    discovery.Type
	SERVICE_ALIAS    discovery.Type
	SERVICE_TAG      discovery.Type
	RULE             discovery.Type
	RULE_INDEX       discovery.Type
	DEPENDENCY       discovery.Type
	DEPENDENCY_RULE  discovery.Type
	DEPENDENCY_QUEUE discovery.Type
	SCHEMA           discovery.Type
	SCHEMA_SUMMARY   discovery.Type
	INSTANCE         discovery.Type
	LEASE            discovery.Type
)

var (
	newService         discovery.CreateValueFunc = func() interface{} { return new(pb.MicroService) }
	newInstance        discovery.CreateValueFunc = func() interface{} { return new(pb.MicroServiceInstance) }
	newRule            discovery.CreateValueFunc = func() interface{} { return new(pb.ServiceRule) }
	newDependencyRule  discovery.CreateValueFunc = func() interface{} { return new(pb.MicroServiceDependency) }
	newDependencyQueue discovery.CreateValueFunc = func() interface{} { return new(pb.ConsumerDependency) }

	ServiceParser         = &discovery.CommonParser{newService, discovery.JsonUnmarshal}
	InstanceParser        = &discovery.CommonParser{newInstance, discovery.JsonUnmarshal}
	RuleParser            = &discovery.CommonParser{newRule, discovery.JsonUnmarshal}
	DependencyRuleParser  = &discovery.CommonParser{newDependencyRule, discovery.JsonUnmarshal}
	DependencyQueueParser = &discovery.CommonParser{newDependencyQueue, discovery.JsonUnmarshal}
)

func KvToResponse(kv *discovery.KeyValue) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
}

func GetInfoFromSvcKV(kv *discovery.KeyValue) (serviceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromInstKV(kv *discovery.KeyValue) (serviceId, instanceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	instanceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(kv *discovery.KeyValue) (domain string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}
	domain = keys[l-1]
	return
}

func GetInfoFromProjectKV(kv *discovery.KeyValue) (domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-2], keys[l-1])
	return
}

func GetInfoFromRuleKV(kv *discovery.KeyValue) (serviceId, ruleId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	ruleId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromTagKV(kv *discovery.KeyValue) (serviceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 3 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromSvcIndexKV(kv *discovery.KeyValue) (key *pb.MicroServiceKey) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 6 {
		return
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

func GetInfoFromSchemaSummaryKV(kv *discovery.KeyValue) (schemaId string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}

	return keys[l-1]
}

func GetInfoFromSchemaKV(kv *discovery.KeyValue) (schemaId string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}

	return keys[l-1]
}

func GetInfoFromDependencyQueueKV(kv *discovery.KeyValue) (consumerId, domainProject, uuid string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	consumerId = keys[l-2]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	uuid = keys[l-1]
	return
}

func GetInfoFromDependencyRuleKV(kv *discovery.KeyValue) (key *pb.MicroServiceKey) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 5 {
		return
	}
	if keys[l-1] == "*" {
		return &pb.MicroServiceKey{
			Tenant:      fmt.Sprintf("%s/%s", keys[l-5], keys[l-4]),
			Environment: keys[l-2],
			ServiceName: keys[l-1],
		}
	}

	return &pb.MicroServiceKey{
		Tenant:      fmt.Sprintf("%s/%s", keys[l-7], keys[l-6]),
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func registerInnerTypes() {
	SERVICE = Store().MustInstall(discovery.NewAddOn("SERVICE",
		discovery.Configure().WithPrefix(core.GetServiceRootKey("")).
			WithInitSize(500).WithParser(ServiceParser)))
	INSTANCE = Store().MustInstall(discovery.NewAddOn("INSTANCE",
		discovery.Configure().WithPrefix(core.GetInstanceRootKey("")).
			WithInitSize(1000).WithParser(InstanceParser).
			WithDeferHandler(NewInstanceEventDeferHandler())))
	DOMAIN = Store().MustInstall(discovery.NewAddOn("DOMAIN",
		discovery.Configure().WithPrefix(core.GetDomainRootKey()+core.SPLIT).
			WithInitSize(100).WithParser(discovery.StringParser)))
	SCHEMA = Store().MustInstall(discovery.NewAddOn("SCHEMA",
		discovery.Configure().WithPrefix(core.GetServiceSchemaRootKey("")).
			WithInitSize(0)))
	SCHEMA_SUMMARY = Store().MustInstall(discovery.NewAddOn("SCHEMA_SUMMARY",
		discovery.Configure().WithPrefix(core.GetServiceSchemaSummaryRootKey("")).
			WithInitSize(100).WithParser(discovery.StringParser)))
	RULE = Store().MustInstall(discovery.NewAddOn("RULE",
		discovery.Configure().WithPrefix(core.GetServiceRuleRootKey("")).
			WithInitSize(100).WithParser(RuleParser)))
	LEASE = Store().MustInstall(discovery.NewAddOn("LEASE",
		discovery.Configure().WithPrefix(core.GetInstanceLeaseRootKey("")).
			WithInitSize(1000).WithParser(discovery.StringParser)))
	SERVICE_INDEX = Store().MustInstall(discovery.NewAddOn("SERVICE_INDEX",
		discovery.Configure().WithPrefix(core.GetServiceIndexRootKey("")).
			WithInitSize(500).WithParser(discovery.StringParser)))
	SERVICE_ALIAS = Store().MustInstall(discovery.NewAddOn("SERVICE_ALIAS",
		discovery.Configure().WithPrefix(core.GetServiceAliasRootKey("")).
			WithInitSize(100).WithParser(discovery.StringParser)))
	SERVICE_TAG = Store().MustInstall(discovery.NewAddOn("SERVICE_TAG",
		discovery.Configure().WithPrefix(core.GetServiceTagRootKey("")).
			WithInitSize(100).WithParser(discovery.MapParser)))
	RULE_INDEX = Store().MustInstall(discovery.NewAddOn("RULE_INDEX",
		discovery.Configure().WithPrefix(core.GetServiceRuleIndexRootKey("")).
			WithInitSize(100).WithParser(discovery.StringParser)))
	DEPENDENCY = Store().MustInstall(discovery.NewAddOn("DEPENDENCY",
		discovery.Configure().WithPrefix(core.GetServiceDependencyRootKey("")).
			WithInitSize(100)))
	DEPENDENCY_RULE = Store().MustInstall(discovery.NewAddOn("DEPENDENCY_RULE",
		discovery.Configure().WithPrefix(core.GetServiceDependencyRuleRootKey("")).
			WithInitSize(100).WithParser(DependencyRuleParser)))
	DEPENDENCY_QUEUE = Store().MustInstall(discovery.NewAddOn("DEPENDENCY_QUEUE",
		discovery.Configure().WithPrefix(core.GetServiceDependencyQueueRootKey("")).
			WithInitSize(100).WithParser(DependencyQueueParser)))
	PROJECT = Store().MustInstall(discovery.NewAddOn("PROJECT",
		discovery.Configure().WithPrefix(core.GetProjectRootKey("")).
			WithInitSize(100).WithParser(discovery.StringParser)))
}

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
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"time"
)

const (
	leaseProfTimeFmt           = "15:04:05.000"
	eventBlockSize             = 1000
	deferCheckWindow           = 2 * time.Second // instance DELETE event will be delay.
	selfPreservationPercentage = 0.8
	selfPreservationMaxTTL     = 10 * 60 // 10min
	selfPreservationInitCount  = 5
)

var (
	DOMAIN          cache.Type
	PROJECT         cache.Type
	SERVICE         cache.Type
	ServiceIndex    cache.Type
	ServiceAlias    cache.Type
	ServiceTag      cache.Type
	RULE            cache.Type
	RuleIndex       cache.Type
	DependencyRule  cache.Type
	DependencyQueue cache.Type
	SCHEMA          cache.Type
	SchemaSummary   cache.Type
	INSTANCE        cache.Type
	LEASE           cache.Type
)

func registerInnerTypes() {
	SERVICE = Store().MustInstall(NewAddOn("SERVICE",
		cache.Configure().WithPrefix(core.GetServiceRootKey("")).
			WithInitSize(500).WithParser(proto.ServiceParser)))
	INSTANCE = Store().MustInstall(NewAddOn("INSTANCE",
		cache.Configure().WithPrefix(core.GetInstanceRootKey("")).
			WithInitSize(1000).WithParser(proto.InstanceParser).
			WithDeferHandler(NewInstanceEventDeferHandler())))
	DOMAIN = Store().MustInstall(NewAddOn("DOMAIN",
		cache.Configure().WithPrefix(core.GetDomainRootKey()+core.SPLIT).
			WithInitSize(100).WithParser(proto.StringParser)))
	SCHEMA = Store().MustInstall(NewAddOn("SCHEMA",
		cache.Configure().WithPrefix(core.GetServiceSchemaRootKey("")).
			WithInitSize(0)))
	SchemaSummary = Store().MustInstall(NewAddOn("SCHEMA_SUMMARY",
		cache.Configure().WithPrefix(core.GetServiceSchemaSummaryRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	RULE = Store().MustInstall(NewAddOn("RULE",
		cache.Configure().WithPrefix(core.GetServiceRuleRootKey("")).
			WithInitSize(100).WithParser(proto.RuleParser)))
	LEASE = Store().MustInstall(NewAddOn("LEASE",
		cache.Configure().WithPrefix(core.GetInstanceLeaseRootKey("")).
			WithInitSize(1000).WithParser(proto.StringParser)))
	ServiceIndex = Store().MustInstall(NewAddOn("SERVICE_INDEX",
		cache.Configure().WithPrefix(core.GetServiceIndexRootKey("")).
			WithInitSize(500).WithParser(proto.StringParser)))
	ServiceAlias = Store().MustInstall(NewAddOn("SERVICE_ALIAS",
		cache.Configure().WithPrefix(core.GetServiceAliasRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	ServiceTag = Store().MustInstall(NewAddOn("SERVICE_TAG",
		cache.Configure().WithPrefix(core.GetServiceTagRootKey("")).
			WithInitSize(100).WithParser(proto.MapParser)))
	RuleIndex = Store().MustInstall(NewAddOn("RULE_INDEX",
		cache.Configure().WithPrefix(core.GetServiceRuleIndexRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	DependencyRule = Store().MustInstall(NewAddOn("DEPENDENCY_RULE",
		cache.Configure().WithPrefix(core.GetServiceDependencyRuleRootKey("")).
			WithInitSize(100).WithParser(proto.DependencyRuleParser)))
	DependencyQueue = Store().MustInstall(NewAddOn("DEPENDENCY_QUEUE",
		cache.Configure().WithPrefix(core.GetServiceDependencyQueueRootKey("")).
			WithInitSize(100).WithParser(proto.DependencyQueueParser)))
	PROJECT = Store().MustInstall(NewAddOn("PROJECT",
		cache.Configure().WithPrefix(core.GetProjectRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
}

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
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"time"
)

const (
	eventBlockSize             = 1000
	deferCheckWindow           = 2 * time.Second // instance DELETE event will be delay.
	selfPreservationPercentage = 0.8
	selfPreservationMaxTTL     = 10 * 60 // 10min
	selfPreservationInitCount  = 5
)

var (
	DOMAIN          sd.Type
	PROJECT         sd.Type
	SERVICE         sd.Type
	ServiceIndex    sd.Type
	ServiceAlias    sd.Type
	ServiceTag      sd.Type
	RULE            sd.Type
	RuleIndex       sd.Type
	DependencyRule  sd.Type
	DependencyQueue sd.Type
	SCHEMA          sd.Type
	SchemaSummary   sd.Type
	INSTANCE        sd.Type
	LEASE           sd.Type
)

func registerInnerTypes() {
	SERVICE = Store().MustInstall(NewAddOn("SERVICE",
		sd.Configure().WithPrefix(core.GetServiceRootKey("")).
			WithInitSize(500).WithParser(proto.ServiceParser)))
	INSTANCE = Store().MustInstall(NewAddOn("INSTANCE",
		sd.Configure().WithPrefix(core.GetInstanceRootKey("")).
			WithInitSize(1000).WithParser(proto.InstanceParser).
			WithDeferHandler(NewInstanceEventDeferHandler())))
	DOMAIN = Store().MustInstall(NewAddOn("DOMAIN",
		sd.Configure().WithPrefix(core.GetDomainRootKey()+core.SPLIT).
			WithInitSize(100).WithParser(proto.StringParser)))
	SCHEMA = Store().MustInstall(NewAddOn("SCHEMA",
		sd.Configure().WithPrefix(core.GetServiceSchemaRootKey("")).
			WithInitSize(0)))
	SchemaSummary = Store().MustInstall(NewAddOn("SCHEMA_SUMMARY",
		sd.Configure().WithPrefix(core.GetServiceSchemaSummaryRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	RULE = Store().MustInstall(NewAddOn("RULE",
		sd.Configure().WithPrefix(core.GetServiceRuleRootKey("")).
			WithInitSize(100).WithParser(proto.RuleParser)))
	LEASE = Store().MustInstall(NewAddOn("LEASE",
		sd.Configure().WithPrefix(core.GetInstanceLeaseRootKey("")).
			WithInitSize(1000).WithParser(proto.StringParser)))
	ServiceIndex = Store().MustInstall(NewAddOn("SERVICE_INDEX",
		sd.Configure().WithPrefix(core.GetServiceIndexRootKey("")).
			WithInitSize(500).WithParser(proto.StringParser)))
	ServiceAlias = Store().MustInstall(NewAddOn("SERVICE_ALIAS",
		sd.Configure().WithPrefix(core.GetServiceAliasRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	ServiceTag = Store().MustInstall(NewAddOn("SERVICE_TAG",
		sd.Configure().WithPrefix(core.GetServiceTagRootKey("")).
			WithInitSize(100).WithParser(proto.MapParser)))
	RuleIndex = Store().MustInstall(NewAddOn("RULE_INDEX",
		sd.Configure().WithPrefix(core.GetServiceRuleIndexRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
	DependencyRule = Store().MustInstall(NewAddOn("DEPENDENCY_RULE",
		sd.Configure().WithPrefix(core.GetServiceDependencyRuleRootKey("")).
			WithInitSize(100).WithParser(proto.DependencyRuleParser)))
	DependencyQueue = Store().MustInstall(NewAddOn("DEPENDENCY_QUEUE",
		sd.Configure().WithPrefix(core.GetServiceDependencyQueueRootKey("")).
			WithInitSize(100).WithParser(proto.DependencyQueueParser)))
	PROJECT = Store().MustInstall(NewAddOn("PROJECT",
		sd.Configure().WithPrefix(core.GetProjectRootKey("")).
			WithInitSize(100).WithParser(proto.StringParser)))
}

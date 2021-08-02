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
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/value"
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
	DependencyRule  sd.Type
	DependencyQueue sd.Type
	SCHEMA          sd.Type
	SchemaSummary   sd.Type
	INSTANCE        sd.Type
	LEASE           sd.Type
)

func registerInnerTypes() {
	SERVICE = Store().MustInstall(NewAddOn("SERVICE",
		sd.Configure().WithPrefix(path.GetServiceRootKey("")).
			WithInitSize(500).WithParser(value.ServiceParser)))
	INSTANCE = Store().MustInstall(NewAddOn("INSTANCE",
		sd.Configure().WithPrefix(path.GetInstanceRootKey("")).
			WithInitSize(1000).WithParser(value.InstanceParser).
			WithDeferHandler(NewInstanceEventDeferHandler())))
	DOMAIN = Store().MustInstall(NewAddOn("DOMAIN",
		sd.Configure().WithPrefix(path.GenerateDomainKey("")).
			WithInitSize(100).WithParser(value.StringParser)))
	SCHEMA = Store().MustInstall(NewAddOn("SCHEMA",
		sd.Configure().WithPrefix(path.GetServiceSchemaRootKey("")).
			WithInitSize(0)))
	SchemaSummary = Store().MustInstall(NewAddOn("SCHEMA_SUMMARY",
		sd.Configure().WithPrefix(path.GetServiceSchemaSummaryRootKey("")).
			WithInitSize(100).WithParser(value.StringParser)))
	LEASE = Store().MustInstall(NewAddOn("LEASE",
		sd.Configure().WithPrefix(path.GetInstanceLeaseRootKey("")).
			WithInitSize(1000).WithParser(value.StringParser)))
	ServiceIndex = Store().MustInstall(NewAddOn("SERVICE_INDEX",
		sd.Configure().WithPrefix(path.GetServiceIndexRootKey("")).
			WithInitSize(500).WithParser(value.StringParser)))
	ServiceAlias = Store().MustInstall(NewAddOn("SERVICE_ALIAS",
		sd.Configure().WithPrefix(path.GetServiceAliasRootKey("")).
			WithInitSize(100).WithParser(value.StringParser)))
	ServiceTag = Store().MustInstall(NewAddOn("SERVICE_TAG",
		sd.Configure().WithPrefix(path.GetServiceTagRootKey("")).
			WithInitSize(100).WithParser(value.MapParser)))
	DependencyRule = Store().MustInstall(NewAddOn("DEPENDENCY_RULE",
		sd.Configure().WithPrefix(path.GetServiceDependencyRuleRootKey("")).
			WithInitSize(100).WithParser(value.DependencyRuleParser)))
	DependencyQueue = Store().MustInstall(NewAddOn("DEPENDENCY_QUEUE",
		sd.Configure().WithPrefix(path.GetServiceDependencyQueueRootKey("")).
			WithInitSize(100).WithParser(value.DependencyQueueParser)))
	PROJECT = Store().MustInstall(NewAddOn("PROJECT",
		sd.Configure().WithPrefix(path.GetProjectRootKey("")).
			WithInitSize(100).WithParser(value.StringParser)))
}

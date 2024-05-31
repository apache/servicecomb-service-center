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

package sd

import (
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
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
	TypeDomain          kvstore.Type
	TypeProject         kvstore.Type
	TypeService         kvstore.Type
	TypeServiceIndex    kvstore.Type
	TypeServiceAlias    kvstore.Type
	TypeServiceTag      kvstore.Type
	TypeDependencyRule  kvstore.Type
	TypeDependencyQueue kvstore.Type
	TypeInstance        kvstore.Type
	TypeLease           kvstore.Type
	TypeEnvironment     kvstore.Type
)

func RegisterInnerTypes() {
	TypeService = state.MustRegister("SERVICE", path.GetServiceRootKey(""),
		state.WithInitSize(500),
		state.WithParser(value.ServiceParser))
	TypeInstance = state.MustRegister("INSTANCE", path.GetInstanceRootKey(""),
		state.WithInitSize(1000),
		state.WithParser(value.InstanceParser),
		state.WithDeferHandler(NewInstanceEventDeferHandler()))
	TypeDomain = state.MustRegister("DOMAIN", path.GenerateDomainKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.StringParser))
	TypeLease = state.MustRegister("LEASE", path.GetInstanceLeaseRootKey(""),
		state.WithInitSize(1000),
		state.WithParser(parser.StringParser))
	TypeServiceIndex = state.MustRegister("SERVICE_INDEX", path.GetServiceIndexRootKey(""),
		state.WithInitSize(500),
		state.WithParser(parser.StringParser))
	TypeServiceAlias = state.MustRegister("SERVICE_ALIAS", path.GetServiceAliasRootKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.StringParser))
	TypeServiceTag = state.MustRegister("SERVICE_TAG", path.GetServiceTagRootKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.MapParser))
	TypeDependencyRule = state.MustRegister("DEPENDENCY_RULE", path.GetServiceDependencyRuleRootKey(""),
		state.WithInitSize(100),
		state.WithParser(value.DependencyRuleParser))
	TypeDependencyQueue = state.MustRegister("DEPENDENCY_QUEUE", path.GetServiceDependencyQueueRootKey(""),
		state.WithInitSize(100),
		state.WithParser(value.DependencyQueueParser))
	TypeProject = state.MustRegister("PROJECT", path.GetProjectRootKey(""),
		state.WithInitSize(100),
		state.WithParser(parser.StringParser))
	TypeEnvironment = state.MustRegister("ENVIRONMENT", path.GetEnvironmentRootKey(""),
		state.WithInitSize(100),
		state.WithParser(value.EnvironmentParser))
}

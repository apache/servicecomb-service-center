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
package backend

import (
	"errors"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"strconv"
	"time"
)

const (
	// re-list etcd when there is no event coming in more than 1h(=120*30s)
	DEFAULT_MAX_NO_EVENT_INTERVAL = 120
	DEFAULT_LISTWATCH_TIMEOUT     = 30 * time.Second

	DEFAULT_SELF_PRESERVATION_PERCENT = 0.8
	DEFAULT_CACHE_INIT_SIZE           = 100
)

const (
	DEFAULT_METRICS_INTERVAL = 30 * time.Second
	DEFAULT_COMPACT_TIMES    = 2
	DEFAULT_COMPACT_TIMEOUT  = 5 * time.Minute

	minWaitInterval = 1 * time.Second
	eventBlockSize  = 1000
)

const DEFAULT_CHECK_WINDOW = 2 * time.Second // instance DELETE event will be delay.

const TIME_FORMAT = "15:04:05.000"

const EVENT_BUS_MAX_SIZE = 1000

// errors
var (
	ErrNoImpl = errors.New("no implement")
)

var closedCh = make(chan struct{})

func init() {
	close(closedCh)
}

type StoreType int

func (st StoreType) String() string {
	if int(st) < 0 {
		return "TypeError"
	}
	if int(st) < len(TypeNames) {
		return TypeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

const TypeError = StoreType(-1)

const (
	DOMAIN StoreType = iota
	PROJECT
	SERVICE
	SERVICE_INDEX
	SERVICE_ALIAS
	SERVICE_TAG
	RULE
	RULE_INDEX
	DEPENDENCY
	DEPENDENCY_RULE
	DEPENDENCY_QUEUE
	SCHEMA // big data should not be stored in memory.
	SCHEMA_SUMMARY
	INSTANCE
	LEASE
	typeEnd // end of the base store types
)

var TypeNames = []string{
	DOMAIN:           "DOMAIN",
	PROJECT:          "PROJECT",
	SERVICE:          "SERVICE",
	SERVICE_INDEX:    "SERVICE_INDEX",
	SERVICE_ALIAS:    "SERVICE_ALIAS",
	SERVICE_TAG:      "SERVICE_TAG",
	RULE:             "RULE",
	RULE_INDEX:       "RULE_INDEX",
	DEPENDENCY:       "DEPENDENCY",
	DEPENDENCY_RULE:  "DEPENDENCY_RULE",
	DEPENDENCY_QUEUE: "DEPENDENCY_QUEUE",
	SCHEMA:           "SCHEMA",
	SCHEMA_SUMMARY:   "SCHEMA_SUMMARY",
	INSTANCE:         "INSTANCE",
	LEASE:            "LEASE",
	typeEnd:          "TYPEEND",
}

var TypeConfig = map[StoreType]*Config{
	SERVICE: Configure().WithPrefix(apt.GetServiceRootKey("")).
		WithInitSize(500).WithParser(ServiceParser),

	INSTANCE: Configure().WithPrefix(apt.GetInstanceRootKey("")).
		WithInitSize(1000).WithParser(InstanceParser).
		WithDeferHandler(NewInstanceEventDeferHandler()),

	DOMAIN: Configure().WithPrefix(apt.GetDomainRootKey() + "/").
		WithInitSize(100).WithParser(StringParser),

	SCHEMA: Configure().WithPrefix(apt.GetServiceSchemaRootKey("")).
		WithInitSize(0),

	SCHEMA_SUMMARY: Configure().WithPrefix(apt.GetServiceSchemaSummaryRootKey("")).
		WithInitSize(100).WithParser(StringParser),

	RULE: Configure().WithPrefix(apt.GetServiceRuleRootKey("")).
		WithInitSize(100).WithParser(RuleParser),

	LEASE: Configure().WithPrefix(apt.GetInstanceLeaseRootKey("")).
		WithInitSize(1000).WithParser(StringParser),

	SERVICE_INDEX: Configure().WithPrefix(apt.GetServiceIndexRootKey("")).
		WithInitSize(500).WithParser(StringParser),

	SERVICE_ALIAS: Configure().WithPrefix(apt.GetServiceAliasRootKey("")).
		WithInitSize(100).WithParser(StringParser),

	SERVICE_TAG: Configure().WithPrefix(apt.GetServiceTagRootKey("")).
		WithInitSize(100).WithParser(MapParser),

	RULE_INDEX: Configure().WithPrefix(apt.GetServiceRuleIndexRootKey("")).
		WithInitSize(100).WithParser(StringParser),

	DEPENDENCY: Configure().WithPrefix(apt.GetServiceDependencyRootKey("")).
		WithInitSize(100),

	DEPENDENCY_RULE: Configure().WithPrefix(apt.GetServiceDependencyRuleRootKey("")).
		WithInitSize(100).WithParser(DependencyRuleParser),

	DEPENDENCY_QUEUE: Configure().WithPrefix(apt.GetServiceDependencyQueueRootKey("")).
		WithInitSize(100).WithParser(DependencyQueueParser),

	PROJECT: Configure().WithPrefix(apt.GetProjectRootKey("")).
		WithInitSize(100).WithParser(StringParser),
}

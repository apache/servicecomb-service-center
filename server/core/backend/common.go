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
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"strconv"
	"time"
)

type StoreType int

func (st StoreType) String() string {
	if int(st) < 0 {
		return "NONEXIST"
	}
	if int(st) < len(TypeNames) {
		return TypeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

const NONEXIST = StoreType(-1)

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
	ENDPOINTS
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
	ENDPOINTS:        "ENDPOINTS",
	typeEnd:          "TYPEEND",
}

var TypeRoots = map[StoreType]string{
	SERVICE:          apt.GetServiceRootKey(""),
	INSTANCE:         apt.GetInstanceRootKey(""),
	DOMAIN:           apt.GetDomainRootKey() + "/",
	SCHEMA:           apt.GetServiceSchemaRootKey(""),
	SCHEMA_SUMMARY:   apt.GetServiceSchemaSummaryRootKey(""),
	RULE:             apt.GetServiceRuleRootKey(""),
	LEASE:            apt.GetInstanceLeaseRootKey(""),
	SERVICE_INDEX:    apt.GetServiceIndexRootKey(""),
	SERVICE_ALIAS:    apt.GetServiceAliasRootKey(""),
	SERVICE_TAG:      apt.GetServiceTagRootKey(""),
	RULE_INDEX:       apt.GetServiceRuleIndexRootKey(""),
	DEPENDENCY:       apt.GetServiceDependencyRootKey(""),
	DEPENDENCY_RULE:  apt.GetServiceDependencyRuleRootKey(""),
	DEPENDENCY_QUEUE: apt.GetServiceDependencyQueueRootKey(""),
	PROJECT:          apt.GetProjectRootKey(""),
	ENDPOINTS:        apt.GetEndpointsRootKey(""),
}

var TypeInitSize = map[StoreType]int{
	SERVICE:          500,
	INSTANCE:         1000,
	DOMAIN:           100,
	SCHEMA:           0,
	SCHEMA_SUMMARY:   100,
	RULE:             100,
	LEASE:            1000,
	SERVICE_INDEX:    500,
	SERVICE_ALIAS:    100,
	SERVICE_TAG:      100,
	RULE_INDEX:       100,
	DEPENDENCY:       100,
	DEPENDENCY_RULE:  100,
	DEPENDENCY_QUEUE: 100,
	PROJECT:          100,
	ENDPOINTS:        1000,
}

const (
	// re-list etcd when there is no event coming in more than 15m(=30*30s)
	DEFAULT_MAX_NO_EVENT_INTERVAL = 30
	DEFAULT_LISTWATCH_TIMEOUT     = 30 * time.Second

	DEFAULT_SELF_PRESERVATION_PERCENT = 0.8
	DEFAULT_CACHE_INIT_SIZE           = 100
)

const (
	DEFAULT_METRICS_INTERVAL = 30 * time.Second

	DEFAULT_COMPACT_TIMES   = 2
	DEFAULT_COMPACT_TIMEOUT = 5 * time.Minute
	minWaitInterval         = 1 * time.Second
	eventBlockSize          = 1000
)

const (
	DEFAULT_MAX_EVENT_COUNT   = 1000
	DEFAULT_ADD_QUEUE_TIMEOUT = 5 * time.Second
)

const DEFAULT_CHECK_WINDOW = 2 * time.Second // instance DELETE event will be delay.

const TIME_FORMAT = "15:04:05.000"

const EVENT_BUS_MAX_SIZE = 1000

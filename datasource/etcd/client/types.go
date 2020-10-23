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

package client

import (
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"strconv"
)

type ImplName string

type ActionType int

func (at ActionType) String() string {
	switch at {
	case ActionGet:
		return "GET"
	case ActionPut:
		return "PUT"
	case ActionDelete:
		return "DELETE"
	default:
		return "ACTION" + strconv.Itoa(int(at))
	}
}

type CacheMode int

func (cm CacheMode) String() string {
	switch cm {
	case ModeBoth:
		return "MODE_BOTH"
	case ModeCache:
		return "MODE_CACHE"
	case ModeNoCache:
		return "MODE_NO_CACHE"
	default:
		return "MODE" + strconv.Itoa(int(cm))
	}
}

type SortOrder int

func (so SortOrder) String() string {
	switch so {
	case SortNone:
		return "SORT_NONE"
	case SortAscend:
		return "SORT_ASCEND"
	case SortDescend:
		return "SORT_DESCEND"
	default:
		return "SORT" + strconv.Itoa(int(so))
	}
}

type CompareType int

func (ct CompareType) String() string {
	switch ct {
	case CmpVersion:
		return "CMP_VERSION"
	case CmpCreate:
		return "CMP_CREATE"
	case CmpMod:
		return "CMP_MOD"
	case CmpValue:
		return "CMP_VALUE"
	default:
		return "CMP_TYPE" + strconv.Itoa(int(ct))
	}
}

type CompareResult int

func (cr CompareResult) String() string {
	switch cr {
	case CmpEqual:
		return "CMP_EQUAL"
	case CmpGreater:
		return "CMP_GREATER"
	case CmpLess:
		return "CMP_LESS"
	case CmpNotEqual:
		return "CMP_NOT_EQUAL"
	default:
		return "CMP_RESULT" + strconv.Itoa(int(cr))
	}
}

type PluginResponse struct {
	Action    ActionType
	Kvs       []*mvccpb.KeyValue
	Count     int64
	Revision  int64
	Succeeded bool
}

func (pr *PluginResponse) MaxModRevision() (max int64) {
	for _, kv := range pr.Kvs {
		if max < kv.ModRevision {
			max = kv.ModRevision
		}
	}
	return
}

func (pr *PluginResponse) String() string {
	return fmt.Sprintf("{action: %s, count: %d/%d, rev: %d, succeed: %v}",
		pr.Action, len(pr.Kvs), pr.Count, pr.Revision, pr.Succeeded)
}

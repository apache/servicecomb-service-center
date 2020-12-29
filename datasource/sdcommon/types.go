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

package sdcommon

import (
	"strconv"
)

const (
	ActionCreate ActionType = iota
	ActionUpdate
	ActionDelete
	// ActionPUT may be create or update
	ActionPUT
)

type ActionType int

func (at ActionType) String() string {
	switch at {
	case ActionCreate:
		return "CREATE"
	case ActionUpdate:
		return "UPDATE"
	case ActionDelete:
		return "DELETE"
	case ActionPUT:
		return "PUT"
	default:
		return "ACTION" + strconv.Itoa(int(at))
	}
}

type ListWatchResp struct {
	Action ActionType
	// Revision is only for etcd
	Revision int64
	// Resources may be list of Instance or Service
	Resources []*Resource
}

type Resource struct {
	// Key in etcd is prefix, in mongo is resourceId
	Key string
	// DocumentID is only for mongo
	DocumentID string

	// this is only for etcd
	CreateRevision int64
	ModRevision    int64
	Version        int64

	Value interface{}
}

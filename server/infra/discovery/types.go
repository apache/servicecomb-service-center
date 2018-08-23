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

package discovery

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"strconv"
)

var typeNames []string

const (
	TypeError = Type(-1)
)

type Type int

func (st Type) String() string {
	if int(st) < 0 {
		return "TypeError"
	}
	if int(st) < len(typeNames) {
		return typeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

func Types() (ids []Type) {
	for i := range typeNames {
		ids = append(ids, Type(i))
	}
	return
}

func RegisterType(name string) (newId Type) {
	newId = Type(len(typeNames))
	typeNames = append(typeNames, name)
	return
}

type KeyValue struct {
	Key            []byte
	Value          interface{}
	Version        int64
	CreateRevision int64
	ModRevision    int64
}

type Response struct {
	Kvs   []*KeyValue
	Count int64
}

type KvEvent struct {
	Revision int64
	Type     proto.EventType
	Prefix   string
	KV       *KeyValue
}

type KvEventFunc func(evt KvEvent)

type KvEventHandler interface {
	Type() Type
	OnEvent(evt KvEvent)
}

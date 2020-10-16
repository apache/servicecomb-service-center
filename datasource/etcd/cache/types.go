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

package cache

import (
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/pkg/types"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"strconv"
	"time"
)

var (
	Types     []Type
	typeNames []string
)

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

func RegisterType(name string) (newID Type, err error) {
	for _, n := range Types {
		if n.String() == name {
			return TypeError, fmt.Errorf("redeclare store type '%s'", n)
		}
	}
	newID = Type(len(Types))
	Types = append(Types, newID)
	typeNames = append(typeNames, name)
	return
}

type KeyValue struct {
	Key            []byte
	Value          interface{}
	Version        int64
	CreateRevision int64
	ModRevision    int64
	ClusterName    string
}

func (kv *KeyValue) String() string {
	b, _ := json.Marshal(kv.Value)
	return fmt.Sprintf("{key: '%s', value: %s, version: %d, cluster: '%s'}",
		util.BytesToStringWithNoCopy(kv.Key), util.BytesToStringWithNoCopy(b), kv.Version, kv.ClusterName)
}

func NewKeyValue() *KeyValue {
	return &KeyValue{ClusterName: etcd.Configuration().ClusterName}
}

type Response struct {
	Kvs   []*KeyValue
	Count int64
}

type KvEvent struct {
	Revision int64
	Type     types.EventType
	KV       *KeyValue
	CreateAt simple.Time
}

type KvEventFunc func(evt KvEvent)

type KvEventHandler interface {
	Type() Type
	OnEvent(evt KvEvent)
}

func NewKvEvent(action types.EventType, kv *KeyValue, rev int64) KvEvent {
	return KvEvent{Type: action, KV: kv, Revision: rev, CreateAt: simple.FromTime(time.Now())}
}

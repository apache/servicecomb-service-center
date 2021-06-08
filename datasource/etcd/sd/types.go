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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var Types []Type

const (
	TypeError = Type("ERROR")
)

type Type string

type Kind string

func (st Type) String() string {
	return string(st)
}

func RegisterType(name string) (Type, error) {
	for _, n := range Types {
		if string(n) == name {
			return TypeError, fmt.Errorf("redeclare store type '%s'", n)
		}
	}
	t := Type(name)
	Types = append(Types, t)
	return t, nil
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
	return &KeyValue{ClusterName: client.DefaultClusterName}
}

type Response struct {
	Kvs   []*KeyValue
	Count int64
}

type KvEvent struct {
	Revision int64
	Type     discovery.EventType
	KV       *KeyValue
	CreateAt simple.Time
}

type KvEventFunc func(evt KvEvent)

type KvEventHandler interface {
	Type() Type
	OnEvent(evt KvEvent)
}

func NewKvEvent(action discovery.EventType, kv *KeyValue, rev int64) KvEvent {
	return KvEvent{Type: action, KV: kv, Revision: rev, CreateAt: simple.FromTime(time.Now())}
}

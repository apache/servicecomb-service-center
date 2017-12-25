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
package store

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

var (
	evtProxies map[StoreType]*KvEventProxy
)

func init() {
	evtProxies = make(map[StoreType]*KvEventProxy)
	for i := StoreType(0); i != typeEnd; i++ {
		evtProxies[i] = &KvEventProxy{
			evtHandleFuncs: make([]KvEventFunc, 0, 5),
		}
	}
}

type KvEventFunc func(evt *KvEvent)

type KvEvent struct {
	Revision int64
	Action   proto.EventType
	KV       *mvccpb.KeyValue
}

type KvEventHandler interface {
	Type() StoreType
	OnEvent(evt *KvEvent)
}

type KvEventProxy struct {
	evtHandleFuncs []KvEventFunc
	lock           sync.RWMutex
}

func (h *KvEventProxy) AddHandleFunc(f KvEventFunc) {
	h.lock.Lock()
	h.evtHandleFuncs = append(h.evtHandleFuncs, f)
	h.lock.Unlock()
}

func (h *KvEventProxy) OnEvent(evt *KvEvent) {
	h.lock.RLock()
	for _, f := range h.evtHandleFuncs {
		f(evt)
	}
	h.lock.RUnlock()
}

func EventProxy(t StoreType) *KvEventProxy {
	return evtProxies[t]
}

func AddEventHandleFunc(t StoreType, f KvEventFunc) {
	EventProxy(t).AddHandleFunc(f)
}

func AddEventHandler(h KvEventHandler) {
	AddEventHandleFunc(h.Type(), h.OnEvent)
}

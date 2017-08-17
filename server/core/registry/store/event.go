//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

var (
	evtHandlers map[StoreType]*KvEventHandler
)

func init() {
	evtHandlers = make(map[StoreType]*KvEventHandler)
	for i := StoreType(0); i != typeEnd; i++ {
		evtHandlers[i] = &KvEventHandler{
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

type KvEventHandler struct {
	evtHandleFuncs []KvEventFunc
	lock           sync.RWMutex
}

func (h *KvEventHandler) AddHandleFunc(f KvEventFunc) {
	h.lock.Lock()
	h.evtHandleFuncs = append(h.evtHandleFuncs, f)
	h.lock.Unlock()
}

func (h *KvEventHandler) OnEvent(evt *KvEvent) {
	h.lock.RLock()
	for _, f := range h.evtHandleFuncs {
		f(evt)
	}
	h.lock.RUnlock()
}

func EventHandler(t StoreType) *KvEventHandler {
	return evtHandlers[t]
}

func AddEventHandleFunc(t StoreType, f KvEventFunc) {
	EventHandler(t).AddHandleFunc(f)
}

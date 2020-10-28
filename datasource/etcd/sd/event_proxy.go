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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"sync"
)

var (
	eventProxies = make(map[Type]*KvEventProxy)
)

type KvEventProxy struct {
	evtHandleFuncs []KvEventFunc
	lock           sync.RWMutex
}

func (h *KvEventProxy) AddHandleFunc(f KvEventFunc) {
	h.lock.Lock()
	h.evtHandleFuncs = append(h.evtHandleFuncs, f)
	h.lock.Unlock()
}

func (h *KvEventProxy) OnEvent(evt KvEvent) {
	h.lock.RLock()
	for _, f := range h.evtHandleFuncs {
		f(evt)
	}
	h.lock.RUnlock()
}

// InjectConfig will inject a resource changed event callback function in Config
func (h *KvEventProxy) InjectConfig(cfg *Config) *Config {
	return cfg.AppendEventFunc(h.OnEvent)
}

// unsafe
func EventProxy(t Type) *KvEventProxy {
	proxy, ok := eventProxies[t]
	if !ok {
		proxy = &KvEventProxy{}
		eventProxies[t] = proxy
	}
	return proxy
}

// the event handler/func must be good performance, or will block the event bus.
func AddEventHandleFunc(t Type, f KvEventFunc) {
	EventProxy(t).AddHandleFunc(f)
	log.Infof("register event handle function[%s] %s", t, util.FuncName(f))
}

func AddEventHandler(h KvEventHandler) {
	EventProxy(h.Type()).AddHandleFunc(h.OnEvent)
	log.Infof("register event handler[%s] %s", h.Type(), util.Reflect(h).Name())
}

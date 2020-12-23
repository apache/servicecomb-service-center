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
	"fmt"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var (
	eventProxies = &sync.Map{}
)

type MongoEventProxy struct {
	evtHandleFuncs []MongoEventFunc
	lock           sync.RWMutex
}

func (h *MongoEventProxy) AddHandleFunc(f MongoEventFunc) {
	h.lock.Lock()
	h.evtHandleFuncs = append(h.evtHandleFuncs, f)
	h.lock.Unlock()
}

func (h *MongoEventProxy) OnEvent(evt MongoEvent) {
	h.lock.RLock()
	for _, f := range h.evtHandleFuncs {
		f(evt)
	}
	h.lock.RUnlock()
}

// InjectOptions will inject a resource changed event callback function in Config
func (h *MongoEventProxy) InjectOptions(options *Options) *Options {
	return options.AppendEventFunc(h.OnEvent)
}

func EventProxy(t string) *MongoEventProxy {
	proxy, ok := eventProxies.Load(t)
	if !ok {
		proxy = &MongoEventProxy{}
		eventProxies.Store(t, proxy)
	}
	return proxy.(*MongoEventProxy)
}

func AddEventHandler(h MongoEventHandler) {
	EventProxy(h.Type()).AddHandleFunc(h.OnEvent)
	log.Info(fmt.Sprintf("register event handler[%s] %s", h.Type(), util.Reflect(h).Name()))
}

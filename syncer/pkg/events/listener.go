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
package events

import (
	"context"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
)

var (
	lock        sync.RWMutex
	listenerMap = map[string][]Listener{}
)

type ContextEvent interface {
	Type() string
	Context() context.Context
}

type Listener interface {
	OnEvent(event ContextEvent)
}

func Clean() {
	lock.Lock()
	listenerMap = map[string][]Listener{}
	lock.Unlock()
}

func AddListener(eventType string, listener Listener) {
	lock.RLock()
	list, ok := listenerMap[eventType]
	lock.RUnlock()
	if !ok {
		list = make([]Listener, 0, 10)
	}

	list = append(list, listener)
	lock.Lock()
	listenerMap[eventType] = list
	lock.Unlock()
}

func RemoveListener(eventType string, listener Listener) {
	lock.RLock()
	list, ok := listenerMap[eventType]
	lock.RUnlock()
	if !ok {
		return
	}

	for index, val := range list {
		if val == listener {
			if index == len(list)-1 {
				list = list[:index]
			} else {
				list = append(list[:index], list[index+1:]...)
			}
			break
		}
	}
	lock.Lock()
	listenerMap[eventType] = list
	lock.Unlock()
}

func Dispatch(event ContextEvent) {
	lock.RLock()
	list, ok := listenerMap[event.Type()]
	lock.RUnlock()
	if !ok {
		return
	}
	for _, listener := range list {
		gopool.Go(func(ctx context.Context) {
			listener.OnEvent(event)
		})
	}
}

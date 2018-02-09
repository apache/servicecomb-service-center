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
package util

import "sync"

type ConcurrentMap struct {
	items map[interface{}]interface{}
	size  int
	mux   sync.RWMutex
	once  sync.Once
}

func (cm *ConcurrentMap) init() {
	cm.once.Do(func() {
		cm.items = make(map[interface{}]interface{}, cm.size)
	})
}

func (cm *ConcurrentMap) Put(key, val interface{}) (old interface{}) {
	cm.init()
	cm.mux.Lock()
	old, cm.items[key] = cm.items[key], val
	cm.mux.Unlock()
	return
}

func (cm *ConcurrentMap) PutIfAbsent(key, val interface{}) (old interface{}) {
	var b bool
	cm.init()
	cm.mux.Lock()
	old, b = cm.items[key]
	if !b {
		cm.items[key] = val
	}
	cm.mux.Unlock()
	return
}

func (cm *ConcurrentMap) Get(key interface{}) (val interface{}, b bool) {
	cm.init()
	cm.mux.RLock()
	val, b = cm.items[key]
	cm.mux.RUnlock()
	return
}

func NewConcurrentMap(size int) *ConcurrentMap {
	return &ConcurrentMap{size: size}
}

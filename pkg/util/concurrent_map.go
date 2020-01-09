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

// +build !go1.9

package util

import "sync"

type MapItem struct {
	Key   interface{}
	Value interface{}
}

type ConcurrentMap struct {
	items map[interface{}]interface{}
	size  int
	mux   sync.RWMutex
	once  sync.Once
}

func (cm *ConcurrentMap) resize() {
	cm.items = make(map[interface{}]interface{}, cm.size)
}

func (cm *ConcurrentMap) init() {
	cm.once.Do(cm.resize)
}

func (cm *ConcurrentMap) Put(key, val interface{}) {
	cm.init()
	cm.mux.Lock()
	cm.items[key] = val
	cm.mux.Unlock()
	return
}

func (cm *ConcurrentMap) PutIfAbsent(key, val interface{}) (exist interface{}) {
	cm.init()
	cm.mux.Lock()
	var b bool
	exist, b = cm.items[key]
	if !b {
		cm.items[key], exist = val, val
	}
	cm.mux.Unlock()
	return
}

func (cm *ConcurrentMap) Fetch(key interface{}, f func() (interface{}, error)) (exist interface{}, err error) {
	cm.init()
	cm.mux.RLock()
	var b bool
	exist, b = cm.items[key]
	cm.mux.RUnlock()
	if !b {
		cm.mux.Lock()
		exist, b = cm.items[key]
		if !b {
			if exist, err = f(); err == nil {
				cm.items[key] = exist
			}
		}
		cm.mux.Unlock()
	}
	return
}

func (cm *ConcurrentMap) Get(key interface{}) (val interface{}, b bool) {
	cm.init()
	cm.mux.RLock()
	val, b = cm.items[key]
	cm.mux.RUnlock()
	return
}

func (cm *ConcurrentMap) Remove(key interface{}) {
	cm.init()
	cm.mux.Lock()
	delete(cm.items, key)
	cm.mux.Unlock()
	return
}

func (cm *ConcurrentMap) Clear() {
	cm.mux.Lock()
	cm.resize()
	cm.mux.Unlock()
}

func (cm *ConcurrentMap) Size() (s int) {
	return len(cm.items)
}

func (cm *ConcurrentMap) ForEach(f func(item MapItem) (next bool)) {
	cm.mux.RLock()
	s := len(cm.items)
	if s == 0 {
		cm.mux.RUnlock()
		return
	}
	// avoid dead lock in function 'f'
	ch := make([]MapItem, 0, s)
	for k, v := range cm.items {
		ch = append(ch, MapItem{k, v})
	}
	cm.mux.RUnlock()

	for _, i := range ch {
		if b := f(i); b {
			continue
		}
		break
	}
}

func NewConcurrentMap(size int) *ConcurrentMap {
	c := &ConcurrentMap{size: size}
	c.init()
	return c
}

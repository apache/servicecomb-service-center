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
	"sync"

	cmp "github.com/orcaman/concurrent-map"
)

type IndexCache struct {
	l sync.RWMutex
	// store cmp.ConcurrentMap.
	store map[string]cmp.ConcurrentMap
}

func NewIndexCache() IndexCache {
	return IndexCache{
		store: make(map[string]cmp.ConcurrentMap),
	}
}

func (m *IndexCache) Get(key string) []string {
	m.l.RLock()
	defer m.l.RUnlock()
	cmap, exist := m.store[key]
	if !exist {
		return []string{}
	}
	return cmap.Keys()
}

func (m *IndexCache) Put(key string, value string) {
	m.l.Lock()
	defer m.l.Unlock()
	cmap, exist := m.store[key]
	if !exist {
		cmap = cmp.New()
		m.store[key] = cmap
	}
	cmap.Set(value, nil)
}

func (m *IndexCache) Delete(key string, value string) {
	m.l.Lock()
	defer m.l.Unlock()
	cmap, exist := m.store[key]
	if !exist {
		return
	}
	cmap.Remove(value)
	if cmap.Count() == 0 {
		delete(m.store, key)
	}
}

func (m *IndexCache) Clear() {
	m.l.Lock()
	defer m.l.Unlock()
	for k, v := range m.store {
		v.Clear()
		delete(m.store, k)
	}
}

//go:build go1.9
// +build go1.9

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

type MapItem struct {
	Key   interface{}
	Value interface{}
}

type ConcurrentMap struct {
	mapper    sync.Map
	fetchLock sync.RWMutex
}

func (cm *ConcurrentMap) Put(key, val interface{}) {
	cm.fetchLock.RLock()
	cm.mapper.Store(key, val)
	cm.fetchLock.RUnlock()
}

func (cm *ConcurrentMap) PutIfAbsent(key, val interface{}) (exist interface{}) {
	cm.fetchLock.RLock()
	exist, _ = cm.mapper.LoadOrStore(key, val)
	cm.fetchLock.RUnlock()
	return
}

func (cm *ConcurrentMap) Fetch(key interface{}, f func() (interface{}, error)) (v interface{}, err error) {
	if exist, b := cm.mapper.Load(key); b {
		return exist, nil
	}

	cm.fetchLock.Lock()
	if exist, b := cm.mapper.Load(key); b {
		cm.fetchLock.Unlock()
		return exist, nil
	}

	if v, err = f(); err != nil {
		cm.fetchLock.Unlock()
		return nil, err
	}

	cm.mapper.Store(key, v)
	cm.fetchLock.Unlock()
	return
}

func (cm *ConcurrentMap) Get(key interface{}) (val interface{}, b bool) {
	return cm.mapper.Load(key)
}

func (cm *ConcurrentMap) Remove(key interface{}) {
	cm.fetchLock.RLock()
	cm.mapper.Delete(key)
	cm.fetchLock.RUnlock()
}

func (cm *ConcurrentMap) Clear() {
	cm.fetchLock.RLock()
	cm.mapper = sync.Map{}
	cm.fetchLock.RUnlock()
}

func (cm *ConcurrentMap) Size() (s int) {
	cm.mapper.Range(func(_, _ interface{}) bool {
		s++
		return true
	})
	return
}

func (cm *ConcurrentMap) ForEach(f func(item MapItem) (next bool)) {
	cm.mapper.Range(func(key, value interface{}) bool {
		return f(MapItem{key, value})
	})
}

func NewConcurrentMap(_ int) *ConcurrentMap {
	return new(ConcurrentMap)
}

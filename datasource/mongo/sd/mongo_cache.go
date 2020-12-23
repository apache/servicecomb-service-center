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

	"github.com/apache/servicecomb-service-center/pkg/util"
)

// MongoCache implements Cache.
// MongoCache is dedicated to stores service discovery data,
// e.g. service, instance, lease.
type MongoCache struct {
	Options       *Options
	name          string
	store         map[string]interface{}
	documentStore map[string]string
	rwMux         sync.RWMutex
	dirty         bool
}

func (c *MongoCache) Name() string {
	return c.name
}

func (c *MongoCache) Size() (l int) {
	c.rwMux.RLock()
	l = int(util.Sizeof(c.store))
	c.rwMux.RUnlock()
	return
}

func (c *MongoCache) Get(id string) (v interface{}) {
	c.rwMux.RLock()
	if p, ok := c.store[id]; ok {
		v = p
	}
	c.rwMux.RUnlock()
	return
}

func (c *MongoCache) GetKeyByDocumentID(documentKey string) (id string) {
	c.rwMux.RLock()
	id = c.documentStore[documentKey]
	c.rwMux.RUnlock()
	return
}

func (c *MongoCache) GetDocumentIDByID(id string) (documentID string) {
	c.rwMux.RLock()
	for k, v := range c.documentStore {
		if v == id {
			documentID = k
			break
		}
	}
	c.rwMux.RUnlock()
	return
}

func (c *MongoCache) Put(id string, v interface{}) {
	c.rwMux.Lock()
	c.store[id] = v
	c.rwMux.Unlock()
}

func (c *MongoCache) PutDocumentID(id string, documentID string) {
	c.rwMux.Lock()
	c.documentStore[documentID] = id
	c.rwMux.Unlock()
}

func (c *MongoCache) Remove(id string) {
	c.rwMux.Lock()
	delete(c.store, id)
	c.rwMux.Unlock()
}

func (c *MongoCache) RemoveDocumentID(documentID string) {
	c.rwMux.Lock()

	delete(c.documentStore, documentID)

	c.rwMux.Unlock()
}

func (c *MongoCache) MarkDirty() {
	c.dirty = true
}

func (c *MongoCache) Dirty() bool { return c.dirty }

func (c *MongoCache) Clear() {
	c.rwMux.Lock()
	c.dirty = false
	c.store = make(map[string]interface{})
	c.rwMux.Unlock()
}

func (c *MongoCache) ForEach(iter func(k string, v interface{}) (next bool)) {
	c.rwMux.RLock()
loopParent:
	for k, v := range c.store {
		if v == nil {
			continue loopParent
		}
		if !iter(k, v) {
			break loopParent
		}
	}
	c.rwMux.RUnlock()
}

func NewMongoCache(name string, options *Options) *MongoCache {
	return &MongoCache{
		Options:       options,
		name:          name,
		store:         make(map[string]interface{}),
		documentStore: make(map[string]string),
	}
}

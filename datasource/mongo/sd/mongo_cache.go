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

// MongoCache implements Cache.
// MongoCache is dedicated to stores service discovery data,
// e.g. service, instance, lease.
type MongoCache struct {
	Cfg   *Config
	name  string
	store map[string]interface{}
	documentStore map[string]string
	rwMux sync.RWMutex
	dirty bool
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

func (c *MongoCache) GetKeyByDocumentId(documentKey string) (id string) {
	c.rwMux.RLock()
	id = c.documentStore[documentKey]
	c.rwMux.RUnlock()
	return
}

func (c *MongoCache) GetDocumentIdById (id string) (documentId string){
	c.rwMux.RLock()
	for k, v := range c.documentStore {
		if v == id {
			documentId = k
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

func (c *MongoCache) PutDocumentId(id string, documentId string) {
	c.rwMux.Lock()
	c.documentStore[documentId] = id
	c.rwMux.Unlock()
}

func (c *MongoCache) Remove(id string) {
	c.rwMux.Lock()

	delete(c.store, id)

	log.Debugf("end of remove id:%s from cache", id)
	c.rwMux.Unlock()

}

func (c *MongoCache) RemoveDocumentId(documentId string) {
	c.rwMux.Lock()

	delete(c.documentStore, documentId)

	c.rwMux.Unlock()
	log.Debugf("end remove document id:%s ", documentId)
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
		if v == nil{
			continue loopParent
		}
		if !iter(k, v) {
			break loopParent
		}
	}
	c.rwMux.RUnlock()
}

func NewMongoCache(name string, cfg *Config) *MongoCache {
	return &MongoCache{
		Cfg:   cfg,
		name:  name,
		store: make(map[string]interface{}),
		documentStore: make(map[string]string),
	}
}

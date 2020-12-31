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
	"github.com/patrickmn/go-cache"
)

// MongoCache implements Cache.
// MongoCache is dedicated to stores service discovery data,
// e.g. service, instance, lease.
// the docStore consists of two parts.
// 1. documentID --> bussinessID
// 2. bussinessID --> documentID
// the store consists of two parts.
// 1. index --> bussinessID list
// 2. bussinessID --> index
type MongoCache struct {
	Options    *Options
	name       string
	store      *cache.Cache
	docStore   *cache.Cache
	indexStore *cache.Cache
	dirty      bool
}

func (c *MongoCache) Name() string {
	return c.name
}

func (c *MongoCache) Size() (l int) {
	return c.store.ItemCount()
}

func (c *MongoCache) Get(id string) (v interface{}) {
	v, _ = c.store.Get(id)
	return
}

func (c *MongoCache) GetKeyByDocumentID(documentKey string) (id string) {
	if v, f := c.docStore.Get(documentKey); f {
		t, ok := v.(string)
		if ok {
			id = t
		}
	}
	return
}

func (c *MongoCache) GetDocumentIDByBussinessID(id string) (documentID string) {
	v, f := c.docStore.Get(id)
	if f {
		if id, ok := v.(string); ok {
			documentID = id
		}
	}
	return
}

func (c *MongoCache) Put(id string, v interface{}) {
	c.store.Set(id, v, cache.NoExpiration)
}

func (c *MongoCache) PutDocumentID(id string, documentID string) {
	//store docID-->ID&ID-->docID
	c.docStore.Set(documentID, id, cache.NoExpiration)
	c.docStore.Set(id, documentID, cache.NoExpiration)
}

func (c *MongoCache) Remove(id string) {
	c.store.Delete(id)
	c.docStore.Delete(id)
	c.indexStore.Delete(id)
}

func (c *MongoCache) RemoveDocumentID(documentID string) {
	c.docStore.Delete(documentID)
}

func (c *MongoCache) GetIndexData(index string) (res []string) {
	if p, found := c.indexStore.Get(index); found {
		res, ok := p.([]string)
		if ok {
			return res
		}
	}
	return
}

func (c *MongoCache) GetIndexByBussinessID(id string) (index string) {
	if v, found := c.indexStore.Get(id); found {
		if t, ok := v.(string); ok {
			index = t
		}
	}
	return
}

func (c *MongoCache) PutIndex(index string, newID string) {
	v, found := c.indexStore.Get(index)
	if !found {
		c.indexStore.Set(index, []string{newID}, cache.NoExpiration)
	} else {
		if ids, ok := v.([]string); ok {
			for _, id := range ids {
				if id == newID {
					return
				}
			}
			ids = append(ids, newID)
			c.indexStore.Set(index, ids, cache.NoExpiration)
		}
	}
	//set id-->index for filterdelete
	c.indexStore.Set(newID, index, cache.NoExpiration)
}

func (c *MongoCache) RemoveIndex(index string, oldID string) {
	if v, found := c.indexStore.Get(index); found {
		ids, ok := v.([]string)
		if ok {
			var newIDs []string
			for _, id := range ids {
				if id == oldID {
					continue
				}
				newIDs = append(newIDs, id)
			}
			if len(newIDs) == 0 {
				c.indexStore.Delete(index)
			} else {
				c.indexStore.Set(index, newIDs, cache.NoExpiration)
			}
		}
	}
}

func (c *MongoCache) MarkDirty() {
	c.dirty = true
}

func (c *MongoCache) Dirty() bool { return c.dirty }

func (c *MongoCache) Clear() {
	c.dirty = false
	c.store.Flush()
	c.docStore.Flush()
	c.indexStore.Flush()
}

func (c *MongoCache) ForEach(iter func(k string, v interface{}) (next bool)) {
	items := c.store.Items()
	for k, v := range items {
		if v.Object == nil {
			continue
		}
		if !iter(k, v) {
			break
		}
	}
}

func NewMongoCache(name string, options *Options) *MongoCache {
	return &MongoCache{
		Options:    options,
		name:       name,
		store:      cache.New(cache.NoExpiration, 0),
		docStore:   cache.New(cache.NoExpiration, 0),
		indexStore: cache.New(cache.NoExpiration, 0),
	}
}

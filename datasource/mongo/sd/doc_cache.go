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

import "github.com/patrickmn/go-cache"

type DocStore struct {
	store *cache.Cache
}

func NewDocStore() *DocStore {
	return &DocStore{store: cache.New(cache.NoExpiration, 0)}
}

func (c *DocStore) Get(key string) (v interface{}) {
	v, ok := c.store.Get(key)
	if !ok {
		return nil
	}
	return
}

func (c *DocStore) Put(key string, v interface{}) {
	c.store.Set(key, v, cache.NoExpiration)
}

func (c *DocStore) DeleteDoc(key string) {
	c.store.Delete(key)
}

func (c *DocStore) ForEach(iter func(k string, v interface{}) (next bool)) {
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

func (c *DocStore) Size() int {
	return c.store.ItemCount()
}

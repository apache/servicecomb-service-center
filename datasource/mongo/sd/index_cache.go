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

type indexCache struct {
	store *cache.Cache
}

func NewIndexCache() *indexCache {
	return &indexCache{
		store: cache.New(cache.NoExpiration, 0),
	}
}

func (i *indexCache) Get(key string) []string {
	if v, found := i.store.Get(key); found {
		hset, ok := v.(*Hset)
		if ok {
			return hset.Iter()
		}
	}
	return nil
}

func (i *indexCache) Put(key string, value string) {
	//todo this should be atomic
	v, found := i.store.Get(key)
	if !found {
		i.store.Set(key, Newhset(value), cache.NoExpiration)
		return
	}
	set, ok := v.(*Hset)
	if !ok {
		return
	}
	set.Insert(value)
}

func (i *indexCache) Delete(key string, value string) {
	v, found := i.store.Get(key)
	if !found {
		return
	}
	set, ok := v.(*Hset)
	if !ok {
		return
	}
	set.Del(value)
	if set.Len() == 0 {
		i.store.Delete(key)
	}
}

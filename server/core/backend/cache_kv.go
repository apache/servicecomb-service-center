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
package backend

import (
	"strings"
	"sync"
	"time"
)

type KvCache struct {
	Cfg         *Config
	name        string
	store       map[string]map[string]*KeyValue
	rwMux       sync.RWMutex
	lastRefresh time.Time
	lastMaxSize int
}

func (c *KvCache) Name() string {
	return c.name
}

func (c *KvCache) Size() (l int) {
	l = c.GetAll(c.Cfg.Prefix, nil)
	return
}

func (c *KvCache) Get(key string) (v *KeyValue) {
	c.rwMux.RLock()
	prefix := c.prefix(key)
	if p, ok := c.store[prefix]; ok {
		v, _ = p[key]
	}
	c.rwMux.RUnlock()
	return
}

func (c *KvCache) GetAll(prefix string, arr *[]*KeyValue) (count int) {
	c.rwMux.RLock()
	count = c.getPrefixKey(arr, prefix)
	c.rwMux.RUnlock()
	return
}

func (c *KvCache) Put(key string, v *KeyValue) {
	c.rwMux.Lock()
	c.addPrefixKey(key, v)
	c.rwMux.Unlock()
}

func (c *KvCache) Remove(key string) {
	c.rwMux.Lock()
	c.deletePrefixKey(key)
	c.rwMux.Unlock()
}

func (c *KvCache) ForEach(iter func(k string, v *KeyValue) (next bool)) {
	c.rwMux.RLock()
loopParent:
	for _, p := range c.store {
		for k, v := range p {
			if v == nil {
				continue loopParent
			}
			if !iter(k, v) {
				break loopParent
			}
		}
	}
	c.rwMux.RUnlock()
}

func (c *KvCache) prefix(key string) string {
	if len(key) == 0 {
		return ""
	}
	return key[:strings.LastIndex(key[:len(key)-1], "/")+1]
}

func (c *KvCache) getPrefixKey(arr *[]*KeyValue, prefix string) (count int) {
	keysRef, ok := c.store[prefix]
	if !ok {
		return 0
	}

	for key := range keysRef {
		n := c.getPrefixKey(arr, key)
		if n == 0 {
			count += len(keysRef)
			if arr != nil {
				// TODO support sort option
				for _, v := range keysRef {
					*arr = append(*arr, v)
				}
			}
			break
		}
		count += n
	}
	return count
}

func (c *KvCache) addPrefixKey(key string, val *KeyValue) {
	if len(c.Cfg.Prefix) >= len(key) {
		return
	}
	prefix := c.prefix(key)
	if len(prefix) == 0 {
		return
	}
	keys, ok := c.store[prefix]
	if !ok {
		// build parent index key and new child nodes
		keys = make(map[string]*KeyValue)
		c.store[prefix] = keys
	} else if _, ok := keys[key]; ok {
		if val != nil {
			// override the value
			keys[key] = val
		}
		return
	}

	keys[key], key = val, prefix
	c.addPrefixKey(key, nil)
}

func (c *KvCache) deletePrefixKey(key string) {
	prefix := c.prefix(key)
	m, ok := c.store[prefix]
	if !ok {
		return
	}
	delete(m, key)

	// remove parent which has no child
	if len(m) == 0 {
		delete(c.store, prefix)
		c.deletePrefixKey(prefix)
	}
}

func NewKvCache(name string, cfg *Config) *KvCache {
	return &KvCache{
		Cfg:         cfg,
		name:        name,
		store:       make(map[string]map[string]*KeyValue),
		lastRefresh: time.Now(),
	}
}

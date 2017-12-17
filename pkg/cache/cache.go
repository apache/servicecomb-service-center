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
package cache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Cache is a goroutine-safe K/V cache.
type Cache struct {
	sync.RWMutex
	items             map[string]*Item
	defaultExpiration time.Duration
}

type Item struct {
	Object     interface{}
	Expiration *time.Time
}

// Returns true if the item has expired.
func (item *Item) Expired() bool {
	if item.Expiration == nil {
		return false
	}
	return item.Expiration.Before(time.Now())
}

// New create a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than 1, the items in the cache
// never expire (by default), and must be deleted manually. If the cleanup
// interval is less than one, expired items are not deleted from the cache
// before calling DeleteExpired.
func New(defaultExpiration, cleanInterval time.Duration) *Cache {
	c := &Cache{
		items:             map[string]*Item{},
		defaultExpiration: defaultExpiration,
	}
	if cleanInterval > 0 {
		go func() {
			for {
				time.Sleep(cleanInterval)
				c.DeleteExpired()
			}
		}()
	}
	return c
}

// Get return an item or nil, and a bool indicating whether
// the key was found.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	item, ok := c.items[key]
	if !ok || item.Expired() {
		c.RUnlock()
		return nil, false
	}
	c.RUnlock()
	return item.Object, true
}

// Get all cache keys
func (c *Cache) Keys() []string {
	c.RLock()
	defer c.RUnlock()
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// Set add a new key or replace an exist key. If the dur is 0, we will
// use the defaultExpiration.
func (c *Cache) Set(key string, val interface{}, dur time.Duration) {
	var t *time.Time
	c.Lock()
	if dur == 0 {
		dur = c.defaultExpiration
	}
	if dur > 0 {
		tmp := time.Now().Add(dur)
		t = &tmp
	}
	c.items[key] = &Item{
		Object:     val,
		Expiration: t,
	}
	c.Unlock()
}

// Delete a key-value pair if the key is existed.
func (c *Cache) Delete(key string) {
	c.Lock()
	delete(c.items, key)
	c.Unlock()
}

// Delete all cache.
func (c *Cache) Flush() {
	c.Lock()
	c.items = map[string]*Item{}
	c.Unlock()
}

// Add a number to a key-value pair.
func (c *Cache) Increment(key string, x int64) error {
	c.Lock()
	val, ok := c.items[key]
	if !ok || val.Expired() {
		c.Unlock()
		return fmt.Errorf("Item %s not found", key)
	}
	switch val.Object.(type) {
	case int:
		val.Object = val.Object.(int) + int(x)
	case int8:
		val.Object = val.Object.(int8) + int8(x)
	case int16:
		val.Object = val.Object.(int16) + int16(x)
	case int32:
		val.Object = val.Object.(int32) + int32(x)
	case int64:
		val.Object = val.Object.(int64) + x
	case uint:
		val.Object = val.Object.(uint) + uint(x)
	case uint8:
		val.Object = val.Object.(uint8) + uint8(x)
	case uint16:
		val.Object = val.Object.(uint16) + uint16(x)
	case uint32:
		val.Object = val.Object.(uint32) + uint32(x)
	case uint64:
		val.Object = val.Object.(uint64) + uint64(x)
	case uintptr:
		val.Object = val.Object.(uintptr) + uintptr(x)
	default:
		c.Unlock()
		return fmt.Errorf("The value type error")
	}
	c.Unlock()
	return nil
}

// Sub a number to a key-value pair.
func (c *Cache) Decrement(key string, x int64) error {
	c.Lock()
	val, ok := c.items[key]
	if !ok || val.Expired() {
		c.Unlock()
		return fmt.Errorf("Item %s not found", key)
	}
	switch val.Object.(type) {
	case int:
		val.Object = val.Object.(int) - int(x)
	case int8:
		val.Object = val.Object.(int8) - int8(x)
	case int16:
		val.Object = val.Object.(int16) - int16(x)
	case int32:
		val.Object = val.Object.(int32) - int32(x)
	case int64:
		val.Object = val.Object.(int64) - x
	case uint:
		val.Object = val.Object.(uint) - uint(x)
	case uint8:
		val.Object = val.Object.(uint8) - uint8(x)
	case uint16:
		val.Object = val.Object.(uint16) - uint16(x)
	case uint32:
		val.Object = val.Object.(uint32) - uint32(x)
	case uint64:
		val.Object = val.Object.(uint64) - uint64(x)
	case uintptr:
		val.Object = val.Object.(uintptr) - uintptr(x)
	default:
		c.Unlock()
		return fmt.Errorf("The value type error")
	}
	c.Unlock()
	return nil
}

// Return the number of item in cache.
func (c *Cache) ItemCount() int {
	c.RLock()
	counts := len(c.items)
	c.RUnlock()
	return counts
}

// Delete all expired items.
func (c *Cache) DeleteExpired() {
	c.Lock()
	for k, v := range c.items {
		if v.Expired() {
			delete(c.items, k)
		}
	}
	c.Unlock()
}

// The LRUCache is a goroutine-safe cache.
type LRUCache struct {
	sync.RWMutex
	maxEntries int
	items      map[string]*list.Element
	cacheList  *list.List
}

type entry struct {
	key   string
	value interface{}
}

// NewLRU create a LRUCache with max size. The size is 0 means no limit.
func NewLRU(size int) (*LRUCache, error) {
	if size < 0 {
		return nil, errors.New("The size of LRU Cache must no less than 0")
	}
	lru := &LRUCache{
		maxEntries: size,
		items:      make(map[string]*list.Element, size),
		cacheList:  list.New(),
	}
	return lru, nil
}

// Add a new key-value pair to the LRUCache.
func (c *LRUCache) Add(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if ent, hit := c.items[key]; hit {
		c.cacheList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return
	}
	ent := &entry{
		key:   key,
		value: value,
	}
	entry := c.cacheList.PushFront(ent)
	c.items[key] = entry

	if c.maxEntries > 0 && c.cacheList.Len() > c.maxEntries {
		c.removeOldestElement()
	}
}

// Get a value from the LRUCache. And a bool indicating
// whether found or not.
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	if ent, hit := c.items[key]; hit {
		c.cacheList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return nil, false
}

// Remove a key-value pair in LRUCache. If the key is not existed,
// nothing will happen.
func (c *LRUCache) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	if ent, hit := c.items[key]; hit {
		c.removeElement(ent)
	}
}

// Return the number of key-value pair in LRUCache.
func (c *LRUCache) Len() int {
	c.RLock()
	length := c.cacheList.Len()
	c.RUnlock()
	return length
}

// Delete all entry in the LRUCache. But the max size will hold.
func (c *LRUCache) Clear() {
	c.Lock()
	c.cacheList = list.New()
	c.items = make(map[string]*list.Element, c.maxEntries)
	c.Unlock()
}

// Resize the max limit.
func (c *LRUCache) SetMaxEntries(max int) error {
	if max < 0 {
		return errors.New("The max limit of entryies must no less than 0")
	}
	c.Lock()
	c.maxEntries = max
	c.Unlock()
	return nil
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.cacheList.Remove(e)
	ent := e.Value.(*entry)
	delete(c.items, ent.key)
}

func (c *LRUCache) removeOldestElement() {
	ent := c.cacheList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

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
	"context"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"sync"
)

type Cache struct {
	data context.Context
	lock sync.RWMutex
}

func (c *Cache) Set(k string, v interface{}) {
	c.lock.Lock()
	c.data = context.WithValue(c.data, util.CtxKey(k), v)
	c.lock.Unlock()
}

func (c *Cache) Get(k string) (v interface{}) {
	c.lock.RLock()
	v = c.data.Value(util.CtxKey(k))
	c.lock.RUnlock()
	return
}

func NewCache() *Cache {
	return &Cache{
		data: context.Background(),
	}
}

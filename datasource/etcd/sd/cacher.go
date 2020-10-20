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
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/types"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// CommonCacher implements pkg.Cacher.
// CommonCacher is universal to manage cache of any cache.
// Use Cfg to set it's behavior.
type CommonCacher struct {
	Cfg *Config
	// cache for indexer
	cache Cache

	ready chan struct{}
}

func (c *CommonCacher) Cache() CacheReader {
	return c.cache
}

func (c *CommonCacher) Notify(action types.EventType, key string, kv *KeyValue) {
	switch action {
	case registry.EVT_DELETE:
		c.cache.Remove(key)
	default:
		c.cache.Put(key, kv)
	}
	c.OnEvent(NewKvEvent(action, kv, kv.ModRevision))
}

func (c *CommonCacher) OnEvent(evt KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer log.Recover()
	c.Cfg.OnEvent(evt)
}

func (c *CommonCacher) Run() {
	util.SafeCloseChan(c.ready)
}

func (c *CommonCacher) Stop() {
}

func (c *CommonCacher) Ready() <-chan struct{} {
	return c.ready
}

func NewCommonCacher(cfg *Config, cache Cache) *CommonCacher {
	return &CommonCacher{
		Cfg:   cfg,
		cache: cache,
		ready: make(chan struct{}),
	}
}

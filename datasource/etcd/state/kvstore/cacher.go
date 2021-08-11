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

package kvstore

import (
	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// CommonCacher implements kvstore.Cacher.
// CommonCacher is universal to manage cache of any registry.
// Use Cfg to set it's behavior.
type CommonCacher struct {
	Cfg *Options
	// cache for indexer
	cache Cache

	ready chan struct{}
}

func (c *CommonCacher) Cache() CacheReader {
	return c.cache
}

func (c *CommonCacher) Notify(action discovery.EventType, key string, kv *KeyValue) {
	switch action {
	case discovery.EVT_DELETE:
		c.cache.Remove(key)
	default:
		c.cache.Put(key, kv)
	}
	c.OnEvent(NewEvent(action, kv, kv.ModRevision))
}

func (c *CommonCacher) OnEvent(evt Event) {
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

func NewCommonCacher(cfg *Options, cache Cache) *CommonCacher {
	return &CommonCacher{
		Cfg:   cfg,
		cache: cache,
		ready: make(chan struct{}),
	}
}

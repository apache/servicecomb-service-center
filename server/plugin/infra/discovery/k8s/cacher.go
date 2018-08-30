// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"golang.org/x/net/context"
)

type K8sCacher struct {
	Cfg *discovery.Config
	// cache for indexer
	cache discovery.Cache

	ready     chan struct{}
	goroutine *gopool.Pool
}

func (c *K8sCacher) Cache() discovery.Cache {
	return c.cache
}

func (c *K8sCacher) Notify(action proto.EventType, key string, kv *discovery.KeyValue) {
	switch action {
	case proto.EVT_DELETE:
		c.cache.Remove(key)
	default:
		c.cache.Put(key, kv)
	}
	c.OnEvents(discovery.KvEvent{Type: action, KV: kv})
}

func (c *K8sCacher) OnEvents(evt discovery.KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer log.Recover()
	c.Cfg.OnEvent(evt)
}

func (c *K8sCacher) Run() {
	util.SafeCloseChan(c.ready)
}

func (c *K8sCacher) Stop() {
	c.goroutine.Close(true)
}

func (c *K8sCacher) Ready() <-chan struct{} {
	return c.ready
}

func NewK8sCacher(cfg *discovery.Config, cache discovery.Cache) *K8sCacher {
	return &K8sCacher{
		Cfg:       cfg,
		cache:     cache,
		ready:     make(chan struct{}),
		goroutine: gopool.New(context.Background()),
	}
}

func BuildCacher(t discovery.Type, cfg *discovery.Config, cache discovery.Cache) discovery.Cacher {
	kc := NewK8sCacher(cfg, cache)
	switch t {
	case backend.SERVICE:
		return NewServiceCacher(kc)
	case backend.SERVICE_INDEX:
		return NewServiceIndexCacher(kc)
	case backend.INSTANCE:
		return NewInstanceCacher(kc)
	default:
		return kc
	}
}

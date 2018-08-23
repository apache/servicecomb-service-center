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
package etcd

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"sync"
)

var (
	defaultEtcdAdaptor *EtcdAdaptor
	newEtcdAdaptorOnce sync.Once
)

type EtcdAdaptor struct {
	discovery.Cacher
	discovery.Indexer
}

func (se *EtcdAdaptor) Run() {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		r.Run()
	}
}

func (se *EtcdAdaptor) Stop() {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		r.Stop()
	}
}

func (se *EtcdAdaptor) Ready() <-chan struct{} {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		return r.Ready()
	}
	return closedCh
}

func NewEtcdAdaptor(name string, cfg *discovery.Config) *EtcdAdaptor {
	var adaptor EtcdAdaptor
	switch {
	case core.ServerInfo.Config.EnableCache && cfg.InitSize > 0:
		cache := discovery.NewKvCache(name, cfg)
		adaptor.Cacher = NewKvCacher(cfg, cache)
		adaptor.Indexer = NewCacheIndexer(cfg, cache)
	default:
		log.Infof(
			"core will not cache '%s' and ignore all events of it, cache enabled: %v, init size: %d",
			name, core.ServerInfo.Config.EnableCache, cfg.InitSize)
		adaptor.Indexer = NewEtcdIndexer(cfg.Key, cfg.Parser)
	}
	return &adaptor
}

func DefaultKvEntity() *EtcdAdaptor {
	newEtcdAdaptorOnce.Do(func() {
		defaultEtcdAdaptor = &EtcdAdaptor{
			Indexer: NewEtcdIndexer(discovery.Configure().Key, discovery.BytesParser),
		}
	})
	return defaultEtcdAdaptor
}

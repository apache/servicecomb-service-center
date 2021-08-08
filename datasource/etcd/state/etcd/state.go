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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// State implements kvstore.State.
// State does service pkg with etcd as it's cache.
type State struct {
	kvstore.Cacher
	kvstore.Indexer
}

func (se *State) Run() {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		r.Run()
	}
}

func (se *State) Stop() {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		r.Stop()
	}
}

func (se *State) Ready() <-chan struct{} {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		return r.Ready()
	}
	return closedCh
}

func NewEtcdState(name string, cfg *kvstore.Options) *State {
	var adaptor State
	switch {
	case cfg.InitSize > 0:
		kvCache := kvstore.NewKvCache(name, cfg)
		adaptor.Cacher = NewKvCacher(cfg, kvCache)
		adaptor.Indexer = NewCacheIndexer(cfg, kvCache)
	default:
		log.Info(fmt.Sprintf("core will not cache '%s' and ignore all events of it, init size: %d",
			name, cfg.InitSize))
		adaptor.Cacher = kvstore.NullCacher
		adaptor.Indexer = NewEtcdIndexer(cfg.Key, cfg.Parser)
	}
	return &adaptor
}

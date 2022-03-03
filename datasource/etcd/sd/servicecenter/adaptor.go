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

package servicecenter

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
)

// Adaptor implements state.State.
// Adaptor does service pkg with other service-centers
// as it's cache.
type Adaptor struct {
	kvstore.Cacher
	kvstore.Indexer
}

func (se *Adaptor) Run() {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		r.Run()
	}
}

func (se *Adaptor) Stop() {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		r.Stop()
	}
}

func (se *Adaptor) Ready() <-chan struct{} {
	if r, ok := se.Cacher.(kvstore.Runnable); ok {
		return r.Ready()
	}
	return closedCh
}

func NewServiceCenterAdaptor(t kvstore.Type, cfg *kvstore.Options) *Adaptor {
	if t == sd.TypeSchema {
		return &Adaptor{
			Indexer: NewClusterIndexer(t, kvstore.NullCache),
			Cacher:  kvstore.NullCacher,
		}
	}
	cache := kvstore.NewKvCache(t.String(), cfg)
	return &Adaptor{
		Indexer: NewClusterIndexer(t, cache),
		Cacher:  BuildCacher(t, cfg, cache),
	}
}

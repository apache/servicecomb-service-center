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

package adaptor

import (
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
)

// K8sAdaptor implements discovery.Adaptor.
// K8sAdaptor does service discovery with kubernetes as it's registry.
type K8sAdaptor struct {
	discovery.Cacher
	discovery.Indexer
}

func (se *K8sAdaptor) Run() {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		r.Run()
	}
}

func (se *K8sAdaptor) Stop() {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		r.Stop()
	}
}

func (se *K8sAdaptor) Ready() <-chan struct{} {
	if r, ok := se.Cacher.(discovery.Runnable); ok {
		return r.Ready()
	}
	return closedCh
}

func NewK8sAdaptor(t discovery.Type, cfg *discovery.Config) *K8sAdaptor {
	cache := discovery.NewKvCache(t.String(), cfg)
	return &K8sAdaptor{
		Indexer: discovery.NewCacheIndexer(cache),
		Cacher:  BuildCacher(t, cfg, cache),
	}
}

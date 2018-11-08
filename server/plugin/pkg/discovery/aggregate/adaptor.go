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

package aggregate

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"
)

// Aggregator is a discovery service adaptor implement of one registry cluster
type Aggregator struct {
	discovery.Indexer
	Adaptors []discovery.Adaptor
}

func (as *Aggregator) Cache() discovery.Cache {
	var cache Cache
	for _, a := range as.Adaptors {
		cache = append(cache, a.Cache())
	}
	return cache
}

func (as *Aggregator) Run() {
	for _, a := range as.Adaptors {
		a.Run()
	}
}

func (as *Aggregator) Stop() {
	for _, a := range as.Adaptors {
		a.Stop()
	}
}

func (as *Aggregator) Ready() <-chan struct{} {
	for _, a := range as.Adaptors {
		<-a.Ready()
	}
	return closedCh
}

func getLogConflictFunc(t discovery.Type) func(origin, conflict *discovery.KeyValue) {
	switch t {
	case backend.SERVICE_INDEX:
		return func(origin, conflict *discovery.KeyValue) {
			if serviceId, conflictId := origin.Value.(string), conflict.Value.(string); conflictId != serviceId {
				key := core.GetInfoFromSvcIndexKV(conflict.Key)
				log.Warnf("conflict! can not merge microservice index[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
					conflict.ClusterName, conflictId, key.Environment, key.AppId, key.ServiceName, key.Version,
					serviceId, origin.ClusterName)
			}
		}
	case backend.SERVICE_ALIAS:
		return func(origin, conflict *discovery.KeyValue) {
			if serviceId, conflictId := origin.Value.(string), conflict.Value.(string); conflictId != serviceId {
				key := core.GetInfoFromSvcAliasKV(conflict.Key)
				log.Warnf("conflict! can not merge microservice alias[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
					conflict.ClusterName, conflictId, key.Environment, key.AppId, key.ServiceName, key.Version,
					serviceId, origin.ClusterName)
			}
		}
	}
	return nil
}

func NewAggregator(t discovery.Type, cfg *discovery.Config) *Aggregator {
	as := &Aggregator{}
	for _, name := range repos {
		repo := mgr.Plugins().Get(mgr.DISCOVERY, name).New().(discovery.AdaptorRepository)
		as.Adaptors = append(as.Adaptors, repo.New(t, cfg))
	}

	switch t {
	case backend.SCHEMA:
		// schema does not been cached, so new the adaptor indexer
		as.Indexer = NewAdaptorsIndexer(as.Adaptors)
	case backend.SERVICE_INDEX, backend.SERVICE_ALIAS:
		NewConflictChecker(as.Cache(), getLogConflictFunc(t))
		fallthrough
	default:
		as.Indexer = discovery.NewCacheIndexer(as.Cache())
	}
	return as
}

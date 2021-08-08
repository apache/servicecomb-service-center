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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// Aggregator implements state.State.
// Aggregator is an aggregator of multi Adaptors, and it aggregates all the
// Adaptors' data as it's result.
type Aggregator struct {
	// Indexer searches data from all the adapters
	kvstore.Indexer
	Type     kvstore.Type
	Adaptors []state.State
}

// Cache gets all the adapters' cache
func (as *Aggregator) Cache() kvstore.CacheReader {
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

func getLogConflictFunc(t kvstore.Type) func(origin, conflict *kvstore.KeyValue) {
	switch t {
	case sd.TypeServiceIndex:
		return func(origin, conflict *kvstore.KeyValue) {
			if serviceID, conflictID := origin.Value.(string), conflict.Value.(string); conflictID != serviceID {
				key := path.GetInfoFromSvcIndexKV(conflict.Key)
				log.Warn(fmt.Sprintf("conflict! can not merge microservice index[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
					conflict.ClusterName, conflictID, key.Environment, key.AppId, key.ServiceName, key.Version,
					serviceID, origin.ClusterName))
			}
		}
	case sd.TypeServiceAlias:
		return func(origin, conflict *kvstore.KeyValue) {
			if serviceID, conflictID := origin.Value.(string), conflict.Value.(string); conflictID != serviceID {
				key := path.GetInfoFromSvcAliasKV(conflict.Key)
				log.Warn(fmt.Sprintf("conflict! can not merge microservice alias[%s][%s][%s/%s/%s/%s], found one[%s] in cluster[%s]",
					conflict.ClusterName, conflictID, key.Environment, key.AppId, key.ServiceName, key.Version,
					serviceID, origin.ClusterName))
			}
		}
	}
	return nil
}

func NewAggregator(t kvstore.Type, cfg *kvstore.Options) *Aggregator {
	as := &Aggregator{Type: t}
	for _, name := range repos {
		// create and get all plugin instances
		repo, err := state.NewRepository(state.Config{Kind: name})
		if err != nil {
			log.Error(fmt.Sprintf("failed to new plugin instance[%s]", name), err)
			continue
		}
		as.Adaptors = append(as.Adaptors, repo.New(t, cfg))
	}
	as.Indexer = NewAggregatorIndexer(as)

	switch t {
	case sd.TypeServiceIndex, sd.TypeServiceAlias:
		NewConflictChecker(as.Cache(), getLogConflictFunc(t))
	}
	return as
}

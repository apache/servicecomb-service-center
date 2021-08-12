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
	"context"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/little-cui/etcdadpt"
)

// AdaptorsIndexer implements kvstore.Indexer.
// AdaptorsIndexer is an aggregator of multi Indexers, and it aggregates all the
// Indexers' data as it's result.
type AdaptorsIndexer struct {
	Adaptors []state.State
}

// Search implements kvstore.Indexer#Search.
// AdaptorsIndexer ignores the errors during search to ensure availability, so
// it always searches successfully, no matter how many Adaptors are abnormal.
// But at the cost of that, AdaptorsIndexer doesn't guarantee the correctness
// of the search results.
func (i *AdaptorsIndexer) Search(ctx context.Context, opts ...etcdadpt.OpOption) (*kvstore.Response, error) {
	var (
		response kvstore.Response
		exists   = make(map[string]struct{})
	)
	for _, a := range i.Adaptors {
		resp, err := a.Search(ctx, opts...)
		if err != nil {
			continue
		}
		for _, kv := range resp.Kvs {
			key := util.BytesToStringWithNoCopy(kv.Key)
			if _, ok := exists[key]; !ok {
				exists[key] = struct{}{}
				response.Kvs = append(response.Kvs, kv)
			}
		}
		response.Count += resp.Count
	}
	return &response, nil
}

// Creditable implements kvstore.Indexer#Creditable.
// AdaptorsIndexer's search result's are not creditable as it ignores the
// errors. In other words, AdaptorsIndexer makes the best efforts to search
// data, but it does not ensure the correctness.
func (i *AdaptorsIndexer) Creditable() bool {
	return false
}

func NewAdaptorsIndexer(as []state.State) *AdaptorsIndexer {
	return &AdaptorsIndexer{Adaptors: as}
}

// AggregatorIndexer implements kvstore.Indexer.
// AggregatorIndexer consists of multi Indexers and it decides which Indexer to
// use based on it's mechanism.
type AggregatorIndexer struct {
	// CacheIndexer searches data from all the adaptors's cache.
	*kvstore.CacheIndexer
	// AdaptorsIndexer searches data from all the adaptors.
	AdaptorsIndexer kvstore.Indexer
	// LocalIndexer data from local adaptor.
	LocalIndexer kvstore.Indexer
}

// Search implements kvstore.Indexer#Search.
func (i *AggregatorIndexer) Search(ctx context.Context, opts ...etcdadpt.OpOption) (resp *kvstore.Response, err error) {
	op := etcdadpt.OpGet(opts...)

	indexer := i.LocalIndexer
	if op.Global {
		// request with global param then do not use local indexer
		indexer = i.AdaptorsIndexer
	}

	if op.NoCache() || !op.Global {
		return indexer.Search(ctx, opts...)
	}

	resp, err = i.CacheIndexer.Search(ctx, opts...)
	if err != nil {
		return
	}

	if resp.Count > 0 || op.CacheOnly() {
		return resp, nil
	}

	return indexer.Search(ctx, opts...)
}

// Creditable implements kvstore.Indexer#Creditable.
func (i *AggregatorIndexer) Creditable() bool {
	return i.AdaptorsIndexer.Creditable() &&
		i.LocalIndexer.Creditable() &&
		i.CacheIndexer.Creditable()
}

func NewAggregatorIndexer(as *Aggregator) *AggregatorIndexer {
	indexer := NewAdaptorsIndexer(as.Adaptors)
	ai := &AggregatorIndexer{
		CacheIndexer:    kvstore.NewCacheIndexer(as.Cache()),
		AdaptorsIndexer: indexer,
		LocalIndexer:    indexer,
	}
	if registryIndex >= 0 {
		ai.LocalIndexer = as.Adaptors[registryIndex]
	}
	return ai
}

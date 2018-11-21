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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"golang.org/x/net/context"
)

type AdaptorsIndexer struct {
	Adaptors []discovery.Adaptor
}

func (i *AdaptorsIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*discovery.Response, error) {
	var (
		response discovery.Response
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

func NewAdaptorsIndexer(as []discovery.Adaptor) *AdaptorsIndexer {
	return &AdaptorsIndexer{Adaptors: as}
}

type AggregatorIndexer struct {
	*discovery.CacheIndexer
	AdaptorsIndexer discovery.Indexer
	LocalIndexer    discovery.Indexer
}

func (i *AggregatorIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (resp *discovery.Response, err error) {
	op := registry.OpGet(opts...)

	if op.NoCache() || !op.Global {
		return i.search(ctx, opts...)
	}

	resp, err = i.CacheIndexer.Search(ctx, opts...)
	if err != nil {
		return
	}

	if resp.Count > 0 || op.CacheOnly() {
		return resp, nil
	}

	return i.search(ctx, opts...)
}

func (i *AggregatorIndexer) search(ctx context.Context, opts ...registry.PluginOpOption) (*discovery.Response, error) {
	op := registry.OptionsToOp(opts...)
	if !op.Global {
		return i.LocalIndexer.Search(ctx, opts...)
	}

	return i.AdaptorsIndexer.Search(ctx, opts...)
}

func NewAggregatorIndexer(as *Aggregator) *AggregatorIndexer {
	indexer := NewAdaptorsIndexer(as.Adaptors)
	ai := &AggregatorIndexer{
		CacheIndexer:    discovery.NewCacheIndexer(as.Cache()),
		AdaptorsIndexer: indexer,
		LocalIndexer:    indexer,
	}
	if registryIndex >= 0 {
		ai.LocalIndexer = as.Adaptors[registryIndex]
	}
	return ai
}

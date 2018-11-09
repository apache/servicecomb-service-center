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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"golang.org/x/net/context"
)

type CacheIndexer struct {
	*EtcdIndexer
	*discovery.CacheIndexer
}

func (i *CacheIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*discovery.Response, error) {
	op := registry.OpGet(opts...)
	key := util.BytesToStringWithNoCopy(op.Key)

	if i.Cache == nil ||
		op.Mode == registry.MODE_NO_CACHE ||
		op.Revision > 0 ||
		(op.Offset >= 0 && op.Limit > 0) {
		return i.EtcdIndexer.Search(ctx, opts...)
	}

	if err := i.CheckPrefix(key); err != nil {
		return nil, fmt.Errorf("%s, cache is '%s'", err.Error(), i.Cache.Name())
	}

	resp, err := i.CacheIndexer.Search(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if resp.Count > 0 || op.Mode == registry.MODE_CACHE {
		return resp, nil
	}

	return i.EtcdIndexer.Search(ctx, opts...)
}

func NewCacheIndexer(cfg *discovery.Config, cache discovery.Cache) *CacheIndexer {
	return &CacheIndexer{
		EtcdIndexer:  NewEtcdIndexer(cfg.Key, cfg.Parser),
		CacheIndexer: discovery.NewCacheIndexer(cache),
	}
}

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
package backend

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type CacheIndexer struct {
	*CommonIndexer

	Cache Cache

	lock sync.Mutex
}

func (i *CacheIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*Response, error) {
	op := registry.OpGet(opts...)

	key := util.BytesToStringWithNoCopy(op.Key)

	if i.Cache == nil ||
		op.Mode == registry.MODE_NO_CACHE ||
		op.Revision > 0 ||
		(op.Offset >= 0 && op.Limit > 0) {
		util.Logger().Debugf("search '%s' match special options, request etcd server, opts: %s",
			key, op)
		return i.CommonIndexer.Search(ctx, opts...)
	}

	if err := i.CheckPrefix(key); err != nil {
		return nil, fmt.Errorf("%s, cache is '%s'", err.Error(), i.Cache.Name())
	}

	var resp *Response
	if op.Prefix {
		resp = i.searchByPrefix(op)
	} else {
		resp = i.search(op)
	}

	if resp.Count > 0 || op.Mode == registry.MODE_CACHE {
		return resp, nil
	}

	util.Logger().Debugf("can not find any key from %s cache, request etcd server, key: %s",
		i.Cache.Name(), key)
	return i.CommonIndexer.Search(ctx, opts...)
}

func (i *CacheIndexer) search(op registry.PluginOp) *Response {
	resp := new(Response)

	key := util.BytesToStringWithNoCopy(op.Key)

	kv := i.Cache.Get(key)
	if kv != nil {
		resp.Count = 1
	}
	if kv == nil || op.CountOnly {
		return resp
	}

	resp.Kvs = []*KeyValue{kv}
	return resp
}

func (i *CacheIndexer) searchByPrefix(op registry.PluginOp) *Response {
	resp := new(Response)

	prefix := util.BytesToStringWithNoCopy(op.Key)

	resp.Count = int64(i.Cache.GetPrefix(prefix, nil))
	if resp.Count == 0 || op.CountOnly {
		return resp
	}

	t := time.Now()
	kvs := make([]*KeyValue, 0, resp.Count)
	i.Cache.GetPrefix(prefix, &kvs)

	util.LogNilOrWarnf(t, "too long to index data[%d] from cache '%s'", len(kvs), i.Cache.Name())

	resp.Kvs = kvs
	return resp
}

func NewCacheIndexer(cfg *Config, cache Cache) *CacheIndexer {
	return &CacheIndexer{
		CommonIndexer: NewCommonIndexer(cfg.Key, cfg.Parser),
		Cache:         cache,
	}
}

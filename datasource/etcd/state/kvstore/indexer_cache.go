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

package kvstore

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// CacheIndexer implements kvstore.Indexer.
// CacheIndexer searches data from cache.
type CacheIndexer struct {
	Cache CacheReader
}

func (i *CacheIndexer) Search(_ context.Context, opts ...etcdadpt.OpOption) (resp *Response, _ error) {
	op := etcdadpt.OpGet(opts...)
	if op.Prefix {
		resp = i.searchByPrefix(op)
	} else {
		resp = i.search(op)
	}
	return
}

func (i *CacheIndexer) search(op etcdadpt.OpOptions) *Response {
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

func (i *CacheIndexer) searchByPrefix(op etcdadpt.OpOptions) *Response {
	resp := new(Response)

	prefix := util.BytesToStringWithNoCopy(op.Key)

	resp.Count = int64(i.Cache.GetPrefix(prefix, nil))
	if resp.Count == 0 || op.CountOnly {
		return resp
	}

	t := time.Now()
	kvs := make([]*KeyValue, 0, resp.Count)
	i.Cache.GetPrefix(prefix, &kvs)
	log.NilOrWarn(t, fmt.Sprintf("too long to index data[%d] from cache '%s'", len(kvs), i.Cache.Name()))

	resp.Kvs = kvs
	return resp
}

// Creditable implements pkg.Indexer.Creditable.
func (i *CacheIndexer) Creditable() bool {
	return true
}

func NewCacheIndexer(cache CacheReader) *CacheIndexer {
	return &CacheIndexer{
		Cache: cache,
	}
}

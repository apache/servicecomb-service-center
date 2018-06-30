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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type CacheIndexer struct {
	*baseIndexer

	cacher  Cacher
	lock    sync.Mutex
	isClose bool
}

func (i *CacheIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*Response, error) {
	op := registry.OpGet(opts...)

	key := util.BytesToStringWithNoCopy(op.Key)

	if !core.ServerInfo.Config.EnableCache ||
		op.Mode == registry.MODE_NO_CACHE ||
		op.Revision > 0 ||
		(op.Offset >= 0 && op.Limit > 0) {
		util.Logger().Debugf("search %s match special options, request etcd server, opts: %s",
			i.cacher.Cache().Name(), op)
		return i.baseIndexer.Search(ctx, opts...)
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

	util.Logger().Debugf("can not find any key from %s cache with prefix, request etcd server, key: %s",
		i.cacher.Cache().Name(), key)
	return i.baseIndexer.Search(ctx, opts...)
}

func (i *CacheIndexer) search(op registry.PluginOp) *Response {
	resp := new(Response)

	key := util.BytesToStringWithNoCopy(op.Key)

	kv := i.cacher.Cache().Get(key)
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

	resp.Count = int64(i.cacher.Cache().GetAll(prefix, nil))
	if resp.Count == 0 || op.CountOnly {
		return resp
	}

	t := time.Now()
	kvs := make([]*KeyValue, 0, resp.Count)
	i.cacher.Cache().GetAll(prefix, &kvs)

	util.LogNilOrWarnf(t, "too long to index data[%d] from cache '%s'", len(kvs), i.cacher.Cache().Name())

	resp.Kvs = kvs
	return resp
}

func (i *CacheIndexer) Run() {
	i.lock.Lock()
	if !i.isClose {
		i.lock.Unlock()
		return
	}
	i.isClose = false
	i.lock.Unlock()

	i.cacher.Run()
}

func (i *CacheIndexer) Stop() {
	i.lock.Lock()
	if i.isClose {
		i.lock.Unlock()
		return
	}
	i.isClose = true
	i.lock.Unlock()

	i.cacher.Stop()
}

func (i *CacheIndexer) Ready() <-chan struct{} {
	return i.cacher.Ready()
}

func NewCacheIndexer(name string, cfg *Config) (indexer *CacheIndexer) {
	indexer = &CacheIndexer{
		baseIndexer: NewBaseIndexer(cfg),
		cacher:      NullCacher,
		isClose:     true,
	}

	switch {
	case core.ServerInfo.Config.EnableCache && cfg.InitSize > 0:
		indexer.cacher = NewKvCacher(name, cfg)
	default:
		util.Logger().Infof("service center will not cache '%s'", name)
	}
	return
}

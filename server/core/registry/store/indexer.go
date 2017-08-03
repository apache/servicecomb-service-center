//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"errors"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
)

type Indexer interface {
	Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error)
	KeepAlive(ctx context.Context, op *registry.PluginOp) (ttl int64, err error)
}

type KvCacheIndexer struct {
	cache       Cache
	localCache  Cache             // TODO
	prefixIndex map[string]string // TODO
	asyncs      map[int64]AsyncTasker
	lock        sync.Mutex
}

func (i *KvCacheIndexer) Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	if op.Action != registry.GET {
		return nil, errors.New("unexpected action")
	}

	if op.WithNoCache || op.WithRev > 0 {
		util.LOGGER.Debugf("match WitchNoCache or WitchRev, request etcd server, op: %s", op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	key := registry.BytesToStringWithNoCopy(op.Key)
	resp := &registry.PluginResponse{
		Action:    op.Action,
		Kvs:       []*mvccpb.KeyValue{},
		Count:     0,
		Revision:  i.cache.Version(),
		Succeeded: true,
	}
	if op.CountOnly {
		if i.cache.Have(key) {
			resp.Count = 1
			return resp, nil
		}
		util.LOGGER.Debugf("cache does not store this key, request etcd server, op: %s", op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	cacheData := i.cache.Data(key)
	if cacheData == nil {
		util.LOGGER.Debugf("do not match any key in cache store, request etcd server, op: %s", op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	resp.Count = 1
	resp.Kvs = []*mvccpb.KeyValue{cacheData.(*mvccpb.KeyValue)}
	return resp, nil
}

func (i *KvCacheIndexer) KeepAlive(ctx context.Context, op *registry.PluginOp) (ttl int64, err error) {
	t := NewLeaseAsyncTask(op.Lease)
	if op.WithNoCache {
		util.LOGGER.Debugf("keep alive lease WitchNoCache, request etcd server, op: %s", op)
		err = t.Do(ctx)
		ttl = t.TTL
		return
	}

	i.lock.Lock()
	at, ok := i.asyncs[op.Lease]
	if !ok {
		at = NewLeaseAsyncTasker(registry.BytesToStringWithNoCopy(op.Key))
		at.Run()
		i.asyncs[op.Lease] = at
	}
	i.lock.Unlock()

	err = at.AddTask(ctx, t)
	ttl = at.(*LeaseAsyncTasker).TTL
	return
}

func (i *KvCacheIndexer) RemoveAsyncLeaseTasker(leaseID int64) {
	i.lock.Lock()
	at, ok := i.asyncs[leaseID]
	if ok {
		at.Stop()
		delete(i.asyncs, leaseID)
	}
	i.lock.Unlock()
}

func NewKvCacheIndexer(c Cache) *KvCacheIndexer {
	return &KvCacheIndexer{
		cache:  c,
		asyncs: make(map[int64]AsyncTasker),
	}
}

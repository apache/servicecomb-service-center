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
package store

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_EVENT_COUNT   = 1000
	DEFAULT_ADD_QUEUE_TIMEOUT = 5 * time.Second
)

var defaultRootKeys map[string]struct{}

func init() {
	defaultRootKeys = make(map[string]struct{}, len(defaultRootKeys))
	for _, root := range TypeRoots {
		defaultRootKeys[root] = struct{}{}
	}
}

type Indexer struct {
	BuildTimeout     time.Duration
	cacher           Cacher
	cacheType        StoreType
	prefixIndex      map[string]map[string]struct{}
	prefixLock       sync.RWMutex
	prefixBuildQueue chan *KvEvent
	goroutine        *util.GoRoutine
	ready            chan struct{}
	isClose          bool
}

func (i *Indexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*registry.PluginResponse, error) {
	op := registry.OpGet(opts...)

	key := util.BytesToStringWithNoCopy(op.Key)

	if op.Mode == registry.MODE_NO_CACHE ||
		op.Revision > 0 ||
		(op.Offset >= 0 && op.Limit > 0) {
		util.Logger().Debugf("search %s match special options, request etcd server, opts: %s",
			i.cacheType, op)
		return backend.Registry().Do(ctx, opts...)
	}

	if op.Prefix {
		resp, err := i.searchPrefixKeyWithCache(ctx, op)
		if err != nil {
			return nil, err
		}

		if len(resp.Kvs) > 0 || op.Mode == registry.MODE_CACHE {
			return resp, nil
		}

		util.Logger().Debugf("can not find any key from %s cache with prefix, request etcd server, key: %s",
			i.cacheType, key)
		return backend.Registry().Do(ctx, opts...)
	}

	resp := &registry.PluginResponse{
		Action:    op.Action,
		Count:     0,
		Revision:  i.Cache().Version(),
		Succeeded: true,
	}

	if op.CountOnly {
		if i.Cache().Have(key) {
			resp.Count = 1
			return resp, nil
		}
		if op.Mode == registry.MODE_CACHE {
			return resp, nil
		}

		util.Logger().Debugf("%s cache does not store this key, request etcd server, key: %s", i.cacheType, key)
		return backend.Registry().Do(ctx, opts...)
	}

	cacheData := i.Cache().Data(key)
	if cacheData == nil {
		if op.Mode == registry.MODE_CACHE {
			return resp, nil
		}

		util.Logger().Debugf("do not match any key in %s cache store, request etcd server, key: %s",
			i.cacheType, key)
		return backend.Registry().Do(ctx, opts...)
	}

	resp.Count = 1
	resp.Kvs = []*mvccpb.KeyValue{cacheData.(*mvccpb.KeyValue)}
	return resp, nil
}

func (i *Indexer) Cache() Cache {
	return i.cacher.Cache()
}

func (i *Indexer) searchPrefixKeyWithCache(ctx context.Context, op registry.PluginOp) (*registry.PluginResponse, error) {
	resp := &registry.PluginResponse{
		Action:    op.Action,
		Kvs:       []*mvccpb.KeyValue{},
		Count:     0,
		Revision:  i.Cache().Version(),
		Succeeded: true,
	}

	prefix := util.BytesToStringWithNoCopy(op.Key)

	i.prefixLock.RLock()
	resp.Count = int64(i.getPrefixKey(nil, prefix))
	if resp.Count == 0 || op.CountOnly {
		i.prefixLock.RUnlock()
		return resp, nil
	}

	t := time.Now()
	keys := make([]string, 0, resp.Count)
	i.getPrefixKey(&keys, prefix)
	i.prefixLock.RUnlock()

	kvs := make([]*mvccpb.KeyValue, resp.Count)
	idx := 0
	for _, key := range keys {
		c := i.Cache().Data(key) // TODO too slow when big data is requested
		if c == nil {
			// it means resp.Count is not equal to len(keys)
			util.Logger().Warnf(nil, "unexpected nil cache, maybe it is removed, key is %s", key)
			continue
		}
		kvs[idx] = c.(*mvccpb.KeyValue)
		idx++
	}
	util.LogNilOrWarnf(t, "too long to copy data[%d] from cache with prefix %s", idx, prefix)

	resp.Kvs = kvs[:idx]
	return resp, nil
}

func (i *Indexer) OnCacheEvent(evt *KvEvent) {
	switch evt.Action {
	case pb.EVT_INIT, pb.EVT_CREATE, pb.EVT_DELETE:
	default:
		return
	}

	if i.isClose {
		return
	}
	defer util.RecoverAndReport()

	ctx, _ := context.WithTimeout(context.Background(), i.BuildTimeout)
	select {
	case <-ctx.Done():
		key := util.BytesToStringWithNoCopy(evt.KV.Key)
		util.Logger().Warnf(nil, "add event to build index queue timed out(%s), key is %s [%s] event",
			i.BuildTimeout, key, evt.Action)
	case i.prefixBuildQueue <- evt:
	}
}

func (i *Indexer) buildIndex() {
	i.goroutine.Do(func(stopCh <-chan struct{}) {
		util.SafeCloseChan(i.ready)
		for {
			select {
			case <-stopCh:
				return
			case evt, ok := <-i.prefixBuildQueue:
				if !ok {
					return
				}
				key := util.BytesToStringWithNoCopy(evt.KV.Key)
				prefix := key[:strings.LastIndex(key[:len(key)-1], "/")+1]

				i.prefixLock.Lock()
				switch evt.Action {
				case pb.EVT_DELETE:
					i.deletePrefixKey(prefix, key)
				default:
					i.addPrefixKey(prefix, key)
				}
				i.prefixLock.Unlock()

			}
		}
		util.Logger().Debugf("build %s index goroutine is stopped", i.cacheType)
	})
}

func (i *Indexer) getPrefixKey(arr *[]string, prefix string) (count int) {
	keysRef, ok := i.prefixIndex[prefix]
	if !ok {
		return 0
	}

	for key := range keysRef {
		var childs *[]string = nil
		if arr != nil {
			childs = &[]string{}
		}
		n := i.getPrefixKey(childs, key)
		if n == 0 {
			count += len(keysRef)
			if arr != nil {
				for k := range keysRef {
					*arr = append(*arr, k)
				}
			}
			break
		}
		count += n
		if arr != nil {
			*arr = append(*arr, *childs...)
		}
	}
	return count
}

func (i *Indexer) addPrefixKey(prefix, key string) {
	_, ok := defaultRootKeys[key]
	if ok {
		return
	}

	keys, ok := i.prefixIndex[prefix]
	if !ok {
		keys = make(map[string]struct{})
		i.prefixIndex[prefix] = keys
	} else if _, ok := keys[key]; ok {
		return
	}

	keys[key], key = struct{}{}, prefix
	prefix = key[:strings.LastIndex(key[:len(key)-1], "/")+1]

	i.addPrefixKey(prefix, key)
}

func (i *Indexer) deletePrefixKey(prefix, key string) {
	m, ok := i.prefixIndex[prefix]
	if !ok {
		return
	}

	for k := range i.prefixIndex[key] {
		i.deletePrefixKey(key, k)
	}
	delete(m, key)
}

func (i *Indexer) Run() {
	i.prefixLock.Lock()
	if !i.isClose {
		i.prefixLock.Unlock()
		return
	}
	i.isClose = false
	i.prefixLock.Unlock()

	i.buildIndex()

	i.cacher.Run()
}

func (i *Indexer) Stop() {
	i.prefixLock.Lock()
	if i.isClose {
		i.prefixLock.Unlock()
		return
	}
	i.isClose = true
	i.prefixLock.Unlock()

	i.cacher.Stop()

	i.goroutine.Close(true)

	close(i.prefixBuildQueue)

	util.SafeCloseChan(i.ready)
}

func (i *Indexer) Ready() <-chan struct{} {
	<-i.cacher.Ready()
	return i.ready
}

func NewCacheIndexer(t StoreType, cr Cacher) *Indexer {
	return &Indexer{
		BuildTimeout:     DEFAULT_ADD_QUEUE_TIMEOUT,
		cacher:           cr,
		cacheType:        t,
		prefixIndex:      make(map[string]map[string]struct{}, DEFAULT_MAX_EVENT_COUNT),
		prefixBuildQueue: make(chan *KvEvent, DEFAULT_MAX_EVENT_COUNT),
		goroutine:        util.NewGo(make(chan struct{})),
		ready:            make(chan struct{}),
		isClose:          true,
	}
}

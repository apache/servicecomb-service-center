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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
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

type Indexer interface {
	Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error)
	Run()
	Stop()
	Ready() <-chan struct{}
}

type KvCacheIndexer struct {
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

func (i *KvCacheIndexer) Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	if op.Action != registry.GET {
		return nil, errors.New("unexpected action")
	}

	if op.WithNoCache || op.WithRev > 0 {
		util.LOGGER.Debugf("search %s match WitchNoCache or WitchRev, request etcd server, op: %s",
			i.cacheType, op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	key := registry.BytesToStringWithNoCopy(op.Key)

	if op.WithPrefix {
		return i.searchPrefixKeyFromCacheOrRemote(ctx, op)
	}

	resp := &registry.PluginResponse{
		Action:    op.Action,
		Kvs:       []*mvccpb.KeyValue{},
		Count:     0,
		Revision:  i.Cache().Version(),
		Succeeded: true,
	}

	if op.CountOnly {
		if i.Cache().Have(key) {
			resp.Count = 1
			return resp, nil
		}
		util.LOGGER.Debugf("%s cache does not store this key, request etcd server, op: %s", i.cacheType, op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	cacheData := i.Cache().Data(key)
	if cacheData == nil {
		util.LOGGER.Debugf("do not match any key in %s cache store, request etcd server, op: %s",
			i.cacheType, op)
		return registry.GetRegisterCenter().Do(ctx, op)
	}

	resp.Count = 1
	resp.Kvs = []*mvccpb.KeyValue{cacheData.(*mvccpb.KeyValue)}
	return resp, nil
}

func (i *KvCacheIndexer) Cache() Cache {
	return i.cacher.Cache()
}

func (i *KvCacheIndexer) searchPrefixKeyFromCacheOrRemote(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	resp := &registry.PluginResponse{
		Action:    op.Action,
		Kvs:       []*mvccpb.KeyValue{},
		Count:     0,
		Revision:  i.Cache().Version(),
		Succeeded: true,
	}

	prefix := registry.BytesToStringWithNoCopy(op.Key)

	i.prefixLock.RLock()
	keys, ok := i.prefixIndex[prefix]
	i.prefixLock.RUnlock()
	if !ok {
		util.LOGGER.Debugf("can not find any key from %s cache with prefix, request etcd server, op: %s",
			i.cacheType, op)
		prefixResp, err := registry.GetRegisterCenter().Do(ctx, op)
		if err != nil {
			return nil, err
		}

		resp.Kvs = prefixResp.Kvs
		resp.Count = prefixResp.Count
		resp.Revision = prefixResp.Revision
		return resp, nil
	}

	resp.Count = int64(len(keys))
	if op.CountOnly {
		return resp, nil
	}

	t := time.Now()
	kvs := make([]*mvccpb.KeyValue, len(keys))
	idx := 0
	for key := range keys {
		c := i.Cache().Data(key) // TODO too slow when big data is requested
		if c == nil {
			// it means resp.Count is not equal to len(keys)
			util.LOGGER.Debugf("unexpected nil cache, maybe it is removed, key is %s", key)
			continue
		}
		kvs[idx] = c.(*mvccpb.KeyValue)
		idx++
	}
	if time.Now().Sub(t) > time.Second {
		util.LOGGER.Warnf(nil, "too long(%s vs 1s) to copy data from cache", time.Now().Sub(t))
	}
	resp.Kvs = kvs[:idx]
	return resp, nil
}

func (i *KvCacheIndexer) onCacheEvent(evt *KvEvent) {
	if evt.Action != pb.EVT_DELETE && evt.Action != pb.EVT_CREATE {
		return
	}

	i.prefixLock.RLock()

	if i.isClose {
		i.prefixLock.RUnlock()
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), i.BuildTimeout)
	select {
	case <-ctx.Done():
		key := registry.BytesToStringWithNoCopy(evt.KV.Key)
		util.LOGGER.Warnf(nil, "add event to build index queue timed out(%s), key is %s [%s] event",
			i.BuildTimeout, key, evt.Action)
	case i.prefixBuildQueue <- evt:
	}

	i.prefixLock.RUnlock()
}

func (i *KvCacheIndexer) buildIndex() {
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
				key := registry.BytesToStringWithNoCopy(evt.KV.Key)

				i.prefixLock.Lock()
				prefix := key[:strings.LastIndex(key, "/")+1]
				keys, ok := i.prefixIndex[prefix]
				if !ok {
					keys = make(map[string]struct{})
					i.prefixIndex[prefix] = keys
				}
				switch evt.Action {
				case pb.EVT_CREATE:
					keys[key] = struct{}{}
				case pb.EVT_DELETE:
					delete(keys, key)
				}
				i.prefixLock.Unlock()

			}
		}
		util.LOGGER.Debugf("build %s index goroutine is stopped", i.cacheType)
	})
}

func (i *KvCacheIndexer) Run() {
	i.prefixLock.Lock()
	if !i.isClose {
		i.prefixLock.Unlock()
		return
	}
	i.isClose = false
	i.prefixLock.Unlock()

	AddKvStoreEventFunc(i.cacheType, i.onCacheEvent)

	i.buildIndex()

	i.cacher.Run()
}

func (i *KvCacheIndexer) Stop() {
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

	util.LOGGER.Debugf("%s indexer is stopped", i.cacheType)
}

func (i *KvCacheIndexer) Ready() <-chan struct{} {
	<-i.cacher.Ready()
	return i.ready
}

func NewKvCacheIndexer(t StoreType, cr Cacher) *KvCacheIndexer {
	return &KvCacheIndexer{
		BuildTimeout:     DEFAULT_ADD_QUEUE_TIMEOUT,
		cacher:           cr,
		cacheType:        t,
		prefixIndex:      make(map[string]map[string]struct{}),
		prefixBuildQueue: make(chan *KvEvent, DEFAULT_MAX_EVENT_COUNT),
		goroutine:        util.NewGo(make(chan struct{})),
		ready:            make(chan struct{}),
		isClose:          true,
	}
}

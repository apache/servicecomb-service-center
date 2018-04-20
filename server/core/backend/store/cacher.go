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
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	DEFAULT_COMPACT_TIMES   = 3
	DEFAULT_COMPACT_TIMEOUT = 5 * time.Minute
	event_block_size        = 1000
)

var (
	NullCache  = &nullCache{}
	NullCacher = &nullCacher{}
)

type Cache interface {
	Version() int64
	Data(interface{}) interface{}
	Have(interface{}) bool
	Size() int
}

type Cacher interface {
	Cache() Cache
	Run()
	Stop()
	Ready() <-chan struct{}
}

type nullCache struct {
}

func (n *nullCache) Version() int64 {
	return 0
}

func (n *nullCache) Data(interface{}) interface{} {
	return nil
}

func (n *nullCache) Have(interface{}) bool {
	return false
}

func (n *nullCache) Size() int {
	return 0
}

type nullCacher struct {
}

func (n *nullCacher) Cache() Cache {
	return NullCache
}

func (n *nullCacher) Run() {}

func (n *nullCacher) Stop() {}

func (n *nullCacher) Ready() <-chan struct{} {
	c := make(chan struct{})
	close(c)
	return c
}

type KvCacheSafeRFunc func()

type KvCache struct {
	owner       *KvCacher
	size        int
	store       map[string]*mvccpb.KeyValue
	rwMux       sync.RWMutex
	lastRefresh time.Time
	lastMaxSize int
}

func (c *KvCache) Version() int64 {
	return c.owner.lw.Revision()
}

func (c *KvCache) Data(k interface{}) interface{} {
	c.rwMux.RLock()
	kv, ok := c.store[k.(string)]
	c.rwMux.RUnlock()
	if !ok {
		return nil
	}
	copied := *kv
	return &copied
}

func (c *KvCache) Have(k interface{}) (ok bool) {
	c.rwMux.RLock()
	_, ok = c.store[k.(string)]
	c.rwMux.RUnlock()
	return
}

func (c *KvCache) RLock() map[string]*mvccpb.KeyValue {
	c.rwMux.RLock()
	return c.store
}

func (c *KvCache) RUnlock() {
	c.rwMux.RUnlock()
}

func (c *KvCache) Lock() map[string]*mvccpb.KeyValue {
	c.rwMux.Lock()
	return c.store
}

func (c *KvCache) Unlock() {
	l := len(c.store)
	if l > c.lastMaxSize {
		c.lastMaxSize = l
	}
	if c.size >= l &&
		c.lastMaxSize > c.size*DEFAULT_COMPACT_TIMES &&
		time.Now().Sub(c.lastRefresh) >= DEFAULT_COMPACT_TIMEOUT {
		util.Logger().Infof("cache %s is not in use over %s, compact capacity to size %d->%d",
			c.owner.Cfg.Key, DEFAULT_COMPACT_TIMEOUT, c.lastMaxSize, c.size)
		// gc
		newCache := make(map[string]*mvccpb.KeyValue, c.size)
		for k, v := range c.store {
			newCache[k] = v
		}
		c.store = newCache
		c.lastMaxSize = l
		c.lastRefresh = time.Now()
	}
	c.rwMux.Unlock()
}

func (c *KvCache) Size() (l int) {
	c.rwMux.RLock()
	l = len(c.store)
	c.rwMux.RUnlock()
	return
}

type KvCacher struct {
	Cfg KvCacherCfg

	lastRev      int64
	noEventCount int

	ready     chan struct{}
	lw        ListWatcher
	mux       sync.Mutex
	once      sync.Once
	cache     *KvCache
	goroutine *util.GoRoutine
}

func (c *KvCacher) needList() bool {
	rev := c.lw.Revision()
	defer func() { c.lastRev = rev }()

	if rev == 0 {
		c.noEventCount = 0
		return true
	}
	if c.lastRev != rev {
		c.noEventCount = 0
		return false
	}
	c.noEventCount++
	if c.noEventCount < c.Cfg.NoEventMaxInterval {
		return false
	}

	util.Logger().Debugf("no events come in more then %s, need to list key %s, rev: %d",
		time.Duration(c.noEventCount)*c.Cfg.Timeout, c.Cfg.Key, rev)
	c.noEventCount = 0
	return true
}

func (c *KvCacher) doList(listOps ListOptions) error {
	kvs, err := c.lw.List(listOps)
	if err != nil {
		return err
	}

	start := time.Now()
	c.sync(c.filter(c.lw.Revision(), kvs))
	util.LogDebugOrWarnf(start, "finish to cache key %s, %d items, rev: %d", c.Cfg.Key, len(kvs), c.lw.Revision())

	return nil
}

func (c *KvCacher) doWatch(listOps ListOptions) error {
	watcher := c.lw.Watch(listOps)
	return c.handleWatcher(watcher)
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	listOps := ListOptions{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}
	if c.needList() {
		err := c.doList(listOps)

		util.SafeCloseChan(c.ready)

		if err != nil {
			return err
		}
	}
	return c.doWatch(listOps)
}

func (c *KvCacher) handleWatcher(watcher *Watcher) error {
	defer watcher.Stop()
	for evts := range watcher.EventBus() {
		if evts[0].Type == proto.EVT_ERROR {
			err := evts[0].Object.(error)
			return err
		}
		c.sync(evts)
	}
	return nil
}

func (c *KvCacher) needDeferHandle(evts []Event) bool {
	if c.Cfg.DeferHandler == nil || !c.IsReady() {
		return false
	}

	return c.Cfg.DeferHandler.OnCondition(c.Cache(), evts)
}

func (c *KvCacher) refresh(ctx context.Context) {
	util.Logger().Debugf("start to list and watch %s", c.Cfg)
	for {
		start := time.Now()
		c.ListAndWatch(ctx)
		watchDuration := time.Since(start)
		nextPeriod := c.Cfg.Period
		if watchDuration > 0 && c.Cfg.Period > watchDuration {
			nextPeriod = c.Cfg.Period - watchDuration
		}
		select {
		case <-ctx.Done():
			util.Logger().Debugf("stop to list and watch %s", c.Cfg)
			return
		case <-time.After(nextPeriod):
		}
	}
}

func (c *KvCacher) deferHandle(ctx context.Context) {
	if c.Cfg.DeferHandler == nil {
		return
	}

	i, evts := 0, make([]Event, event_block_size)
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-c.Cfg.DeferHandler.HandleChan():
			if !ok {
				<-time.After(time.Second)
				continue
			}

			if i >= event_block_size {
				c.onEvents(evts[:i])
				i = 0
			}

			evts[i] = evt
			i++
		case <-time.After(300 * time.Millisecond):
			if i == 0 {
				continue
			}

			c.onEvents(evts[:i])
			i = 0
		}
	}
}

func (c *KvCacher) sync(evts []Event) {
	if len(evts) == 0 {
		return
	}

	if c.needDeferHandle(evts) {
		return
	}

	c.onEvents(evts)
}

func (c *KvCacher) filter(rev int64, items []*mvccpb.KeyValue) []Event {
	store := c.cache.RLock()
	defer c.cache.RUnlock()

	oc, nc := len(store), len(items)
	tc := oc + nc
	if tc == 0 {
		return nil
	}
	max := oc
	if nc > oc {
		max = nc
	}

	newStore := make(map[string]*mvccpb.KeyValue, nc)
	for _, kv := range items {
		newStore[util.BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan []Event, max/event_block_size+2)

	go c.filterDelete(store, newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(store, newStore, rev, eventsCh, filterStopCh)

	evts := make([]Event, 0, max)
	for block := range eventsCh {
		evts = append(evts, block...)
	}
	return evts
}

func (c *KvCacher) filterDelete(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan []Event, filterStopCh chan struct{}) {
	var block [event_block_size]Event
	i := 0
	for k, v := range store {
		_, ok := newStore[k]
		if ok {
			continue
		}

		if i >= event_block_size {
			eventsCh <- block[:i]
			i = 0
		}

		block[i] = Event{
			Revision: rev,
			Type:     proto.EVT_DELETE,
			Prefix:   c.Cfg.Key,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block[:i]
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan []Event, filterStopCh chan struct{}) {
	var block [event_block_size]Event
	i := 0
	for k, v := range newStore {
		ov, ok := store[k]
		if !ok {
			if i >= event_block_size {
				eventsCh <- block[:i]
				i = 0
			}

			block[i] = Event{
				Revision: rev,
				Type:     proto.EVT_CREATE,
				Prefix:   c.Cfg.Key,
				Object:   v,
			}
			i++
			continue
		}

		if ov.CreateRevision == v.CreateRevision && ov.ModRevision == v.ModRevision {
			continue
		}

		if i >= event_block_size {
			eventsCh <- block[:i]
			i = 0
		}

		block[i] = Event{
			Revision: rev,
			Type:     proto.EVT_UPDATE,
			Prefix:   c.Cfg.Key,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block[:i]
	}

	<-filterStopCh

	close(eventsCh)
}

func (c *KvCacher) onEvents(evts []Event) {
	idx, init := 0, !c.IsReady()
	kvEvts := make([]KvEvent, len(evts))
	store := c.cache.Lock()
	for _, evt := range evts {
		kv := evt.Object.(*mvccpb.KeyValue)
		key := util.BytesToStringWithNoCopy(kv.Key)
		prevKv, ok := store[key]

		switch evt.Type {
		case proto.EVT_CREATE, proto.EVT_UPDATE:
			t := evt.Type

			switch {
			case init:
				t = proto.EVT_INIT
			case !ok && evt.Type != proto.EVT_CREATE:
				util.Logger().Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				t = proto.EVT_CREATE
			case ok && evt.Type != proto.EVT_UPDATE:
				util.Logger().Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				t = proto.EVT_UPDATE
			}

			store[key] = kv
			kvEvts[idx] = KvEvent{
				Revision: evt.Revision,
				Action:   t,
				KV:       kv,
			}
		case proto.EVT_DELETE:
			if !ok {
				util.Logger().Warnf(nil, "unexpected %s event! key %s does not exist", evt.Type, key)
				continue
			}

			delete(store, key)
			kvEvts[idx] = KvEvent{
				Revision: evt.Revision,
				Action:   evt.Type,
				KV:       prevKv,
			}
		}

		idx++
	}
	c.cache.Unlock()

	c.onKvEvents(kvEvts[:idx])
}

func (c *KvCacher) onKvEvents(evts []KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}
	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
}

func (c *KvCacher) Cache() Cache {
	return c.cache
}

func (c *KvCacher) Run() {
	c.once.Do(func() {
		c.goroutine.Do(c.refresh)
		c.goroutine.Do(c.deferHandle)
	})
}

func (c *KvCacher) Stop() {
	c.goroutine.Close(true)

	util.SafeCloseChan(c.ready)
}

func (c *KvCacher) Ready() <-chan struct{} {
	return c.ready
}

func (c *KvCacher) IsReady() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

func NewKvCache(c *KvCacher, size int) *KvCache {
	return &KvCache{
		owner:       c,
		size:        size,
		store:       make(map[string]*mvccpb.KeyValue, size),
		lastRefresh: time.Now(),
	}
}

func NewKvCacher(opts ...KvCacherCfgOption) *KvCacher {
	cfg := DefaultKvCacherConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	cacher := &KvCacher{
		Cfg:   cfg,
		ready: make(chan struct{}),
		lw: ListWatcher{
			Client: backend.Registry(),
			Key:    cfg.Key,
		},
		goroutine: util.NewGo(context.Background()),
	}
	cacher.cache = NewKvCache(cacher, cfg.InitSize)
	return cacher
}

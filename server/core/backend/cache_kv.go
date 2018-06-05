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
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

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
	return kv
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
		c.compact()
		c.lastMaxSize = l
		c.lastRefresh = time.Now()
	}

	c.rwMux.Unlock()
}

func (c *KvCache) compact() {
	// gc
	newCache := make(map[string]*mvccpb.KeyValue, c.size)
	for k, v := range c.store {
		newCache[k] = v
	}
	c.store = newCache

	util.Logger().Infof("cache %s is not in use over %s, compact capacity to size %d->%d",
		c.owner.Cfg.Prefix, DEFAULT_COMPACT_TIMEOUT, c.lastMaxSize, c.size)
}

func (c *KvCache) Size() (l int) {
	c.rwMux.RLock()
	l = len(c.store)
	c.rwMux.RUnlock()
	return
}

type KvCacher struct {
	Cfg Config

	name         string
	lastRev      int64
	noEventCount int

	ready     chan struct{}
	lw        ListWatcher
	mux       sync.Mutex
	once      sync.Once
	cache     *KvCache
	goroutine *util.GoRoutine
}

func (c *KvCacher) Name() string {
	return c.name
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
		time.Duration(c.noEventCount)*c.Cfg.Timeout, c.Cfg.Prefix, rev)
	c.noEventCount = 0
	return true
}

func (c *KvCacher) doList(cfg ListWatchConfig) error {
	kvs, err := c.lw.List(cfg)
	if err != nil {
		return err
	}

	start := time.Now()
	evts := c.filter(c.lw.Revision(), kvs)
	if ec, kc := len(evts), len(kvs); c.Cfg.DeferHandler != nil && ec == 0 && kc != 0 &&
		c.Cfg.DeferHandler.Reset() {
		util.Logger().Warnf(nil, "most of the protected data(%d/%d) are recovered", kc, c.cache.Size())
	}
	c.sync(evts)
	util.LogDebugOrWarnf(start, "finish to cache key %s, %d items, rev: %d",
		c.Cfg.Prefix, len(kvs), c.lw.Revision())

	return nil
}

func (c *KvCacher) doWatch(cfg ListWatchConfig) error {
	watcher := c.lw.Watch(cfg)
	return c.handleWatcher(watcher)
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	cfg := ListWatchConfig{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}
	if c.needList() {
		c.doList(cfg)
	}

	util.SafeCloseChan(c.ready)

	return c.doWatch(cfg)
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

func (c *KvCacher) needDeferHandle(evts []KvEvent) bool {
	if c.Cfg.DeferHandler == nil || !c.IsReady() {
		return false
	}

	return c.Cfg.DeferHandler.OnCondition(c.Cache(), evts)
}

func (c *KvCacher) refresh(ctx context.Context) {
	util.Logger().Debugf("start to list and watch %s", c.Cfg)
	retries := 0
	ticker := time.NewTicker(DEFAULT_METRICS_INTERVAL)
	for {
		nextPeriod := minWaitInterval
		if err := c.ListAndWatch(ctx); err != nil {
			nextPeriod = util.GetBackoff().Delay(retries)
			retries++
		} else {
			retries = 0
		}
		select {
		case <-ctx.Done():
			util.Logger().Debugf("stop to list and watch %s", c.Cfg)
			return
		case <-ticker.C:
			ReportCacheMetrics(c.Name(), "raw", c.cache.RLock())
			c.cache.RUnlock()
		case <-time.After(nextPeriod):
		}
	}
}

func (c *KvCacher) deferHandle(ctx context.Context) {
	if c.Cfg.DeferHandler == nil {
		return
	}

	var (
		evts = make([]KvEvent, eventBlockSize)
		i    int
	)
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-c.Cfg.DeferHandler.HandleChan():
			if !ok {
				<-time.After(time.Second)
				continue
			}

			if i >= eventBlockSize {
				c.onEvents(evts[:i])
				evts = make([]KvEvent, eventBlockSize)
				i = 0
			}

			evts[i] = evt
			i++
		case <-time.After(300 * time.Millisecond):
			if i == 0 {
				continue
			}

			c.onEvents(evts[:i])
			evts = make([]KvEvent, eventBlockSize)
			i = 0
		}
	}
}

func (c *KvCacher) sync(evts []KvEvent) {
	if len(evts) == 0 {
		return
	}

	if c.needDeferHandle(evts) {
		return
	}

	c.onEvents(evts)
}

func (c *KvCacher) filter(rev int64, items []*mvccpb.KeyValue) []KvEvent {
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
	eventsCh := make(chan [eventBlockSize]KvEvent, 2)

	go c.filterDelete(store, newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(store, newStore, rev, eventsCh, filterStopCh)

	evts := make([]KvEvent, 0, max)
	for block := range eventsCh {
		for _, e := range block {
			if e.Object == nil {
				break
			}
			evts = append(evts, e)
		}
	}
	return evts
}

func (c *KvCacher) filterDelete(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan [eventBlockSize]KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]KvEvent
	i := 0
	for k, v := range store {
		_, ok := newStore[k]
		if ok {
			continue
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]KvEvent{}
			i = 0
		}

		block[i] = KvEvent{
			Revision: rev,
			Type:     proto.EVT_DELETE,
			Prefix:   c.Cfg.Prefix,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan [eventBlockSize]KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]KvEvent
	i := 0
	for k, v := range newStore {
		ov, ok := store[k]
		if !ok {
			if i >= eventBlockSize {
				eventsCh <- block
				block = [eventBlockSize]KvEvent{}
				i = 0
			}

			block[i] = KvEvent{
				Revision: rev,
				Type:     proto.EVT_CREATE,
				Prefix:   c.Cfg.Prefix,
				Object:   v,
			}
			i++
			continue
		}

		if ov.CreateRevision == v.CreateRevision && ov.ModRevision == v.ModRevision {
			continue
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]KvEvent{}
			i = 0
		}

		block[i] = KvEvent{
			Revision: rev,
			Type:     proto.EVT_UPDATE,
			Prefix:   c.Cfg.Prefix,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block
	}

	<-filterStopCh

	close(eventsCh)
}

func (c *KvCacher) onEvents(evts []KvEvent) {
	init := !c.IsReady()
	store := c.cache.Lock()
	for i, evt := range evts {
		kv := evt.Object.(*mvccpb.KeyValue)
		key := util.BytesToStringWithNoCopy(kv.Key)
		prevKv, ok := store[key]

		switch evt.Type {
		case proto.EVT_CREATE, proto.EVT_UPDATE:
			switch {
			case init:
				evt.Type = proto.EVT_INIT
			case !ok && evt.Type != proto.EVT_CREATE:
				util.Logger().Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				evt.Type = proto.EVT_CREATE
			case ok && evt.Type != proto.EVT_UPDATE:
				util.Logger().Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				evt.Type = proto.EVT_UPDATE
			}

			store[key] = kv
			evts[i] = evt
		case proto.EVT_DELETE:
			if !ok {
				util.Logger().Warnf(nil, "unexpected %s event! key %s does not cache", evt.Type, key)
			} else {
				delete(store, key)
			}
			evt.Object = prevKv // maybe nil
			evts[i] = evt
		}
	}

	c.cache.Unlock()

	c.onKvEvents(evts)
}

func (c *KvCacher) onKvEvents(evts []KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}
	for _, evt := range evts {
		if evt.Object == nil {
			continue
		}
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

func NewKvCacher(name string, opts ...ConfigOption) *KvCacher {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	cacher := &KvCacher{
		name:  name,
		Cfg:   cfg,
		ready: make(chan struct{}),
		lw: ListWatcher{
			Client: Registry(),
			Prefix: cfg.Prefix,
		},
		goroutine: util.NewGo(context.Background()),
	}
	cacher.cache = NewKvCache(cacher, cfg.InitSize)
	return cacher
}

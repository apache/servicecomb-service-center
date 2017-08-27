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
	"fmt"
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_NO_EVENT_INTERVAL = 1 // TODO it should be set to 1 for prevent etcd data is lost accidentally.
	DEFAULT_LISTWATCH_TIMEOUT     = 30 * time.Second
	DEFAULT_COMPACT_TIMES         = 3
	DEFAULT_COMPACT_TIMEOUT       = 5 * time.Minute
	event_block_size              = 1000
)

var (
	NullCache  = &nullCache{}
	NullCacher = &nullCacher{}
)

type Cache interface {
	Version() int64
	Data(interface{}) interface{}
	Have(interface{}) bool
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
		util.LOGGER.Infof("cache is empty and not in use over %s, compact capacity to size %d->%d",
			DEFAULT_COMPACT_TIMEOUT, c.lastMaxSize, c.size)
		// gc
		newCache := make(map[string]*mvccpb.KeyValue, c.size)
		for k, v := range c.store {
			newCache[k] = v
		}
		c.store = newCache
		c.lastMaxSize = c.size
		c.lastRefresh = time.Now()
	}
	c.rwMux.Unlock()
}

type KvCacher struct {
	Cfg *KvCacherConfig

	lastRev            int64
	noEventInterval    int
	noEventMaxInterval int

	ready   chan struct{}
	lw      ListWatcher
	mux     sync.Mutex
	once    sync.Once
	cache   *KvCache
	goroute *util.GoRoutine
}

func (c *KvCacher) needList() bool {
	rev := c.lw.Revision()
	defer func() { c.lastRev = rev }()

	if rev == 0 {
		c.noEventInterval = 0
		return true
	}
	if c.lastRev != rev {
		c.noEventInterval = 0
		return false
	}
	c.noEventInterval++
	if c.noEventInterval < c.noEventMaxInterval {
		return false
	}

	util.LOGGER.Debugf("no events come in more then %s, need to list key %s",
		time.Duration(c.noEventInterval)*c.Cfg.Timeout, c.Cfg.Key)
	c.noEventInterval = 0
	return true
}

func (c *KvCacher) doList(listOps *ListOptions) error {
	start := time.Now()
	kvs, err := c.lw.List(listOps)
	if err != nil {
		return err
	}
	lastRev := c.lastRev
	c.lastRev = c.lw.Revision()
	c.sync(c.filter(c.lastRev, kvs))
	syncDuration := time.Now().Sub(start)

	if syncDuration > 5*time.Second {
		util.LOGGER.Warnf(nil, "finish to cache key %s, %d items took %s! opts: %s, rev: %d",
			c.Cfg.Key, len(kvs), syncDuration, listOps, c.lastRev)
		return nil
	}
	if lastRev != c.lastRev {
		util.LOGGER.Infof("finish to cache key %s, %d items took %s, opts: %s, rev: %d",
			c.Cfg.Key, len(kvs), syncDuration, listOps, c.lastRev)
	}
	return nil
}

func (c *KvCacher) doWatch(listOps *ListOptions) error {
	watcher := c.lw.Watch(listOps)
	util.LOGGER.Debugf("finish to new watcher, key %s, opts: %s, start rev: %d+1", c.Cfg.Key, listOps, c.lastRev)
	return c.handleWatcher(watcher)
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()

	listOps := &ListOptions{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}
	if c.needList() {
		err := c.doList(listOps)
		if err != nil {
			util.LOGGER.Errorf(err, "list key %s failed, opts: %s, rev: %d",
				c.Cfg.Key, listOps, c.lastRev)
			// do not return err, continue to watch
		}
		util.SafeCloseChan(c.ready)
	}

	err := c.doWatch(listOps)

	c.mux.Unlock()

	if err != nil {
		util.LOGGER.Errorf(err, "handle watcher failed, watch key %s, opts: %s, start rev: %d+1",
			c.Cfg.Key, listOps, c.lastRev)
		return err
	}
	return nil
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

func (c *KvCacher) sync(evts []*Event) {
	if len(evts) == 0 {
		return
	}

	cache := c.Cache().(*KvCache)
	idx := 0
	kvEvts := make([]*KvEvent, len(evts))
	store := cache.Lock()
	for _, evt := range evts {
		kv := evt.Object.(*mvccpb.KeyValue)
		key := util.BytesToStringWithNoCopy(kv.Key)
		prevKv, ok := store[key]
		switch evt.Type {
		case proto.EVT_CREATE, proto.EVT_UPDATE:
			util.LOGGER.Debugf("sync %s event and notify watcher, cache key %s, %+v", evt.Type, key, kv)
			store[key] = kv
			t := evt.Type
			if !ok && evt.Type != proto.EVT_CREATE {
				util.LOGGER.Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				t = proto.EVT_CREATE
			}
			if ok && evt.Type != proto.EVT_UPDATE {
				util.LOGGER.Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				t = proto.EVT_UPDATE
			}
			kvEvts[idx] = &KvEvent{
				Revision: evt.Revision,
				Action:   t,
				KV:       kv,
			}
			idx++
		case proto.EVT_DELETE:
			if ok {
				util.LOGGER.Debugf("sync %s event and notify watcher, remove key %s, %+v", evt.Type, key, kv)
				delete(store, key)
				kvEvts[idx] = &KvEvent{
					Revision: evt.Revision,
					Action:   evt.Type,
					KV:       prevKv,
				}
				idx++
				continue
			}
			util.LOGGER.Warnf(nil, "unexpected %s event! nonexistent key %s", evt.Type, key)
		}
	}
	cache.Unlock()

	c.onKvEvents(kvEvts[:idx])
}

func (c *KvCacher) filter(rev int64, items []*mvccpb.KeyValue) []*Event {
	cache := c.Cache().(*KvCache)
	store := cache.Lock()
	defer cache.Unlock()
	oc, nc := len(store), len(items)
	tc := oc + nc
	if tc == 0 {
		return nil
	}
	max := oc
	if nc > oc {
		max = nc
	}

	newStore := make(map[string]*mvccpb.KeyValue)
	for _, kv := range items {
		newStore[util.BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan [event_block_size]*Event, max/event_block_size+2)

	go c.filterDelete(store, newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(store, newStore, rev, eventsCh, filterStopCh)

	evts := make([]*Event, 0, max)
	for block := range eventsCh {
		for _, evt := range block {
			if evt == nil {
				break
			}
			evts = append(evts, evt)
		}
	}
	return evts
}

func (c *KvCacher) filterDelete(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue, rev int64, eventsCh chan [event_block_size]*Event, filterStopCh chan struct{}) {
	var block [event_block_size]*Event
	i := 0
	for k, v := range store {
		_, ok := newStore[k]
		if ok {
			continue
		}

		if i >= event_block_size {
			eventsCh <- block
			block = [event_block_size]*Event{}
			i = 0
		}

		block[i] = &Event{
			Revision: rev,
			Type:     proto.EVT_DELETE,
			Key:      c.Cfg.Key,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(store map[string]*mvccpb.KeyValue, newStore map[string]*mvccpb.KeyValue, rev int64, eventsCh chan [event_block_size]*Event, filterStopCh chan struct{}) {
	var block [event_block_size]*Event
	i := 0
	for k, v := range newStore {
		ov, ok := store[k]
		if !ok {
			if i >= event_block_size {
				eventsCh <- block
				block = [event_block_size]*Event{}
				i = 0
			}

			block[i] = &Event{
				Revision: rev,
				Type:     proto.EVT_CREATE,
				Key:      c.Cfg.Key,
				Object:   v,
			}
			i++
			continue
		}

		if ov.ModRevision >= v.ModRevision {
			continue
		}

		if i >= event_block_size {
			eventsCh <- block
			block = [event_block_size]*Event{}
			i = 0
		}

		block[i] = &Event{
			Revision: rev,
			Type:     proto.EVT_UPDATE,
			Key:      c.Cfg.Key,
			Object:   v,
		}
		i++
	}

	if i > 0 {
		eventsCh <- block
	}

	select {
	case <-filterStopCh:
		close(eventsCh)
	}
}

func (c *KvCacher) onKvEvents(evts []*KvEvent) {
	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
}

func (c *KvCacher) run() {
	c.goroute.Do(func(stopCh <-chan struct{}) {
		util.LOGGER.Debugf("start to list and watch %s", c.Cfg)
		ctx, cancel := context.WithCancel(context.Background())
		c.goroute.Do(func(stopCh <-chan struct{}) {
			defer cancel()
			<-stopCh
		})
		for {
			start := time.Now()
			c.ListAndWatch(ctx)
			watchDuration := time.Now().Sub(start)
			nextPeriod := 0 * time.Second
			if watchDuration > 0 && c.Cfg.Period > watchDuration {
				nextPeriod = c.Cfg.Period - watchDuration
			}
			select {
			case <-stopCh:
				util.LOGGER.Warnf(nil, "stop to list and watch %s", c.Cfg)
				return
			case <-time.After(nextPeriod):
			}
		}
	})
}

func (c *KvCacher) Cache() Cache {
	return c.cache
}

func (c *KvCacher) Run() {
	c.once.Do(c.run)
}

func (c *KvCacher) Stop() {
	c.goroute.Close(true)

	util.SafeCloseChan(c.ready)

	util.LOGGER.Debugf("cacher is stopped, %s", c.Cfg)
}

func (c *KvCacher) Ready() <-chan struct{} {
	return c.ready
}

type KvCacherConfig struct {
	Key      string
	InitSize int
	Timeout  time.Duration
	Period   time.Duration
	OnEvent  KvEventFunc
}

func (cfg *KvCacherConfig) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		cfg.Key, cfg.Timeout, cfg.Period)
}

func NewKvCache(c *KvCacher, size int) *KvCache {
	return &KvCache{
		owner:       c,
		size:        size,
		lastMaxSize: size,
		store:       make(map[string]*mvccpb.KeyValue, size),
		lastRefresh: time.Now(),
	}
}

func NewKvCacher(cfg *KvCacherConfig) Cacher {
	cacher := &KvCacher{
		Cfg:   cfg,
		ready: make(chan struct{}),
		lw: ListWatcher{
			Client: registry.GetRegisterCenter(),
			Key:    cfg.Key,
		},
		goroute:            util.NewGo(make(chan struct{})),
		noEventMaxInterval: DEFAULT_MAX_NO_EVENT_INTERVAL,
	}
	cacher.cache = NewKvCache(cacher, cfg.InitSize)
	return cacher
}

func NewCacher(initSize int, prefix string, callback KvEventFunc) Cacher {
	return NewTimeoutCacher(DEFAULT_LISTWATCH_TIMEOUT, initSize, prefix, callback)
}

func NewTimeoutCacher(ot time.Duration, initSize int, prefix string, callback KvEventFunc) Cacher {
	return NewKvCacher(&KvCacherConfig{
		Key:      prefix,
		InitSize: initSize,
		Timeout:  ot,
		Period:   time.Second,
		OnEvent:  callback,
	})
}

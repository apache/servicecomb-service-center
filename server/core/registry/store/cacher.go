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
	DEFAULT_MAX_NO_EVENT_INTERVAL = 4
	DEFAULT_LISTWATCH_TIMEOUT     = 30 * time.Second
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

type KvCacheSafeRFunc func()

type KvCache struct {
	owner *KvCacher
	store map[string]*mvccpb.KeyValue
	rwMux sync.RWMutex
}

func (c *KvCache) Version() int64 {
	return c.owner.lw.Revision()
}

func (c *KvCache) Data(k interface{}) interface{} {
	c.rwMux.RLock()
	defer c.rwMux.RUnlock()
	kv, ok := c.store[k.(string)]
	if !ok {
		return nil
	}
	var copy mvccpb.KeyValue
	util.DeepCopy(&copy, kv)
	return &copy
}

func (c *KvCache) Have(k interface{}) (ok bool) {
	c.rwMux.RLock()
	defer c.rwMux.RUnlock()
	_, ok = c.store[k.(string)]
	return
}

func (c *KvCache) Lock() map[string]*mvccpb.KeyValue {
	c.rwMux.Lock()
	return c.store
}

func (c *KvCache) Unlock() {
	c.rwMux.Unlock()
}

type KvCacher struct {
	Cfg *KvCacherConfig

	lastRev            int64
	noEventInterval    int
	noEventMaxInterval int

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
	c.lastRev = c.lw.Revision()
	c.sync(c.filter(c.lastRev, kvs))
	syncDuration := time.Now().Sub(start)

	if syncDuration > 5*time.Second {
		util.LOGGER.Warnf(nil, "finish to cache key %s, %d items took %s! list options: %+v, rev: %d",
			c.Cfg.Key, len(kvs), syncDuration, listOps, c.lastRev)
		return nil
	}

	util.LOGGER.Infof("finish to cache key %s, %d items took %s, list options: %+v, rev: %d",
		c.Cfg.Key, len(kvs), syncDuration, listOps, c.lastRev)
	return nil
}

func (c *KvCacher) doWatch(listOps *ListOptions) error {
	watcher := c.lw.Watch(listOps)
	util.LOGGER.Debugf("finish to new watcher, key %s, list options: %+v, start rev: %d+1",
		c.Cfg.Key, listOps, c.lastRev)
	return c.handleWatcher(watcher)
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	listOps := &ListOptions{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}
	if c.needList() {
		err := c.doList(listOps)
		if err != nil {
			util.LOGGER.Errorf(err, "list key %s failed, list options: %+v", c.Cfg.Key, listOps)
			return err
		}
	}

	err := c.doWatch(listOps)
	if err != nil {
		util.LOGGER.Errorf(err, "handle watcher failed, watch key %s, list options: %+v", c.Cfg.Key, listOps)
		return err
	}
	return nil
}

func (c *KvCacher) handleWatcher(watcher Watcher) error {
	defer watcher.Stop()
	for evt := range watcher.EventBus() {
		if evt.Type == proto.EVT_ERROR {
			err := evt.Object.(error)
			return err
		}
		c.sync([]*Event{evt})
	}
	return nil
}

func (c *KvCacher) sync(evts []*Event) {
	cache := c.Cache().(*KvCache)
	store := cache.Lock()
	defer cache.Unlock()
	for _, evt := range evts {
		kv := evt.Object.(*mvccpb.KeyValue)
		key := registry.BytesToStringWithNoCopy(kv.Key)
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
			c.Cfg.OnEvent(&KvEvent{
				Revision: evt.Revision,
				Action:   t,
				KV:       kv,
			})
		case proto.EVT_DELETE:
			if ok {
				util.LOGGER.Debugf("remove key %s and notify watcher, %+v", key, kv)
				delete(store, key)
				c.Cfg.OnEvent(&KvEvent{
					Revision: evt.Revision,
					Action:   evt.Type,
					KV:       prevKv,
				})
				continue
			}
			util.LOGGER.Warnf(nil, "unexpected %s event! nonexistent key %s", evt.Type, key)
		}
	}
}

func (c *KvCacher) filter(rev int64, items []interface{}) []*Event {
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
	for _, itf := range items {
		kv := itf.(*mvccpb.KeyValue)
		newStore[registry.BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan *Event, max)
	go func() {
		for k, v := range store {
			_, ok := newStore[k]
			if !ok {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_DELETE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
			}
		}
		close(filterStopCh)
	}()

	go func() {
		for k, v := range newStore {
			ov, ok := store[k]
			if !ok {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_CREATE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
				continue
			}
			if ov.ModRevision < v.ModRevision {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_UPDATE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
				continue
			}
		}
		select {
		case <-filterStopCh:
			close(eventsCh)
		}
	}()

	evts := make([]*Event, 0, max)
	for evt := range eventsCh {
		evts = append(evts, evt)
	}
	return evts
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
	util.LOGGER.Debugf("cacher is stopped, %s", c.Cfg)
}

type KvEvent struct {
	Revision int64
	Action   proto.EventType
	KV       *mvccpb.KeyValue
}

type KvEventFunc func(evt *KvEvent)

type KvCacherConfig struct {
	Key     string
	Timeout time.Duration
	Period  time.Duration
	OnEvent KvEventFunc
}

func (cfg *KvCacherConfig) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		cfg.Key, cfg.Timeout, cfg.Period)
}

func NewKvCache(c *KvCacher) *KvCache {
	return &KvCache{
		owner: c,
		store: make(map[string]*mvccpb.KeyValue),
	}
}

func NewKvCacher(cfg *KvCacherConfig) Cacher {
	cacher := &KvCacher{
		Cfg: cfg,
		lw: &KvListWatcher{
			Client: registry.GetRegisterCenter(),
			Key:    cfg.Key,
		},
		goroute:            util.NewGo(make(chan struct{})),
		noEventMaxInterval: DEFAULT_MAX_NO_EVENT_INTERVAL,
	}
	cacher.cache = NewKvCache(cacher)
	return cacher
}

func NewCacher(prefix string, callback KvEventFunc) Cacher {
	return NewTimeoutCacher(DEFAULT_LISTWATCH_TIMEOUT, prefix, callback)
}

func NewTimeoutCacher(ot time.Duration, prefix string, callback KvEventFunc) Cacher {
	return NewKvCacher(&KvCacherConfig{
		Key:     prefix,
		Timeout: ot,
		Period:  time.Second,
		OnEvent: callback,
	})
}

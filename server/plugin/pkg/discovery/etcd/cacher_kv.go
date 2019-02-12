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
package etcd

import (
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type KvCacher struct {
	Cfg *discovery.Config

	reListCount int

	ready     chan struct{}
	lw        ListWatch
	mux       sync.Mutex
	once      sync.Once
	cache     discovery.Cache
	goroutine *gopool.Pool
}

func (c *KvCacher) Config() *discovery.Config {
	return c.Cfg
}

func (c *KvCacher) needList() bool {
	rev := c.lw.Revision()
	if rev == 0 {
		c.reListCount = 0
		return true
	}
	c.reListCount++
	if c.reListCount < DEFAULT_FORCE_LIST_INTERVAL {
		return false
	}
	c.reListCount = 0
	return true
}

func (c *KvCacher) doList(cfg ListWatchConfig) error {
	resp, err := c.lw.List(cfg)
	if err != nil {
		return err
	}

	kvs := resp.Kvs
	start := time.Now()
	evts := c.filter(c.lw.Revision(), kvs)
	if ec, kc := len(evts), len(kvs); c.Cfg.DeferHandler != nil && ec == 0 && kc != 0 &&
		c.Cfg.DeferHandler.Reset() {
		log.Warnf("most of the protected data(%d/%d) are recovered",
			kc, c.cache.GetAll(nil))
	}
	c.sync(evts)
	log.LogDebugOrWarnf(start, "finish to cache key %s, %d items, rev: %d",
		c.Cfg.Key, len(kvs), c.lw.Revision())

	return nil
}

func (c *KvCacher) doWatch(cfg ListWatchConfig) error {
	if watcher := c.lw.Watch(cfg); watcher != nil {
		return c.handleWatcher(watcher)
	}
	return fmt.Errorf("handle a nil watcher")
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	cfg := ListWatchConfig{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}

	// the scenario need to list etcd:
	// 1. Initial: cache is building, the lister's revision is 0.
	// 2. Runtime: error occurs in previous watch operation, the lister's revision is set to 0.
	// 3. Runtime: no event comes in watch operation over DEFAULT_FORCE_LIST_INTERVAL times.
	if c.needList() {
		if err := c.doList(cfg); err != nil && (!c.IsReady() || c.lw.Revision() == 0) {
			return err // do retry to list etcd
		}
		// keep going to next step:
		// 1. doList return OK.
		// 2. some traps in etcd client, like the limitation of max response body(4MB),
		//    doList always return error. So call doWatch to compensate it if cacher is ready.
	}

	util.SafeCloseChan(c.ready)

	return c.doWatch(cfg)
}

func (c *KvCacher) handleWatcher(watcher Watcher) error {
	defer watcher.Stop()
	for resp := range watcher.EventBus() {
		if resp == nil {
			return errors.New("handle watcher error")
		}

		start := time.Now()
		rev := resp.Revision
		evts := make([]discovery.KvEvent, 0, len(resp.Kvs))
		for _, kv := range resp.Kvs {
			evt := discovery.NewKvEvent(proto.EVT_CREATE, nil, kv.ModRevision)
			switch {
			case resp.Action == registry.Put && kv.Version == 1:
				evt.Type, evt.KV = proto.EVT_CREATE, c.doParse(kv)
			case resp.Action == registry.Put:
				evt.Type, evt.KV = proto.EVT_UPDATE, c.doParse(kv)
			case resp.Action == registry.Delete:
				evt.Type = proto.EVT_DELETE
				if kv.Value == nil {
					// it will happen in embed mode, and then need to get the cache value to unmarshal
					evt.KV = c.cache.Get(util.BytesToStringWithNoCopy(kv.Key))
				} else {
					evt.KV = c.doParse(kv)
				}
			default:
				log.Errorf(nil, "unknown KeyValue %v", kv)
				continue
			}
			if evt.KV == nil {
				continue
			}
			evts = append(evts, evt)
		}
		c.sync(evts)
		log.LogDebugOrWarnf(start, "finish to handle %d events, prefix: %s, rev: %d",
			len(evts), c.Cfg.Key, rev)
	}
	return nil
}

func (c *KvCacher) needDeferHandle(evts []discovery.KvEvent) bool {
	if c.Cfg.DeferHandler == nil || !c.IsReady() {
		return false
	}

	return c.Cfg.DeferHandler.OnCondition(c.Cache(), evts)
}

func (c *KvCacher) refresh(ctx context.Context) {
	log.Debugf("start to list and watch %s", c.Cfg)
	retries := 0

	timer := time.NewTimer(minWaitInterval)
	defer timer.Stop()
	for {
		nextPeriod := minWaitInterval
		if err := c.ListAndWatch(ctx); err != nil {
			retries++
			nextPeriod = backoff.GetBackoff().Delay(retries)
		} else {
			retries = 0
		}

		select {
		case <-ctx.Done():
			log.Debugf("stop to list and watch %s", c.Cfg)
			return
		case <-timer.C:
			timer.Reset(nextPeriod)
		}
	}
}

// keep the evts valid when call sync
func (c *KvCacher) sync(evts []discovery.KvEvent) {
	if len(evts) == 0 {
		return
	}

	if c.needDeferHandle(evts) {
		return
	}

	c.onEvents(evts)
}

func (c *KvCacher) filter(rev int64, items []*mvccpb.KeyValue) []discovery.KvEvent {
	nc := len(items)
	newStore := make(map[string]*mvccpb.KeyValue, nc)
	for _, kv := range items {
		newStore[util.BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan [eventBlockSize]discovery.KvEvent, 2)

	go c.filterDelete(newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(newStore, rev, eventsCh, filterStopCh)

	evts := make([]discovery.KvEvent, 0, nc)
	for block := range eventsCh {
		for _, e := range block {
			if e.KV == nil {
				break
			}
			evts = append(evts, e)
		}
	}
	return evts
}

func (c *KvCacher) filterDelete(newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan [eventBlockSize]discovery.KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]discovery.KvEvent
	i := 0

	c.cache.ForEach(func(k string, v *discovery.KeyValue) (next bool) {
		next = true

		_, ok := newStore[k]
		if ok {
			return
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]discovery.KvEvent{}
			i = 0
		}

		block[i] = discovery.NewKvEvent(proto.EVT_DELETE, v, rev)
		i++
		return
	})

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan [eventBlockSize]discovery.KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]discovery.KvEvent
	i := 0

	for k, v := range newStore {
		ov := c.cache.Get(k)
		if ov == nil {
			if i >= eventBlockSize {
				eventsCh <- block
				block = [eventBlockSize]discovery.KvEvent{}
				i = 0
			}

			if kv := c.doParse(v); kv != nil {
				block[i] = discovery.NewKvEvent(proto.EVT_CREATE, kv, rev)
				i++
			}
			continue
		}

		if ov.CreateRevision == v.CreateRevision && ov.ModRevision == v.ModRevision {
			continue
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]discovery.KvEvent{}
			i = 0
		}

		if kv := c.doParse(v); kv != nil {
			block[i] = discovery.NewKvEvent(proto.EVT_UPDATE, kv, rev)
			i++
		}
	}

	if i > 0 {
		eventsCh <- block
	}

	<-filterStopCh

	close(eventsCh)
}

func (c *KvCacher) deferHandle(ctx context.Context) {
	if c.Cfg.DeferHandler == nil {
		return
	}

	var (
		evts = make([]discovery.KvEvent, eventBlockSize)
		i    int
	)
	interval := 300 * time.Millisecond
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-c.Cfg.DeferHandler.HandleChan():
			if !ok {
				return
			}

			if i >= eventBlockSize {
				c.onEvents(evts[:i])
				evts = make([]discovery.KvEvent, eventBlockSize)
				i = 0
			}

			evts[i] = evt
			i++

			util.ResetTimer(timer, interval)
		case <-timer.C:
			timer.Reset(interval)

			if i == 0 {
				continue
			}

			c.onEvents(evts[:i])
			evts = make([]discovery.KvEvent, eventBlockSize)
			i = 0
		}
	}
}

func (c *KvCacher) onEvents(evts []discovery.KvEvent) {
	start := time.Now()
	init := !c.IsReady()
	for i, evt := range evts {
		key := util.BytesToStringWithNoCopy(evt.KV.Key)
		prevKv := c.cache.Get(key)
		ok := prevKv != nil

		switch evt.Type {
		case proto.EVT_CREATE, proto.EVT_UPDATE:
			switch {
			case init:
				evt.Type = proto.EVT_INIT
			case !ok && evt.Type != proto.EVT_CREATE:
				log.Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				evt.Type = proto.EVT_CREATE
			case ok && evt.Type != proto.EVT_UPDATE:
				log.Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				evt.Type = proto.EVT_UPDATE
			}

			c.cache.Put(key, evt.KV)
			evts[i] = evt
		case proto.EVT_DELETE:
			if !ok {
				log.Warnf("unexpected %s event! key %s does not cache",
					evt.Type, key)
			} else {
				evt.KV = prevKv
				c.cache.Remove(key)
			}
			evts[i] = evt
		}
	}

	c.notify(evts)

	discovery.ReportProcessEventCompleted(evts, start)
}

func (c *KvCacher) notify(evts []discovery.KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer log.Recover()
	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
}

func (c *KvCacher) doParse(src *mvccpb.KeyValue) (kv *discovery.KeyValue) {
	kv = discovery.NewKeyValue()
	if err := FromEtcdKeyValue(kv, src, c.Cfg.Parser); err != nil {
		log.Errorf(err, "parse %s value failed", util.BytesToStringWithNoCopy(src.Key))
		return nil
	}
	return
}

func (c *KvCacher) Cache() discovery.Cache {
	return c.cache
}

func (c *KvCacher) Run() {
	c.once.Do(func() {
		c.goroutine.Do(c.refresh)
		c.goroutine.Do(c.deferHandle)
		c.goroutine.Do(c.reportMetrics)
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

func (c *KvCacher) reportMetrics(ctx context.Context) {
	if !core.ServerInfo.Config.EnablePProf {
		return
	}
	timer := time.NewTimer(DEFAULT_METRICS_INTERVAL)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			ReportCacheSize(c.cache.Name(), "raw", c.cache.Size())
			timer.Reset(DEFAULT_METRICS_INTERVAL)
		}
	}
}

func NewKvCacher(cfg *discovery.Config, cache discovery.Cache) *KvCacher {
	return &KvCacher{
		Cfg:   cfg,
		cache: cache,
		ready: make(chan struct{}),
		lw: &innerListWatch{
			Client: backend.Registry(),
			Prefix: cfg.Key,
		},
		goroutine: gopool.New(context.Background()),
	}
}

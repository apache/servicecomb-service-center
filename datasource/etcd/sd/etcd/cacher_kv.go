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
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	rmodel "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
)

// KvCacher implements sd.Cacher.
// KvCacher manages etcd cache.
// To update cache, KvCacher watch etcd event and pull data periodly from etcd.
// When the cache data changes, KvCacher creates events and notifies it's
// subscribers.
// Use Cfg to set it's behaviors.
type KvCacher struct {
	Cfg *sd.Config

	reListCount int

	ready     chan struct{}
	lw        sdcommon.ListWatch
	mux       sync.Mutex
	once      sync.Once
	cache     sd.Cache
	goroutine *gopool.Pool
}

func (c *KvCacher) Config() *sd.Config {
	return c.Cfg
}

func (c *KvCacher) needList() bool {
	rev := c.getRevision()
	if rev == 0 {
		c.reListCount = 0
		return true
	}
	c.reListCount++
	if c.reListCount < sdcommon.DefaultForceListInterval {
		return false
	}
	c.reListCount = 0
	return true
}

func (c *KvCacher) doList(cfg sdcommon.ListWatchConfig) error {
	resp, err := c.lw.List(cfg)
	if err != nil {
		return err
	}

	rev := c.getRevision()
	resources := resp.Resources
	start := time.Now()
	defer log.DebugOrWarnf(start, "finish to cache key %s, %d items, rev: %d",
		c.Cfg.Key, len(resources), rev)

	// just reset the cacher if cache marked dirty
	if c.cache.Dirty() {
		c.reset(rev, resources)
		log.Warnf("Cache[%s] is reset!", c.cache.Name())
		return nil
	}

	// calc and return the diff between cache and ETCD
	evts := c.filter(rev, resources)
	// there is no change between List() and cache, then stop the self preservation
	if ec, kc := len(evts), len(resources); c.Cfg.DeferHandler != nil && ec == 0 && kc != 0 &&
		c.Cfg.DeferHandler.Reset() {
		log.Warnf("most of the protected data(%d/%d) are recovered",
			kc, c.cache.GetAll(nil))
	}

	// notify the subscribers
	c.sync(evts)
	return nil
}

func (c *KvCacher) reset(rev int64, kvs []*sdcommon.Resource) {
	if c.Cfg.DeferHandler != nil {
		c.Cfg.DeferHandler.Reset()
	}
	// clear cache before Set is safe, because the watch operation is stop,
	// but here will make all API requests go to ETCD directly.
	c.cache.Clear()
	// do not notify when cacher is dirty status,
	// otherwise, too many events will notify to downstream.
	c.buildCache(c.filter(rev, kvs))
}

func (c *KvCacher) doWatch(cfg sdcommon.ListWatchConfig) error {
	if eventBus := c.lw.EventBus(cfg); eventBus != nil {
		return c.handleEventBus(eventBus)
	}
	return fmt.Errorf("handle a nil watcher")
}

func (c *KvCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	defer log.Recover() // ensure ListAndWatch never raise panic

	cfg := sdcommon.ListWatchConfig{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}

	// the scenario need to list etcd:
	// 1. Initial: cache is building, the lister's revision is 0.
	// 2. Runtime: error occurs in previous watch operation, the lister's revision is set to 0.
	// 3. Runtime: watch operation timed out over DEFAULT_FORCE_LIST_INTERVAL times.
	if c.needList() {
		if err := c.doList(cfg); err != nil && (!c.IsReady() || c.getRevision() == 0) {
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

func (c *KvCacher) handleEventBus(eventBus *sdcommon.EventBus) error {
	defer eventBus.Stop()
	for resp := range eventBus.ResourceEventBus() {
		if resp == nil {
			return errors.New("handle watcher error")
		}

		start := time.Now()
		rev := resp.Revision
		evts := make([]sd.KvEvent, 0, len(resp.Resources))
		for _, resource := range resp.Resources {
			evt := sd.NewKvEvent(rmodel.EVT_CREATE, nil, resource.ModRevision)
			switch {
			case resp.Action == sdcommon.ActionPUT && resource.Version == 1:
				evt.Type, evt.KV = rmodel.EVT_CREATE, c.doParse(resource)
			case resp.Action == sdcommon.ActionPUT:
				evt.Type, evt.KV = rmodel.EVT_UPDATE, c.doParse(resource)
			case resp.Action == sdcommon.ActionDelete:
				evt.Type = rmodel.EVT_DELETE
				if resource.Value == nil {
					// it will happen in embed mode, and then need to get the cache value not unmarshal
					evt.KV = c.cache.Get(resource.Key)
				} else {
					evt.KV = c.doParse(resource)
				}
			default:
				log.Errorf(nil, "unknown KeyValue %v", resource)
				continue
			}
			if evt.KV == nil {
				log.Errorf(nil, "failed to parse KeyValue %v", resource)
				continue
			}
			evts = append(evts, evt)
		}
		c.sync(evts)
		log.DebugOrWarnf(start, "finish to handle %d events, prefix: %s, rev: %d",
			len(evts), c.Cfg.Key, rev)
	}
	return nil
}

func (c *KvCacher) needDeferHandle(evts []sd.KvEvent) bool {
	if c.Cfg.DeferHandler == nil || !c.IsReady() {
		return false
	}

	return c.Cfg.DeferHandler.OnCondition(c.Cache(), evts)
}

func (c *KvCacher) refresh(ctx context.Context) {
	log.Debugf("start to list and watch %s", c.Cfg)
	retries := 0

	timer := time.NewTimer(sdcommon.MinWaitInterval)
	defer timer.Stop()
	for {
		nextPeriod := sdcommon.MinWaitInterval
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
func (c *KvCacher) sync(evts []sd.KvEvent) {
	if len(evts) == 0 {
		return
	}

	if c.needDeferHandle(evts) {
		return
	}

	c.onEvents(evts)
}

func (c *KvCacher) filter(rev int64, items []*sdcommon.Resource) []sd.KvEvent {
	nc := len(items)
	newStore := make(map[string]*sdcommon.Resource, nc)
	for _, kv := range items {
		newStore[kv.Key] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan [sdcommon.EventBlockSize]sd.KvEvent, 2)

	go c.filterDelete(newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(newStore, rev, eventsCh, filterStopCh)

	evts := make([]sd.KvEvent, 0, nc)
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

func (c *KvCacher) filterDelete(newStore map[string]*sdcommon.Resource,
	rev int64, eventsCh chan [sdcommon.EventBlockSize]sd.KvEvent, filterStopCh chan struct{}) {
	var block [sdcommon.EventBlockSize]sd.KvEvent
	i := 0

	c.cache.ForEach(func(k string, v *sd.KeyValue) (next bool) {
		next = true

		_, ok := newStore[k]
		if ok {
			return
		}

		if i >= sdcommon.EventBlockSize {
			eventsCh <- block
			block = [sdcommon.EventBlockSize]sd.KvEvent{}
			i = 0
		}

		block[i] = sd.NewKvEvent(rmodel.EVT_DELETE, v, rev)
		i++
		return
	})

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(newStore map[string]*sdcommon.Resource,
	rev int64, eventsCh chan [sdcommon.EventBlockSize]sd.KvEvent, filterStopCh chan struct{}) {
	var block [sdcommon.EventBlockSize]sd.KvEvent
	i := 0

	for k, v := range newStore {
		ov := c.cache.Get(k)
		if ov == nil {
			if i >= sdcommon.EventBlockSize {
				eventsCh <- block
				block = [sdcommon.EventBlockSize]sd.KvEvent{}
				i = 0
			}

			if kv := c.doParse(v); kv != nil {
				block[i] = sd.NewKvEvent(rmodel.EVT_CREATE, kv, rev)
				i++
			}
			continue
		}

		if ov.CreateRevision == v.CreateRevision && ov.ModRevision == v.ModRevision {
			continue
		}

		if i >= sdcommon.EventBlockSize {
			eventsCh <- block
			block = [sdcommon.EventBlockSize]sd.KvEvent{}
			i = 0
		}

		if kv := c.doParse(v); kv != nil {
			block[i] = sd.NewKvEvent(rmodel.EVT_UPDATE, kv, rev)
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
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.handleDeferEvents(ctx)
		}
	}
}

func (c *KvCacher) handleDeferEvents(ctx context.Context) {
	defer log.Recover()
	var (
		evts = make([]sd.KvEvent, sdcommon.EventBlockSize)
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
				log.Error("replay channel is closed", nil)
				return
			}

			if i >= sdcommon.EventBlockSize {
				c.onEvents(evts[:i])
				evts = make([]sd.KvEvent, sdcommon.EventBlockSize)
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
			evts = make([]sd.KvEvent, sdcommon.EventBlockSize)
			i = 0
		}
	}
}

func (c *KvCacher) onEvents(evts []sd.KvEvent) {
	c.buildCache(evts)
	c.notify(evts)
}

func (c *KvCacher) buildCache(evts []sd.KvEvent) {
	init := !c.IsReady()
	for i, evt := range evts {
		key := util.BytesToStringWithNoCopy(evt.KV.Key)
		prevKv := c.cache.Get(key)
		ok := prevKv != nil

		switch evt.Type {
		case rmodel.EVT_CREATE, rmodel.EVT_UPDATE:
			switch {
			case init:
				evt.Type = rmodel.EVT_INIT
			case !ok && evt.Type != rmodel.EVT_CREATE:
				log.Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, rmodel.EVT_CREATE, key)
				evt.Type = rmodel.EVT_CREATE
			case ok && evt.Type != rmodel.EVT_UPDATE:
				log.Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, rmodel.EVT_UPDATE, key)
				evt.Type = rmodel.EVT_UPDATE
			}

			c.cache.Put(key, evt.KV)
			evts[i] = evt
		case rmodel.EVT_DELETE:
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
	sd.ReportProcessEventCompleted(c.Cfg.Key, evts)
}

func (c *KvCacher) notify(evts []sd.KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer log.Recover()
	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
	sd.ReportDispatchEventCompleted(c.Cfg.Key, evts)
}

func (c *KvCacher) doParse(src *sdcommon.Resource) (kv *sd.KeyValue) {
	kv = sd.NewKeyValue()
	if err := ParseResourceToEtcdKeyValue(kv, src, c.Cfg.Parser); err != nil {
		log.Errorf(err, "parse %s value failed", src.Key)
		return nil
	}
	return
}

func (c *KvCacher) Cache() sd.CacheReader {
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
	if !config.GetServer().EnablePProf {
		return
	}
	timer := time.NewTimer(sdcommon.DefaultMetricsInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			ReportCacheSize(c.cache.Name(), "raw", c.cache.Size())
			timer.Reset(sdcommon.DefaultMetricsInterval)
		}
	}
}

func NewKvCacher(cfg *sd.Config, cache sd.Cache) *KvCacher {
	return &KvCacher{
		Cfg:   cfg,
		cache: cache,
		ready: make(chan struct{}),
		lw: &innerListWatch{
			Client: client.Instance(),
			Prefix: cfg.Key,
		},
		goroutine: gopool.New(context.Background()),
	}
}

func (c *KvCacher) getRevision() int64 {
	lw := reflect.ValueOf(c.lw)
	revMethod := lw.MethodByName(revMethodName)

	if !revMethod.IsValid() {
		log.Warn("get revision failed")
		return 0
	}

	revs := revMethod.Call(make([]reflect.Value, 0))
	rev := revs[0].Int()
	return rev
}

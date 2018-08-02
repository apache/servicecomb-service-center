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
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type KvCacher struct {
	Cfg *Config

	latestListRev  int64
	noEventPeriods int

	ready     chan struct{}
	lw        ListWatch
	mux       sync.Mutex
	once      sync.Once
	cache     Cache
	goroutine *util.GoRoutine
}

func (c *KvCacher) Config() *Config {
	return c.Cfg
}

func (c *KvCacher) needList() bool {
	rev := c.lw.Revision()
	if rev == 0 {
		c.noEventPeriods = 0
		return true
	}
	if c.latestListRev != rev {
		c.noEventPeriods = 0
		return false
	}
	c.noEventPeriods++
	if c.Cfg.NoEventPeriods == 0 || c.noEventPeriods < c.Cfg.NoEventPeriods {
		return false
	}

	util.Logger().Debugf("no events come in more then %s, need to list key %s, rev: %d",
		time.Duration(c.noEventPeriods)*c.Cfg.Timeout, c.Cfg.Key, rev)
	c.noEventPeriods = 0
	return true
}

func (c *KvCacher) doList(cfg ListWatchConfig) error {
	resp, err := c.lw.List(cfg)
	if err != nil {
		return err
	}
	c.latestListRev = c.lw.Revision()

	kvs := resp.Kvs
	start := time.Now()
	evts := c.filter(c.lw.Revision(), kvs)
	if ec, kc := len(evts), len(kvs); c.Cfg.DeferHandler != nil && ec == 0 && kc != 0 &&
		c.Cfg.DeferHandler.Reset() {
		util.Logger().Warnf("most of the protected data(%d/%d) are recovered",
			kc, c.cache.GetAll(nil))
	}
	c.sync(evts)
	util.LogDebugOrWarnf(start, "finish to cache key %s, %d items, rev: %d",
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
	if c.needList() {
		if err := c.doList(cfg); err != nil && !c.IsReady() {
			// cacher is not ready, so it need to retry util the cache is created
			return err
		}
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
		evts := make([]KvEvent, 0, len(resp.Kvs))
		for _, kv := range resp.Kvs {
			evt := KvEvent{Prefix: c.Cfg.Key, Revision: kv.ModRevision}
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
				util.Logger().Errorf(nil, "unknown KeyValue %v", kv)
				continue
			}
			if evt.KV == nil {
				continue
			}
			evts = append(evts, evt)
		}
		c.sync(evts)
		util.LogDebugOrWarnf(start, "finish to handle %d events, prefix: %s, rev: %d",
			len(evts), c.Cfg.Key, rev)
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

	timer := time.NewTimer(minWaitInterval)
	defer timer.Stop()
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
		case <-timer.C:
			timer.Reset(nextPeriod)
		}
	}
}

// keep the evts valid when call sync
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
	nc := len(items)
	newStore := make(map[string]*mvccpb.KeyValue, nc)
	for _, kv := range items {
		newStore[util.BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan [eventBlockSize]KvEvent, 2)

	go c.filterDelete(newStore, rev, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(newStore, rev, eventsCh, filterStopCh)

	evts := make([]KvEvent, 0, nc)
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
	rev int64, eventsCh chan [eventBlockSize]KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]KvEvent
	i := 0

	c.cache.ForEach(func(k string, v *KeyValue) (next bool) {
		next = true

		_, ok := newStore[k]
		if ok {
			return
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]KvEvent{}
			i = 0
		}

		block[i] = KvEvent{
			Revision: rev,
			Type:     proto.EVT_DELETE,
			Prefix:   c.Cfg.Key,
			KV:       v,
		}
		i++
		return
	})

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *KvCacher) filterCreateOrUpdate(newStore map[string]*mvccpb.KeyValue,
	rev int64, eventsCh chan [eventBlockSize]KvEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]KvEvent
	i := 0

	for k, v := range newStore {
		ov := c.cache.Get(k)
		if ov == nil {
			if i >= eventBlockSize {
				eventsCh <- block
				block = [eventBlockSize]KvEvent{}
				i = 0
			}

			if kv := c.doParse(v); kv != nil {
				block[i] = KvEvent{
					Revision: rev,
					Type:     proto.EVT_CREATE,
					Prefix:   c.Cfg.Key,
					KV:       kv,
				}
				i++
			}
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

		if kv := c.doParse(v); kv != nil {
			block[i] = KvEvent{
				Revision: rev,
				Type:     proto.EVT_UPDATE,
				Prefix:   c.Cfg.Key,
				KV:       kv,
			}
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
		evts = make([]KvEvent, eventBlockSize)
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
				evts = make([]KvEvent, eventBlockSize)
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
			evts = make([]KvEvent, eventBlockSize)
			i = 0
		}
	}
}

func (c *KvCacher) onEvents(evts []KvEvent) {
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
				util.Logger().Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				evt.Type = proto.EVT_CREATE
			case ok && evt.Type != proto.EVT_UPDATE:
				util.Logger().Warnf("unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				evt.Type = proto.EVT_UPDATE
			}

			c.cache.Put(key, evt.KV)
			evts[i] = evt
		case proto.EVT_DELETE:
			if !ok {
				util.Logger().Warnf("unexpected %s event! key %s does not cache",
					evt.Type, key)
			} else {
				evt.KV = prevKv
				c.cache.Remove(key)
			}
			evts[i] = evt
		}
	}

	c.onKvEvents(evts)
}

func (c *KvCacher) onKvEvents(evts []KvEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer util.RecoverAndReport()

	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
}

func (c *KvCacher) doParse(src *mvccpb.KeyValue) (kv *KeyValue) {
	kv = new(KeyValue)
	if err := kv.From(c.Cfg.Parser, src); err != nil {
		util.Logger().Errorf(err, "parse %s value failed", util.BytesToStringWithNoCopy(src.Key))
		return nil
	}
	return
}

func (c *KvCacher) Cache() Cache {
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

func NewKvCacher(name string, cfg *Config) *KvCacher {
	cacher := &KvCacher{
		Cfg:   cfg,
		ready: make(chan struct{}),
		lw: &innerListWatch{
			Client: Registry(),
			Prefix: cfg.Key,
		},
		goroutine: util.NewGo(context.Background()),
	}
	cacher.cache = NewKvCache(name, cfg)
	return cacher
}

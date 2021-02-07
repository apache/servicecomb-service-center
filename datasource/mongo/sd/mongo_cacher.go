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

package sd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rmodel "github.com/go-chassis/cari/discovery"
)

// MongoCacher manages mongo cache.
// To updateOp cache, MongoCacher watch mongo event and pull data periodly from mongo.
// When the cache data changes, MongoCacher creates events and notifies it's
// subscribers.
// Use Options to set it's behaviors.
type MongoCacher struct {
	Options     *Options
	reListCount int
	isFirstTime bool
	cache       *MongoCache
	ready       chan struct{}
	lw          sdcommon.ListWatch
	mux         sync.Mutex
	once        sync.Once
	goroutine   *gopool.Pool
}

func (c *MongoCacher) Cache() *MongoCache {
	return c.cache
}

func (c *MongoCacher) Run() {
	c.once.Do(func() {
		c.goroutine.Do(c.refresh)
	})
}

func (c *MongoCacher) Stop() {
	c.goroutine.Close(true)

	util.SafeCloseChan(c.ready)
}

func (c *MongoCacher) Ready() <-chan struct{} {
	return c.ready
}

func (c *MongoCacher) IsReady() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

func (c *MongoCacher) needList() bool {
	if c.isFirstTime {
		c.isFirstTime = false
		return true
	}
	c.reListCount++
	if c.reListCount < sdcommon.DefaultForceListInterval {
		return false
	}
	c.reListCount = 0
	return true
}

func (c *MongoCacher) doList(cfg sdcommon.ListWatchConfig) error {
	resp, err := c.lw.List(cfg)
	if err != nil {
		return err
	}

	resources := resp.Resources

	defer log.Debug(fmt.Sprintf("finish to cache key %s, %d items",
		c.Options.Key, len(resources)))

	//just reset the cacher if cache marked dirty
	if c.cache.Dirty() {
		c.reset(resources)
		log.Warn(fmt.Sprintf("Cache[%s] is reset!", c.cache.Name()))
		return nil
	}

	// calc and return the diff between cache and mongodb
	events := c.filter(resources)

	//notify the subscribers
	c.sync(events)
	return nil
}

func (c *MongoCacher) reset(infos []*sdcommon.Resource) {
	// clear cache before Set is safe, because the watch operation is stop,
	// but here will make all API requests go to MONGO directly.
	c.cache.Clear()
	// do not notify when cacher is dirty status,
	// otherwise, too many events will notify to downstream.
	c.buildCache(c.filter(infos))
}

func (c *MongoCacher) doWatch(cfg sdcommon.ListWatchConfig) error {
	if eventbus := c.lw.EventBus(cfg); eventbus != nil {
		return c.handleEventBus(eventbus)
	}
	return fmt.Errorf("handle a nil watcher")
}

func (c *MongoCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	defer log.Recover() // ensure ListAndWatch never raise panic

	cfg := sdcommon.ListWatchConfig{
		Timeout: c.Options.Timeout,
		Context: ctx,
	}

	// first time should initial cache, set watch timeout less
	if c.isFirstTime {
		cfg.Timeout = FirstTimeout
	}

	err := c.doWatch(cfg)
	if err != nil {
		log.Error("doWatch err", err)
	}

	// the scenario need to list mongo:
	// 1. Initial: cache is building, the lister's is first time to run.
	// 2. Runtime: error occurs in previous watch operation, the lister's status is set to error.
	// 3. Runtime: watch operation timed out over DEFAULT_FORCE_LIST_INTERVAL times.
	if c.needList() {
		// recover timeout for list
		if c.isFirstTime {
			cfg.Timeout = c.Options.Timeout
		}

		if err := c.doList(cfg); err != nil && (!c.IsReady()) {
			log.Error("doList error", err)
			return err // do retry to list mongo
		}
		// keep going to next step:
		// 1. doList return OK.
		// 2. some traps in mongo client
	}

	util.SafeCloseChan(c.ready)

	return nil
}

func (c *MongoCacher) handleEventBus(eventbus *sdcommon.EventBus) error {
	defer eventbus.Stop()

	if eventbus.Bus == nil {
		return nil
	}

	for resp := range eventbus.ResourceEventBus() {
		events := make([]MongoEvent, 0)

		if resp == nil {
			return errors.New("handle watcher error")
		}

		for _, resource := range resp.Resources {
			action := resp.Action
			var event MongoEvent
			switch action {
			case sdcommon.ActionCreate:
				event = NewMongoEventByResource(resource, rmodel.EVT_CREATE)
			case sdcommon.ActionUpdate:
				event = NewMongoEventByResource(resource, rmodel.EVT_UPDATE)
			case sdcommon.ActionDelete:
				resource.Key = c.cache.GetKeyByDocumentID(resource.DocumentID)
				resource.Value = c.cache.Get(resource.Key)
				event = NewMongoEventByResource(resource, rmodel.EVT_DELETE)
			}
			events = append(events, event)
		}

		c.sync(events)
		log.Debug(fmt.Sprintf("finish to handle %d events, table: %s", len(events), c.Options.Key))
	}

	return nil
}

func (c *MongoCacher) refresh(ctx context.Context) {
	log.Debug(fmt.Sprintf("start to list and watch %s", c.Options))
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
			log.Debug(fmt.Sprintf("stop to list and watch %s", c.Options))
			return
		case <-timer.C:
			timer.Reset(nextPeriod)
		}
	}
}

// keep the evts valID when call sync
func (c *MongoCacher) sync(evts []MongoEvent) {
	if len(evts) == 0 {
		return
	}

	c.onEvents(evts)
}

func (c *MongoCacher) filter(infos []*sdcommon.Resource) []MongoEvent {
	nc := len(infos)
	newStore := make(map[string]interface{}, nc)
	documentIDRecord := make(map[string]string, nc)
	indexRecord := make(map[string]string, nc)

	for _, info := range infos {
		event := NewMongoEventByResource(info, rmodel.EVT_CREATE)
		newStore[event.ResourceID] = info.Value
		documentIDRecord[event.ResourceID] = info.DocumentID
		indexRecord[event.ResourceID] = info.Index
	}

	filterStopCh := make(chan struct{})
	eventsCh := make(chan [sdcommon.EventBlockSize]MongoEvent, 2)

	go c.filterDelete(newStore, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(newStore, documentIDRecord, indexRecord, eventsCh, filterStopCh)

	events := make([]MongoEvent, 0, nc)
	for block := range eventsCh {
		for _, e := range block {
			if e.Value == nil {
				break
			}
			events = append(events, e)
		}
	}
	return events
}

func (c *MongoCacher) filterDelete(newStore map[string]interface{},
	eventsCh chan [sdcommon.EventBlockSize]MongoEvent, filterStopCh chan struct{}) {
	var block [sdcommon.EventBlockSize]MongoEvent
	i := 0

	c.cache.ForEach(func(k string, v interface{}) (next bool) {
		next = true

		_, ok := newStore[k]
		if ok {
			// k in store, also in new store, is not deleted, return
			return
		}

		// k in store but not in new store, it means k is deleted
		if i >= sdcommon.EventBlockSize {
			eventsCh <- block
			block = [sdcommon.EventBlockSize]MongoEvent{}
			i = 0
		}

		documentID := c.cache.GetDocumentIDByBussinessID(k)
		index := c.cache.GetIndexByBussinessID(k)
		block[i] = NewMongoEvent(k, documentID, index, rmodel.EVT_DELETE, v)
		i++
		return
	})

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *MongoCacher) filterCreateOrUpdate(newStore map[string]interface{}, newDocumentStore map[string]string, indexRecord map[string]string,
	eventsCh chan [sdcommon.EventBlockSize]MongoEvent, filterStopCh chan struct{}) {
	var block [sdcommon.EventBlockSize]MongoEvent
	i := 0

	for k, v := range newStore {
		ov := c.cache.Get(k)
		if ov == nil {
			if i >= sdcommon.EventBlockSize {
				eventsCh <- block
				block = [sdcommon.EventBlockSize]MongoEvent{}
				i = 0
			}

			block[i] = NewMongoEvent(k, newDocumentStore[k], indexRecord[k], rmodel.EVT_CREATE, v)
			i++

			continue
		}

		if c.isValueNotUpdated(v, ov) {
			continue
		}

		log.Debug(fmt.Sprintf("value is updateOp of key:%s, old value is:%s, new value is:%s", k, ov, v))

		if i >= sdcommon.EventBlockSize {
			eventsCh <- block
			block = [sdcommon.EventBlockSize]MongoEvent{}
			i = 0
		}

		block[i] = NewMongoEvent(k, newDocumentStore[k], indexRecord[k], rmodel.EVT_UPDATE, v)
		i++
	}

	if i > 0 {
		eventsCh <- block
	}

	<-filterStopCh

	close(eventsCh)
}

func (c *MongoCacher) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	var modTime string
	var newModTime string

	switch c.Options.Key {
	case instance:
		instance := value.(Instance)
		newInstance := newValue.(Instance)
		if instance.Instance == nil || newInstance.Instance == nil {
			return true
		}
		modTime = instance.Instance.ModTimestamp
		newModTime = newInstance.Instance.ModTimestamp
	case service:
		service := value.(Service)
		newService := newValue.(Service)
		if service.Service == nil || newService.Service == nil {
			return true
		}
		modTime = service.Service.ModTimestamp
		newModTime = newService.Service.ModTimestamp
	}

	if newModTime == "" || modTime == newModTime {
		return true
	}

	return false
}

func (c *MongoCacher) onEvents(events []MongoEvent) {
	c.buildCache(events)

	c.notify(events)
}

func (c *MongoCacher) buildCache(events []MongoEvent) {
	for i, evt := range events {
		key := evt.ResourceID
		value := c.cache.Get(key)
		ok := value != nil

		switch evt.Type {
		case rmodel.EVT_CREATE, rmodel.EVT_UPDATE:
			switch {
			case !c.IsReady():
				evt.Type = rmodel.EVT_INIT
			case !ok && evt.Type != rmodel.EVT_CREATE:
				log.Warn(fmt.Sprintf("unexpected %s event! it should be %s key %s",
					evt.Type, rmodel.EVT_CREATE, key))
				evt.Type = rmodel.EVT_CREATE
			case ok && evt.Type != rmodel.EVT_UPDATE:
				log.Warn(fmt.Sprintf("unexpected %s event! it should be %s key %s",
					evt.Type, rmodel.EVT_UPDATE, key))
				evt.Type = rmodel.EVT_UPDATE
			}

			c.cache.Put(key, evt.Value)
			c.cache.PutDocumentID(key, evt.DocumentID)
			c.cache.PutIndex(evt.Index, evt.ResourceID)

			events[i] = evt
		case rmodel.EVT_DELETE:
			if !ok {
				log.Warn(fmt.Sprintf("unexpected %s event! key %s does not cache",
					evt.Type, key))
			} else {
				evt.Value = value

				c.cache.Remove(key)
				c.cache.RemoveDocumentID(evt.DocumentID)
				c.cache.RemoveIndex(evt.Index, evt.ResourceID)
			}
			events[i] = evt
		}
	}
}

func (c *MongoCacher) notify(evts []MongoEvent) {
	eventProxy := EventProxy(c.Options.Key)

	if eventProxy == nil {
		return
	}

	defer log.Recover()

	for _, evt := range evts {
		if evt.Type == rmodel.EVT_DELETE && evt.Value == nil {
			log.Warn(fmt.Sprintf("caught delete event:%s, but value can't get from caches, it may be deleted by last list", evt.ResourceID))
			continue
		}
		eventProxy.OnEvent(evt)
	}
}

func NewMongoCacher(options *Options, cache *MongoCache) *MongoCacher {
	return &MongoCacher{
		Options:     options,
		isFirstTime: true,
		cache:       cache,
		ready:       make(chan struct{}),
		lw: &mongoListWatch{
			Key: options.Key,
		},
		goroutine: gopool.New(context.Background()),
	}
}

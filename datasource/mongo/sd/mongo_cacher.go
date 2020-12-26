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
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rmodel "github.com/go-chassis/cari/discovery"
	"sync"
	"time"
)

// MongoCacher manages mongo cache.
// To update cache, MongoCacher watch mongo event and pull data periodly from mongo.
// When the cache data changes, MongoCacher creates events and notifies it's
// subscribers.
// Use Cfg to set it's behaviors.
type MongoCacher struct {
	Cfg         *Config
	reListCount int
	cache       *MongoCache
	ready       chan struct{}
	lw          ListWatch
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

func (c *MongoCacher) Config() *Config {
	return c.Cfg
}

func (c *MongoCacher) needList() bool {
	if c.lw.ResumeToken() == nil {
		return true
	}
	c.reListCount++
	if c.reListCount < DefaultForceListInterval {
		return false
	}
	c.reListCount = 0
	return true
}

func (c *MongoCacher) doList(cfg ListWatchConfig) error {
	resp, err := c.lw.List(cfg)
	if err != nil {
		return err
	}

	infos := resp.Infos
	start := time.Now()
	defer log.DebugOrWarnf(start, "finish to cache key %s, %d items",
		c.Cfg.Key, len(infos))

	//just reset the cacher if cache marked dirty
	if c.cache.Dirty() {
		c.reset(infos)
		log.Warnf("Cache[%s] is reset!", c.cache.Name())
		return nil
	}

	// calc and return the diff between cache and mongodb
	events := c.filter(infos)

	//notify the subscribers
	c.sync(events)
	return nil
}

func (c *MongoCacher) reset(infos []MongoInfo) {
	// clear cache before Set is safe, because the watch operation is stop,
	// but here will make all API requests go to MONGO directly.
	c.cache.Clear()
	// do not notify when cacher is dirty status,
	// otherwise, too many events will notify to downstream.
	c.buildCache(c.filter(infos))
}

func (c *MongoCacher) doWatch(cfg ListWatchConfig) error {
	if watcher := c.lw.Watch(cfg); watcher != nil {
		return c.handleWatcher(watcher)
	}
	return fmt.Errorf("handle a nil watcher")
}

func (c *MongoCacher) ListAndWatch(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	defer log.Recover() // ensure ListAndWatch never raise panic

	cfg := ListWatchConfig{
		Timeout: c.Cfg.Timeout,
		Context: ctx,
	}

	// the scenario need to list mongo:
	// 1. Initial: cache is building, the lister's resume token is nil.
	// 2. Runtime: error occurs in previous watch operation, the lister's status is set to error.
	// 3. Runtime: watch operation timed out over DEFAULT_FORCE_LIST_INTERVAL times.
	if c.needList() {
		if err := c.doList(cfg); err != nil && (!c.IsReady()) {
			return err // do retry to list etcd
		}
		// keep going to next step:
		// 1. doList return OK.
		// 2. some traps in mongo client
	}

	util.SafeCloseChan(c.ready)

	return c.doWatch(cfg)
}

func (c *MongoCacher) handleWatcher(watcher Watcher) error {
	defer watcher.Stop()

	for resp := range watcher.EventBus() {
		events := make([]MongoEvent, 0)

		if resp == nil {
			return errors.New("handle watcher error")
		}

		for _, info := range resp.Infos {
			action := resp.OperationType
			var event MongoEvent
			switch action {
			case INSERT:
				event = NewMongoEventByMongoInf(info, rmodel.EVT_CREATE)
			case UPDATE, REPLACE:
				event = NewMongoEventByMongoInf(info, rmodel.EVT_UPDATE)
			case DELETE:
				info.BusinessId = c.cache.GetKeyByDocumentId(info.DocumentId)
				info.Value = c.cache.Get(info.BusinessId)
				event = NewMongoEventByMongoInf(info, rmodel.EVT_DELETE)
			}
			events = append(events, event)
		}

		c.sync(events)
		log.Debugf("finish to handle %d events, table: %s", len(events), c.Cfg.Key)
	}

	return nil
}

func (c *MongoCacher) refresh(ctx context.Context) {
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
func (c *MongoCacher) sync(evts []MongoEvent) {
	if len(evts) == 0 {
		return
	}

	c.onEvents(evts)
}

func (c *MongoCacher) filter(infos []MongoInfo) []MongoEvent {
	nc := len(infos)
	newStore := make(map[string]interface{}, nc)
	documentIdRecord := make(map[string]string, nc)

	for _, info := range infos {
		event := NewMongoEventByMongoInf(info, rmodel.EVT_CREATE)
		newStore[event.BusinessId] = info.Value
		documentIdRecord[event.BusinessId] = info.DocumentId
	}

	filterStopCh := make(chan struct{})
	eventsCh := make(chan [eventBlockSize]MongoEvent, 2)

	go c.filterDelete(newStore, eventsCh, filterStopCh)

	go c.filterCreateOrUpdate(newStore, documentIdRecord, eventsCh, filterStopCh)

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
	eventsCh chan [eventBlockSize]MongoEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]MongoEvent
	i := 0

	c.cache.ForEach(func(k string, v interface{}) (next bool) {
		next = true

		_, ok := newStore[k]
		if ok {
			// k in store, also in new store, is not deleted, return
			return
		}

		// k in store but not in new store, it means k is deleted
		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]MongoEvent{}
			i = 0
		}

		// todo for test
		documentId := c.cache.GetDocumentIdById(k)
		block[i] = NewMongoEvent(k, documentId, rmodel.EVT_DELETE, v)
		i++
		return
	})

	if i > 0 {
		eventsCh <- block
	}

	close(filterStopCh)
}

func (c *MongoCacher) filterCreateOrUpdate(newStore map[string]interface{}, newDocumentStore map[string]string,
	eventsCh chan [eventBlockSize]MongoEvent, filterStopCh chan struct{}) {
	var block [eventBlockSize]MongoEvent
	i := 0

	for k, v := range newStore {
		ov := c.cache.Get(k)
		if ov == nil {
			if i >= eventBlockSize {
				eventsCh <- block
				block = [eventBlockSize]MongoEvent{}
				i = 0
			}

			block[i] = NewMongoEvent(k, newDocumentStore[k], rmodel.EVT_CREATE, v)
			i++

			continue
		}

		if c.isValueNotUpdated(v, ov) {
			continue
		}

		if i >= eventBlockSize {
			eventsCh <- block
			block = [eventBlockSize]MongoEvent{}
			i = 0
		}

		block[i] = NewMongoEvent(k, newDocumentStore[k], rmodel.EVT_UPDATE, v)
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

	switch c.Cfg.Key {
	case INSTANCE:
		modTime = value.(Instance).InstanceInfo.ModTimestamp
		newModTime = newValue.(Instance).InstanceInfo.ModTimestamp
	case SERVICE:
		modTime = value.(Service).ServiceInfo.ModTimestamp
		newModTime = newValue.(Service).ServiceInfo.ModTimestamp
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
		key := evt.BusinessId
		value := c.cache.Get(key)
		log.Debugf("****key:%s,  value:%s", key, value)
		ok := value != nil

		switch evt.Type {
		case rmodel.EVT_CREATE, rmodel.EVT_UPDATE:
			switch {
			case !c.IsReady():
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

			c.cache.Put(evt.BusinessId, evt.Value)
			c.cache.PutDocumentId(evt.BusinessId, evt.DocumentId)

			events[i] = evt
		case rmodel.EVT_DELETE:
			if !ok {
				log.Warnf("unexpected %s event! key %s does not cache",
					evt.Type, key)
			} else {
				evt.Value = value

				c.cache.Remove(key)
				c.cache.RemoveDocumentId(evt.DocumentId)
			}
			events[i] = evt
		}
	}
}

func (c *MongoCacher) notify(evts []MongoEvent) {
	if c.Cfg.OnEvent == nil {
		return
	}

	defer log.Recover()

	for _, evt := range evts {
		c.Cfg.OnEvent(evt)
	}
}

func NewMongoCacher(cfg *Config, cache *MongoCache) *MongoCacher {
	return &MongoCacher{
		Cfg:   cfg,
		cache: cache,
		ready: make(chan struct{}),
		lw: &innerListWatch{
			Client: client.GetMongoClient(),
			Key:    cfg.Key,
		},
		goroutine: gopool.New(context.Background()),
	}
}

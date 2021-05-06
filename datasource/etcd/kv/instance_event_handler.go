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

package kv

import (
	"context"
	"sync"
	"time"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type deferItem struct {
	ReplayAfter int32 // in seconds
	event       sd.KvEvent
}

type InstanceEventDeferHandler struct {
	Percent float64

	cache    sd.CacheReader
	once     sync.Once
	enabled  bool
	items    map[string]*deferItem
	evts     chan []sd.KvEvent
	replayCh chan sd.KvEvent
	resetCh  chan struct{}
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache sd.CacheReader, evts []sd.KvEvent) bool {
	if iedh.Percent <= 0 {
		return false
	}

	iedh.once.Do(func() {
		iedh.cache = cache
		iedh.items = make(map[string]*deferItem)
		iedh.evts = make(chan []sd.KvEvent, eventBlockSize)
		iedh.replayCh = make(chan sd.KvEvent, eventBlockSize)
		iedh.resetCh = make(chan struct{})
		gopool.Go(iedh.check)
	})

	iedh.evts <- evts
	return true
}

func (iedh *InstanceEventDeferHandler) recoverOrDefer(evt sd.KvEvent) {
	if evt.KV == nil {
		log.Errorf(nil, "defer or replayEvent a %s nil KV", evt.Type)
		return
	}
	kv := evt.KV
	key := util.BytesToStringWithNoCopy(kv.Key)
	_, ok := iedh.items[key]
	switch evt.Type {
	case discovery.EVT_CREATE, discovery.EVT_UPDATE:
		if ok {
			log.Infof("recovered key %s events", key)
			// return nil // no need to publish event to subscribers?
		}
		iedh.replayEvent(evt)
	case discovery.EVT_DELETE:
		if ok {
			return
		}

		instance := kv.Value.(*discovery.MicroServiceInstance)
		if instance == nil {
			log.Errorf(nil, "defer or replayEvent a %s nil Value, KV is %v", evt.Type, kv)
			return
		}
		ttl := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
		if ttl <= 0 || ttl > selfPreservationMaxTTL {
			ttl = selfPreservationMaxTTL
		}
		iedh.items[key] = &deferItem{
			ReplayAfter: ttl,
			event:       evt,
		}
	}
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan sd.KvEvent {
	return iedh.replayCh
}

func (iedh *InstanceEventDeferHandler) check(ctx context.Context) {
	defer log.Recover()
	t, n := time.NewTimer(deferCheckWindow), false
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Error("self preservation routine dead", nil)
			return
		case evts := <-iedh.evts:
			for _, evt := range evts {
				iedh.recoverOrDefer(evt)
			}

			del := len(iedh.items)
			if del == 0 {
				continue
			}

			if iedh.enabled {
				continue
			}

			total := iedh.cache.GetAll(nil)
			if total > selfPreservationInitCount && float64(del) >= float64(total)*iedh.Percent {
				iedh.enabled = true
				log.Warnf("self preservation is enabled, caught %d/%d(>=%.0f%%) DELETE events",
					del, total, iedh.Percent*100)
			}

			if !n {
				util.ResetTimer(t, deferCheckWindow)
				n = true
			}
		case <-t.C:
			n = false
			t.Reset(deferCheckWindow)

			if !iedh.enabled {
				for _, item := range iedh.items {
					iedh.replayEvent(item.event)
				}
				continue
			}

			iedh.ReplayEvents()
		case <-iedh.resetCh:
			iedh.ReplayEvents()
			iedh.enabled = false
			util.ResetTimer(t, deferCheckWindow)
		}
	}
}

func (iedh *InstanceEventDeferHandler) ReplayEvents() {
	interval := int32(deferCheckWindow / time.Second)
	for key, item := range iedh.items {
		item.ReplayAfter -= interval
		if item.ReplayAfter > 0 {
			continue
		}
		log.Warnf("replay delete event, remove key: %s", key)
		iedh.replayEvent(item.event)
	}
	if len(iedh.items) == 0 {
		iedh.enabled = false
		log.Warnf("self preservation stopped")
	}
}

func (iedh *InstanceEventDeferHandler) replayEvent(evt sd.KvEvent) {
	key := util.BytesToStringWithNoCopy(evt.KV.Key)
	delete(iedh.items, key)
	iedh.replayCh <- evt
}

func (iedh *InstanceEventDeferHandler) Reset() bool {
	if iedh.enabled || len(iedh.items) != 0 {
		log.Warnf("self preservation is reset")
		iedh.resetCh <- struct{}{}
		return true
	}
	return false
}

func NewInstanceEventDeferHandler() *InstanceEventDeferHandler {
	return &InstanceEventDeferHandler{Percent: selfPreservationPercentage}
}

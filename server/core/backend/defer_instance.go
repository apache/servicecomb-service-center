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
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type deferItem struct {
	ttl   *time.Timer
	event discovery.KvEvent
}

type InstanceEventDeferHandler struct {
	Percent float64

	cache     discovery.Cache
	once      sync.Once
	enabled   bool
	items     map[string]deferItem
	pendingCh chan []discovery.KvEvent
	deferCh   chan discovery.KvEvent
	resetCh   chan struct{}
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache discovery.Cache, evts []discovery.KvEvent) bool {
	if iedh.Percent <= 0 {
		return false
	}

	iedh.once.Do(func() {
		iedh.cache = cache
		iedh.items = make(map[string]deferItem)
		iedh.pendingCh = make(chan []discovery.KvEvent, eventBlockSize)
		iedh.deferCh = make(chan discovery.KvEvent, eventBlockSize)
		iedh.resetCh = make(chan struct{})
		gopool.Go(iedh.check)
	})

	iedh.pendingCh <- evts
	return true
}

func (iedh *InstanceEventDeferHandler) recoverOrDefer(evt discovery.KvEvent) error {
	kv := evt.KV
	key := util.BytesToStringWithNoCopy(kv.Key)
	_, ok := iedh.items[key]
	switch evt.Type {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		if ok {
			log.Infof("recovered key %s events", key)
			// return nil // no need to publish event to subscribers?
		}
		iedh.recover(evt)
	case pb.EVT_DELETE:
		if ok {
			return nil
		}

		instance := kv.Value.(*pb.MicroServiceInstance)
		iedh.items[key] = deferItem{
			ttl: time.NewTimer(
				time.Duration(instance.HealthCheck.Interval*(instance.HealthCheck.Times+1)) * time.Second),
			event: evt,
		}
	}
	return nil
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan discovery.KvEvent {
	return iedh.deferCh
}

func (iedh *InstanceEventDeferHandler) check(ctx context.Context) {
	defer log.Recover()

	t, n := time.NewTimer(deferCheckWindow), false
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case evts := <-iedh.pendingCh:
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
			if total > 5 && float64(del) >= float64(total)*iedh.Percent {
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

			for key, item := range iedh.items {
				if iedh.enabled {
					select {
					case <-item.ttl.C:
					default:
						continue
					}
					log.Warnf("defer handle timed out, removed key is %s", key)
				}
				iedh.recover(item.event)
			}

			if iedh.enabled && len(iedh.items) == 0 {
				iedh.renew()
				log.Warnf("self preservation is stopped")
			}

			t.Reset(deferCheckWindow)
		case <-iedh.resetCh:
			iedh.renew()
			log.Warnf("self preservation is reset")

			util.ResetTimer(t, deferCheckWindow)
		}
	}
}

func (iedh *InstanceEventDeferHandler) recover(evt discovery.KvEvent) {
	key := util.BytesToStringWithNoCopy(evt.KV.Key)
	delete(iedh.items, key)
	iedh.deferCh <- evt
}

func (iedh *InstanceEventDeferHandler) renew() {
	iedh.enabled = false
	for _, item := range iedh.items {
		item.ttl.Stop()
	}
	iedh.items = make(map[string]deferItem)
}

func (iedh *InstanceEventDeferHandler) Reset() bool {
	if iedh.enabled {
		iedh.resetCh <- struct{}{}
		return true
	}
	return false
}

func NewInstanceEventDeferHandler() *InstanceEventDeferHandler {
	return &InstanceEventDeferHandler{Percent: selfPreservationPercentage}
}

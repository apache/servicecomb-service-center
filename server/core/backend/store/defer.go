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
package store

import (
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
	"time"
)

type DeferHandler interface {
	OnCondition(Cache, []*Event) bool
	HandleChan() <-chan *Event
}

type deferItem struct {
	ttl   *time.Timer
	event *Event
}

type InstanceEventDeferHandler struct {
	Percent float64

	cache     Cache
	once      sync.Once
	enabled   bool
	items     map[string]*deferItem
	pendingCh chan []*Event
	deferCh   chan *Event
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache Cache, evts []*Event) bool {
	if iedh.Percent <= 0 {
		return false
	}

	iedh.once.Do(func() {
		iedh.cache = cache
		iedh.items = make(map[string]*deferItem, event_block_size)
		iedh.pendingCh = make(chan []*Event, event_block_size)
		iedh.deferCh = make(chan *Event, event_block_size)
		util.Go(iedh.check)
	})

	iedh.pendingCh <- evts
	return true
}

func (iedh *InstanceEventDeferHandler) recoverOrDefer(evt *Event) error {
	kv, ok := evt.Object.(*mvccpb.KeyValue)
	if !ok {
		// error type event
		return nil
	}
	key := util.BytesToStringWithNoCopy(kv.Key)
	_, ok = iedh.items[key]
	switch evt.Type {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		if ok {
			util.Logger().Infof("recovered key %s events", key)
			// return nil // no need to publish event to subscribers?
		}
		iedh.recover(evt)
	case pb.EVT_DELETE:
		if ok {
			return nil
		}

		var instance pb.MicroServiceInstance
		err := json.Unmarshal(kv.Value, &instance)
		if err != nil {
			util.Logger().Errorf(err, "unmarshal instance file failed, key is %s", key)
			return err
		}
		iedh.items[key] = &deferItem{
			ttl: time.NewTimer(
				time.Duration(instance.HealthCheck.Interval*(instance.HealthCheck.Times+1)) * time.Second),
			event: evt,
		}
	}
	return nil
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan *Event {
	return iedh.deferCh
}

func (iedh *InstanceEventDeferHandler) check(stopCh <-chan struct{}) {
	defer util.RecoverAndReport()
	t, n := iedh.newTimer(), false
	for {
		select {
		case <-stopCh:
			return
		case evts := <-iedh.pendingCh:
			for _, evt := range evts {
				iedh.recoverOrDefer(evt)
			}

			del := len(iedh.items)
			if del > 0 && !n {
				t.Stop()
				t, n = iedh.newTimer(), true
			}

			total := iedh.cache.Size()
			if del > 0 && total > 0 && float64(del) >= float64(total)*iedh.Percent {
				iedh.enabled = true
				util.Logger().Warnf(nil, "self preservation is enabled, caught %d/%d(>=%.0f%%) DELETE events",
					del, total, iedh.Percent*100)
			}
		case <-t.C:
			t, n = iedh.newTimer(), false

			if !iedh.enabled {
				for _, item := range iedh.items {
					iedh.recover(item.event)
				}
				continue
			}

			for key, item := range iedh.items {
				select {
				case <-item.ttl.C:
				default:
					continue
				}
				iedh.recover(item.event)
				util.Logger().Warnf(nil, "defer handle timed out, removed key is %s", key)
			}

			if len(iedh.items) == 0 {
				iedh.enabled = false
				util.Logger().Warnf(nil, "self preservation is stopped")
			}
		}
	}
}

func (iedh *InstanceEventDeferHandler) newTimer() *time.Timer {
	return time.NewTimer(time.Second)
}

func (iedh *InstanceEventDeferHandler) recover(evt *Event) {
	key := util.BytesToStringWithNoCopy(evt.Object.(*mvccpb.KeyValue).Key)
	delete(iedh.items, key)
	iedh.deferCh <- evt
}

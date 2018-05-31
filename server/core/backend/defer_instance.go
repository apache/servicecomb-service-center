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
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type deferItem struct {
	ttl   *time.Timer
	event KvEvent
}

type InstanceEventDeferHandler struct {
	Percent float64

	cache     Cache
	once      sync.Once
	enabled   bool
	items     map[string]deferItem
	pendingCh chan []KvEvent
	deferCh   chan KvEvent
	resetCh   chan struct{}
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache Cache, evts []KvEvent) bool {
	if iedh.Percent <= 0 {
		return false
	}

	iedh.once.Do(func() {
		iedh.cache = cache
		iedh.items = make(map[string]deferItem, eventBlockSize)
		iedh.pendingCh = make(chan []KvEvent, eventBlockSize)
		iedh.deferCh = make(chan KvEvent, eventBlockSize)
		iedh.resetCh = make(chan struct{})
		util.Go(iedh.check)
	})

	iedh.pendingCh <- evts
	return true
}

func (iedh *InstanceEventDeferHandler) recoverOrDefer(evt KvEvent) error {
	kv := evt.Object.(*mvccpb.KeyValue)
	key := util.BytesToStringWithNoCopy(kv.Key)
	_, ok := iedh.items[key]
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

		// it will happen in embed mode, and then need to get the cache value to unmarshal
		if kv.Value == nil {
			if c, ok := iedh.cache.Data(key).(*mvccpb.KeyValue); ok {
				kv.Value = c.Value
			}
		}
		err := json.Unmarshal(kv.Value, &instance)
		if err != nil {
			util.Logger().Errorf(err, "unmarshal instance file failed, key is %s, value is %s", key,
				util.BytesToStringWithNoCopy(kv.Value))
			return err
		}
		iedh.items[key] = deferItem{
			ttl: time.NewTimer(
				time.Duration(instance.HealthCheck.Interval*(instance.HealthCheck.Times+1)) * time.Second),
			event: evt,
		}
	}
	return nil
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan KvEvent {
	return iedh.deferCh
}

func (iedh *InstanceEventDeferHandler) check(ctx context.Context) {
	defer util.RecoverAndReport()
	t, n := time.NewTimer(DEFAULT_CHECK_WINDOW), false
	for {
		select {
		case <-ctx.Done():
			return
		case evts := <-iedh.pendingCh:
			for _, evt := range evts {
				iedh.recoverOrDefer(evt)
			}

			del := len(iedh.items)
			if del > 0 && !n {
				if !t.Stop() {
					<-t.C
				}
				t.Reset(DEFAULT_CHECK_WINDOW)
				n = true
			}

			total := iedh.cache.Size()
			if !iedh.enabled && del > 0 && total > 5 && float64(del) >= float64(total)*iedh.Percent {
				iedh.enabled = true
				util.Logger().Warnf(nil, "self preservation is enabled, caught %d/%d(>=%.0f%%) DELETE events",
					del, total, iedh.Percent*100)
			}
		case <-t.C:
			t.Reset(DEFAULT_CHECK_WINDOW)
			n = false

			for key, item := range iedh.items {
				if iedh.enabled {
					select {
					case <-item.ttl.C:
					default:
						continue
					}
					util.Logger().Warnf(nil, "defer handle timed out, removed key is %s", key)
				}
				iedh.recover(item.event)
			}

			if iedh.enabled && len(iedh.items) == 0 {
				iedh.renew()
				util.Logger().Warnf(nil, "self preservation is stopped")
			}
		case <-iedh.resetCh:
			iedh.renew()
			util.Logger().Warnf(nil, "self preservation is reset")
		}
	}
}

func (iedh *InstanceEventDeferHandler) recover(evt KvEvent) {
	key := util.BytesToStringWithNoCopy(evt.Object.(*mvccpb.KeyValue).Key)
	delete(iedh.items, key)
	iedh.deferCh <- evt
}

func (iedh *InstanceEventDeferHandler) renew() {
	iedh.enabled = false
	iedh.items = make(map[string]deferItem, eventBlockSize)
}

func (iedh *InstanceEventDeferHandler) Reset() bool {
	if iedh.enabled {
		iedh.resetCh <- struct{}{}
		return true
	}
	return false
}

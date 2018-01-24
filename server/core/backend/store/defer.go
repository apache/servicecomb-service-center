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
	Defer([]*Event)
	HandleChan() <-chan *Event
}

type InstanceEventDeferHandler struct {
	Percent float64

	enabled bool
	t       *time.Ticker
	cache   Cache
	deferCh chan *Event
	events  map[string]*Event
	ttls    map[string]int64
	mux     sync.RWMutex
	once    sync.Once
}

func (iedh *InstanceEventDeferHandler) deferMode(total int, del int) bool {
	return iedh.Percent > 0 && total > 0 && del > 0 &&
		float64(del/total) >= iedh.Percent
}

func (iedh *InstanceEventDeferHandler) needDefer(cache Cache, evts []*Event) bool {
	if !iedh.deferMode(cache.Size(), len(evts)) {
		return false
	}

	for _, evt := range evts {
		if evt.Type == pb.EVT_DELETE && !iedh.ticking() {
			iedh.t = time.NewTicker(time.Second)
		}
		if evt.Type != pb.EVT_DELETE {
			return false
		}
	}
	return true
}

func (iedh *InstanceEventDeferHandler) ticking() bool {
	return iedh.t != nil
}

func (iedh *InstanceEventDeferHandler) startTick() {
	iedh.t = time.NewTicker(10 * time.Second)
	util.Go(func(stopCh <-chan struct{}) {
		select {
		case <-stopCh:
			return
		case <-iedh.t.C:
		}
		iedh.t = nil

		iedh.mux.Lock()
		t, d := iedh.cache.Size(), len(iedh.events)
		if iedh.deferMode(t, d) {
			util.Logger().Warnf(nil, "self preservation is enabled, caught %d/%d(>=%.0f%%) DELETE events",
				d, t, iedh.Percent*100)
			iedh.enabled = true
		}
		iedh.mux.Unlock()
	})
}

func (iedh *InstanceEventDeferHandler) init() {
	if iedh.deferCh == nil {
		iedh.deferCh = make(chan *Event, event_block_size)
	}

	if iedh.events == nil {
		iedh.events = make(map[string]*Event, event_block_size)
		iedh.ttls = make(map[string]int64, event_block_size)
		util.Go(iedh.check)
	}
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache Cache, evts []*Event) bool {
	iedh.mux.Lock()
	if !iedh.enabled && iedh.needDefer(cache, evts) {
		util.Logger().Warnf(nil, "self preservation is enabled, caught %d(>=%.0f%%) DELETE events",
			len(evts), iedh.Percent*100)
		iedh.enabled = true
	}
	iedh.mux.Unlock()

	return iedh.enabled
}

func (iedh *InstanceEventDeferHandler) Defer(evts []*Event) {
	iedh.mux.Lock()

	iedh.once.Do(iedh.init)

	for _, evt := range evts {
		iedh.recoverOrDefer(evt)
	}

	iedh.mux.Unlock()
}

func (iedh *InstanceEventDeferHandler) recoverOrDefer(evt *Event) error {
	kv, ok := evt.Object.(*mvccpb.KeyValue)
	if !ok {
		return nil
	}
	key := util.BytesToStringWithNoCopy(kv.Key)
	switch evt.Type {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		delete(iedh.events, key)
		delete(iedh.ttls, key)

		util.Logger().Infof("recovered key %s events", key)

		iedh.deferCh <- evt
	case pb.EVT_DELETE:
		var instance pb.MicroServiceInstance
		err := json.Unmarshal(kv.Value, &instance)
		if err != nil {
			util.Logger().Errorf(err, "unmarshal instance file failed, key is %s", key)
			return err
		}
		if _, ok := iedh.ttls[key]; ok {
			return nil
		}
		iedh.events[key] = evt
		iedh.ttls[key] = int64(instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1))
	}
	return nil
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan *Event {
	return iedh.deferCh
}

func (iedh *InstanceEventDeferHandler) check(stopCh <-chan struct{}) {
	defer util.RecoverAndReport()
	for {
		select {
		case <-stopCh:
			return
		case <-time.After(time.Second):
			iedh.mux.Lock()
			for key, ttl := range iedh.ttls {
				ttl--
				if ttl > 0 {
					iedh.ttls[key] = ttl
					continue
				}

				evt := iedh.events[key]
				delete(iedh.events, key)
				delete(iedh.ttls, key)

				util.Logger().Warnf(nil, "defer handle timed out, removed key is %s", key)

				iedh.deferCh <- evt
			}
			if iedh.enabled && len(iedh.ttls) == 0 {
				iedh.enabled = false
				util.Logger().Warnf(nil, "self preservation is stopped")
			}
			iedh.mux.Unlock()
		}
	}
}

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
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	rmodel "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

type LeaseEventDeferHandler struct {
	once      sync.Once
	pendingCh chan []discovery.KvEvent
	deferCh   chan discovery.KvEvent
}

func (iedh *LeaseEventDeferHandler) OnCondition(_ discovery.CacheReader, evts []discovery.KvEvent) bool {
	iedh.once.Do(func() {
		iedh.pendingCh = make(chan []discovery.KvEvent, eventBlockSize)
		iedh.deferCh = make(chan discovery.KvEvent, eventBlockSize)
		gopool.Go(iedh.recoverLoop)
	})

	pendingEvts := make([]discovery.KvEvent, 0, eventBlockSize)
	for _, evt := range evts {
		//
		iedh.deferCh <- evt
		if evt.Type != rmodel.EVT_DELETE && evt.KV != nil && evt.KV.Lease == 0 {
			pendingEvts = append(pendingEvts, evt)
		}
	}
	if len(pendingEvts) == 0 {
		return true
	}

	iedh.pendingCh <- pendingEvts
	return true
}

func (iedh *LeaseEventDeferHandler) recoverLoop(ctx context.Context) {
	defer log.Recover()

	for {
		select {
		case <-ctx.Done():
			return
		case evts, ok := <-iedh.pendingCh:
			if !ok {
				log.Error("lease pending chan is closed!", nil)
				return
			}
			log.Info(fmt.Sprintf("start to recover leases[%d]", len(evts)))

			l := 0
			for _, evt := range evts {
				if iedh.recoverLease(evt) {
					l++
				}
			}
			if l > 0 {
				log.Info(fmt.Sprintf("recover leases[%d] successfully", l))
			}
		}
	}
}

func (iedh *LeaseEventDeferHandler) recoverLease(evt discovery.KvEvent) (ok bool) {
	ctx := context.Background()
	kv := evt.KV
	key := util.BytesToStringWithNoCopy(kv.Key)
	instance, ok := kv.Value.(*rmodel.MicroServiceInstance)
	if !ok {
		log.Error(fmt.Sprintf("[%s] value covert to MicroServiceInstance failed", key), nil)
		return
	}
	data, err := json.Marshal(kv.Value)
	if err != nil {
		log.Error(fmt.Sprintf("[%s]Marshal instance failed", key), err)
		return
	}

	ttl := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
	leaseID, err := Registry().LeaseGrant(ctx, int64(ttl))
	if err != nil {
		log.Error(fmt.Sprintf("[%s]grant lease failed", key), err)
		return
	}
	serviceID, instanceID, domainProject := apt.GetInfoFromInstKV(kv.Key)
	hbKey := apt.GenerateInstanceLeaseKey(domainProject, serviceID, instanceID)
	opts := []registry.PluginOp{
		registry.OpPut(registry.WithStrKey(key), registry.WithValue(data),
			registry.WithLease(leaseID)),
		registry.OpPut(registry.WithStrKey(hbKey), registry.WithStrValue(fmt.Sprintf("%d", leaseID)),
			registry.WithLease(leaseID)),
	}

	resp, err := Registry().TxnWithCmp(ctx, opts,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
			registry.CmpNotEqual, 0)}, nil)
	if err != nil {
		log.Error(fmt.Sprintf("recover instance lease failed, serviceID %s, instanceID %s", serviceID, instanceID), nil)
		return
	}
	if !resp.Succeeded {
		log.Error(fmt.Sprintf("recover instance lease failed, serviceID %s, instanceID %s: service does not exist", serviceID, instanceID), nil)
		return
	}
	return true
}

func (iedh *LeaseEventDeferHandler) HandleChan() <-chan discovery.KvEvent {
	return iedh.deferCh
}

func (iedh *LeaseEventDeferHandler) Reset() bool {
	return false
}

func NewLeaseEventDeferHandler() *LeaseEventDeferHandler {
	return &LeaseEventDeferHandler{}
}

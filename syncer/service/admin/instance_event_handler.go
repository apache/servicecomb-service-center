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

package admin

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/disco"
)

// InstanceEventHandler 对端SC同步异常期间，无法查询到在对端SC注册的实例，因此这期间对端SC的实例的事件，对客户端而言是无效事件，
// 需要在对端SC同步恢复后，处理一次
type InstanceEventHandler struct {
	handler       kvstore.EventHandler
	interval      time.Duration
	healthChecker *HealthChecker

	events     map[string]kvstore.Event
	eventsLock sync.Mutex
}

func (h *InstanceEventHandler) Type() kvstore.Type {
	return sd.TypeInstance
}

func (h *InstanceEventHandler) OnEvent(evt kvstore.Event) {
	if h.healthChecker.ShouldTrustPeerServer() {
		return
	}
	instance, ignore := checkShouldIgnoreEvent(evt)
	if ignore {
		return
	}
	h.add(evt, instance)
}

func (h *InstanceEventHandler) exportAndClearEvents() []kvstore.Event {
	h.eventsLock.Lock()
	defer h.eventsLock.Unlock()
	result := make([]kvstore.Event, 0, len(h.events))
	for _, evt := range h.events {
		result = append(result, evt)
	}
	h.events = make(map[string]kvstore.Event)
	return result
}

func (h *InstanceEventHandler) add(evt kvstore.Event, instance *pb.MicroServiceInstance) {
	h.eventsLock.Lock()
	defer h.eventsLock.Unlock()
	_, ok := h.events[instance.ServiceId] // 每个服务仅保留一个事件
	if !ok {
		evt.Type = pb.EVT_UPDATE // 统一改为刷新事件，避免影响配额
		h.events[instance.ServiceId] = evt
	}
}

func (h *InstanceEventHandler) run() {
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(h.interval):
				if !h.healthChecker.ShouldTrustPeerServer() {
					continue
				}
				// 对端SC恢复后，再重新处理事件
				events := h.exportAndClearEvents()
				h.handleAgain(events)
			}
		}
	})
}
func (h *InstanceEventHandler) handleAgain(events []kvstore.Event) {
	for _, evt := range events {
		h.handler.OnEvent(evt)
	}
}

func NewInstanceEventHandler(healthChecker *HealthChecker, interval time.Duration, handler kvstore.EventHandler) *InstanceEventHandler {
	h := &InstanceEventHandler{
		handler:       handler,
		events:        make(map[string]kvstore.Event),
		eventsLock:    sync.Mutex{},
		interval:      interval,
		healthChecker: healthChecker,
	}
	h.run()
	return h
}

func checkShouldIgnoreEvent(evt kvstore.Event) (storedInstance *pb.MicroServiceInstance, ignore bool) {
	action := evt.Type
	// 初始化是内部事件，忽略
	// 本来就查不到，推了删除事件后还是查不到，因此删除事件也忽略
	if action == pb.EVT_INIT || action == pb.EVT_DELETE {
		return nil, true
	}
	instance, ok := evt.KV.Value.(*pb.MicroServiceInstance)
	if !ok {
		log.Error("failed to assert microServiceInstance", datasource.ErrAssertFail)
		return nil, true
	}
	// 连接在本SC的实例，忽略
	if util.IsMapFullyMatch(instance.Properties, disco.GetInnerProperties()) {
		return nil, true
	}
	log.Info(fmt.Sprintf("caught [%s] service[%s] instance[%s] event, endpoints %v, will handle it again when peer sc recovery",
		action, instance.ServiceId, instance.InstanceId, instance.Endpoints))
	return instance, false
}

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

package event

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	pb "github.com/go-chassis/cari/discovery"
)

const (
	msKeyPrefix = "/cse-sr/ms/files/"
	sep         = "/"
)

// InstanceEventHandler is the handler to handle:
// 1. report instance metrics
// 2. recover the instance quota
// 3. publish the instance events to the subscribers
// 4. reset the find instance cache
type InstanceEventHandler struct {
}

func (h *InstanceEventHandler) Type() kvstore.Type {
	return sd.TypeInstance
}

func (h *InstanceEventHandler) OnEvent(evt kvstore.Event) {
	action := evt.Type
	instance, ok := evt.KV.Value.(*pb.MicroServiceInstance)
	if !ok {
		log.Error("failed to assert microServiceInstance", datasource.ErrAssertFail)
		return
	}
	providerID, providerInstanceID, domainProject := path.GetInfoFromInstKV(evt.KV.Key)
	ctx := util.WithGlobal(util.WithCacheOnly(context.Background()))

	if action == pb.EVT_INIT {
		return
	}

	if action == pb.EVT_DELETE && !datasource.IsDefaultDomainProject(domainProject) {
		domain, project := path.SplitDomainProject(domainProject)
		quotasvc.RemandInstance(util.SetDomainProject(context.Background(), domain, project))
	}

	// 查询服务版本信息
	ms, err := serviceUtil.GetService(ctx, domainProject, providerID)
	if err != nil {
		log.Error(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, get cached provider's file failed",
			action, providerID, providerInstanceID, instance.Endpoints), err)
		return
	}

	if !syncernotify.GetSyncerNotifyCenter().Closed() {
		NotifySyncerInstanceEvent(evt, domainProject, ms)
	}

	if event.Center().Closed() {
		log.Warn(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, but notify service is closed",
			action, providerID, providerInstanceID, instance.Endpoints))
		return
	}

	log.Info(fmt.Sprintf("caught [%s] service[%s][%s/%s/%s/%s] instance[%s] event, endpoints %v",
		action, providerID, ms.Environment, ms.AppId, ms.ServiceName, ms.Version,
		providerInstanceID, instance.Endpoints))

	// 查询所有consumer
	consumerIDs, err := serviceUtil.GetConsumerIds(ctx, domainProject, ms)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s consumerIDs failed",
			providerID, ms.Environment, ms.AppId, ms.ServiceName, ms.Version), err)
		return
	}

	PublishInstanceEvent(evt, pb.MicroServiceToKey(domainProject, ms), consumerIDs)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt kvstore.Event, serviceKey *pb.MicroServiceKey, subscribers []string) {
	defer cache.FindInstances.Remove(serviceKey)

	if len(subscribers) == 0 {
		return
	}

	response := &pb.WatchInstanceResponse{
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.KV.Value.(*pb.MicroServiceInstance),
	}
	for _, consumerID := range subscribers {
		evt := event.NewInstanceEvent(consumerID, evt.Revision, evt.CreateAt, response)
		err := event.Center().Fire(evt)
		if err != nil {
			log.Error(fmt.Sprintf("publish event[%v] into channel failed", evt), err)
		}
	}
}

func NotifySyncerInstanceEvent(evt kvstore.Event, domainProject string, ms *pb.MicroService) {
	msInstance := evt.KV.Value.(*pb.MicroServiceInstance)

	serviceKey := msKeyPrefix + domainProject + sep + ms.ServiceId
	msKV := &dump.KV{Key: serviceKey, ClusterName: evt.KV.ClusterName}
	service := &dump.Microservice{KV: msKV, Value: ms}

	instKey := string(evt.KV.Key)
	instKV := &dump.KV{Key: instKey, ClusterName: evt.KV.ClusterName}
	instance := &dump.Instance{KV: instKV, Value: msInstance}

	instEvent := &dump.WatchInstanceChangedEvent{
		Action:   string(evt.Type),
		Service:  service,
		Instance: instance,
	}

	syncernotify.GetSyncerNotifyCenter().AddEvent(instEvent)

	log.Debug(fmt.Sprintf("success to add instance change event: %v to event queue", instEvent))
}

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
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	pb "github.com/go-chassis/cari/discovery"
)

const (
	increaseOne = 1
	decreaseOne = -1
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

func (h *InstanceEventHandler) Type() sd.Type {
	return kv.INSTANCE
}

func (h *InstanceEventHandler) OnEvent(evt sd.KvEvent) {
	action := evt.Type
	instance := evt.KV.Value.(*pb.MicroServiceInstance)
	providerID, providerInstanceID, domainProject := path.GetInfoFromInstKV(evt.KV.Key)
	idx := strings.Index(domainProject, "/")
	domainName := domainProject[:idx]
	projectName := domainProject[idx+1:]

	ctx := util.WithGlobal(util.WithCacheOnly(context.Background()))

	var count float64 = increaseOne
	if action == pb.EVT_INIT {
		metrics.ReportInstances(domainName, count)
		ms, err := serviceUtil.GetService(ctx, domainProject, providerID)
		if err != nil {
			log.Warnf("caught [%s] instance[%s/%s] event, endpoints %v, get cached provider's file failed",
				action, providerID, providerInstanceID, instance.Endpoints)
			return
		}
		frameworkName, frameworkVersion := getFramework(ms)
		metrics.ReportFramework(domainName, projectName, frameworkName, frameworkVersion, count)
		return
	}

	if action == pb.EVT_DELETE {
		count = decreaseOne
		if !core.IsDefaultDomainProject(domainProject) {
			projectName := domainProject[idx+1:]
			serviceUtil.RemandInstanceQuota(
				util.SetDomainProject(context.Background(), domainName, projectName))
		}
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

	if action != pb.EVT_UPDATE {
		frameworkName, frameworkVersion := getFramework(ms)
		metrics.ReportInstances(domainName, count)
		metrics.ReportFramework(domainName, projectName, frameworkName, frameworkVersion, count)
	}

	log.Infof("caught [%s] service[%s][%s/%s/%s/%s] instance[%s] event, endpoints %v",
		action, providerID, ms.Environment, ms.AppId, ms.ServiceName, ms.Version,
		providerInstanceID, instance.Endpoints)

	// 查询所有consumer
	consumerIDs, _, err := serviceUtil.GetAllConsumerIds(ctx, domainProject, ms)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s consumerIDs failed",
			providerID, ms.Environment, ms.AppId, ms.ServiceName, ms.Version)
		return
	}

	PublishInstanceEvent(evt, domainProject, pb.MicroServiceToKey(domainProject, ms), consumerIDs)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt sd.KvEvent, domainProject string, serviceKey *pb.MicroServiceKey, subscribers []string) {
	defer cache.FindInstances.Remove(serviceKey)

	if len(subscribers) == 0 {
		return
	}

	response := &pb.WatchInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Watch instance successfully."),
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.KV.Value.(*pb.MicroServiceInstance),
	}
	for _, consumerID := range subscribers {
		evt := event.NewInstanceEventWithTime(consumerID, domainProject, evt.Revision, evt.CreateAt, response)
		err := event.Center().Fire(evt)
		if err != nil {
			log.Errorf(err, "publish event[%v] into channel failed", evt)
		}
	}
}

func NotifySyncerInstanceEvent(evt sd.KvEvent, domainProject string, ms *pb.MicroService) {
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

	log.Debugf("success to add instance change event:%s to event queue", instEvent)
}

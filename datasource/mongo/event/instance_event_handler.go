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
	"time"

	simple "github.com/apache/servicecomb-service-center/pkg/time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	"github.com/go-chassis/cari/discovery"
)

// InstanceEventHandler is the handler to handle events
//as instance registry or instance delete, and notify syncer
type InstanceEventHandler struct {
}

func (h InstanceEventHandler) Type() string {
	return model.CollectionInstance
}

func (h InstanceEventHandler) OnEvent(evt sd.MongoEvent) {
	action := evt.Type
	instance, ok := evt.Value.(model.Instance)
	if !ok {
		log.Error("failed to assert instance", datasource.ErrAssertFail)
		return
	}
	providerID := instance.Instance.ServiceId
	providerInstanceID := instance.Instance.InstanceId
	domainProject := instance.Domain + "/" + instance.Project
	ctx := util.SetDomainProject(context.Background(), instance.Domain, instance.Project)

	res, err := mongo.GetServiceByID(ctx, providerID)
	if err != nil {
		log.Error(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, get provider's file failed from db\n",
			action, providerID, providerInstanceID, instance.Instance.Endpoints), err)
	}
	if res == nil {
		return
	}
	microService := res.Service
	if action == discovery.EVT_INIT {
		return
	}
	if !syncernotify.GetSyncerNotifyCenter().Closed() {
		NotifySyncerInstanceEvent(evt, microService)
	}
	consumerIDs, err := mongo.GetConsumerIDs(ctx, microService)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s consumerIDs failed",
			providerID, microService.Environment, microService.AppId, microService.ServiceName, microService.Version), err)
		return
	}
	PublishInstanceEvent(evt, discovery.MicroServiceToKey(domainProject, microService), consumerIDs)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt sd.MongoEvent, serviceKey *discovery.MicroServiceKey, subscribers []string) {
	if len(subscribers) == 0 {
		return
	}
	response := &discovery.WatchInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Watch instance successfully."),
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.Value.(model.Instance).Instance,
	}
	for _, consumerID := range subscribers {
		evt := event.NewInstanceEvent(consumerID, -1, simple.FromTime(time.Now()), response)
		err := event.Center().Fire(evt)
		if err != nil {
			log.Error(fmt.Sprintf("publish event[%v] into channel failed", evt), err)
		}
	}
}

func NotifySyncerInstanceEvent(event sd.MongoEvent, microService *discovery.MicroService) {
	instance := event.Value.(model.Instance).Instance
	log.Info(fmt.Sprintf("instanceId : %s and serviceId : %s in NotifySyncerInstanceEvent", instance.InstanceId, instance.ServiceId))
	instanceKey := util.StringJoin([]string{datasource.InstanceKeyPrefix, event.Value.(model.Instance).Domain,
		event.Value.(model.Instance).Project, instance.ServiceId, instance.InstanceId}, datasource.SPLIT)

	instanceKv := dump.KV{
		Key:   instanceKey,
		Value: instance,
	}

	dumpInstance := dump.Instance{
		KV:    &instanceKv,
		Value: instance,
	}
	serviceKey := util.StringJoin([]string{datasource.ServiceKeyPrefix, event.Value.(model.Instance).Domain,
		event.Value.(model.Instance).Project, instance.ServiceId}, datasource.SPLIT)
	serviceKv := dump.KV{
		Key:   serviceKey,
		Value: microService,
	}

	dumpService := dump.Microservice{
		KV:    &serviceKv,
		Value: microService,
	}

	instEvent := &dump.WatchInstanceChangedEvent{
		Action:   string(event.Type),
		Service:  &dumpService,
		Instance: &dumpInstance,
	}
	syncernotify.GetSyncerNotifyCenter().AddEvent(instEvent)

	log.Debug(fmt.Sprintf("success to add instance change event action [%s], instanceKey : %s to event queue",
		instEvent.Action, instanceKey))
}

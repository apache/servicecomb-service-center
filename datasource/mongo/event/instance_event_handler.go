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
	"errors"
	"fmt"
	"time"

	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/server/notify"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
)

// InstanceEventHandler is the handler to handle events
//as instance registry or instance delete, and notify syncer
type InstanceEventHandler struct {
}

func (h InstanceEventHandler) Type() string {
	return mongo.CollectionInstance
}

func (h InstanceEventHandler) OnEvent(evt sd.MongoEvent) {
	action := evt.Type
	instance := evt.Value.(sd.Instance)
	providerID := instance.Instance.ServiceId
	providerInstanceID := instance.Instance.InstanceId
	domainProject := instance.Domain + "/" + instance.Project
	cacheService := sd.Store().Service().Cache().Get(providerID)
	var microService *discovery.MicroService
	if cacheService != nil {
		microService = cacheService.(sd.Service).Service
	}
	if microService == nil {
		log.Info("get cached service failed, then get from database")
		service, err := mongo.GetService(context.Background(), bson.M{"serviceinfo.serviceid": providerID})
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Warn(fmt.Sprintf("there is no service with id [%s] in the database", providerID))
			} else {
				log.Error("query database error", err)
			}
			return
		}
		microService = service.Service // service in the cache may not ready, query from db once
		if microService == nil {
			log.Warn(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, get provider's file failed from db\n",
				action, providerID, providerInstanceID, instance.Instance.Endpoints))
			return
		}
	}
	if !syncernotify.GetSyncerNotifyCenter().Closed() {
		NotifySyncerInstanceEvent(evt, microService)
	}
	ctx := util.SetDomainProject(context.Background(), instance.Domain, instance.Project)
	consumerIDS, _, err := mongo.GetAllConsumerIds(ctx, microService)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s consumerIDs failed",
			providerID, microService.Environment, microService.AppId, microService.ServiceName, microService.Version), err)
		return
	}
	PublishInstanceEvent(evt, domainProject, discovery.MicroServiceToKey(domainProject, microService), consumerIDS)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt sd.MongoEvent, domainProject string, serviceKey *discovery.MicroServiceKey, subscribers []string) {
	if len(subscribers) == 0 {
		return
	}
	response := &discovery.WatchInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Watch instance successfully."),
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.Value.(sd.Instance).Instance,
	}
	for _, consumerID := range subscribers {
		evt := notify.NewInstanceEventWithTime(consumerID, domainProject, -1, simple.FromTime(time.Now()), response)
		err := notify.Center().Publish(evt)
		if err != nil {
			log.Error(fmt.Sprintf("publish event[%v] into channel failed", evt), err)
		}
	}
}

func NotifySyncerInstanceEvent(event sd.MongoEvent, microService *discovery.MicroService) {
	instance := event.Value.(sd.Instance).Instance
	log.Info(fmt.Sprintf("instanceId : %s and serviceId : %s in NotifySyncerInstanceEvent", instance.InstanceId, instance.ServiceId))
	instanceKey := util.StringJoin([]string{datasource.InstanceKeyPrefix, event.Value.(sd.Instance).Domain,
		event.Value.(sd.Instance).Project, instance.ServiceId, instance.InstanceId}, datasource.SPLIT)

	instanceKv := dump.KV{
		Key:   instanceKey,
		Value: instance,
	}

	dumpInstance := dump.Instance{
		KV:    &instanceKv,
		Value: instance,
	}
	serviceKey := util.StringJoin([]string{datasource.ServiceKeyPrefix, event.Value.(sd.Instance).Domain,
		event.Value.(sd.Instance).Project, instance.ServiceId}, datasource.SPLIT)
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

	log.Debug(fmt.Sprintf("success to add instance change event action [%s], instanceKey : %s to event queue", instEvent.Action, instanceKey))
}

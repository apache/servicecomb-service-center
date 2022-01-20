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

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
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

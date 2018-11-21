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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	nf "github.com/apache/servicecomb-service-center/server/service/notification"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

type TagsChangedTask struct {
	key string
	err error

	DomainProject string
	consumerId    string
	Rev           int64
}

func (apt *TagsChangedTask) Key() string {
	return apt.key
}

func (apt *TagsChangedTask) Do(ctx context.Context) error {
	apt.err = apt.publish(ctx, apt.DomainProject, apt.consumerId, apt.Rev)
	return apt.err
}

func (apt *TagsChangedTask) Err() error {
	return apt.err
}

func (apt *TagsChangedTask) publish(ctx context.Context, domainProject, consumerId string, rev int64) error {
	ctx = context.WithValue(context.WithValue(ctx,
		serviceUtil.CTX_CACHEONLY, "1"),
		serviceUtil.CTX_GLOBAL, "1")

	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerId)
	if err != nil {
		log.Errorf(err, "get consumer[%s] for publish event failed", consumerId)
		return err
	}
	if consumer == nil {
		log.Errorf(nil, "consumer[%s] does not exist", consumerId)
		return fmt.Errorf("consumer[%s] does not exist", consumerId)
	}

	serviceKey := pb.MicroServiceToKey(domainProject, consumer)
	cache.FindInstances.Remove(serviceKey)

	providerIds, err := serviceUtil.GetProviderIds(ctx, domainProject, consumer)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s providerIds failed",
			consumerId, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
		return err
	}

	for _, providerId := range providerIds {
		provider, err := serviceUtil.GetService(ctx, domainProject, providerId)
		if provider == nil {
			log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s provider[%s] file failed",
				consumerId, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version, providerId)
			continue
		}

		providerKey := pb.MicroServiceToKey(domainProject, provider)
		PublishInstanceEvent(domainProject, pb.EVT_EXPIRE, providerKey, nil, rev, []string{consumerId})
	}
	return nil
}

type TagEventHandler struct {
}

func (h *TagEventHandler) Type() discovery.Type {
	return backend.SERVICE_TAG
}

func (h *TagEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	consumerId, domainProject := core.GetInfoFromTagKV(evt.KV.Key)

	if nf.GetNotifyService().Closed() {
		log.Warnf("caught [%s] service tags[%s/%s] event, but notify service is closed",
			action, consumerId, evt.KV.Value)
		return
	}
	log.Infof("caught [%s] service tags[%s/%s] event", action, consumerId, evt.KV.Value)

	task.Service().Add(context.Background(),
		NewTagsChangedAsyncTask(domainProject, consumerId, evt.Revision))
}

func NewTagEventHandler() *TagEventHandler {
	return &TagEventHandler{}
}

func NewTagsChangedAsyncTask(domainProject, consumerId string, rev int64) *TagsChangedTask {
	return &TagsChangedTask{
		key:           "TagsChangedAsyncTask_" + consumerId,
		DomainProject: domainProject,
		consumerId:    consumerId,
		Rev:           rev,
	}
}

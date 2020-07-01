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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

type TagsChangedTask struct {
	discovery.KvEvent

	key string
	err error

	DomainProject string
	ConsumerID    string
}

func (apt *TagsChangedTask) Key() string {
	return apt.key
}

func (apt *TagsChangedTask) Do(ctx context.Context) error {
	apt.err = apt.publish(ctx, apt.DomainProject, apt.ConsumerID)
	return apt.err
}

func (apt *TagsChangedTask) Err() error {
	return apt.err
}

func (apt *TagsChangedTask) publish(ctx context.Context, domainProject, consumerID string) error {
	ctx = context.WithValue(context.WithValue(ctx,
		util.CtxCacheOnly, "1"),
		util.CtxGlobal, "1")

	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerID)
	if err != nil {
		log.Errorf(err, "get consumer[%s] for publish event failed", consumerID)
		return err
	}
	if consumer == nil {
		log.Errorf(nil, "consumer[%s] does not exist", consumerID)
		return fmt.Errorf("consumer[%s] does not exist", consumerID)
	}

	serviceKey := pb.MicroServiceToKey(domainProject, consumer)
	cache.FindInstances.Remove(serviceKey)

	providerIDs, err := serviceUtil.GetProviderIds(ctx, domainProject, consumer)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s providerIDs failed",
			consumerID, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
		return err
	}

	for _, providerID := range providerIDs {
		provider, err := serviceUtil.GetService(ctx, domainProject, providerID)
		if provider == nil {
			log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s provider[%s] file failed",
				consumerID, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version, providerID)
			continue
		}

		providerKey := pb.MicroServiceToKey(domainProject, provider)
		PublishInstanceEvent(apt.KvEvent, domainProject, providerKey, []string{consumerID})
	}
	return nil
}

// TagEventHandler is the handler to handle:
// 1. publish the EVT_EXPIRE event to subscribers when tag is changed
// 2. reset the find instance cache
type TagEventHandler struct {
}

func (h *TagEventHandler) Type() discovery.Type {
	return backend.ServiceTag
}

func (h *TagEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	consumerID, domainProject := core.GetInfoFromTagKV(evt.KV.Key)

	if notify.GetNotifyCenter().Closed() {
		log.Warnf("caught [%s] service tags[%s/%s] event, but notify service is closed",
			action, consumerID, evt.KV.Value)
		return
	}
	log.Infof("caught [%s] service tags[%s/%s] event", action, consumerID, evt.KV.Value)

	err := task.GetService().Add(context.Background(),
		NewTagsChangedAsyncTask(domainProject, consumerID, evt))
	if err != nil {
		log.Error("", err)
	}
}

func NewTagEventHandler() *TagEventHandler {
	return &TagEventHandler{}
}

func NewTagsChangedAsyncTask(domainProject, consumerID string, evt discovery.KvEvent) *TagsChangedTask {
	evt.Type = pb.EVT_EXPIRE
	return &TagsChangedTask{
		KvEvent:       evt,
		key:           "TagsChangedAsyncTask_" + consumerID,
		DomainProject: domainProject,
		ConsumerID:    consumerID,
	}
}

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

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
)

type TagsChangedTask struct {
	sd.KvEvent

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
	ctx = util.WithGlobal(util.WithCacheOnly(ctx))

	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("consumer[%s] does not exist in db", consumerID))
		} else {
			log.Error(fmt.Sprintf("get consumer[%s] for publish event failed", consumerID), err)
		}
		return err
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
		if err != nil {
			log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s provider[%s] file failed",
				consumerID, consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version, providerID), err)
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

func (h *TagEventHandler) Type() sd.Type {
	return kv.ServiceTag
}

func (h *TagEventHandler) OnEvent(evt sd.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	consumerID, domainProject := path.GetInfoFromTagKV(evt.KV.Key)

	if event.Center().Closed() {
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

func NewTagsChangedAsyncTask(domainProject, consumerID string, evt sd.KvEvent) *TagsChangedTask {
	evt.Type = pb.EVT_EXPIRE
	return &TagsChangedTask{
		KvEvent:       evt,
		key:           "TagsChangedAsyncTask_" + consumerID,
		DomainProject: domainProject,
		ConsumerID:    consumerID,
	}
}

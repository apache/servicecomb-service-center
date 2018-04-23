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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	nf "github.com/apache/incubator-servicecomb-service-center/server/service/notification"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
)

type TagsChangedAsyncTask struct {
	key string
	err error

	DomainProject string
	consumerId    string
	Rev           int64
}

func (apt *TagsChangedAsyncTask) Key() string {
	return apt.key
}

func (apt *TagsChangedAsyncTask) Do(ctx context.Context) error {
	defer store.AsyncTaskService().DeferRemove(apt.Key())
	apt.err = apt.publish(ctx, apt.DomainProject, apt.consumerId, apt.Rev)
	return apt.err
}

func (apt *TagsChangedAsyncTask) Err() error {
	return apt.err
}

func (apt *TagsChangedAsyncTask) publish(ctx context.Context, domainProject, consumerId string, rev int64) error {
	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerId)
	if err != nil {
		util.Logger().Errorf(err, "get comsumer for publish event %s failed", consumerId)
		return err
	}
	if consumer == nil {
		util.Logger().Errorf(nil, "service not exist, %s", consumerId)
		return fmt.Errorf("service not exist, %s", consumerId)
	}
	providerIds, err := serviceUtil.GetProvidersInCache(ctx, domainProject, consumer)
	if err != nil {
		util.Logger().Errorf(err, "get provider services by consumer %s failed", consumerId)
		return err
	}

	for _, providerId := range providerIds {
		provider, err := serviceUtil.GetService(ctx, domainProject, providerId)
		if provider == nil {
			util.Logger().Warnf(err, "get service %s file failed", providerId)
			continue
		}
		nf.PublishInstanceEvent(domainProject, pb.EVT_EXPIRE,
			&pb.MicroServiceKey{
				Environment: provider.Environment,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
				Version:     provider.Version,
			}, nil, rev, []string{consumerId})
	}
	return nil
}

type TagEventHandler struct {
}

func (h *TagEventHandler) Type() store.StoreType {
	return store.SERVICE_TAG
}

func (h *TagEventHandler) OnEvent(evt store.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	kv := evt.Object.(*mvccpb.KeyValue)
	consumerId, domainProject, data := pb.GetInfoFromTagKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal service rule file failed, service %s tags [%s] event, data is nil",
			consumerId, action)
		return
	}

	if nf.GetNotifyService().Closed() {
		util.Logger().Warnf(nil, "caught service %s tags [%s] event, but notify service is closed",
			consumerId, action)
		return
	}
	util.Logger().Infof("caught service %s tags [%s] event", consumerId, action)

	var rule pb.ServiceRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		util.Logger().Errorf(err, "unmarshal service %s tags file failed", consumerId)
		return
	}

	store.AsyncTaskService().Add(context.Background(),
		NewTagsChangedAsyncTask(domainProject, consumerId, evt.Revision))
}

func NewTagEventHandler() *TagEventHandler {
	return &TagEventHandler{}
}

func NewTagsChangedAsyncTask(domainProject, consumerId string, rev int64) *TagsChangedAsyncTask {
	return &TagsChangedAsyncTask{
		key:           "TagsChangedAsyncTask_" + consumerId,
		DomainProject: domainProject,
		consumerId:    consumerId,
		Rev:           rev,
	}
}

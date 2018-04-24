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

type RulesChangedAsyncTask struct {
	key string
	err error

	DomainProject string
	ProviderId    string
	Rev           int64
}

func (apt *RulesChangedAsyncTask) Key() string {
	return apt.key
}

func (apt *RulesChangedAsyncTask) Do(ctx context.Context) error {
	defer store.AsyncTaskService().DeferRemove(apt.Key())
	apt.err = apt.publish(ctx, apt.DomainProject, apt.ProviderId, apt.Rev)
	return apt.err
}

func (apt *RulesChangedAsyncTask) Err() error {
	return apt.err
}

func (apt *RulesChangedAsyncTask) publish(ctx context.Context, domainProject, providerId string, rev int64) error {
	provider, err := serviceUtil.GetService(ctx, domainProject, providerId)
	if err != nil {
		util.Logger().Errorf(err, "get provider %s service file failed", providerId)
		return err
	}
	if provider == nil {
		util.Logger().Errorf(nil, "provider %s does not exist", providerId)
		return fmt.Errorf("provider %s does not exist", providerId)
	}

	consumerIds, err := serviceUtil.GetConsumersInCache(ctx, domainProject, provider)
	if err != nil {
		util.Logger().Errorf(err, "get consumer services by provider %s failed", providerId)
		return err
	}
	providerKey := pb.MicroServiceToKey(domainProject, provider)

	nf.PublishInstanceEvent(domainProject, pb.EVT_EXPIRE, providerKey, nil, rev, consumerIds)
	return nil
}

type RuleEventHandler struct {
}

func (h *RuleEventHandler) Type() store.StoreType {
	return store.RULE
}

func (h *RuleEventHandler) OnEvent(evt store.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	kv := evt.Object.(*mvccpb.KeyValue)
	providerId, ruleId, domainProject, data := pb.GetInfoFromRuleKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal service rule file failed, service %s rule %s [%s] event, data is nil",
			providerId, ruleId, action)
		return
	}

	if nf.GetNotifyService().Closed() {
		util.Logger().Warnf(nil, "caught service %s rule %s [%s] event, but notify service is closed",
			providerId, ruleId, action)
		return
	}
	util.Logger().Infof("caught service %s rule %s [%s] event", providerId, ruleId, action)

	var rule pb.ServiceRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		util.Logger().Errorf(err, "unmarshal service %s rule %s file failed",
			providerId, ruleId)
		return
	}

	store.AsyncTaskService().Add(context.Background(),
		NewRulesChangedAsyncTask(domainProject, providerId, evt.Revision))
}

func NewRuleEventHandler() *RuleEventHandler {
	return &RuleEventHandler{}
}

func NewRulesChangedAsyncTask(domainProject, providerId string, rev int64) *RulesChangedAsyncTask {
	return &RulesChangedAsyncTask{
		key:           "RulesChangedAsyncTask_" + providerId,
		DomainProject: domainProject,
		ProviderId:    providerId,
		Rev:           rev,
	}
}

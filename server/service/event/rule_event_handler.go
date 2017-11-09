//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package event

import (
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
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
	defer store.Store().AsyncTasker().DeferRemoveTask(apt.Key())
	apt.err = apt.publish(ctx, apt.DomainProject, apt.ProviderId, apt.Rev)
	return apt.err
}

func (apt *RulesChangedAsyncTask) Err() error {
	return apt.err
}

func (apt *RulesChangedAsyncTask) publish(ctx context.Context, domainProject, providerId string, rev int64) error {
	provider, err := serviceUtil.GetService(ctx, domainProject, providerId)
	if err != nil {
		util.Logger().Errorf(err, "get service %s file failed", providerId)
		return err
	}
	if provider == nil {
		tmpProvider, found := serviceUtil.MsCache().Get(providerId)
		if !found {
			util.Logger().Errorf(nil, "service not exist, %s", providerId)
			return fmt.Errorf("service not exist, %s", providerId)
		}
		provider = tmpProvider.(*pb.MicroService)
	}

	consumerIds, err := serviceUtil.GetConsumersInCache(ctx, domainProject, providerId, provider)
	if err != nil {
		util.Logger().Errorf(err, "get consumer services by provider %s failed", providerId)
		return err
	}
	providerKey := pb.ToMicroServiceKey(domainProject, provider)

	nf.PublishInstanceEvent(domainProject, pb.EVT_EXPIRE, providerKey, nil, rev, consumerIds)
	return nil
}

type RuleEventHandler struct {
}

func (h *RuleEventHandler) Type() store.StoreType {
	return store.RULE
}

func (h *RuleEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action
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

	store.Store().AsyncTasker().AddTask(context.Background(),
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

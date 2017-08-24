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
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
)

type RulesChangedAsyncTask struct {
	key string
	err error

	Tenant     string
	ProviderId string
	Rev        int64
	Service    *nf.NotifyService
}

func (apt *RulesChangedAsyncTask) Key() string {
	return apt.key
}

func (apt *RulesChangedAsyncTask) Do(ctx context.Context) error {
	defer store.Store().AsyncTasker().DeferRemoveTask(apt.Key())
	apt.err = publish(ctx, apt.Service, apt.Tenant, apt.ProviderId, apt.Rev)
	return apt.err
}

func (apt *RulesChangedAsyncTask) Err() error {
	return apt.err
}

func publish(ctx context.Context, service *nf.NotifyService, tenant, providerId string, rev int64) error {
	provider, err := ms.GetService(ctx, tenant, providerId)
	if provider == nil {
		util.LOGGER.Errorf(err, "get service %s file failed", providerId)
		return err
	}

	allow, deny, err := serviceUtil.GetConsumerIds(ctx, tenant, provider)
	if err != nil {
		util.LOGGER.Errorf(err, "get consumer services by provider %s failed", providerId)
		return err
	}

	providerKey := &pb.MicroServiceKey{
		AppId:       provider.AppId,
		ServiceName: provider.ServiceName,
		Version:     provider.Version,
	}

	instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, tenant, providerId, "")
	if err != nil || len(instances) == 0 {
		util.LOGGER.Errorf(err, "get provider service %s instance failed", providerId)
		return err
	}

	for _, instance := range instances {
		nf.PublishInstanceEvent(service, tenant, pb.EVT_UPDATE, providerKey, instance, rev, allow)
		nf.PublishInstanceEvent(service, tenant, pb.EVT_DELETE, providerKey, instance, rev, deny)
	}
	return nil
}

type RuleEventHandler struct {
	service *nf.NotifyService
}

func (h *RuleEventHandler) Type() store.StoreType {
	return store.RULE
}

func (h *RuleEventHandler) OnEvent(evt *store.KvEvent) {
	kv := evt.KV
	action := evt.Action
	providerId, ruleId, tenant, data := pb.GetInfoFromRuleKV(kv)
	if data == nil {
		util.LOGGER.Errorf(nil,
			"unmarshal service rule file failed, service %s rule %s [%s] event, data is nil",
			providerId, ruleId, action)
		return
	}

	if h.service.Closed() {
		util.LOGGER.Warnf(nil, "caught service %s rule %s [%s] event, but notify service is closed",
			providerId, ruleId, action)
		return
	}
	util.LOGGER.Infof("caught service %s rule %s [%s] event", providerId, ruleId, action)

	var rule pb.ServiceRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		util.LOGGER.Errorf(err, "unmarshal service %s rule %s file failed",
			providerId, ruleId)
		return
	}

	store.Store().AsyncTasker().AddTask(context.Background(),
		NewRulesChangedAsyncTask(h.service, tenant, providerId, evt.Revision))
}

func NewRuleEventHandler(s *nf.NotifyService) *RuleEventHandler {
	h := &RuleEventHandler{
		service: s,
	}
	return h
}

func NewRulesChangedAsyncTask(service *nf.NotifyService, tenant, providerId string, rev int64) *RulesChangedAsyncTask {
	return &RulesChangedAsyncTask{
		key:        "RulesChangedAsyncTask_" + providerId,
		Tenant:     tenant,
		ProviderId: providerId,
		Rev:        rev,
		Service:    service,
	}
}

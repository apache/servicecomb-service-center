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
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	nf "github.com/ServiceComb/service-center/server/service/notification"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strings"
)

type RuleFilter struct {
	Tenant       string
	Provider     *pb.MicroService
	ProviderRule *pb.ServiceRule
}

func (rf *RuleFilter) Filter(ctx context.Context, consumerId string) (bool, error) {
	consumer, err := ms.GetServiceByServiceId(ctx, rf.Tenant, consumerId)

	tags, err := serviceUtil.GetTagsUtils(context.Background(), rf.Tenant, consumerId)
	if err != nil {
		return false, err
	}
	matchErr := serviceUtil.MatchRules([]*pb.ServiceRule{rf.ProviderRule}, consumer, tags)
	switch matchErr.(type) {
	case serviceUtil.NotMatchWhiteListError, serviceUtil.MatchBlackListError:
		return false, nil
	default:
	}
	return true, nil
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

	h.publish(context.Background(), tenant, providerId, &rule, evt.Action, evt.Revision)
}

func (h *RuleEventHandler) publish(ctx context.Context, tenant, providerId string, rule *pb.ServiceRule, action pb.EventType, rev int64) {
	provider, err := ms.GetServiceByServiceId(ctx, tenant, providerId)
	if provider == nil {
		util.LOGGER.Errorf(err, "get service %s file failed", providerId)
		return
	}
	providerKey := &pb.MicroServiceKey{
		AppId:       provider.AppId,
		ServiceName: provider.ServiceName,
		Version:     provider.Version,
	}

	instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, tenant, providerId, "")
	if err != nil {
		util.LOGGER.Errorf(err, "get provider service %s instance failed", providerId)
		return
	}
	if len(instances) == 0 {
		return
	}

	rf := RuleFilter{
		Tenant:       tenant,
		Provider:     provider,
		ProviderRule: rule,
	}

	allow, deny, err := GetConsumerIdsWithFilter(ctx, tenant, providerId, rf.Filter)
	if err != nil {
		util.LOGGER.Errorf(err, "get consumer services failed, provider %s rule %s",
			providerId, rule.RuleId)
		return
	}

	for _, instance := range instances {
		nf.PublishInstanceEvent(h.service, tenant, pb.EVT_UPDATE, providerKey, instance, rev, allow)
		nf.PublishInstanceEvent(h.service, tenant, pb.EVT_DELETE, providerKey, instance, rev, deny)
	}
}

func GetConsumerIdsWithFilter(ctx context.Context, tenant, providerId string,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	kvs, err := dependency.GetConsumersInCache(ctx, tenant, providerId)
	if err != nil {
		return nil, nil, err
	}
	l := len(kvs)
	if l == 0 {
		return nil, nil, nil
	}
	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, kv := range kvs {
		consumerId := util.BytesToStringWithNoCopy(kv.Key)
		consumerId = consumerId[strings.LastIndex(consumerId, "/")+1:]

		ok, err := filter(ctx, consumerId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerId
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerId
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func NewRuleEventHandler(s *nf.NotifyService) *RuleEventHandler {
	h := &RuleEventHandler{
		service: s,
	}
	return h
}

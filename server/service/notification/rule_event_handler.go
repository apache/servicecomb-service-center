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
package notification

import (
	"encoding/json"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"github.com/ServiceComb/service-center/server/service/dependency"
	"strings"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/server/service/microservice"
)

type RuleEventHandler struct {
	service *NotifyService
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
	util.LOGGER.Infof("caught service %s rule %s [%s] event",
		providerId, ruleId, action)

	var rule pb.ServiceRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		util.LOGGER.Errorf(err, "unmarshal service %s rule %s file failed",
			providerId, ruleId)
		return
	}

	kvs, err := dependency.GetConsumersInCache(context.Background(), tenant, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "get consumer services failed, provider %s rule %s",
			providerId, ruleId)
		return
	}

	type serviceInfo struct {
		Service *pb.MicroService
		Tags map[string]string
	}
	consumers := make([]*serviceInfo, 0, len(kvs))
	for _, kv := range kvs {
		consumerId := string(kv.Key)
		consumerId = consumerId[strings.LastIndex(consumerId, "/")+1:]

		ms, err := microservice.GetServiceByServiceId(context.Background(), tenant, consumerId)
		if err != nil {
			util.LOGGER.Errorf(err, "get consumer %s service file failed, provider %s rule %s",
				consumerId, providerId, ruleId)
			return
		}

		tags, err := serviceUtil.GetTagsUtils(context.Background(), tenant, consumerId)
		if err != nil {
			util.LOGGER.Errorf(err, "get consumer %s service tag failed, provider %s rule %s",
				consumerId, providerId, ruleId)
			return
		}

		matchErr := serviceUtil.MatchRules([]*pb.ServiceRule{&rule}, ms, tags)
		switch matchErr.(type) {
		case serviceUtil.NotMatchWhiteListError:
			switch evt.Action {
			case pb.EVT_CREATE,pb.EVT_UPDATE:
			case pb.EVT_DELETE:
			}
		case serviceUtil.MatchBlackListError:
			switch evt.Action {
			case pb.EVT_CREATE,pb.EVT_UPDATE:
			case pb.EVT_DELETE:
			}
		default:
			util.LOGGER.Errorf(err, "match service %s rules %s failed",
				consumerId, providerId, ruleId)
			return
		}

		consumers = append(consumers, &serviceInfo{ms, tags})
	}
}

func NewRuleEventHandler(s *NotifyService) EventHandler {
	h := &RuleEventHandler{
		service: s,
	}
	return h
}

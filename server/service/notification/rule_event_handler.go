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
	serviceId, ruleId, _, data := pb.GetInfoFromRuleKV(kv)
	if data == nil {
		util.LOGGER.Errorf(nil,
			"unmarshal service rule file failed, service %s rule %s [%s] event, data is nil",
			serviceId, ruleId, action)
		return
	}

	if h.service.Closed() {
		util.LOGGER.Warnf(nil, "caught service %s rule %s [%s] event, but notify service is closed",
			serviceId, ruleId, action)
		return
	}
	util.LOGGER.Infof("caught service %s rule %s [%s] event",
		serviceId, ruleId, action)

	var rule pb.ServiceRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		util.LOGGER.Errorf(err, "unmarshal service %s rule %s file failed",
			serviceId, ruleId)
		return
	}
	// TODO
}

func NewRuleEventHandler(s *NotifyService) EventHandler {
	h := &RuleEventHandler{
		service: s,
	}
	return h
}

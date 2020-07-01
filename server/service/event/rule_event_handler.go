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
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

type RulesChangedTask struct {
	discovery.KvEvent

	key string
	err error

	DomainProject string
	ProviderId    string
}

func (apt *RulesChangedTask) Key() string {
	return apt.key
}

func (apt *RulesChangedTask) Do(ctx context.Context) error {
	apt.err = apt.publish(ctx, apt.DomainProject, apt.ProviderId)
	return apt.err
}

func (apt *RulesChangedTask) Err() error {
	return apt.err
}

func (apt *RulesChangedTask) publish(ctx context.Context, domainProject, providerId string) error {
	ctx = context.WithValue(context.WithValue(ctx,
		serviceUtil.CTX_CACHEONLY, "1"),
		serviceUtil.CTX_GLOBAL, "1")

	provider, err := serviceUtil.GetService(ctx, domainProject, providerId)
	if err != nil {
		log.Errorf(err, "get provider[%s] service file failed", providerId)
		return err
	}
	if provider == nil {
		log.Errorf(nil, "provider[%s] does not exist", providerId)
		return fmt.Errorf("provider %s does not exist", providerId)
	}

	consumerIds, err := serviceUtil.GetConsumerIds(ctx, domainProject, provider)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s consumerIds failed",
			providerId, provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
		return err
	}
	providerKey := pb.MicroServiceToKey(domainProject, provider)

	PublishInstanceEvent(apt.KvEvent, domainProject, providerKey, consumerIds)
	return nil
}

// RuleEventHandler is the handler to handle:
// 1. publish the EVT_EXPIRE event to subscribers when rule is changed
// 2. reset the find instance cache
type RuleEventHandler struct {
}

func (h *RuleEventHandler) Type() discovery.Type {
	return backend.RULE
}

func (h *RuleEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	if action == pb.EVT_INIT {
		return
	}

	providerId, ruleId, domainProject := core.GetInfoFromRuleKV(evt.KV.Key)
	if notify.NotifyCenter().Closed() {
		log.Warnf("caught [%s] service rule[%s/%s] event, but notify service is closed",
			action, providerId, ruleId)
		return
	}
	log.Infof("caught [%s] service rule[%s/%s] event", action, providerId, ruleId)

	err := task.Service().Add(context.Background(),
		NewRulesChangedAsyncTask(domainProject, providerId, evt))
	if err != nil {
		log.Error("", err)
	}
}

func NewRuleEventHandler() *RuleEventHandler {
	return &RuleEventHandler{}
}

func NewRulesChangedAsyncTask(domainProject, providerId string, evt discovery.KvEvent) *RulesChangedTask {
	evt.Type = pb.EVT_EXPIRE
	return &RulesChangedTask{
		KvEvent:       evt,
		key:           "RulesChangedAsyncTask_" + providerId,
		DomainProject: domainProject,
		ProviderId:    providerId,
	}
}

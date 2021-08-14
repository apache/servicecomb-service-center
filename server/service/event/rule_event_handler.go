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
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

type RulesChangedTask struct {
	discovery.KvEvent

	key string
	err error

	DomainProject string
	ProviderID    string
}

func (apt *RulesChangedTask) Key() string {
	return apt.key
}

func (apt *RulesChangedTask) Do(ctx context.Context) error {
	apt.err = apt.publish(ctx, apt.DomainProject, apt.ProviderID)
	return apt.err
}

func (apt *RulesChangedTask) Err() error {
	return apt.err
}

func (apt *RulesChangedTask) publish(ctx context.Context, domainProject, providerID string) error {
	ctx = context.WithValue(context.WithValue(ctx,
		util.CtxCacheOnly, "1"),
		util.CtxGlobal, "1")

	provider, err := serviceUtil.GetService(ctx, domainProject, providerID)
	if err != nil {
		log.Errorf(err, "get provider[%s] service file failed", providerID)
		return err
	}
	if provider == nil {
		log.Errorf(nil, "provider[%s] does not exist", providerID)
		return fmt.Errorf("provider %s does not exist", providerID)
	}

	consumerIds, err := serviceUtil.GetConsumerIds(ctx, domainProject, provider)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s consumerIds failed",
			providerID, provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
		return err
	}
	providerKey := proto.MicroServiceToKey(domainProject, provider)

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

	providerID, ruleID, domainProject := core.GetInfoFromRuleKV(evt.KV.Key)
	if event.Center().Closed() {
		log.Warnf("caught [%s] service rule[%s/%s] event, but notify service is closed",
			action, providerID, ruleID)
		return
	}
	log.Infof("caught [%s] service rule[%s/%s] event", action, providerID, ruleID)

	err := task.GetService().Add(context.Background(),
		NewRulesChangedAsyncTask(domainProject, providerID, evt))
	if err != nil {
		log.Error("", err)
	}
}

func NewRuleEventHandler() *RuleEventHandler {
	return &RuleEventHandler{}
}

func NewRulesChangedAsyncTask(domainProject, providerID string, evt discovery.KvEvent) *RulesChangedTask {
	evt.Type = pb.EVT_EXPIRE
	return &RulesChangedTask{
		KvEvent:       evt,
		key:           "RulesChangedAsyncTask_" + providerID,
		DomainProject: domainProject,
		ProviderID:    providerID,
	}
}

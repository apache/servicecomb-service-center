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
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/mux"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"net/http"
)

type DependencyEventHandler struct {
}

func (h *DependencyEventHandler) Type() store.StoreType {
	return store.SERVICE
}

func (h *DependencyEventHandler) OnEvent(evt *store.KvEvent) {
	action := evt.Action
	if action != pb.EVT_CREATE && action != pb.EVT_INIT {
		return
	}

	kv := evt.KV
	method, domainProject, data := pb.GetInfoFromDependencyKV(kv)
	if data == nil {
		util.Logger().Errorf(nil,
			"unmarshal dependency file failed, method %s [%s] event, data is nil",
			method, action)
		return
	}

	// maintain dependency rules.
	var (
		r   pb.AddDependenciesRequest
		ctx context.Context = context.Background()
	)

	err := json.Unmarshal(data, &r.Dependencies)
	if err != nil {
		util.Logger().Errorf(err, "unmarshal dependency file failed, method %s [%s] event",
			method, action)
		return
	}

	//建立依赖规则，用于维护依赖关系
	lock, err := mux.Try(mux.GLOBAL_LOCK)
	if err != nil {
		// retry
		util.Logger().Errorf(err, "%s dependency failed, [%s] event", method, action)
		return
	}
	if lock == nil {
		return
	}
	defer lock.Unlock()

	for _, dependencyInfo := range r.Dependencies {
		util.Logger().Infof("start %s dependency, [%s] event, data info %v", method, action, dependencyInfo)

		serviceUtil.SetDependencyDefaultValue(dependencyInfo)

		consumerFlag := util.StringJoin([]string{dependencyInfo.Consumer.AppId, dependencyInfo.Consumer.ServiceName, dependencyInfo.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.DependencyKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		consumerId, err := serviceUtil.GetServiceId(ctx, consumerInfo)
		if err != nil {
			// retry
			util.Logger().Errorf(err, "%s dependency failed because of getting service %s error, [%s] event",
				method, consumerFlag, action)
			return
		}
		if len(consumerId) == 0 {
			util.Logger().Errorf(nil, "%s dependency failed, consumer %s: consumer does not exist, [%s] event",
				method, consumerFlag, action)
			return
		}

		var dep serviceUtil.Dependency
		dep.DomainProject = domainProject
		dep.Consumer = consumerInfo
		dep.ProvidersRule = providersInfo
		dep.ConsumerId = consumerId
		switch method {
		case http.MethodPost:
			err = serviceUtil.AddDependencyRule(ctx, &dep)
		default:
			err = serviceUtil.CreateDependencyRule(ctx, &dep)
		}

		if err != nil {
			// retry
			util.Logger().Errorf(err, "%s dependency failed: consumer %s, [%s] event",
				method, consumerFlag, action)
			return
		}
	}
	return
}

func NewDependencyEventHandler() *DependencyEventHandler {
	return &DependencyEventHandler{}
}

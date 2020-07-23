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
package util

import (
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/log"
	rmodel "github.com/apache/servicecomb-service-center/pkg/registry"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

type Dependency struct {
	DomainProject string
	// store the consumer Dependency from dep-queue object
	Consumer      *rmodel.MicroServiceKey
	ProvidersRule []*rmodel.MicroServiceKey
	// store the parsed rules from Dependency object
	DeleteDependencyRuleList []*rmodel.MicroServiceKey
	CreateDependencyRuleList []*rmodel.MicroServiceKey
}

func (dep *Dependency) removeConsumerOfProviderRule(ctx context.Context) ([]registry.PluginOp, error) {
	opts := make([]registry.PluginOp, 0, len(dep.DeleteDependencyRuleList))
	for _, providerRule := range dep.DeleteDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		log.Debugf("This proProkey is %s", proProkey)
		consumerValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			return nil, err
		}
		for key, tmp := range consumerValue.Dependency {
			if ok := equalServiceDependency(tmp, dep.Consumer); ok {
				consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
				break
			}
			log.Debugf("tmp and dep.Consumer not equal, tmp %v, consumer %v", tmp, dep.Consumer)
		}
		//删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
		if len(consumerValue.Dependency) == 0 {
			opts = append(opts, registry.OpDel(registry.WithStrKey(proProkey)))
			continue
		}
		data, err := json.Marshal(consumerValue)
		if err != nil {
			log.Errorf(err, "Marshal MicroServiceDependency failed")
			return nil, err
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
	}
	return opts, nil
}

func (dep *Dependency) addConsumerOfProviderRule(ctx context.Context) ([]registry.PluginOp, error) {
	opts := make([]registry.PluginOp, 0, len(dep.CreateDependencyRuleList))
	for _, providerRule := range dep.CreateDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		tmpValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			return nil, err
		}
		tmpValue.Dependency = append(tmpValue.Dependency, dep.Consumer)

		data, errMarshal := json.Marshal(tmpValue)
		if errMarshal != nil {
			log.Errorf(errMarshal, "Marshal MicroServiceDependency failed")
			return nil, errMarshal
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
		if providerRule.ServiceName == "*" {
			break
		}
	}
	return opts, nil
}

func (dep *Dependency) updateProvidersRuleOfConsumer(_ context.Context) ([]registry.PluginOp, error) {
	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	if len(dep.ProvidersRule) == 0 {
		return []registry.PluginOp{registry.OpDel(registry.WithStrKey(conKey))}, nil
	}

	dependency := &rmodel.MicroServiceDependency{
		Dependency: dep.ProvidersRule,
	}
	data, err := json.Marshal(dependency)
	if err != nil {
		log.Errorf(err, "Marshal MicroServiceDependency failed")
		return nil, err
	}
	return []registry.PluginOp{registry.OpPut(registry.WithStrKey(conKey), registry.WithValue(data))}, nil
}

func (dep *Dependency) Commit(ctx context.Context) error {
	dopts, err := dep.removeConsumerOfProviderRule(ctx)
	if err != nil {
		return err
	}
	copts, err := dep.addConsumerOfProviderRule(ctx)
	if err != nil {
		return err
	}
	uopts, err := dep.updateProvidersRuleOfConsumer(ctx)
	if err != nil {
		return err
	}
	return backend.BatchCommit(ctx, append(append(dopts, copts...), uopts...))
}

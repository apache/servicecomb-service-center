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
	"fmt"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// Dependency contains dependency rules
type Dependency struct {
	DomainProject string
	// store the consumer Dependency from dep-queue object
	Consumer      *discovery.MicroServiceKey
	ProvidersRule []*discovery.MicroServiceKey
	// store the parsed rules from Dependency object
	DeleteDependencyRuleList []*discovery.MicroServiceKey
	CreateDependencyRuleList []*discovery.MicroServiceKey
}

func (dep *Dependency) removeConsumerOfProviderRule(ctx context.Context) ([]etcdadpt.OpOptions, error) {
	opts := make([]etcdadpt.OpOptions, 0, len(dep.DeleteDependencyRuleList))
	for _, providerRule := range dep.DeleteDependencyRuleList {
		proProkey := path.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		consumerValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			return nil, err
		}
		for key, tmp := range consumerValue.Dependency {
			if ok := EqualServiceDependency(tmp, dep.Consumer); ok {
				consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
				break
			}
			log.Debug(fmt.Sprintf("tmp and dep.Consumer not equal, tmp %v, consumer %v", tmp, dep.Consumer))
		}
		// 删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
		if len(consumerValue.Dependency) == 0 {
			opts = append(opts, etcdadpt.OpDel(etcdadpt.WithStrKey(proProkey)))
			continue
		}
		data, err := json.Marshal(consumerValue)
		if err != nil {
			log.Error("Marshal MicroServiceDependency failed", err)
			return nil, err
		}
		opts = append(opts, etcdadpt.OpPut(
			etcdadpt.WithStrKey(proProkey),
			etcdadpt.WithValue(data)))
	}
	return opts, nil
}

func (dep *Dependency) addConsumerOfProviderRule(ctx context.Context) ([]etcdadpt.OpOptions, error) {
	opts := make([]etcdadpt.OpOptions, 0, len(dep.CreateDependencyRuleList))
	for _, providerRule := range dep.CreateDependencyRuleList {
		providerRuleKey := path.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		tmpValue, err := TransferToMicroServiceDependency(ctx, providerRuleKey)
		if err != nil {
			return nil, err
		}
		tmpValue.Dependency = append(tmpValue.Dependency, dep.Consumer)

		data, errMarshal := json.Marshal(tmpValue)
		if errMarshal != nil {
			log.Error("Marshal MicroServiceDependency failed", errMarshal)
			return nil, errMarshal
		}
		opts = append(opts, etcdadpt.OpPut(
			etcdadpt.WithStrKey(providerRuleKey),
			etcdadpt.WithValue(data)))
	}
	return opts, nil
}

func (dep *Dependency) updateProvidersRuleOfConsumer(_ context.Context) ([]etcdadpt.OpOptions, error) {
	conKey := path.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	if len(dep.ProvidersRule) == 0 {
		return []etcdadpt.OpOptions{etcdadpt.OpDel(etcdadpt.WithStrKey(conKey))}, nil
	}

	dependency := &discovery.MicroServiceDependency{
		Dependency: dep.ProvidersRule,
	}
	data, err := json.Marshal(dependency)
	if err != nil {
		log.Error("Marshal MicroServiceDependency failed", err)
		return nil, err
	}
	return []etcdadpt.OpOptions{etcdadpt.OpPut(etcdadpt.WithStrKey(conKey), etcdadpt.WithValue(data))}, nil
}

// Commit is dependent rule operations
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
	return etcdadpt.Txn(ctx, append(append(dopts, copts...), uopts...))
}

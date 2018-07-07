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
	"encoding/json"
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
)

type Dependency struct {
	ConsumerId                string
	DomainProject             string
	removedDependencyRuleList []*pb.MicroServiceKey
	NewDependencyRuleList     []*pb.MicroServiceKey
	err                       chan error
	chanNum                   int8
	Consumer                  *pb.MicroServiceKey
	ProvidersRule             []*pb.MicroServiceKey
}

func (dep *Dependency) RemoveConsumerOfProviderRule() {
	dep.chanNum++
	util.Go(dep.removeConsumerOfProviderRule)
}

func (dep *Dependency) removeConsumerOfProviderRule(ctx context.Context) {
	opts := make([]registry.PluginOp, 0, len(dep.removedDependencyRuleList))
	for _, providerRule := range dep.removedDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		util.Logger().Debugf("This proProkey is %s.", proProkey)
		consumerValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		for key, tmp := range consumerValue.Dependency {
			if ok := equalServiceDependency(tmp, dep.Consumer); ok {
				consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
				break
			}
			util.Logger().Debugf("tmp and dep.Consumer not equal, tmp %v, consumer %v", tmp, dep.Consumer)
		}
		//删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
		if len(consumerValue.Dependency) == 0 {
			opts = append(opts, registry.OpDel(registry.WithStrKey(proProkey)))
			continue
		}
		data, err := json.Marshal(consumerValue)
		if err != nil {
			util.Logger().Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- err
			return
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
	}
	if len(opts) != 0 {
		err := backend.BatchCommit(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) AddConsumerOfProviderRule() {
	dep.chanNum++
	util.Go(dep.addConsumerOfProviderRule)
}

func (dep *Dependency) addConsumerOfProviderRule(ctx context.Context) {
	opts := []registry.PluginOp{}
	for _, providerRule := range dep.NewDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		tmpValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		tmpValue.Dependency = append(tmpValue.Dependency, dep.Consumer)

		data, errMarshal := json.Marshal(tmpValue)
		if errMarshal != nil {
			util.Logger().Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- errors.New("Marshal tmpValue failed.")
			return
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
		if providerRule.ServiceName == "*" {
			break
		}
	}
	if len(opts) != 0 {
		err := backend.BatchCommit(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) UpdateProvidersRuleOfConsumer(conKey string) error {
	if len(dep.ProvidersRule) == 0 {
		_, err := backend.Registry().Do(context.TODO(),
			registry.DEL,
			registry.WithStrKey(conKey),
		)
		if err != nil {
			util.Logger().Errorf(nil, "Upload dependency rule failed.")
			return err
		}
		return nil
	}

	dependency := &pb.MicroServiceDependency{
		Dependency: dep.ProvidersRule,
	}
	data, err := json.Marshal(dependency)
	if err != nil {
		util.Logger().Errorf(nil, "Marshal tmpValue fialed.")
		return err
	}
	_, err = backend.Registry().Do(context.TODO(),
		registry.PUT,
		registry.WithStrKey(conKey),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(nil, "Upload dependency rule failed.")
		return err
	}
	return nil
}

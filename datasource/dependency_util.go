/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package datasource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/discovery"
)

type Dependency struct {
	DomainProject string
	// store the consumer Dependency from dep-queue object
	Consumer      *discovery.MicroServiceKey
	ProvidersRule []*discovery.MicroServiceKey
	// store the parsed rules from Dependency object
	DeleteDependencyRuleList []*discovery.MicroServiceKey
	CreateDependencyRuleList []*discovery.MicroServiceKey
}

func ParamsChecker(consumerInfo *discovery.MicroServiceKey, providersInfo []*discovery.MicroServiceKey) *discovery.CreateDependenciesResponse {
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			break
		}
		if len(providerInfo.AppId) == 0 {
			providerInfo.AppId = consumerInfo.AppId
		}

		version := providerInfo.Version
		if len(version) == 0 {
			return BadParamsResponse("Required provider version")
		}

		providerInfo.Version = ""
		if _, ok := flag[toString(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider or (serviceName and appId is same).")
		}
		flag[toString(providerInfo)] = true
		providerInfo.Version = version
	}
	return nil
}

func BadParamsResponse(detailErr string) *discovery.CreateDependenciesResponse {
	log.Error(fmt.Sprintf("request params is invalid. %s", detailErr), nil)
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &discovery.CreateDependenciesResponse{
		Response: discovery.CreateResponse(discovery.ErrInvalidParams, detailErr),
	}
}

func toString(in *discovery.MicroServiceKey) string {
	return path.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func ParseAddOrUpdateRules(ctx context.Context, dep *Dependency, oldProviderRules *discovery.MicroServiceDependency) {
	deleteDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	createDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := ContainServiceDependency(oldProviderRules.Dependency, tmpProviderRule); ok {
			continue
		}
		createDependencyRuleList = append(createDependencyRuleList, tmpProviderRule)
		old := IsNeedUpdate(oldProviderRules.Dependency, tmpProviderRule)
		if old != nil {
			deleteDependencyRuleList = append(deleteDependencyRuleList, old)
		}
	}
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := ContainServiceDependency(deleteDependencyRuleList, oldProviderRule); !ok {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	dep.ProvidersRule = append(createDependencyRuleList, existDependencyRuleList...)
	setDep(dep, createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList)
}

func ParseOverrideRules(ctx context.Context, dep *Dependency, oldProviderRules *discovery.MicroServiceDependency) {
	deleteDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	createDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList := make([]*discovery.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := ContainServiceDependency(dep.ProvidersRule, oldProviderRule); !ok {
			deleteDependencyRuleList = append(deleteDependencyRuleList, oldProviderRule)
		} else {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := ContainServiceDependency(existDependencyRuleList, tmpProviderRule); !ok {
			createDependencyRuleList = append(createDependencyRuleList, tmpProviderRule)
		}
	}
	setDep(dep, createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList)
}

func setDep(dep *Dependency, createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*discovery.MicroServiceKey) {
	consumerFlag := strings.Join([]string{dep.Consumer.Environment, dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	if len(createDependencyRuleList) == 0 && len(existDependencyRuleList) == 0 && len(deleteDependencyRuleList) == 0 {
		return
	}

	if len(deleteDependencyRuleList) != 0 {
		log.Info(fmt.Sprintf("delete consumer[%s]'s dependency rule %v", consumerFlag, deleteDependencyRuleList))
		dep.DeleteDependencyRuleList = deleteDependencyRuleList
	}

	if len(createDependencyRuleList) != 0 {
		log.Info(fmt.Sprintf("create consumer[%s]'s dependency rule %v", consumerFlag, createDependencyRuleList))
		dep.CreateDependencyRuleList = createDependencyRuleList
	}
}

func ContainServiceDependency(services []*discovery.MicroServiceKey, service *discovery.MicroServiceKey) (bool, error) {
	if services == nil || service == nil {
		return false, errors.New("invalid params input")
	}
	for _, value := range services {
		rst := EqualServiceDependency(service, value)
		if rst {
			return true, nil
		}
	}
	return false, nil
}

func EqualServiceDependency(serviceA *discovery.MicroServiceKey, serviceB *discovery.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	return stringA == stringB
}

func IsNeedUpdate(services []*discovery.MicroServiceKey, service *discovery.MicroServiceKey) *discovery.MicroServiceKey {
	for _, tmp := range services {
		if DiffServiceVersion(tmp, service) {
			return tmp
		}
	}
	return nil
}

func DiffServiceVersion(serviceA *discovery.MicroServiceKey, serviceB *discovery.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA != stringB &&
		stringA[:strings.LastIndex(stringA, "/")+1] == stringB[:strings.LastIndex(stringB, "/")+1] {
		return true
	}
	return false
}

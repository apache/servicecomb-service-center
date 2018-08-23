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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"strings"
)

func GetConsumerIds(ctx context.Context, domainProject string, provider *pb.MicroService) ([]string, error) {
	// 查询所有consumer
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	consumerIds, err := dr.GetDependencyConsumerIds()
	if err != nil {
		log.Errorf(err, "Get dependency consumerIds failed.%s", provider.ServiceId)
		return nil, err
	}
	return consumerIds, nil
}

func GetProviderIds(ctx context.Context, domainProject string, consumer *pb.MicroService) ([]string, error) {
	// 查询所有provider
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer)
	providerIds, err := dr.GetDependencyProviderIds()
	if err != nil {
		log.Errorf(err, "Get dependency providerIds failed.%s", consumer.ServiceId)
		return nil, err
	}
	return providerIds, nil
}

// GetAllConsumerIds is the function get from dependency rule and filter with service rules
func GetAllConsumerIds(ctx context.Context, domainProject string, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, err := GetRulesUtil(ctx, domainProject, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}

	rf := &RuleFilter{
		DomainProject: domainProject,
		ProviderRules: providerRules,
	}

	allow, deny, err = getConsumerIdsWithFilter(ctx, domainProject, provider, rf)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func getConsumerIdsWithFilter(ctx context.Context, domainProject string, provider *pb.MicroService, rf *RuleFilter) (allow []string, deny []string, err error) {
	consumerIds, err := GetConsumerIds(ctx, domainProject, provider)
	if err != nil {
		return nil, nil, err
	}
	return rf.FilterAll(ctx, consumerIds)
}

func GetAllProviderIds(ctx context.Context, domainProject string, service *pb.MicroService) (allow []string, deny []string, _ error) {
	providerIdsInCache, err := GetProviderIds(ctx, domainProject, service)
	if err != nil {
		return nil, nil, err
	}
	l := len(providerIdsInCache)
	rf := RuleFilter{
		DomainProject: domainProject,
	}
	allowIdx, denyIdx := 0, l
	providerIds := make([]string, l)
	copyCtx := util.SetContext(util.CloneContext(ctx), CTX_CACHEONLY, "1")
	for _, providerId := range providerIdsInCache {
		providerRules, err := GetRulesUtil(copyCtx, domainProject, providerId)
		if err != nil {
			return nil, nil, err
		}
		if len(providerRules) == 0 {
			providerIds[allowIdx] = providerId
			allowIdx++
			continue
		}
		rf.ProviderRules = providerRules
		ok, err := rf.Filter(ctx, service.ServiceId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			providerIds[allowIdx] = providerId
			allowIdx++
		} else {
			denyIdx--
			providerIds[denyIdx] = providerId
		}
	}
	return providerIds[:allowIdx], providerIds[denyIdx:], nil
}

func DependencyRuleExist(ctx context.Context, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}

	consumerKey := apt.GenerateConsumerDependencyRuleKey(consumer.Tenant, consumer)
	existed, err := dependencyRuleExistUtil(ctx, consumerKey, provider)
	if err != nil || existed {
		return existed, err
	}

	providerKey := apt.GenerateProviderDependencyRuleKey(targetDomainProject, provider)
	return dependencyRuleExistUtil(ctx, providerKey, consumer)
}

func dependencyRuleExistUtil(ctx context.Context, key string, target *pb.MicroServiceKey) (bool, error) {
	compareData, err := TransferToMicroServiceDependency(ctx, key)
	if err != nil {
		return false, err
	}
	if len(compareData.Dependency) != 0 {
		isEqual, err := containServiceDependency(compareData.Dependency, target)
		if err != nil {
			return false, err
		}
		if isEqual {
			//删除之前的依赖
			return true, nil
		}
	}
	return false, nil
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	//创建依赖一致
	consumerKey := pb.MicroServiceToKey(domainProject, consumer)
	exist, err := DependencyRuleExist(ctx, provider, consumerKey)
	if exist || err != nil {
		return err
	}

	r := &pb.ConsumerDependency{
		Consumer:  consumerKey,
		Providers: []*pb.MicroServiceKey{provider},
		Override:  false,
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	id := util.StringJoin([]string{provider.AppId, provider.ServiceName}, "_")
	key := apt.GenerateConsumerDependencyQueueKey(domainProject, consumer.ServiceId, id)
	resp, err := backend.Registry().TxnWithCmp(ctx,
		nil,
		[]registry.CompareOp{registry.OpCmp(registry.CmpStrVal(key), registry.CMP_EQUAL, util.BytesToStringWithNoCopy(data))},
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))})
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		log.Infof("find request into dependency queue successfully, %s: %v", key, r)
	}
	return nil
}

func TransferToMicroServiceDependency(ctx context.Context, key string) (*pb.MicroServiceDependency, error) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}

	opts := append(FromContext(ctx), registry.WithStrKey(key))
	res, err := backend.Store().DependencyRule().Search(ctx, opts...)
	if err != nil {
		log.Errorf(nil, "Get dependency rule failed.")
		return nil, err
	}
	if len(res.Kvs) != 0 {
		return res.Kvs[0].Value.(*pb.MicroServiceDependency), nil
	}
	return microServiceDependency, nil
}

func equalServiceDependency(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA == stringB {
		return true
	}
	return false
}

func diffServiceVersion(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA != stringB &&
		stringA[:strings.LastIndex(stringA, "/")+1] == stringB[:strings.LastIndex(stringB, "/")+1] {
		return true
	}
	return false
}

func toString(in *pb.MicroServiceKey) string {
	return apt.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func parseAddOrUpdateRules(ctx context.Context, dep *Dependency) (newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		log.Errorf(err, "maintain dependency rule failed, consumer %s/%s/%s: get consumer depedency rule failed.",
			dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containServiceDependency(oldProviderRules.Dependency, tmpProviderRule); ok {
			continue
		}

		if tmpProviderRule.ServiceName == "*" {
			newDependencyRuleList = append([]*pb.MicroServiceKey{}, tmpProviderRule)
			deleteDependencyRuleList = oldProviderRules.Dependency
			break
		}

		newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		old := isNeedUpdate(oldProviderRules.Dependency, tmpProviderRule)
		if old != nil {
			deleteDependencyRuleList = append(deleteDependencyRuleList, old)
		}
	}
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if oldProviderRule.ServiceName == "*" {
			newDependencyRuleList = nil
			deleteDependencyRuleList = nil
			return
		}
		if ok, _ := containServiceDependency(deleteDependencyRuleList, oldProviderRule); !ok {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}

	dep.ProvidersRule = append(newDependencyRuleList, existDependencyRuleList...)
	return
}

func parseOverrideRules(ctx context.Context, dep *Dependency) (newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		log.Errorf(err, "maintain dependency rule failed, consumer %s/%s/%s: get consumer depedency rule failed.",
			dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := containServiceDependency(dep.ProvidersRule, oldProviderRule); !ok {
			deleteDependencyRuleList = append(deleteDependencyRuleList, oldProviderRule)
		} else {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containServiceDependency(existDependencyRuleList, tmpProviderRule); !ok {
			newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		}
	}
	return
}

func syncDependencyRule(ctx context.Context, dep *Dependency, filter func(context.Context, *Dependency) (_, _, _ []*pb.MicroServiceKey)) error {
	//更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList := filter(ctx, dep)
	if len(newDependencyRuleList) == 0 && len(existDependencyRuleList) == 0 && len(deleteDependencyRuleList) == 0 {
		return nil
	}

	dep.err = make(chan error, 5)
	dep.chanNum = 0
	if len(deleteDependencyRuleList) != 0 {
		log.Infof("Delete dependency rule remove for consumer %s, %v, ", consumerFlag, deleteDependencyRuleList)
		dep.removedDependencyRuleList = deleteDependencyRuleList
		dep.RemoveConsumerOfProviderRule()
	}

	if len(newDependencyRuleList) != 0 {
		log.Infof("New dependency rule add for consumer %s, %v, ", consumerFlag, newDependencyRuleList)
		dep.newDependencyRuleList = newDependencyRuleList
		dep.AddConsumerOfProviderRule()
	}

	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	err := dep.UpdateProvidersRuleOfConsumer(ctx, conKey)
	if err != nil {
		return err
	}

	if dep.chanNum != 0 {
		for tmpErr := range dep.err {
			dep.chanNum--
			if tmpErr != nil {
				return tmpErr
			}
			if 0 == dep.chanNum {
				close(dep.err)
			}
		}
	}
	return nil
}

func AddDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseAddOrUpdateRules)
}

func CreateDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseOverrideRules)
}

func isNeedUpdate(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) *pb.MicroServiceKey {
	for _, tmp := range services {
		if diffServiceVersion(tmp, service) {
			return tmp
		}
	}
	return nil
}

func containServiceDependency(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) (bool, error) {
	if services == nil || service == nil {
		return false, errors.New("Invalid params input.")
	}
	for _, value := range services {
		rst := equalServiceDependency(service, value)
		if rst {
			return true, nil
		}
	}
	return false, nil
}

func BadParamsResponse(detailErr string) *pb.CreateDependenciesResponse {
	log.Errorf(nil, "Request params is invalid. %s", detailErr)
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(scerr.ErrInvalidParams, detailErr),
	}
}

func ParamsChecker(consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey) *pb.CreateDependenciesResponse {
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			log.Debugf("%s 's provider contains *.", consumerInfo.ServiceName)
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
		} else {
			flag[toString(providerInfo)] = true
		}
		providerInfo.Version = version
	}
	return nil
}

func DeleteDependencyForDeleteService(domainProject string, serviceId string, service *pb.MicroServiceKey) (registry.PluginOp, error) {
	key := apt.GenerateConsumerDependencyQueueKey(domainProject, serviceId, apt.DEPS_QUEUE_UUID)
	conDep := new(pb.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*pb.MicroServiceKey{}
	conDep.Override = true
	data, err := json.Marshal(conDep)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func removeProviderRuleOfConsumer(ctx context.Context, domainProject string, cache map[string]bool) ([]registry.PluginOp, error) {
	key := apt.GenerateConsumerDependencyRuleKey(domainProject, nil) + apt.SPLIT
	resp, err := backend.Store().DependencyRule().Search(ctx, registry.WithStrKey(key), registry.WithPrefix())
	if err != nil {
		return nil, err
	}

	var ops []registry.PluginOp
loop:
	for _, kv := range resp.Kvs {
		var left []*pb.MicroServiceKey
		all := kv.Value.(*pb.MicroServiceDependency).Dependency
		for _, key := range all {
			if key.ServiceName == "*" {
				continue loop
			}

			id := apt.GenerateProviderDependencyRuleKey(key.Tenant, key)
			exist, ok := cache[id]
			if !ok {
				_, exist, err = FindServiceIds(ctx, key.Version, key)
				if err != nil {
					return nil, fmt.Errorf("%v, find service %s/%s/%s/%s",
						err, key.Tenant, key.AppId, key.ServiceName, key.Version)
				}
				cache[id] = exist
			}

			if exist {
				left = append(left, key)
			}
		}

		if len(all) == len(left) {
			continue
		}

		if len(left) == 0 {
			ops = append(ops, registry.OpDel(registry.WithKey(kv.Key)))
		} else {
			val, err := json.Marshal(&pb.MicroServiceDependency{Dependency: left})
			if err != nil {
				return nil, fmt.Errorf("%v, marshal %v", err, left)
			}
			ops = append(ops, registry.OpPut(registry.WithKey(kv.Key), registry.WithValue(val)))
		}
	}
	return ops, nil
}

func removeProviderRuleKeys(ctx context.Context, domainProject string, cache map[string]bool) ([]registry.PluginOp, error) {
	key := apt.GenerateProviderDependencyRuleKey(domainProject, nil) + apt.SPLIT
	resp, err := backend.Store().DependencyRule().Search(ctx, registry.WithStrKey(key), registry.WithPrefix())
	if err != nil {
		return nil, err
	}

	var ops []registry.PluginOp
	for _, kv := range resp.Kvs {
		id := util.BytesToStringWithNoCopy(kv.Key)
		exist, ok := cache[id]
		if !ok {
			key := backend.GetInfoFromDependencyRuleKV(kv)
			if key == nil || key.ServiceName == "*" {
				continue
			}

			_, exist, err = FindServiceIds(ctx, key.Version, key)
			if err != nil {
				return nil, fmt.Errorf("find service %s/%s/%s/%s, %v",
					key.Tenant, key.AppId, key.ServiceName, key.Version, err)
			}
			cache[id] = exist
		}

		if !exist {
			ops = append(ops, registry.OpDel(registry.WithKey(kv.Key)))
		}
	}
	return ops, nil
}

func CleanUpDependencyRules(ctx context.Context, domainProject string) error {
	if len(domainProject) == 0 {
		return errors.New("required domainProject")
	}

	cache := make(map[string]bool)
	pOps, err := removeProviderRuleOfConsumer(ctx, domainProject, cache)
	if err != nil {
		return err
	}

	kOps, err := removeProviderRuleKeys(ctx, domainProject, cache)
	if err != nil {
		return err
	}

	ops := append(append([]registry.PluginOp(nil), pOps...), kOps...)
	if len(ops) == 0 {
		return nil
	}

	return backend.BatchCommit(ctx, ops)
}

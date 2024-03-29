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
	"errors"
	"fmt"
	"strings"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetConsumerIds(ctx context.Context, domainProject string, provider *pb.MicroService) ([]string, error) {
	// 查询所有consumer
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	consumerIds, err := dr.getDependencyConsumerIds()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s consumerIds failed", provider.ServiceId), err)
		return nil, err
	}
	return consumerIds, nil
}

func GetConsumers(ctx context.Context, domainProject string, provider *pb.MicroService,
	opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	return dr.GetDependencyConsumers(opts...)
}

func GetProviderIds(ctx context.Context, domainProject string, consumer *pb.MicroService) ([]string, error) {
	// 查询所有provider
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer)
	providerIDs, err := dr.getDependencyProviderIds()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s providerIDs failed", consumer.ServiceId), err)
		return nil, err
	}
	return providerIDs, nil
}

func GetProviders(ctx context.Context, domainProject string, consumer *pb.MicroService,
	opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer)
	return dr.GetDependencyProviders(opts...)
}

func DependencyRuleExist(ctx context.Context, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}

	consumerKey := path.GenerateConsumerDependencyRuleKey(consumer.Tenant, consumer)
	existed, err := DependencyRuleExistWithKey(ctx, consumerKey, provider)
	if err != nil || existed {
		return existed, err
	}

	providerKey := path.GenerateProviderDependencyRuleKey(targetDomainProject, provider)
	return DependencyRuleExistWithKey(ctx, providerKey, consumer)
}

func DependencyRuleExistWithKey(ctx context.Context, key string, target *pb.MicroServiceKey) (bool, error) {
	compareData, err := TransferToMicroServiceDependency(ctx, key)
	if err != nil {
		return false, err
	}
	if len(compareData.Dependency) != 0 {
		isEqual, err := ContainServiceDependency(compareData.Dependency, target)
		if err != nil {
			return false, err
		}
		if isEqual {
			// 删除之前的依赖
			return true, nil
		}
	}
	return false, nil
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	// 创建依赖一致
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
	key := path.GenerateConsumerDependencyQueueKey(domainProject, consumer.ServiceId, id)
	opts := make([]etcdadpt.OpOptions, 0)
	opts = append(opts, etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)))
	syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceKV, data, esync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		log.Error("fail to create sync opts", err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)
	resp, err := etcdadpt.Instance().TxnWithCmp(ctx, opts, etcdadpt.If(etcdadpt.NotExistKey(key)), nil)
	if err != nil {
		return err
	}
	if resp.Succeeded {
		log.Info(fmt.Sprintf("put in queue[%s/%s]: consumer[%s/%s/%s/%s] -> provider[%s/%s/%s]", consumer.ServiceId, id,
			consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version,
			provider.Environment, provider.AppId, provider.ServiceName))
	}
	return nil
}

func TransferToMicroServiceDependency(ctx context.Context, key string) (*pb.MicroServiceDependency, error) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}

	opts := append(FromContext(ctx), etcdadpt.WithStrKey(key))
	res, err := sd.DependencyRule().Search(ctx, opts...)
	if err != nil {
		log.Error(fmt.Sprintf("get dependency rule[%s] failed", key), nil)
		return nil, err
	}
	if len(res.Kvs) != 0 {
		return res.Kvs[0].Value.(*pb.MicroServiceDependency), nil
	}
	return microServiceDependency, nil
}

func EqualServiceDependency(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	return stringA == stringB
}

func DiffServiceVersion(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA != stringB &&
		stringA[:strings.LastIndex(stringA, "/")+1] == stringB[:strings.LastIndex(stringB, "/")+1] {
		return true
	}
	return false
}

func toString(in *pb.MicroServiceKey) string {
	return path.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func parseAddOrUpdateRules(ctx context.Context, dep *Dependency) (createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := path.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		log.Error(fmt.Sprintf("update dependency rule failed, get consumer[%s/%s/%s/%s]'s dependency rule failed",
			dep.Consumer.Environment, dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version), err)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	createDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
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
	return
}

func parseOverrideRules(ctx context.Context, dep *Dependency) (createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := path.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		log.Error(fmt.Sprintf("override dependency rule failed, get consumer[%s/%s/%s/%s]'s dependency rule failed",
			dep.Consumer.Environment, dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version), err)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	createDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
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
	return
}

func syncDependencyRule(ctx context.Context, dep *Dependency, filter func(context.Context, *Dependency) (_, _, _ []*pb.MicroServiceKey)) error {
	// 更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{dep.Consumer.Environment, dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	createDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList := filter(ctx, dep)
	if len(createDependencyRuleList) == 0 && len(existDependencyRuleList) == 0 && len(deleteDependencyRuleList) == 0 {
		return nil
	}

	if len(deleteDependencyRuleList) != 0 {
		log.Info(fmt.Sprintf("delete consumer[%s]'s dependency rule %v", consumerFlag, deleteDependencyRuleList))
		dep.DeleteDependencyRuleList = deleteDependencyRuleList
	}

	if len(createDependencyRuleList) != 0 {
		log.Info(fmt.Sprintf("create consumer[%s]'s dependency rule %v", consumerFlag, createDependencyRuleList))
		dep.CreateDependencyRuleList = createDependencyRuleList
	}

	return dep.Commit(ctx)
}

func AddDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseAddOrUpdateRules)
}

func CreateDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseOverrideRules)
}

func IsNeedUpdate(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) *pb.MicroServiceKey {
	for _, tmp := range services {
		if DiffServiceVersion(tmp, service) {
			return tmp
		}
	}
	return nil
}

func ContainServiceDependency(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) (bool, error) {
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

func DeleteDependencyForDeleteService(domainProject string, serviceID string, service *pb.MicroServiceKey) (etcdadpt.OpOptions, error) {
	key := path.GenerateConsumerDependencyQueueKey(domainProject, serviceID, path.DepsQueueUUID)
	conDep := new(pb.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*pb.MicroServiceKey{}
	conDep.Override = true
	data, err := json.Marshal(conDep)
	if err != nil {
		return etcdadpt.OpOptions{}, err
	}
	return etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)), nil
}

func removeProviderRuleOfConsumer(ctx context.Context, domainProject string, cache map[string]bool) ([]etcdadpt.OpOptions, error) {
	key := path.GenerateConsumerDependencyRuleKey(domainProject, nil) + path.SPLIT
	resp, err := sd.DependencyRule().Search(ctx,
		etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	if err != nil {
		return nil, err
	}

	var ops []etcdadpt.OpOptions
	for _, keyValue := range resp.Kvs {
		var left []*pb.MicroServiceKey
		all := keyValue.Value.(*pb.MicroServiceDependency).Dependency
		for _, key := range all {
			id := path.GenerateProviderDependencyRuleKey(key.Tenant, key)
			exist, ok := cache[id]
			if !ok {
				_, exist, err = FindServiceIds(ctx, key, false)
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
			ops = append(ops, etcdadpt.OpDel(etcdadpt.WithKey(keyValue.Key)))
		} else {
			val, err := json.Marshal(&pb.MicroServiceDependency{Dependency: left})
			if err != nil {
				return nil, fmt.Errorf("%v, marshal %v", err, left)
			}
			ops = append(ops, etcdadpt.OpPut(etcdadpt.WithKey(keyValue.Key), etcdadpt.WithValue(val)))
		}
	}
	return ops, nil
}

func RemoveProviderRuleKeys(ctx context.Context, domainProject string, cache map[string]bool) ([]etcdadpt.OpOptions, error) {
	key := path.GenerateProviderDependencyRuleKey(domainProject, nil) + path.SPLIT
	resp, err := sd.DependencyRule().Search(ctx,
		etcdadpt.WithStrKey(key), etcdadpt.WithPrefix(), etcdadpt.WithKeyOnly())
	if err != nil {
		return nil, err
	}

	var ops []etcdadpt.OpOptions
	for _, keyValue := range resp.Kvs {
		id := util.BytesToStringWithNoCopy(keyValue.Key)
		exist, ok := cache[id]
		if !ok {
			_, key := path.GetInfoFromDependencyRuleKV(keyValue.Key)
			if key == nil || key.ServiceName == "*" {
				continue
			}

			_, exist, err = FindServiceIds(ctx, key, false)
			if err != nil {
				return nil, fmt.Errorf("find service %s/%s/%s/%s, %v",
					key.Tenant, key.AppId, key.ServiceName, key.Version, err)
			}
			cache[id] = exist
		}

		if !exist {
			ops = append(ops, etcdadpt.OpDel(etcdadpt.WithKey(keyValue.Key)))
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

	opts, err := RemoveProviderRuleKeys(ctx, domainProject, cache)
	if err != nil {
		return err
	}

	ops := append(append([]etcdadpt.OpOptions(nil), pOps...), opts...)
	if len(ops) == 0 {
		return nil
	}

	return etcdadpt.Txn(ctx, ops)
}

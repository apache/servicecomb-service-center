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
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/cache"
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/ServiceComb/service-center/server/mux"
	"golang.org/x/net/context"
	"strings"
	"time"
)

var consumerCache *cache.Cache
var providerCache *cache.Cache

/*
缓存2分钟过期
1分钟周期缓存consumers 遍历所有serviceid并查询consumers 做缓存
当发现新查询到的consumers列表变成0时则不做cache set操作
这样当consumers关系完全被删除也有1分钟的时间窗让实例变化推送到相应的consumers里 1分鐘后緩存也會自動清理
实例推送中的依赖发现实时性为T+1分钟
*/
func init() {
	d, _ := time.ParseDuration("2m")
	consumerCache = cache.New(d, d)
	providerCache = cache.New(d, d)
}

func GetConsumersInCache(ctx context.Context, domainProject string, providerId string, provider *pb.MicroService) ([]string, error) {
	// 查询所有consumer
	dr := NewProviderDependencyRelation(ctx, domainProject, providerId, provider)
	consumerIds, err := dr.GetDependencyConsumerIds()
	if err != nil {
		util.Logger().Errorf(err, "Get dependency consumerIds failed.%s", providerId)
		return nil, err
	}

	if len(consumerIds) == 0 {
		util.Logger().Warnf(nil, "Get consumer for publish from database is empty. %s, get from cache", providerId)
		consumerIds, found := consumerCache.Get(providerId)
		if found && len(consumerIds.([]string)) > 0 {
			return consumerIds.([]string), nil
		}
		return nil, nil
	}

	return consumerIds, nil
}

func GetProvidersInCache(ctx context.Context, domainProject string, consumerId string, consumer *pb.MicroService) ([]string, error) {
	// 查询所有provider
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumerId, consumer)
	providerIds, err := dr.GetDependencyProviderIds()
	if err != nil {
		util.Logger().Errorf(err, "Get dependency providerIds failed.%s", consumerId)
		return nil, err
	}

	if len(providerIds) == 0 {
		util.Logger().Warnf(nil, "Get consumer for publish from database is empty.%s , get from cache", consumerId)
		providerIds, found := providerCache.Get(consumerId)
		if found && len(providerIds.([]string)) > 0 {
			return providerIds.([]string), nil
		}
		return nil, nil
	}

	return providerIds, nil
}

func RefreshDependencyCache(ctx context.Context, domainProject string, serviceId string, service *pb.MicroService) error {
	dr := NewDependencyRelation(ctx, domainProject, serviceId, service, serviceId, service)
	consumerIds, err := dr.GetDependencyConsumerIds()
	if err != nil {
		util.Logger().Errorf(err, "%s,refresh dependency cache failed, get consumerIds failed.", serviceId)
		return err
	}
	providerIds, err := dr.GetDependencyProviderIds()
	if err != nil {
		util.Logger().Errorf(err, "%s,refresh dependency cache failed, get providerIds failed.", serviceId)
		return err
	}
	MsCache().Set(serviceId, service, 5*time.Minute)
	if len(consumerIds) == 0 {
		util.Logger().Infof("refresh dependency cache: this services %s has no consumer dependency.", serviceId)
	} else {
		consumerCache.Set(serviceId, consumerIds, 5*time.Minute)
	}
	if len(providerIds) == 0 {
		util.Logger().Infof("refresh dependency cache: this services %s has no consumer dependency.", serviceId)
	} else {
		providerCache.Set(serviceId, providerIds, 5*time.Minute)
	}
	return nil
}

func GetConsumerIds(ctx context.Context, domainProject string, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, err := GetRulesUtil(util.SetContext(util.CloneContext(ctx), "cacheOnly", "1"),
		domainProject, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}
	if len(providerRules) == 0 {
		return getConsumerIdsWithFilter(ctx, domainProject, provider.ServiceId, provider, noFilter)
	}

	rf := RuleFilter{
		DomainProject: domainProject,
		Provider:      provider,
		ProviderRules: providerRules,
	}

	allow, deny, err = getConsumerIdsWithFilter(ctx, domainProject, provider.ServiceId, provider, rf.Filter)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func getConsumerIdsWithFilter(ctx context.Context, domainProject, providerId string, provider *pb.MicroService,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	consumerIds, err := GetConsumersInCache(ctx, domainProject, providerId, provider)
	if err != nil {
		return nil, nil, err
	}
	return filterConsumerIds(ctx, consumerIds, filter)
}

func filterConsumerIds(ctx context.Context, consumerIds []string,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	l := len(consumerIds)
	if l == 0 {
		return nil, nil, nil
	}
	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerId := range consumerIds {
		ok, err := filter(ctx, consumerId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerId
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerId
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func noFilter(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func GetProviderIdsByConsumerId(ctx context.Context, domainProject, consumerId string, service *pb.MicroService) (allow []string, deny []string, _ error) {
	providerIdsInCache, err := GetProvidersInCache(ctx, domainProject, consumerId, service)
	if err != nil {
		return nil, nil, err
	}
	l := len(providerIdsInCache)
	rf := RuleFilter{
		DomainProject: domainProject,
	}
	allowIdx, denyIdx := 0, l
	providerIds := make([]string, l)
	copyCtx := util.SetContext(util.CloneContext(ctx), "cacheOnly", "1")
	for _, providerId := range providerIdsInCache {
		provider, err := GetService(ctx, domainProject, providerId)
		if provider == nil {
			continue
		}
		providerRules, err := GetRulesUtil(copyCtx, domainProject, provider.ServiceId)
		if err != nil {
			return nil, nil, err
		}
		if len(providerRules) == 0 {
			providerIds[allowIdx] = providerId
			allowIdx++
			continue
		}
		rf.Provider = provider
		rf.ProviderRules = providerRules
		ok, err := rf.Filter(ctx, consumerId)
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

func ProviderDependencyRuleExist(ctx context.Context, domainProject string, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) (bool, error) {
	providerKey := apt.GenerateProviderDependencyRuleKey(domainProject, provider)
	consumers, err := TransferToMicroServiceDependency(ctx, providerKey)
	if err != nil {
		return false, err
	}
	if len(consumers.Dependency) != 0 {
		isEqual, err := containServiceDependency(consumers.Dependency, consumer)
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

func AddServiceVersionRule(ctx context.Context, domainProject string, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey, consumerId string) error {
	//创建依赖一致
	exist, err := ProviderDependencyRuleExist(ctx, domainProject, provider, consumer)
	if exist || err != nil {
		return err
	}

	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		return err
	}
	err = CreateDependencyRuleForFind(ctx, domainProject, provider, consumer)
	lock.Unlock()
	return err
}

func UpdateServiceForAddDependency(ctx context.Context, consumerId string, providers []*pb.DependencyKey, domainProject string) error {
	conServiceKey := apt.GenerateServiceKey(domainProject, consumerId)
	service, err := GetService(ctx, domainProject, consumerId)
	if err != nil {
		util.Logger().Errorf(err, "create dependency faild: get service failed. consumerId %s", consumerId)
		return err
	}
	if service == nil {
		util.Logger().Errorf(nil, "create dependency faild: service not exist.serviceId %s", consumerId)
		return errors.New("Get service is empty")
	}

	service.Providers = providers
	data, err := json.Marshal(service)
	if err != nil {
		util.Logger().Errorf(err, "create dependency faild: marshal service failed.")
		return err
	}
	_, err = backend.Registry().Do(ctx,
		registry.PUT,
		registry.WithStrKey(conServiceKey),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(err, "create dependency faild: commit service data into etcd failed.")
		return err
	}
	return nil
}

func DeleteDependencyForService(ctx context.Context, consumer *pb.MicroServiceKey, serviceId string) ([]registry.PluginOp, error) {
	ops := []registry.PluginOp{}
	opsTmps := []registry.PluginOp{}
	domainProject := consumer.Tenant
	flag := map[string]bool{}
	//删除依赖规则
	conKey := apt.GenerateConsumerDependencyRuleKey(domainProject, consumer)
	providerValue, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		return nil, err
	}
	if providerValue != nil && len(providerValue.Dependency) != 0 {
		proProkey := ""
		for _, providerRule := range providerValue.Dependency {
			proProkey = apt.GenerateProviderDependencyRuleKey(domainProject, providerRule)
			consumers, err := TransferToMicroServiceDependency(ctx, proProkey)
			if err != nil {
				return nil, err
			}
			err = deleteDependencyRuleUtil(ctx, consumers, consumer, proProkey)
			if err != nil {
				return nil, err
			}
		}

		util.Logger().Debugf("conKey is %s.", conKey)
		ops = append(ops, registry.OpDel(registry.WithStrKey(conKey)))
	}
	//作为provider的依赖规则
	providerKey := apt.GenerateProviderDependencyRuleKey(domainProject, consumer)

	util.Logger().Debugf("providerKey is %s", providerKey)
	ops = append(ops, registry.OpDel(registry.WithStrKey(providerKey)))

	//删除依赖关系
	opsTmps, err = deleteDependencyUtil(ctx, "c", domainProject, serviceId, flag)
	if err != nil {
		return nil, err
	}
	ops = append(ops, opsTmps...)
	util.Logger().Debugf("flag is %s", flag)
	opsTmps, err = deleteDependencyUtil(ctx, "p", domainProject, serviceId, flag)
	if err != nil {
		return nil, err
	}
	util.Logger().Debugf("flag is %s", flag)
	ops = append(ops, opsTmps...)
	return ops, nil
}

func TransferToMicroServiceDependency(ctx context.Context, key string) (*pb.MicroServiceDependency, error) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}

	opts := append(FromContext(ctx), registry.WithStrKey(key))
	res, err := store.Store().DependencyRule().Search(ctx, opts...)
	if err != nil {
		util.Logger().Errorf(nil, "Get dependency rule failed.")
		return nil, err
	}
	if len(res.Kvs) != 0 {
		err = json.Unmarshal(res.Kvs[0].Value, microServiceDependency)
		if err != nil {
			util.Logger().Errorf(nil, "Unmarshal res failed.")
			return nil, err
		}
	} else {
		util.Logger().Infof("for key %s, no mircroServiceDependency stored", key)
	}
	return microServiceDependency, nil
}

func deleteDependencyRuleUtil(ctx context.Context, microServiceDependency *pb.MicroServiceDependency, service *pb.MicroServiceKey, serviceKey string) error {
	for key, serviceTmp := range microServiceDependency.Dependency {
		if ok := equalServiceDependency(serviceTmp, service); ok {
			microServiceDependency.Dependency = append(microServiceDependency.Dependency[:key], microServiceDependency.Dependency[key+1:]...)
			util.Logger().Debugf("delete versionRule from %s", serviceTmp.ServiceName)
			break
		}
	}
	opts := []registry.PluginOpOption{}
	if len(microServiceDependency.Dependency) == 0 {
		opts = append(opts, registry.DEL, registry.WithStrKey(serviceKey))
		util.Logger().Debugf("serviceKey is .", serviceKey)
		util.Logger().Debugf("After deleting versionRule from %s,provider's consumer is empty.", serviceKey)

	} else {
		data, err := json.Marshal(microServiceDependency)
		if err != nil {
			util.Logger().Errorf(nil, "Marshal tmpValue failed.")
			return err
		}
		opts = append(opts, registry.PUT, registry.WithStrKey(serviceKey), registry.WithValue(data))
		util.Logger().Debugf("serviceKey is %s.", serviceKey)
	}
	_, err := backend.Registry().Do(ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "Submit update dependency failed.")
		return err
	}
	return nil
}

func equalServiceDependency(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA == stringB {
		return true
	}
	return false
}

func toString(in *pb.MicroServiceKey) string {
	return apt.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func deleteDependencyUtil(ctx context.Context, serviceType string, domainProject string, serviceId string, flag map[string]bool) ([]registry.PluginOp, error) {
	serviceKey := apt.GenerateServiceDependencyKey(serviceType, domainProject, serviceId, "")
	rsp, err := store.Store().Dependency().Search(ctx,
		registry.WithStrKey(serviceKey),
		registry.WithPrefix())
	if err != nil {
		return nil, err
	}
	ops := []registry.PluginOp{}
	if rsp != nil {
		serviceTmpId := ""
		serviceTmpKey := ""
		deleteKey := ""
		for _, kv := range rsp.Kvs {
			tmpKeyArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
			serviceTmpId = tmpKeyArr[len(tmpKeyArr)-1]
			if serviceType == "p" {
				serviceTmpKey = apt.GenerateConsumerDependencyKey(domainProject, serviceTmpId, serviceId)
				deleteKey = util.StringJoin([]string{"c", serviceTmpId, serviceId}, "/")
			} else {
				serviceTmpKey = apt.GenerateProviderDependencyKey(domainProject, serviceTmpId, serviceId)
				deleteKey = util.StringJoin([]string{"p", serviceTmpId, serviceId}, "/")
			}
			if _, ok := flag[serviceTmpKey]; ok {
				util.Logger().Debugf("serviceTmpKey is more exist.%s", serviceTmpKey)
				continue
			}
			flag[serviceTmpKey] = true
			util.Logger().Infof("delete dependency %s", deleteKey)
			ops = append(ops, registry.OpDel(registry.WithStrKey(serviceTmpKey)))
		}
		util.Logger().Infof("delete dependency serviceKey is %s", serviceType+"/"+serviceId)
		ops = append(ops, registry.OpDel(registry.WithStrKey(serviceKey), registry.WithPrefix()))
	}
	return ops, nil
}

func CreateDependencyRule(ctx context.Context, dep *Dependency) error {
	//更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.Logger().Errorf(err, "maintain dependency rule failed, consumer %s: get consumer depedency rule failed.", consumerFlag)
		return err
	}

	unExistDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := containServiceDependency(dep.ProvidersRule, oldProviderRule); !ok {
			unExistDependencyRuleList = append(unExistDependencyRuleList, oldProviderRule)
		} else {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containServiceDependency(existDependencyRuleList, tmpProviderRule); !ok {
			newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		}
	}

	dep.err = make(chan error, 5)
	dep.chanNum = 0
	if len(unExistDependencyRuleList) != 0 {
		util.Logger().Infof("Unexist dependency rule remove for consumer %s, %v, ", consumerFlag, unExistDependencyRuleList)
		dep.removedDependencyRuleList = unExistDependencyRuleList
		dep.RemoveConsumerOfProviderRule()
	}

	if len(newDependencyRuleList) != 0 {
		util.Logger().Infof("New dependency rule add for consumer %s, %v, ", consumerFlag, newDependencyRuleList)
		dep.NewDependencyRuleList = newDependencyRuleList
		dep.AddConsumerOfProviderRule()
	}

	err = dep.UpdateProvidersRuleOfConsumer(conKey)
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

func CreateDependencyRuleForFind(ctx context.Context, domainProject string, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) error {
	//更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{consumer.AppId, consumer.ServiceName, consumer.Version}, "/")
	conKey := apt.GenerateConsumerDependencyRuleKey(domainProject, consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.Logger().Errorf(err, "get dependency rule failed, consumer %s: get consumer depedency rule failed.", consumerFlag)
		return err
	}
	if isDependencyAll(oldProviderRules) {
		util.Logger().Infof("find update dep rule, exist * for %v", consumer)
		return nil
	}
	opts := make([]registry.PluginOp, 0)
	if oldProviderRule := isNeedUpdate(oldProviderRules.Dependency, provider); oldProviderRule != nil {
		opt, err := deleteConsumerDepOfProviderRule(ctx, domainProject, oldProviderRule, consumer)
		if err != nil {
			util.Logger().Errorf(err, "marshal consumerDepRules failed for delete consumer rule from provider rule's dep.%s", consumerFlag)
			return err
		}
		util.Logger().Infof("delete consumer %v from provider dep %v", consumer, oldProviderRule)
		opts = append(opts, opt)

		oldRule, opt, err := updateDepRuleUtil(conKey, oldProviderRules, provider)
		if err != nil {
			util.Logger().Errorf(err, "update provider rule into consumer's dep rule failed, %s", consumerFlag)
			return err
		}
		util.Logger().Infof("update provider %v(from version '%s') into consumer dep %s", provider, oldRule, consumerFlag)
		opts = append(opts, opt)

		opt, err = updateProviderRuleDep(ctx, domainProject, provider, consumer)
		if err != nil {
			return err
		}
		opts = append(opts, opt)
	} else {
		if !isExist(oldProviderRules.Dependency, provider) {
			opt, err := addDepRuleUtil(conKey, oldProviderRules, provider)
			if err != nil {
				util.Logger().Errorf(err, "add provider rule into consumer's dep rule failed, %s", consumerFlag)
				return err
			}
			util.Logger().Infof("add consumer dep, %s, add %v", consumerFlag, provider)
			opts = append(opts, opt)
		}

		proKey := apt.GenerateProviderDependencyRuleKey(domainProject, provider)
		consumerDepRules, err := TransferToMicroServiceDependency(ctx, proKey)
		if err != nil {
			util.Logger().Errorf(err, "get consumer rule of provider failed,%v", provider)
			return err
		}
		if !isExist(consumerDepRules.Dependency, consumer) {
			opt, err := addDepRuleUtil(proKey, consumerDepRules, consumer)
			if err != nil {
				util.Logger().Errorf(err, "add consumer rule into provider's dep rule failed,%s", consumerFlag)
				return err
			}
			util.Logger().Infof("add provider dep, %s, add %v", consumerFlag, provider)
			opts = append(opts, opt)
		}
	}

	if len(opts) != 0 {
		_, err = backend.Registry().Txn(ctx, opts)
		if err != nil {
			util.Logger().Errorf(err, "update dep rule for consumer failed, %s", consumerFlag)
			return err
		}
	}
	return nil
}

func updateProviderRuleDep(ctx context.Context, domainProject string, providerRule , consumer *pb.MicroServiceKey) (registry.PluginOp, error) {
	proKey := apt.GenerateProviderDependencyRuleKey(domainProject, providerRule)
	consumerDepRules, err := TransferToMicroServiceDependency(ctx, proKey)
	opt := registry.PluginOp{}
	if err != nil {
		util.Logger().Errorf(err, "get provider rule's dep failed, providerRule %v, consumer %v", providerRule, consumer)
		return opt, err
	}
	opt, err = addDepRuleUtil(proKey, consumerDepRules, consumer)
	if err != nil {
		util.Logger().Errorf(err, "add consumer into provider's dep rule failed, providerRule %v, consumer %v", providerRule, consumer)
		return opt, err
	}
	return opt, nil
}

func isDependencyAll(dep *pb.MicroServiceDependency) bool {
	for _, servicedep := range dep.Dependency {
		if servicedep.ServiceName == "*" {
			return true
		}
	}
	return false
}

func isExist(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) bool {
	for _, tmp := range services {
		if equalServiceDependency(tmp, service) {
			return true
		}
	}
	return false
}

func deleteConsumerDepOfProviderRule(ctx context.Context, domainProject string, providerRule *pb.MicroServiceKey, deleteConsumer *pb.MicroServiceKey) (registry.PluginOp, error) {
	proKey := apt.GenerateProviderDependencyRuleKey(domainProject, providerRule)
	consumerDepRules, err := TransferToMicroServiceDependency(ctx, proKey)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return deleteDepRuleUtil(proKey, consumerDepRules, deleteConsumer)
}

func deleteDepRuleUtil(key string, deps *pb.MicroServiceDependency, deleteDepRule *pb.MicroServiceKey) (registry.PluginOp, error) {
	for key, consumerDepRule := range deps.Dependency {
		if equalServiceDependency(consumerDepRule, deleteDepRule) {
			deps.Dependency = append(deps.Dependency[:key], deps.Dependency[key+1:]...)
			break
		}
	}
	data, err := json.Marshal(deps)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func addDepRuleUtil(key string, deps *pb.MicroServiceDependency, updateDepRule *pb.MicroServiceKey) (registry.PluginOp, error) {
	deps.Dependency = append(deps.Dependency, updateDepRule)
	data, err := json.Marshal(deps)
	if err != nil {
		util.Logger().Errorf(err, "marshal consumerDepRules failed for delete consumer rule from provider rule's dep.")
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func updateDepRuleUtil(key string, deps *pb.MicroServiceDependency, updateDepRule *pb.MicroServiceKey) (string, registry.PluginOp, error) {
	oldRule := ""
	for _, serviceRule := range deps.Dependency {
		if serviceRule.Environment == updateDepRule.Environment &&
			serviceRule.AppId == updateDepRule.AppId &&
			serviceRule.ServiceName == updateDepRule.ServiceName &&
			serviceRule.Version != updateDepRule.Version {
			oldRule = serviceRule.Version
			serviceRule.Version = updateDepRule.Version
			break
		}
	}
	data, err := json.Marshal(deps)
	if err != nil {
		return oldRule, registry.PluginOp{}, err
	}
	return oldRule, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func isNeedUpdate(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) *pb.MicroServiceKey {
	for _, tmp := range services {
		if tmp.Environment == service.Environment &&
			tmp.AppId == service.AppId &&
			tmp.ServiceName == service.ServiceName &&
			tmp.Version != service.Version {
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

// fuzzyMatch: 是否使用模糊规则
func validateMicroServiceKey(in *pb.MicroServiceKey, fuzzyMatch bool) error {
	if fuzzyMatch {
		// provider的ServiceName, Version支持模糊规则
		return apt.ProviderMsValidator.Validate(in)
	} else {
		return apt.DependencyMSValidator.Validate(in)
	}
}

func BadParamsResponse(detailErr string) *pb.CreateDependenciesResponse {
	util.Logger().Errorf(nil, "Request params is invalid.")
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(scerr.ErrInvalidParams, detailErr),
	}
}

func ParamsChecker(consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey) *pb.CreateDependenciesResponse {
	if err := validateMicroServiceKey(consumerInfo, false); err != nil {
		return BadParamsResponse(err.Error())
	}
	if providersInfo == nil {
		return BadParamsResponse("Invalid request body for provider info.")
	}
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			util.Logger().Debugf("%s 's provider contains *.", consumerInfo.ServiceName)
			break
		}
		if len(providerInfo.AppId) == 0 {
			providerInfo.AppId = consumerInfo.AppId
		}
		if err := validateMicroServiceKey(providerInfo, true); err != nil {
			return BadParamsResponse(err.Error())
		}

		version := providerInfo.Version
		providerInfo.Version = ""
		if _, ok := flag[toString(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider or (serviceName and appid is same).")
		} else {
			flag[toString(providerInfo)] = true
		}
		providerInfo.Version = version
	}
	return nil
}

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
	go dep.removeConsumerOfProviderRule()
}

func (dep *Dependency) removeConsumerOfProviderRule() {
	ctx := context.TODO()
	opts := make([]registry.PluginOp, 0, len(dep.removedDependencyRuleList))
	for _, providerRule := range dep.removedDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(dep.DomainProject, providerRule)
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
		_, err := backend.Registry().Txn(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) AddConsumerOfProviderRule() {
	dep.chanNum++
	go dep.addConsumerOfProviderRule()
}

func (dep *Dependency) addConsumerOfProviderRule() {
	ctx := context.TODO()
	opts := []registry.PluginOp{}
	for _, prividerRule := range dep.NewDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(dep.DomainProject, prividerRule)
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
		if prividerRule.ServiceName == "*" {
			break
		}
	}
	if len(opts) != 0 {
		_, err := backend.Registry().Txn(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) UpdateProvidersRuleOfConsumer(conKey string) error {
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

type DependencyRelation struct {
	ctx           context.Context
	domainProject string
	consumerId    string
	consumer      *pb.MicroService
	providerId    string
	provider      *pb.MicroService
}

func NewProviderDependencyRelation(ctx context.Context, domainProject string, providerId string, provider *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, "", nil, providerId, provider)
}

func NewConsumerDependencyRelation(ctx context.Context, domainProject string, consumerId string, consumer *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, consumerId, consumer, "", nil)
}

func NewDependencyRelation(ctx context.Context, domainProject string, consumerId string, consumer *pb.MicroService, providerId string, provider *pb.MicroService) *DependencyRelation {
	return &DependencyRelation{
		ctx:           ctx,
		domainProject: domainProject,
		consumerId:    consumerId,
		consumer:      consumer,
		providerId:    providerId,
		provider:      provider,
	}
}

func (dr *DependencyRelation) GetDependencyProviders() ([]*pb.MicroService, error) {
	providerIds, err := dr.GetDependencyProviderIds()
	if err != nil {
		return nil, err
	}
	services := make([]*pb.MicroService, 0)
	for _, providerId := range providerIds {
		provider, err := GetService(dr.ctx, dr.domainProject, providerId)
		if err != nil {
			return nil, err
		}
		if provider == nil {
			util.Logger().Warnf(nil, "Provider not exist, %s", providerId)
			continue
		}
		services = append(services, provider)
	}
	return services, nil
}

func (dr *DependencyRelation) GetDependencyProviderIds() ([]string, error) {
	if dr.consumer == nil {
		util.LOGGER.Infof("dr.consumer is nil ------->")
		return nil, fmt.Errorf("Invalid consumer")
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)

	conKey := apt.GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)
	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, conKey)
	if err != nil {
		return nil, err
	}
	return dr.getDependencyProviderIds(consumerDependency.Dependency)
}

func (dr *DependencyRelation) getDependencyProviderIds(providerRules []*pb.MicroServiceKey) ([]string, error) {
	provideServiceIds := make([]string, 0)
	opts := FromContext(dr.ctx)
	for _, provider := range providerRules {
		switch {
		case provider.ServiceName == "*":
			util.Logger().Infof("Rely all service,* type, consumerId %s", dr.consumerId)
			splited := strings.Split(apt.GenerateServiceIndexKey(provider), "/")
			allServiceKey := util.StringJoin(splited[:len(splited)-3], "/") + "/"
			sopts := append(opts,
				registry.WithStrKey(allServiceKey),
				registry.WithPrefix())
			resp, err := store.Store().Service().Search(dr.ctx, sopts...)
			if err != nil {
				util.Logger().Errorf(err, "Add dependency failed, rely all service: get all services failed.")
				return provideServiceIds, err
			}

			for _, kv := range resp.Kvs {
				provideServiceIds = append(provideServiceIds, util.BytesToStringWithNoCopy(kv.Value))
			}
			return provideServiceIds, nil
		default:
			serviceIds, err := FindServiceIds(dr.ctx, provider.Version, provider)
			if err != nil {
				util.Logger().Errorf(err, "Get providerIds failed, service: %s/%s/%s",
					provider.AppId, provider.ServiceName, provider.Version)
				return provideServiceIds, err
			}
			if len(serviceIds) == 0 {
				util.Logger().Warnf(nil, "Get providerIds is empty, service: %s/%s/%s does not exist",
					provider.AppId, provider.ServiceName, provider.Version)
				continue
			}
			provideServiceIds = append(provideServiceIds, serviceIds...)
		}
	}
	return provideServiceIds, nil
}

func (dr *DependencyRelation) GetDependencyConsumers() ([]*pb.MicroService, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		util.Logger().Errorf(err, "Get consumers of provider rule failed, %s", dr.providerId)
		return nil, err
	}
	consumers := make([]*pb.MicroService, 0)

	for _, consumer := range consumerDependAllList {
		service, err := dr.getServiceByMicroServiceKey(dr.domainProject, consumer)
		if err != nil {
			return nil, err
		}
		if service == nil {
			util.Logger().Warnf(nil, "Consumer not exist,%v", service)
			continue
		}
		consumers = append(consumers, service)
	}
	return consumers, nil
}

func (dr *DependencyRelation) getServiceByMicroServiceKey(domainProject string, service *pb.MicroServiceKey) (*pb.MicroService, error) {
	serviceId, err := GetServiceId(dr.ctx, service)
	if err != nil {
		return nil, err
	}
	if len(serviceId) == 0 {
		util.Logger().Warnf(nil, "Service not exist,%v", service)
		return nil, nil
	}
	return GetService(dr.ctx, domainProject, serviceId)
}

func (dr *DependencyRelation) GetDependencyConsumerIds() ([]string, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		return nil, err
	}
	consumerIds := make([]string, 0)
	for _, consumer := range consumerDependAllList {
		consumerId, err := GetServiceId(context.TODO(), consumer)
		if err != nil {
			util.Logger().Errorf(err, "Get consumer failed, %v", consumer)
			return nil, err
		}
		if len(consumerId) == 0 {
			util.Logger().Warnf(nil, "Get consumer not exist, %v", consumer)
			continue
		}
		consumerIds = append(consumerIds, consumerId)
	}
	return consumerIds, nil

}

func (dr *DependencyRelation) getDependencyConsumersOfProvider() ([]*pb.MicroServiceKey, error) {
	if dr.provider == nil {
		util.LOGGER.Infof("dr.provider is nil ------->")
		return nil, fmt.Errorf("Invalid provider")
	}
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		util.Logger().Errorf(err, "Get consumer that depend on all services failed, %s", dr.providerId)
		return nil, err
	}

	consumerDependList, err := dr.getConsumerOfSameServiceNameAndAppId(providerService)
	if err != nil {
		util.Logger().Errorf(err, "Get consumer that depend on same serviceName and appid rule failed, %s", dr.providerId)
		return nil, err
	}
	consumerDependAllList = append(consumerDependAllList, consumerDependList...)
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*pb.MicroServiceKey, error) {
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	relyAllKey := apt.GenerateProviderDependencyRuleKey(dr.domainProject, providerService)
	opts := append(FromContext(dr.ctx), registry.WithStrKey(relyAllKey))
	rsp, err := store.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "get consumer that rely all service failed.")
		return nil, err
	}
	dependency := &pb.MicroServiceDependency{}
	if len(rsp.Kvs) != 0 {
		util.Logger().Infof("consumer that rely all service exist.ServiceName: %s.", dr.provider.ServiceName)
		err = json.Unmarshal(rsp.Kvs[0].Value, dependency)
		if err != nil {
			return nil, err
		}
		return dependency.Dependency, nil
	}
	return dependency.Dependency, nil
}

func (dr *DependencyRelation) getConsumerOfSameServiceNameAndAppId(provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	prefix := apt.GenerateProviderDependencyRuleKey(dr.domainProject, provider)
	provider.Version = providerVersion

	opts := append(FromContext(dr.ctx),
		registry.WithStrKey(prefix),
		registry.WithPrefix())
	rsp, err := store.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "get all dependency rule failed: provider rule key %v.", provider)
		return nil, err
	}

	allConsumers := make([]*pb.MicroServiceKey, 0, len(rsp.Kvs))
	var latestServiceId []string

	for _, kv := range rsp.Kvs {
		dependency := &pb.MicroServiceDependency{
			Dependency: []*pb.MicroServiceKey{},
		}
		providerVersionRuleArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		providerVersionRule := providerVersionRuleArr[len(providerVersionRuleArr)-1]
		if providerVersionRule == "latest" {
			if latestServiceId == nil {
				latestServiceId, err = FindServiceIds(dr.ctx, providerVersionRule, provider)
				if err != nil {
					util.Logger().Errorf(err, "Get latest service failed.")
					return nil, err
				}
			}
			if len(latestServiceId) == 0 {
				util.Logger().Infof("%s 's providerId is empty,no this service.", provider.ServiceName)
				continue
			}
			if dr.providerId != latestServiceId[0] {
				continue
			}

		} else {
			if !VersionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}

		util.Logger().Debugf("providerETCD is %s", providerVersionRuleArr)
		err = json.Unmarshal(kv.Value, dependency)
		if err != nil {
			util.Logger().Errorf(err, "Unmarshal consumers failed.")
			return nil, err
		}
		allConsumers = append(allConsumers, dependency.Dependency...)
	}
	return allConsumers, nil
}

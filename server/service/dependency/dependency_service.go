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
package dependency

import (
	"encoding/json"
	"errors"
	"github.com/ServiceComb/service-center/pkg/common/cache"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strconv"
	"strings"
	"time"
)

var consumerCache *cache.Cache

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
	go autoSyncConsumers()
}

//TODO
func autoSyncConsumers() {
	//ticker := time.NewTicker(time.Minute * 1)
	//for t := range ticker.C {
	//	util.LOGGER.Debug(fmt.Sprintf("sync consumers at %s", t))
	//	keys := microservice.GetServiceWithRev(context.TODO())
	//	for _, v := range keys {
	//		util.LOGGER.Debug(fmt.Sprintf("sync consumers for %s", v))
	//		domainAndId := strings.Split(v, ":::")
	//		// 查询所有consumer
	//		key := apt.GenerateProviderDependencyKey(domainAndId[0], domainAndId[1], "")
	//		resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
	//			Action:     registry.GET,
	//			Key:        util.StringToBytesWithNoCopy(key),
	//			WithPrefix: true,
	//			KeyOnly:    true,
	//		})
	//		if err != nil {
	//			util.LOGGER.Errorf(err, "query service consumers failed, provider id %s", domainAndId[1])
	//		}
	//		if len(resp.Kvs) != 0 {
	//			consumerCache.Set(v, resp.Kvs, 0)
	//		}
	//
	//	}
	//}
}
func GetConsumersInCache(ctx context.Context, tenant string, providerId string) ([]*mvccpb.KeyValue, error) {
	// 查询所有consumer
	key := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	resp, err := store.Store().Dependency().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
		KeyOnly:    true,
		Mode:       registry.MODE_CACHE,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "query service consumers failed, provider id %s", providerId)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		// 如果没有，使用最后一次更新的缓存
		util.LOGGER.Debugf("query service consumers in cache %s", providerId)
		consumerList, found := consumerCache.Get(providerId)
		if found && len(consumerList.([]*mvccpb.KeyValue)) > 0 {
			return consumerList.([]*mvccpb.KeyValue), nil
		}
		return []*mvccpb.KeyValue{}, nil
	}
	consumerCache.Set(providerId, resp.Kvs, 0)

	util.LOGGER.Debugf("query service consumers in center %s", key)
	return resp.Kvs, nil
}

func RefreshDependencyCache(tenant string, providerId string, provider *pb.MicroService) error {
	key := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	resp, err := store.Store().Dependency().Search(context.Background(), &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
		KeyOnly:    true,
		Mode:       registry.MODE_CACHE,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "refresh dependency cache failed: query service consumers failed, provider id %s", providerId)
		return err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		util.LOGGER.Infof("refresh dependency cache: this services %s has no consumer dependency.", providerId)
	} else {
		ms.MsCache().Set(providerId, provider, 5*time.Minute)
		consumerCache.Set(providerId, resp.Kvs, 5*time.Minute)
	}
	return nil
}

func UpdateDependency(dep *Dependency) error {
	addDepencyList, err := dep.GetAddDependencyProviderIds()
	if err != nil {
		return err
	}
	deleteDepencyList, err := dep.GetDeleteDependencyProviderIds()
	if err != nil {
		return err
	}
	opts := make([]*registry.PluginOp, 0, 2*(len(addDepencyList)+len(deleteDepencyList)))
	var isEqual bool
	sameServiceIdList := make([]string, 0, len(addDepencyList))
	for _, addDepencyProviderId := range addDepencyList {
		for _, deleteDependencyProviderId := range deleteDepencyList {
			if addDepencyProviderId == deleteDependencyProviderId {
				sameServiceIdList = append(sameServiceIdList, addDepencyProviderId)
				isEqual = true
				break
			}
		}
		if !isEqual {
			optProsTmps := putServiceDependency(dep.ConsumerId, addDepencyProviderId, dep.Tenant)
			opts = append(opts, optProsTmps...)
		}
		isEqual = false
	}
	for _, deleteDependencyProviderId := range deleteDepencyList {
		for _, tmpServiceId := range sameServiceIdList {
			if tmpServiceId == deleteDependencyProviderId {
				isEqual = true
				break
			}
		}
		if !isEqual {
			optsTmp := deleteDependency(dep.Tenant, dep.ConsumerId, deleteDependencyProviderId)
			opts = append(opts, optsTmp...)
		}
		isEqual = false
	}
	if len(opts) != 0 {
		err := registry.BatchCommit(context.TODO(), opts)
		if err != nil {
			util.LOGGER.Errorf(err, err.Error())
			return err
		}
	}
	return nil
}

func putServiceDependency(consumerId string, providerId string, tenant string) []*registry.PluginOp {
	conProKey := apt.GenerateConsumerDependencyKey(tenant, consumerId, providerId)
	proProKey := apt.GenerateProviderDependencyKey(tenant, providerId, consumerId)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	util.LOGGER.Debugf("%s %s %s", conProKey, " timeStamp is ", timestamp)
	optPro := &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(conProKey),
		Value:  util.StringToBytesWithNoCopy(timestamp),
	}
	optCon := &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(proProKey),
		Value:  util.StringToBytesWithNoCopy(timestamp),
	}
	optPros := []*registry.PluginOp{}
	optPros = append(optPros, optPro)
	optPros = append(optPros, optCon)
	util.LOGGER.Infof("add dependency rule: consumerId %s, providerId %s", consumerId, providerId)
	return optPros
}

func UpdateAsProviderDependency(ctx context.Context, providerServiseId string, provider *pb.MicroServiceKey) error {
	//查询etcd里是否存在带*的情况，则添加与对应的consumer与该provider的依赖关系
	tenant := util.ParseTenantProject(ctx)
	consumerDependAllList, err := getConsumerOfDependAllServices(tenant, provider)
	if err != nil {
		util.LOGGER.Errorf(err, "Get consumer that depend on all services failed, %s", providerServiseId)
		return err
	}

	consumerDependList, err := getConsumersFromAllVersionRuleOfOne(tenant, providerServiseId, provider)
	if err != nil {
		util.LOGGER.Errorf(err, "Get consumer that depend on same serviceName and appid rule failed, %s", providerServiseId)
		return err
	}

	allConsumers := make([]*pb.MicroServiceKey, 0, len(consumerDependAllList)+len(consumerDependList))
	allConsumers = append(allConsumers, consumerDependAllList...)
	allConsumers = append(allConsumers, consumerDependList...)

	if len(allConsumers) == 0 {
		util.LOGGER.Infof("%s as a provider, no consumer use it.", provider.ServiceName)
		return nil
	}
	//更新作为provider的依赖关系
	opts := make([]*registry.PluginOp, 0, 2*len(allConsumers))
	flag := make(map[string]bool, len(allConsumers))
	for _, consumer := range allConsumers {
		consumerServiceid, err := ms.GetServiceId(ctx, consumer)
		if err != nil {
			util.LOGGER.Errorf(err, "Get consumer's serviceId failed.")
			return err
		}
		if len(consumerServiceid) == 0 {
			util.LOGGER.Warnf(nil, "Get consumer's serviceId is empty, skip serviceName: %s, appId: %s, version: %s", consumer.ServiceName, consumer.AppId, consumer.Version)
			continue
		}
		if _, ok := flag[consumerServiceid]; ok {
			util.LOGGER.Infof("consumerServiceid %s more exists.", consumerServiceid)
			continue
		} else {
			flag[consumerServiceid] = true
		}
		optsTmp := putServiceDependency(consumerServiceid, providerServiseId, tenant)
		opts = append(opts, optsTmp...)
	}
	if len(opts) != 0 {
		err = registry.BatchCommit(ctx, opts)
		if err != nil {
			util.LOGGER.Errorf(err, "Add provider dependency rule failed: provider %v", provider)
			return err
		}
	}
	return nil
}

func getConsumersFromAllVersionRuleOfOne(tenant string, providerServiceId string, provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	proKey := apt.GenerateProviderDependencyRuleKey(tenant, provider)
	provider.Version = providerVersion
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(proKey),
		WithPrefix: true,
	}
	rsp, err := store.Store().DependencyRule().Search(context.TODO(), opt)
	if err != nil {
		util.LOGGER.Errorf(err, "get all dependency rule failed: provider rule key %v.", provider)
		return nil, err
	}
	if rsp == nil {
		util.LOGGER.Warnf(nil, "get same appId and serviceName dependency rule is nul: provider rule key %v.", provider)
		return nil, nil
	}
	allConsumers := make([]*pb.MicroServiceKey, 0)
	for _, kv := range rsp.Kvs {
		dependency := &pb.MicroServiceDependency{
			Dependency: []*pb.MicroServiceKey{},
		}
		providerVersionRuleArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		providerVersionRule := providerVersionRuleArr[len(providerVersionRuleArr)-1]
		if providerVersionRule == "latest" {
			latestServiceId, err := ms.FindServiceIds(context.TODO(), providerVersionRule, &pb.MicroServiceKey{
				Tenant:      tenant,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "Get latest service failed.")
				return nil, err
			}
			if len(latestServiceId) == 0 {
				util.LOGGER.Infof("%s 's providerId is empty,no this service.", provider.ServiceName)
				continue
			}
			if providerServiceId != latestServiceId[0] {
				continue
			}
			providerIds, err := getAllVersionServiceIdsForOne(&pb.MicroServiceKey{
				Tenant:      tenant,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
			})
			if err != nil {
				return nil, err
			}
			if len(providerIds) != 0 {
				err = json.Unmarshal(kv.Value, dependency)
				if err != nil {
					return nil, err
				}
				deleteDependencyOptList := make([]*registry.PluginOp, 0)
				consumers := dependency.Dependency
				consumerIds := make([]string, 0, len(consumers))
				for _, consumer := range consumers {
					consumerId, err := ms.GetServiceId(context.TODO(), consumer)
					if err != nil {
						util.LOGGER.Errorf(err, "Delete lastest old version dependency failed: provider %v", provider)
						return nil, err
					}
					if len(consumerId) == 0 {
						util.LOGGER.Errorf(err, "Delete lastest old version dependency, get consumer is empty: consumer %v", consumer)
						continue
					}
					consumerIds = append(consumerIds, consumerId)
				}
				for _, providerId := range providerIds {
					if providerId == providerServiceId {
						continue
					}
					for _, consumerId := range consumerIds {
						tmp := deleteDependency(tenant, consumerId, providerId)
						deleteDependencyOptList = append(deleteDependencyOptList, tmp...)
					}
				}
				if len(deleteDependencyOptList) != 0 {
					err = registry.BatchCommit(context.TODO(), deleteDependencyOptList)
					if err != nil {
						util.LOGGER.Errorf(err, "Add provider dependency rule failed: provider %v", provider)
						return nil, err
					}
				}
			}
		} else {
			if !ms.VersionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}

		util.LOGGER.Debugf("providerETCD is %s", providerVersionRuleArr)
		if len(dependency.Dependency) == 0 {
			err = json.Unmarshal(kv.Value, dependency)
			if err != nil {
				util.LOGGER.Errorf(err, "Unmarshal consumers failed.")
				return nil, err
			}
		}
		util.LOGGER.Infof("Add dependency as provider, provider: serviecName(%s), version(%s) .consumer is %v",
			provider.ServiceName, providerVersionRule, dependency.Dependency)
		allConsumers = append(allConsumers, dependency.Dependency...)
	}
	return allConsumers, nil
}

func getAllVersionServicesForOne(service *pb.MicroServiceKey) (*registry.PluginResponse, error) {
	key := apt.GenerateServiceIndexKey(service)
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
	}
	resp, err := store.Store().ServiceIndex().Search(context.TODO(), opt)
	return resp, err
}

func getAllVersionServiceIdsForOne(service *pb.MicroServiceKey) ([]string, error) {
	resp, err := getAllVersionServicesForOne(service)
	if err != nil {
		return nil, err
	}
	serviceIdList := make([]string, 0)
	for _, kv := range resp.Kvs {
		serviceId := util.BytesToStringWithNoCopy(kv.Value)
		serviceIdList = append(serviceIdList, serviceId)
	}
	return serviceIdList, nil
}

func getConsumerOfDependAllServices(tenant string, provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {

	relyAllKey := apt.GenerateProviderDependencyRuleKey(tenant, &pb.MicroServiceKey{
		ServiceName: "*",
	})
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(relyAllKey),
	}
	rsp, err := store.Store().DependencyRule().Search(context.TODO(), opt)
	if err != nil {
		util.LOGGER.Errorf(err, "get consumer that rely all service failed.")
		return nil, err
	}
	if len(rsp.Kvs) != 0 {
		dependency := &pb.MicroServiceDependency{}
		util.LOGGER.Infof("consumer that rely all service exist.ServiceName: %s.", provider.ServiceName)
		err = json.Unmarshal(rsp.Kvs[0].Value, dependency)
		if err != nil {
			return nil, err
		}
		return dependency.Dependency, nil
	}
	return nil, nil
}

func DeleteDependencyForService(ctx context.Context, consumer *pb.MicroServiceKey, serviceId string) ([]*registry.PluginOp, error) {
	opts := []*registry.PluginOp{}
	optsTmps := []*registry.PluginOp{}
	tenant := consumer.Tenant
	flag := map[string]bool{}
	//删除依赖规则
	conKey := apt.GenerateConsumerDependencyRuleKey(tenant, consumer)
	err, providerValue := transferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		return nil, err
	}
	if providerValue != nil && len(providerValue.Dependency) != 0 {
		proProkey := ""
		for _, providerRule := range providerValue.Dependency {
			proProkey = apt.GenerateProviderDependencyRuleKey(tenant, providerRule)
			err, consumers := transferToMicroServiceDependency(ctx, proProkey)
			if err != nil {
				return nil, err
			}
			err = deleteDependencyRuleUtil(ctx, consumers, consumer, proProkey)
			if err != nil {
				return nil, err
			}
		}

		opt := &registry.PluginOp{
			Action: registry.DELETE,
			Key:    util.StringToBytesWithNoCopy(conKey),
		}
		util.LOGGER.Debugf("conKey is %s.", conKey)
		opts = append(opts, opt)
	}
	//作为provider的依赖规则
	providerKey := apt.GenerateProviderDependencyRuleKey(tenant, consumer)
	opt := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(providerKey),
	}
	util.LOGGER.Debugf("providerKey is %s", providerKey)
	opts = append(opts, opt)

	//删除依赖关系
	optsTmps, err = deleteDependencyUtil(ctx, "c", tenant, serviceId, flag)
	if err != nil {
		return nil, err
	}
	opts = append(opts, optsTmps...)
	util.LOGGER.Debugf("flag is %s", flag)
	optsTmps, err = deleteDependencyUtil(ctx, "p", tenant, serviceId, flag)
	if err != nil {
		return nil, err
	}
	util.LOGGER.Debugf("flag is %s", flag)
	opts = append(opts, optsTmps...)
	return opts, nil
}

func transferToMicroServiceDependency(ctx context.Context, key string) (error, *pb.MicroServiceDependency) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(key),
	}
	res, err := store.Store().DependencyRule().Search(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Get dependency rule failed.")
		return err, nil
	}
	if len(res.Kvs) != 0 {
		err = json.Unmarshal(res.Kvs[0].Value, microServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Unmarshal res failed.")
			return err, nil
		}
	} else {
		util.LOGGER.Infof("for key %s, no mircroServiceDependency stored", key)
	}
	return nil, microServiceDependency
}

func deleteDependencyRuleUtil(ctx context.Context, microServiceDependency *pb.MicroServiceDependency, service *pb.MicroServiceKey, serviceKey string) error {
	for key, serviceTmp := range microServiceDependency.Dependency {
		if ok := equalServiceDependency(serviceTmp, service); ok {
			microServiceDependency.Dependency = append(microServiceDependency.Dependency[:key], microServiceDependency.Dependency[key+1:]...)
			util.LOGGER.Debugf("delete versionRule from %s", serviceTmp.ServiceName)
			break
		}
	}
	opt := &registry.PluginOp{}
	if len(microServiceDependency.Dependency) == 0 {
		opt = &registry.PluginOp{
			Action: registry.DELETE,
			Key:    util.StringToBytesWithNoCopy(serviceKey),
		}
		util.LOGGER.Debugf("serviceKey is .", serviceKey)
		util.LOGGER.Debugf("After deleting versionRule from %s,provider's consumer is empty.", serviceKey)

	} else {
		data, err := json.Marshal(microServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
			return err
		}
		opt = &registry.PluginOp{
			Action: registry.PUT,
			Key:    util.StringToBytesWithNoCopy(serviceKey),
			Value:  data,
		}
		util.LOGGER.Debugf("serviceKey is %s.", serviceKey)
	}
	_, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(err, "Submit update dependency failed.")
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
	return util.StringJoin([]string{
		in.Tenant,
		in.AppId,
		in.Stage,
		in.ServiceName,
		in.Version,
	}, "")
}

func deleteDependencyUtil(ctx context.Context, serviceType string, tenant string, serviceId string, flag map[string]bool) ([]*registry.PluginOp, error) {
	serviceKey := apt.GenerateServiceDependencyKey(serviceType, tenant, serviceId, "")
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(serviceKey),
		WithPrefix: true,
	}
	rsp, err := store.Store().Dependency().Search(ctx, opt)
	if err != nil {
		return nil, err
	}
	opts := []*registry.PluginOp{}
	if rsp != nil {
		serviceTmpId := ""
		serviceTmpKey := ""
		deleteKey := ""
		for _, kv := range rsp.Kvs {
			tmpKeyArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
			serviceTmpId = tmpKeyArr[len(tmpKeyArr)-1]
			if serviceType == "p" {
				serviceTmpKey = apt.GenerateConsumerDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = util.StringJoin([]string{"c", serviceTmpId, serviceId}, "/")
			} else {
				serviceTmpKey = apt.GenerateProviderDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = util.StringJoin([]string{"p", serviceTmpId, serviceId}, "/")
			}
			if _, ok := flag[serviceTmpKey]; ok {
				util.LOGGER.Debugf("serviceTmpKey is more exist.%s", serviceTmpKey)
				continue
			}
			flag[serviceTmpKey] = true
			util.LOGGER.Infof("delete dependency %s", deleteKey)
			opt = &registry.PluginOp{
				Action: registry.DELETE,
				Key:    util.StringToBytesWithNoCopy(serviceTmpKey),
			}
			opts = append(opts, opt)
		}
		opt = &registry.PluginOp{
			Action:     registry.DELETE,
			Key:        util.StringToBytesWithNoCopy(serviceKey),
			WithPrefix: true,
		}
		util.LOGGER.Infof("delete dependency serviceKey is %s", serviceType+"/"+serviceId)
		opts = append(opts, opt)
	}
	return opts, nil
}

func CreateDependencyRule(ctx context.Context, dep *Dependency) error {
	//更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	conKey := apt.GenerateConsumerDependencyRuleKey(dep.Tenant, dep.Consumer)

	err, oldProviderRules := transferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.LOGGER.Errorf(err, "maintain dependency rule failed, consumer %s: get consumer depedency rule failed.", consumerFlag)
		return err
	}

	unExistDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList := make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := containerServiceDependency(dep.ProvidersRule, oldProviderRule); !ok {
			unExistDependencyRuleList = append(unExistDependencyRuleList, oldProviderRule)
		} else {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containerServiceDependency(existDependencyRuleList, tmpProviderRule); !ok {
			newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		}
	}

	dep.err = make(chan error, 5)
	dep.chanNum = 0
	if len(unExistDependencyRuleList) != 0 {
		util.LOGGER.Infof("Unexist dependency rule remove for consumer %s, %v, ", consumerFlag, unExistDependencyRuleList)
		dep.removedDependencyRuleList = unExistDependencyRuleList
		dep.RemoveConsumerOfProviderRule()
	}

	if len(newDependencyRuleList) != 0 {
		util.LOGGER.Infof("New dependency rule add for consumer %s, %v, ", consumerFlag, newDependencyRuleList)
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

func containerServiceDependency(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) (bool, error) {
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

func deleteDependency(tenant string, consumerId string, providerId string) []*registry.PluginOp {
	proKey := apt.GenerateProviderDependencyKey(tenant, providerId, consumerId)
	conKey := apt.GenerateConsumerDependencyKey(tenant, consumerId, providerId)
	opts := []*registry.PluginOp{}
	optPro := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(proKey),
	}
	optCon := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(conKey),
	}
	opts = append(opts, optPro)
	opts = append(opts, optCon)
	return opts
}

func GetDependencies(ctx context.Context, dependencyKey string, tenant string) ([]*pb.MicroService, error) {
	util.LOGGER.Debugf("GetDependencies start.")
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(dependencyKey),
		WithPrefix: true,
	}
	data, err := store.Store().Dependency().Search(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Get dependency failed,%s", err.Error())
		return nil, err
	}
	if len(data.Kvs) == 0 {
		util.LOGGER.Errorf(nil, "data.Kvs len is 0.")
		return nil, nil
	}

	util.LOGGER.Debugf("data.Kvs[0].Value is %s.", data.Kvs[0].Value)

	key := ""
	keySplilt := []string{}
	providerId := ""
	microServices := []*pb.MicroService{}
	for _, kv := range data.Kvs {
		key = util.BytesToStringWithNoCopy(kv.Key)
		util.LOGGER.Debugf("key is %s", key)
		keySplilt = strings.Split(key, "/")
		providerId = keySplilt[len(keySplilt)-1]

		provider, err := ms.GetService(ctx, tenant, providerId)
		if err != nil {
			util.LOGGER.Errorf(nil, "Get service failed, %s", err.Error())
			return nil, err
		}
		if provider == nil {
			util.LOGGER.Errorf(nil, "Get service is empty, serviceId is %s", providerId)
			return nil, errors.New("Get service is empty.")
		}
		microServices = append(microServices, provider)
	}
	return microServices, nil
}

// fuzzyMatch: 是否使用模糊规则
func validateMicroServiceKey(in *pb.MicroServiceKey, fuzzyMatch bool) error {
	var err error
	if fuzzyMatch {
		// provider的ServiceName, Version支持模糊规则
		err = apt.ProviderMsValidator.Validate(in)
	} else {
		err = apt.DependencyMSValidator.Validate(in)
	}
	if err != nil {
		return err
	}
	if len(in.Stage) == 0 {
		in.Stage = "dev"
	}
	return nil
}

func BadParamsResponse(detailErr string) *pb.CreateDependenciesResponse {
	util.LOGGER.Errorf(nil, "Request params is Valid.")
	if len(detailErr) == 0 {
		detailErr = "Request params is Valid."
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(pb.Response_FAIL, detailErr),
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
			util.LOGGER.Debugf("%s 's provider contains *.", consumerInfo.ServiceName)
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

func AddServiceVersionRule(ctx context.Context, provider *pb.MicroServiceKey, tenant string, consumer *pb.MicroServiceKey) (error, bool) {
	if apt.VersionRegex.Match(util.StringToBytesWithNoCopy(provider.Version)) {
		return nil, false
	}

	providerKey := apt.GenerateProviderDependencyRuleKey(tenant, provider)
	err, consumers := transferToMicroServiceDependency(ctx, providerKey)
	if err != nil {
		return err, false
	}
	if len(consumers.Dependency) != 0 {
		isEqual, err := containerServiceDependency(consumers.Dependency, consumer)
		if err != nil {
			return err, false
		}
		if isEqual {
			//删除之前的依赖
			return nil, true
		}
	}
	//添加依赖
	consumers.Dependency = append(consumers.Dependency, consumer)
	data, err := json.Marshal(consumers)
	if err != nil {
		util.LOGGER.Errorf(err, "Marshal dependency of find failed.")
		return err, false
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(providerKey),
		Value:  data,
	}
	_, err = registry.GetRegisterCenter().Do(ctx, opt)
	return err, false
}

func UpdateServiceForAddDependency(ctx context.Context, consumerId string, providers []*pb.DependencyMircroService, tenant string) error {
	conServiceKey := apt.GenerateServiceKey(tenant, consumerId)
	service, err := ms.GetService(ctx, tenant, consumerId)
	if err != nil {
		util.LOGGER.Errorf(err, "create dependency faild: get service failed. consumerId %s", consumerId)
		return err
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "create dependency faild: service not exist.serviceId %s", consumerId)
		return errors.New("Get service is empty")
	}

	service.Providers = providers
	data, err := json.Marshal(service)
	if err != nil {
		util.LOGGER.Errorf(err, "create dependency faild: marshal service failed.")
		return err
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(conServiceKey),
		Value:  data,
	}
	_, err = registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(err, "create dependency faild: commit service data into etcd failed.")
		return err
	}
	return nil
}

type Dependency struct {
	ConsumerId                string
	Tenant                    string
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
	opts := make([]*registry.PluginOp, 0, len(dep.removedDependencyRuleList))
	for _, providerRule := range dep.removedDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(dep.Tenant, providerRule)
		util.LOGGER.Debugf("This proProkey is %s.", proProkey)
		err, consumerValue := transferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		for key, tmp := range consumerValue.Dependency {
			if ok := equalServiceDependency(tmp, dep.Consumer); ok {
				consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
				break
			}
		}
		//删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
		if len(consumerValue.Dependency) == 0 {
			opt := &registry.PluginOp{
				Action: registry.DELETE,
				Key:    []byte(proProkey),
			}
			opts = append(opts, opt)
			continue
		}
		data, err := json.Marshal(consumerValue)
		if err != nil {
			util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- err
			return
		}
		opt := &registry.PluginOp{
			Action: registry.PUT,
			Key:    util.StringToBytesWithNoCopy(proProkey),
			Value:  data,
		}
		opts = append(opts, opt)
	}
	if len(opts) != 0 {
		_, err := registry.GetRegisterCenter().Txn(ctx, opts)
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
	opts := []*registry.PluginOp{}
	for _, prividerRule := range dep.NewDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(dep.Tenant, prividerRule)
		err, tmpValue := transferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		tmpValue.Dependency = append(tmpValue.Dependency, dep.Consumer)

		data, errMarshal := json.Marshal(tmpValue)
		if errMarshal != nil {
			util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- errors.New("Marshal tmpValue failed.")
			return
		}
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    util.StringToBytesWithNoCopy(proProkey),
			Value:  data,
		})
		if prividerRule.ServiceName == "*" {
			break
		}
	}
	if len(opts) != 0 {
		_, err := registry.GetRegisterCenter().Txn(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) updateProvidersRuleOfConsumer(conKey string) error {
	dependency := &pb.MicroServiceDependency{
		Dependency: dep.ProvidersRule,
	}
	data, err := json.Marshal(dependency)
	if err != nil {
		util.LOGGER.Errorf(nil, "Marshal tmpValue fialed.")
		return err
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(conKey),
		Value:  data,
	}
	_, err = registry.GetRegisterCenter().Do(context.TODO(), opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Upload dependency rule failed.")
		return err
	}
	return nil
}

func (dep *Dependency) UpdateProvidersRuleOfConsumer(conKey string) error {
	return dep.updateProvidersRuleOfConsumer(conKey)
}

func (dep *Dependency) GetAddDependencyProviderIds() ([]string, error) {
	return dep.getDependencyProviderIds(dep.NewDependencyRuleList)
}

func (dep *Dependency) GetDeleteDependencyProviderIds() ([]string, error) {
	return dep.getDependencyProviderIds(dep.removedDependencyRuleList)
}

func (dep *Dependency) getDependencyProviderIds(providerRules []*pb.MicroServiceKey) ([]string, error) {
	tenant := dep.Tenant
	provideServiceIds := make([]string, 0)
	for _, provider := range providerRules {
		switch {
		case provider.ServiceName == "*":
			util.LOGGER.Infof("Add dependency, *: rely all service, consumerId %s", dep.ConsumerId)
			allServiceKey := apt.GenerateServiceKey(tenant, "")
			resp, err := store.Store().Service().Search(context.TODO(), &registry.PluginOp{
				Action:     registry.GET,
				Key:        util.StringToBytesWithNoCopy(allServiceKey),
				WithPrefix: true,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "Add dependency failed, rely all service: get all services failed.")
				return provideServiceIds, err
			}
			keyArr := []string{}
			providerId := ""
			for _, kvs := range resp.Kvs {
				keyArr = strings.Split(util.BytesToStringWithNoCopy(kvs.Key), "/")
				providerId = keyArr[len(keyArr)-1]
				provideServiceIds = append(provideServiceIds, providerId)
			}
			return provideServiceIds, nil
		default:
			serviceIds, err := ms.FindServiceIds(context.TODO(), provider.Version, &pb.MicroServiceKey{
				Tenant:      tenant,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "Get providerIds failed, service: %s/%s/%s",
					provider.AppId, provider.ServiceName, provider.Version)
				return provideServiceIds, err
			}
			if len(serviceIds) == 0 {
				util.LOGGER.Warnf(nil, "Get providerIds is empty, service: %s/%s/%s does not exist",
					provider.AppId, provider.ServiceName, provider.Version)
				continue
			}
			provideServiceIds = append(provideServiceIds, serviceIds...)
		}
	}
	return provideServiceIds, nil
}

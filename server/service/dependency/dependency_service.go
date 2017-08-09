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
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	VESION_RULE_REGEX = `^[0-9\.]+$`
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
	//	keys := microservice.GetService(context.TODO())
	//	for _, v := range keys {
	//		util.LOGGER.Debug(fmt.Sprintf("sync consumers for %s", v))
	//		domainAndId := strings.Split(v, ":::")
	//		// 查询所有consumer
	//		key := apt.GenerateProviderDependencyKey(domainAndId[0], domainAndId[1], "")
	//		resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
	//			Action:     registry.GET,
	//			Key:        []byte(key),
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
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		KeyOnly:    true,
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
	resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		KeyOnly:    true,
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

func UpdateAsConsumerDependency(ctx context.Context, consumerId string, providers []*pb.MicroServiceKey, tenant string) error {
	optPros := []*registry.PluginOp{}
	optProsTmps := []*registry.PluginOp{}
	for _, provider := range providers {
		switch {
		case provider.ServiceName == "*":
			util.LOGGER.Infof("Add dependency, *: rely all service, consumerId %s", consumerId)
			allServiceKey := apt.GenerateServiceKey(tenant, "")
			resp, err := store.Store().Service().Search(ctx, &registry.PluginOp{
				Action:     registry.GET,
				Key:        []byte(allServiceKey),
				WithPrefix: true,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "Add dependency failed, rely all service: get all services failed.")
				return err
			}
			keyArr := []string{}
			providerId := ""
			for _, kvs := range resp.Kvs {
				keyArr = strings.Split(string(kvs.Key), "/")
				providerId = keyArr[len(keyArr)-1]
				optProsTmps = putServiceDependency(consumerId, providerId, tenant)
				optPros = append(optPros, optProsTmps...)
			}
		default:
			serviceIds, err := ms.FindServiceIds(ctx, provider.Version, &pb.MicroServiceKey{
				Tenant:      tenant,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "Get providerIds failed, service: %s/%s/%s",
					provider.AppId, provider.ServiceName, provider.Version)
				return err
			}
			if len(serviceIds) == 0 {
				util.LOGGER.Errorf(nil, "Get providerIds failed, service: %s/%s/%s does not exist",
					provider.AppId, provider.ServiceName, provider.Version)
				continue
			}
			for _, serviceId := range serviceIds {
				optProsTmps = putServiceDependency(consumerId, serviceId, tenant)
				optPros = append(optPros, optProsTmps...)
			}
		}
	}
	err := registry.BatchCommit(ctx, optPros)
	if err != nil {
		util.LOGGER.Errorf(nil, "Put consumer's dependency failed.")
		return err
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
		Key:    []byte(conProKey),
		Value:  []byte(timestamp),
	}
	optCon := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(proProKey),
		Value:  []byte(timestamp),
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
	allConsumers := []*pb.MicroServiceKey{}
	relyAllKey := apt.GenerateProviderDependencyRuleKey(tenant, &pb.MicroServiceKey{
		ServiceName: "*",
	})
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(relyAllKey),
	}
	rsp, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(err, "get consumer that rely all service failed.")
		return err
	}
	if len(rsp.Kvs) == 0 {
		util.LOGGER.Debugf("There is no serviceName == * providers situation.")
	} else {
		util.LOGGER.Infof("consumer that rely all service exist.ServiceName: %s.", provider.ServiceName)
		microServiceDependency := &pb.MicroServiceDependency{}
		err = json.Unmarshal(rsp.Kvs[0].Value, microServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Unmarshal res failed.")
			return err
		}
		allConsumers = append(allConsumers, microServiceDependency.Dependency...)
	}
	//获取该服务的所有的版本，找出所有去掉版本的前缀的keyvalue
	tempProvider := *provider
	tempProvider.Version = ""
	providerVersion := provider.Version
	proKey := apt.GenerateProviderDependencyRuleKey(tenant, &tempProvider)
	util.LOGGER.Debugf("proKey is %s", proKey)
	opt = &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(proKey),
		WithPrefix: true,
	}
	rsp, err = registry.GetRegisterCenter().Do(ctx, opt)

	if err != nil {
		util.LOGGER.Errorf(err, "get all dependency rule failed: provider rule key %v.", tempProvider)
		return err
	}
	if len(rsp.Kvs) == 0 && len(allConsumers) == 0 {
		util.LOGGER.Infof("%s as a provider, no consumer use it.", provider.ServiceName)
		return nil
	} else {
		for _, kv := range rsp.Kvs {
			consumers := &pb.MicroServiceDependency{
				Dependency: []*pb.MicroServiceKey{},
			}
			providerVersionRuleArr := strings.Split(string(kv.Key), "/")
			providerVersionRule := providerVersionRuleArr[len(providerVersionRuleArr)-1]
			if providerVersionRule == "latest" {
				latestServiceId, err := ms.FindServiceIds(ctx, providerVersionRule, &pb.MicroServiceKey{
					Tenant:      tenant,
					AppId:       provider.AppId,
					ServiceName: provider.ServiceName,
				})
				if err != nil {
					util.LOGGER.Errorf(err, "Get latest service failed.")
					return err
				}
				if len(latestServiceId) == 0 {
					util.LOGGER.Infof("%s 's providerId is empty,no this service.", provider.ServiceName)
					continue
				}
				if providerServiseId != latestServiceId[0] {
					continue
				}
			} else {
				//当版本号一样，或者存在+（版本不确定）且版本小于当前版本，则将其consumer，加入allConsumers
				if !ms.VersionMatchRule(providerVersion, providerVersionRule) {
					continue
				}
			}

			util.LOGGER.Debugf("providerETCD is %s", providerVersionRuleArr)
			err = json.Unmarshal(kv.Value, consumers)
			if err != nil {
				util.LOGGER.Errorf(nil, "Unmarshal consumers failed.")
				return err
			}
			//fmt.Println("kv.Value is ", consumers)
			util.LOGGER.Infof("Add dependency as provider, provider: serviecName(%s), version(%s) .consumer is", provider.ServiceName, providerVersionRule, consumers.Dependency)
			allConsumers = append(allConsumers, consumers.Dependency...)
		}
	}
	if len(allConsumers) == 0 {
		util.LOGGER.Infof("%s as a provider, no consumer use it.", provider.ServiceName)
		return nil
	}
	//更新作为provider的依赖关系
	opts := []*registry.PluginOp{}
	//标记相同的serviceId是否被加入
	flag := map[string]bool{}
	for _, consumer := range allConsumers {
		consumerServiceid, err := ms.GetServiceId(ctx, consumer)
		if err != nil {
			util.LOGGER.Errorf(nil, "Get consumer's serviceId failed.")
			return err
		}
		if len(consumerServiceid) == 0 {
			util.LOGGER.Errorf(nil, "Get consumer's serviceId is empty, skip serviceName: %s, appId: %s, version: %s", consumer.ServiceName, consumer.AppId, consumer.Version)
			continue
		}
		if _, ok := flag[consumerServiceid]; ok {
			util.LOGGER.Errorf(nil, "consumerServiceid %s more exists.", consumerServiceid)
			continue
		} else {
			flag[consumerServiceid] = true
		}
		optsTmp := putServiceDependency(consumerServiceid, providerServiseId, tenant)
		opts = append(opts, optsTmp...)
	}
	if len(opts) == 0 {
		util.LOGGER.Errorf(nil, "%s as a provider, no consumer use it.", provider.ServiceName)
		return nil
	}
	err = registry.BatchCommit(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "Add provider dependency rule failed: provider %v", provider)
		return err
	}
	return nil
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
			Key:    []byte(conKey),
		}
		util.LOGGER.Debugf("conKey is %s.", conKey)
		opts = append(opts, opt)
	}
	//作为provider的依赖规则
	providerKey := apt.GenerateProviderDependencyRuleKey(tenant, consumer)
	opt := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(providerKey),
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
		Key:    []byte(key),
	}
	res, err := registry.GetRegisterCenter().Do(ctx, opt)
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
		util.LOGGER.Errorf(nil, "Can not get mircroServiceDependency")
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
			Key:    []byte(serviceKey),
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
			Key:    []byte(serviceKey),
			Value:  []byte(data),
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
	return strings.Join([]string{
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
		Key:        []byte(serviceKey),
		WithPrefix: true,
	}
	rsp, err := registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		return nil, err
	}
	opts := []*registry.PluginOp{}
	if rsp != nil {
		serviceTmpId := ""
		serviceTmpKey := ""
		deleteKey := ""
		for _, kv := range rsp.Kvs {
			tmpKeyArr := strings.Split(string(kv.Key), "/")
			serviceTmpId = tmpKeyArr[len(tmpKeyArr)-1]
			if serviceType == "p" {
				serviceTmpKey = apt.GenerateConsumerDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = strings.Join([]string{"c", serviceTmpId, serviceId}, "/")
			} else {
				serviceTmpKey = apt.GenerateProviderDependencyKey(tenant, serviceTmpId, serviceId)
				deleteKey = strings.Join([]string{"p", serviceTmpId, serviceId}, "/")
			}
			if _, ok := flag[serviceTmpKey]; ok {
				util.LOGGER.Debugf("serviceTmpKey is more exist.%s", serviceTmpKey)
				continue
			}
			flag[serviceTmpKey] = true
			util.LOGGER.Infof("delete dependency %s", deleteKey)
			opt = &registry.PluginOp{
				Action: registry.DELETE,
				Key:    []byte(serviceTmpKey),
			}
			opts = append(opts, opt)
		}
		opt = &registry.PluginOp{
			Action:     registry.DELETE,
			Key:        []byte(serviceKey),
			WithPrefix: true,
		}
		util.LOGGER.Infof("delete dependency serviceKey is %s", serviceType+"/"+serviceId)
		opts = append(opts, opt)
	}
	return opts, nil
}

func CreateDependencyRule(ctx context.Context, consumerServiceid string, consumer *pb.MicroServiceKey, providers []*pb.MicroServiceKey) error {
	tenant := util.ParseTenantProject(ctx)
	//更新consumer的providers的值,consumer的版本是确定的
	conKey := apt.GenerateConsumerDependencyRuleKey(tenant, consumer)
	consumerFlag := strings.Join([]string{consumer.AppId, consumer.ServiceName, consumer.Version}, "/")
	err, oldProviderRules := transferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.LOGGER.Errorf(err, "maintain dependency rule failed, consumer %s: get consumer depedency rule failed.", consumerFlag)
		return err
	}

	//更新consumer之前的provider，存在以前依赖一个服务，后面不依赖了，需要删除
	if len(oldProviderRules.Dependency) != 0 {
		proProkey := ""
		consumerValue := &pb.MicroServiceDependency{
			Dependency: []*pb.MicroServiceKey{},
		}
		//provider里是否存在serviceName == *的情况，存在,则更新关系后，直接退出循环
		flag := false
		for _, oldProviderRule := range oldProviderRules.Dependency {
			util.LOGGER.Debugf("This provider's serviceName is %s.", oldProviderRule.ServiceName)
			//以前的provider是否在当前providers里，如果没有，就删除provider对应的依赖关系
			if ok, _ := containerServiceDependency(providers, oldProviderRule); ok {
				continue
			}
			oldProviderRuleFlag := strings.Join([]string{oldProviderRule.AppId, oldProviderRule.ServiceName, oldProviderRule.Version}, "/")
			util.LOGGER.Infof("old dependency rule %s not exist, delete", oldProviderRuleFlag)
			proProkey = apt.GenerateProviderDependencyRuleKey(tenant, oldProviderRule)
			util.LOGGER.Debugf("This proProkey is %s.", proProkey)
			err, consumerValue = transferToMicroServiceDependency(ctx, proProkey)
			if err != nil {
				return err
			}
			for key, tmp := range consumerValue.Dependency {
				if ok := equalServiceDependency(tmp, consumer); ok {
					util.LOGGER.Debugf("delete cunsumer %s.", consumer.ServiceName)
					consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
					versionRule := oldProviderRule.Version
					serviceName := oldProviderRule.ServiceName
					appId := oldProviderRule.AppId
					opts := []*registry.PluginOp{}
					//如果这个provider的版本是确定的，则删除该provider和consumer的依赖关系
					switch {
					case serviceName == "*":
						opts, err = deleteDependencyUtil(ctx, "c", tenant, consumerServiceid, map[string]bool{})
						if err != nil {
							util.LOGGER.Errorf(nil, "delete dependency failed.%s", err.Error())
							return err
						}
						flag = true
					default:
						serviceIds, err := ms.FindServiceIds(ctx, versionRule, &pb.MicroServiceKey{
							Tenant:      tenant,
							AppId:       appId,
							ServiceName: serviceName,
						})
						if err != nil {
							util.LOGGER.Errorf(err, "Get providerIds failed.")
							return err
						}
						if len(serviceIds) == 0 {
							util.LOGGER.Errorf(nil, "Get providerIds is empty.")
						} else {
							for _, serviceId := range serviceIds {
								optsTmp := deleteDependency(tenant, consumerServiceid, serviceId)
								opts = append(opts, optsTmp...)
							}
						}
					}
					if len(opts) != 0 {
						err = registry.BatchCommit(ctx, opts)
						if err != nil {
							util.LOGGER.Errorf(nil, err.Error())
							return err
						}
					}
					//删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
					if len(consumerValue.Dependency) == 0 {
						opt := &registry.PluginOp{
							Action: registry.DELETE,
							Key:    []byte(proProkey),
						}
						_, err = registry.GetRegisterCenter().Do(ctx, opt)
						if err != nil {
							util.LOGGER.Errorf(nil, err.Error())
							return err
						}
						break
					}
					data, err := json.Marshal(consumerValue)
					if err != nil {
						util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
						return err
					}
					opt := &registry.PluginOp{
						Action: registry.PUT,
						Key:    []byte(proProkey),
						Value:  []byte(data),
					}
					_, err = registry.GetRegisterCenter().Do(ctx, opt)
					if err != nil {
						util.LOGGER.Errorf(nil, err.Error())
						return err
					}
					break
				}
			}
			//如果provider里存在serviceName == *的情况，则直接break
			if flag {
				break
			}
		}
	}

	//更新该consumer的providers
	oldProviderRules.Dependency = providers
	data, err := json.Marshal(oldProviderRules)
	if err != nil {
		util.LOGGER.Errorf(nil, "Marshal tmpValue fialed.")
		return err
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(conKey),
		Value:  []byte(data),
	}
	_, err = registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(nil, "Upload dependency rule failed.")
		return err
	}

	util.LOGGER.Infof("add consumer's dependency rule success: %s", consumerFlag)
	//更新providers的consumer,consumer的版本是确定的，provider的版本可能不确定，只是一个版本规则，如1.0.0+

	err = addDependencyRuleOfProvider(ctx, consumer, providers, tenant)
	if err != nil {
		util.LOGGER.Errorf(err, "Add provider's dependency rule failed.")
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
		Key:    []byte(proKey),
	}
	optCon := &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(conKey),
	}
	opts = append(opts, optPro)
	opts = append(opts, optCon)
	return opts
}

func addDependencyRuleOfProvider(ctx context.Context, consumer *pb.MicroServiceKey, providers []*pb.MicroServiceKey, tenant string) error {
	opts := []*registry.PluginOp{}
	//查看provider的consumer是否包含该consumer，不包含则加入
	proProkey := ""
	for _, provider := range providers {
		proProkey = apt.GenerateProviderDependencyRuleKey(tenant, provider)
		err, tmpValue := transferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			return err
		}
		serviceDependency := tmpValue.Dependency
		ok, errCont := containerServiceDependency(serviceDependency, consumer)
		if errCont != nil {
			util.LOGGER.Errorf(nil, errCont.Error())
			return errCont
		}
		if !ok {
			tmpValue.Dependency = append(tmpValue.Dependency, consumer)
			//fmt.Println("serviceDependency is ", serviceDependency)
			//fmt.Println("tmpValue is ", tmpValue)
			data, errMarshal := json.Marshal(tmpValue)
			if errMarshal != nil {
				util.LOGGER.Errorf(nil, "Marshal tmpValue failed.")
				return errors.New("Marshal tmpValue failed.")
			}
			opt := &registry.PluginOp{
				Action: registry.PUT,
				Key:    []byte(proProkey),
				Value:  data,
			}
			opts = append(opts, opt)
			util.LOGGER.Infof("update provider's dependency rule, provider %v, consumer %v", provider, consumer)
		} else {
			util.LOGGER.Infof("createDependencyRule This consumer more exist in provlider's consumer list.")
		}
		if provider.ServiceName == "*" {
			break
		}
	}
	if len(opts) != 0 {
		_, err := registry.GetRegisterCenter().Txn(ctx, opts)
		if err != nil {
			util.LOGGER.Errorf(err, "update providers' dependency rule into etcd failed.consumer %v, providers %v", consumer, providers)
			return err
		}
	}
	return nil
}

func GetDependencies(ctx context.Context, dependencyKey string, tenant string) ([]*pb.MicroService, error) {
	util.LOGGER.Debugf("GetDependencies start.")
	opt := &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(dependencyKey),
		WithPrefix: true,
	}
	data, err := registry.GetRegisterCenter().Do(ctx, opt)
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
		key = string(kv.Key)
		util.LOGGER.Debugf("key is %s", key)
		keySplilt = strings.Split(key, "/")
		providerId = keySplilt[len(keySplilt)-1]

		provider, err := ms.GetServiceByServiceId(ctx, tenant, providerId)
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

func ParamsChecker(ctx context.Context, consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey, tenant string) *pb.CreateDependenciesResponse {
	if err := validateMicroServiceKey(consumerInfo, false); err != nil {
		return BadParamsResponse(err.Error())
	}
	if providersInfo == nil {
		return BadParamsResponse("Invalid request body for provider info.")
	}
	flag := map[string]bool{}
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
		if _, ok := flag[toString(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider.")
		} else {
			flag[toString(providerInfo)] = true
		}
	}
	return nil
}

func AddServiceVersionRule(ctx context.Context, provider *pb.MicroServiceKey, tenant string, consumer *pb.MicroServiceKey) (error, bool) {
	reg := regexp.MustCompile(VESION_RULE_REGEX)
	if reg.Match([]byte(provider.Version)) {
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
		Key:    []byte(providerKey),
		Value:  []byte(data),
	}
	_, err = registry.GetRegisterCenter().Do(ctx, opt)
	return err, false
}

func UpdateServiceForAddDependency(ctx context.Context, consumerId string, providers []*pb.DependencyMircroService, tenant string) error {
	conServiceKey := apt.GenerateServiceKey(tenant, consumerId)
	service, err := ms.GetServiceByServiceId(ctx, tenant, consumerId)
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
		Key:    []byte(conServiceKey),
		Value:  data,
	}
	_, err = registry.GetRegisterCenter().Do(ctx, opt)
	if err != nil {
		util.LOGGER.Errorf(err, "create dependency faild: commit service data into etcd failed.")
		return err
	}
	return nil
}

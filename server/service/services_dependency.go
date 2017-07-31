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
package service

import (
	"encoding/json"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/errors"
	"golang.org/x/net/context"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_TXN_NUMBER_ONE_TIME = 128
	VESION_RULE_REGEX       = `^[0-9\.]+$`
)

func (s *ServiceController) CreateDependenciesForMircServices(ctx context.Context, in *pb.CreateDependenciesRequest) (*pb.CreateDependenciesResponse, error) {
	dependencyInfos := in.Dependencies
	if dependencyInfos == nil {
		return BadParamsResponse("Invalid request body."), nil
	}
	tenant := util.ParaseTenantProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerInfo := transferToMicroServiceKeys([]*pb.DependencyMircroService{dependencyInfo.Consumer}, tenant)[0]
		providersInfo := transferToMicroServiceKeys(dependencyInfo.Providers, tenant)
		consumerFlag := strings.Join([]string{consumerInfo.AppId, consumerInfo.ServiceName, consumerInfo.Version}, "--")
		rsp := s.paramsChecker(ctx, consumerInfo, providersInfo, tenant)
		if rsp != nil {
			util.LOGGER.Errorf(nil, "create dependency faild, conusmer %s: invalid params.%s", consumerFlag, rsp.Response.Message)
			return rsp, nil
		}

		consumerId, err := ms.GetServiceId(ctx, consumerInfo, false)
		util.LOGGER.Debugf("consumerId is %s", consumerId)
		if err != nil {
			util.LOGGER.Errorf(err, "create dependency faild, conusmer %s: get consumer failed.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if len(consumerId) == 0 {
			util.LOGGER.Errorf(nil, "create dependency faild, conusmer %s: consumer not exist.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get consumer's serviceId is empty."),
			}, nil
		}
		//更新服务的内容，把providers加入
		err = s.updateServiceForAddDependency(ctx, consumerId, dependencyInfo.Providers, tenant)
		if err != nil {
			util.LOGGER.Errorf(err, "create dependency faild, conusmer %s: Update service failed.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}

		//建立依赖规则，用于维护依赖关系
		lock, err := mux.Lock(mux.SERVICE_LOCK)
		if err != nil {
			util.LOGGER.Errorf(err, "create dependency faild, conusmer %s: create lock failed.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}

		err = s.createDependencyRule(ctx, consumerId, consumerInfo, providersInfo)
		lock.Unlock()

		if err != nil {
			util.LOGGER.Errorf(err, "create dependency rule failed: consumer %s", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		//更新consumer作为provider的依赖关系
		err = s.updateAsProviderDependency(ctx, consumerId, consumerInfo)
		if err != nil {
			util.LOGGER.Errorf(nil, "Dependency update,as provider,update it's consumer list failed. %s", err.Error())
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		//更新consumer作为consumer的依赖关系
		err = s.updateAsConsumerDependency(ctx, consumerId, providersInfo, tenant)
		if err != nil {
			util.LOGGER.Errorf(nil, "Dependency update,as consumer,update it's provider list failed. %s", err.Error())
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		util.LOGGER.Infof("Create dependency success: consumer %s, %s  from remote %s", consumerFlag, consumerId, util.GetIPFromContext(ctx))
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "create dependency success"),
	}, nil
}

func (s *ServiceController) updateServiceForAddDependency(ctx context.Context, consumerId string, providers []*pb.DependencyMircroService, tenant string) error {
	conServiceKey := apt.GenerateServiceKey(tenant, consumerId)
	service, err := getServiceByServiceId(ctx, tenant, consumerId)
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

func transferToMicroServiceKeys(in []*pb.DependencyMircroService, tenant string) []*pb.MicroServiceKey {
	rst := []*pb.MicroServiceKey{}
	for _, value := range in {
		rst = append(rst, &pb.MicroServiceKey{
			Tenant:      tenant,
			AppId:       value.AppId,
			ServiceName: value.ServiceName,
			Version:     value.Version,
		})
	}
	return rst
}

func getServiceByServiceId(ctx context.Context, tenant string, serviceId string) (*pb.MicroService, error) {
	return ms.GetService(tenant, serviceId, 0)
}

func (s *ServiceController) createDependencyRule(ctx context.Context, consumerServiceid string, consumer *pb.MicroServiceKey, providers []*pb.MicroServiceKey) error {
	tenant := util.ParaseTenantProject(ctx)
	//更新consumer的providers的值,consumer的版本是确定的
	conKey := apt.GenerateConsumerDependencyRuleKey(tenant, consumer)
	consumerFlag := strings.Join([]string{consumer.AppId, consumer.ServiceName, consumer.Version}, "--")
	err, oldProviderRules := transferToMircroServiceDependency(ctx, conKey)
	if err != nil {
		util.LOGGER.Errorf(err, "maintain dependency rule failed, consumer %s: get consumer depedency rule failed.", consumerFlag)
		return err
	}

	//更新consumer之前的provider，存在以前依赖一个服务，后面不依赖了，需要删除
	if len(oldProviderRules.Dependency) != 0 {
		proProkey := ""
		consumerValue := &MircroServiceDependency{
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
			oldProviderRuleFlag := strings.Join([]string{oldProviderRule.AppId, oldProviderRule.ServiceName, oldProviderRule.Version}, "--")
			util.LOGGER.Infof("old dependency rule %s not exist, delete", oldProviderRuleFlag)
			proProkey = apt.GenerateProviderDependencyRuleKey(tenant, oldProviderRule)
			util.LOGGER.Debugf("This proProkey is %s.", proProkey)
			err, consumerValue = transferToMircroServiceDependency(ctx, proProkey)
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
						opts, err = s.deleteDependencyUtil(ctx, "c", tenant, consumerServiceid, map[string]bool{})
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
						err = batchCommit(ctx, opts)
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

	err = s.addDependencyRuleOfProvider(ctx, consumer, providers, tenant)
	if err != nil {
		util.LOGGER.Errorf(err, "Add provider's dependency rule failed.")
	}

	return nil
}

func (s *ServiceController) updateAsProviderDependency(ctx context.Context, providerServiseId string, provider *pb.MicroServiceKey) error {
	//查询etcd里是否存在带*的情况，则添加与对应的consumer与该provider的依赖关系
	tenant := util.ParaseTenantProject(ctx)
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
		mircroServiceDependency := &MircroServiceDependency{}
		err = json.Unmarshal(rsp.Kvs[0].Value, mircroServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Unmarshal res failed.")
			return err
		}
		allConsumers = append(allConsumers, mircroServiceDependency.Dependency...)
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
			consumers := &MircroServiceDependency{
				Dependency: []*pb.MicroServiceKey{},
			}
			providerETCD := strings.Split(string(kv.Key), "/")
			providerVersionETCD := providerETCD[len(providerETCD)-1]
			if providerVersionETCD == "latest" {
				latestServiceId, err := ms.FindServiceIds(ctx, providerVersionETCD, &pb.MicroServiceKey{
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
				if !compareVersion(providerVersion, providerVersionETCD) {
					continue
				}
			}

			util.LOGGER.Debugf("providerETCD is %s", providerETCD)
			err = json.Unmarshal(kv.Value, consumers)
			if err != nil {
				util.LOGGER.Errorf(nil, "Unmarshal consumers failed.")
				return err
			}
			//fmt.Println("kv.Value is ", consumers)
			util.LOGGER.Infof("Add dependency as provider, provider: serviecName(%s), version(%s) .consumer is", provider.ServiceName, providerVersionETCD, consumers.Dependency)
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
		consumerServiceid, err := ms.GetServiceId(ctx, consumer, false)
		if err != nil {
			util.LOGGER.Errorf(nil, "Get consumer's serviceId failed.")
			return err
		}
		if len(consumerServiceid) == 0 {
			util.LOGGER.Errorf(nil, "Get consumer's serviceId is empty.serviceName: %s, appId: %s, version: %s", consumer.ServiceName, consumer.AppId, consumer.Version)
			return errors.New("Get consumer's serviceId is empty.")
		}
		if _, ok := flag[consumerServiceid]; ok {
			util.LOGGER.Errorf(nil, "consumerServiceid %s more exists.", consumerServiceid)
			continue
		} else {
			flag[consumerServiceid] = true
		}
		optsTmp := s.putServiceDependency(consumerServiceid, providerServiseId, tenant)
		opts = append(opts, optsTmp...)
	}
	if len(opts) == 0 {
		util.LOGGER.Errorf(nil, "%s as a provider, no consumer use it.", provider.ServiceName)
		return nil
	}
	err = batchCommit(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "Add provider dependency rule failed: provider %v", provider)
		return err
	}
	return nil
}

func (s *ServiceController) putServiceDependency(consumerId string, providerId string, tenant string) []*registry.PluginOp {
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

func (s *ServiceController) updateAsConsumerDependency(ctx context.Context, consumerId string, providers []*pb.MicroServiceKey, tenant string) error {
	optPros := []*registry.PluginOp{}
	optProsTmps := []*registry.PluginOp{}
	for _, provider := range providers {
		switch {
		case provider.ServiceName == "*":
			util.LOGGER.Infof("Add dependency, *: rely all service, consumerId %s", consumerId)
			allServiceKey := strings.Join([]string{
				apt.GetServiceRootKey(tenant),
				"",
			}, "/")
			resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
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
				optProsTmps = s.putServiceDependency(consumerId, providerId, tenant)
				optPros = append(optPros, optProsTmps...)
			}
		default:
			serviceIds, err := ms.FindServiceIds(ctx, provider.Version, &pb.MicroServiceKey{
				Tenant:      tenant,
				AppId:       provider.AppId,
				ServiceName: provider.ServiceName,
			})
			if err != nil {
				util.LOGGER.Errorf(nil, "Get providerIds failed.")
				return err
			}
			if len(serviceIds) == 0 {
				util.LOGGER.Errorf(nil, "Get providerIds is empty.")
				continue
			}
			for _, serviceId := range serviceIds {
				optProsTmps = s.putServiceDependency(consumerId, serviceId, tenant)
				optPros = append(optPros, optProsTmps...)
			}
		}
	}
	err := batchCommit(ctx, optPros)
	if err != nil {
		util.LOGGER.Errorf(nil, "Put consumer's dependency failed.")
		return err
	}
	return nil
}

func (s *ServiceController) paramsChecker(ctx context.Context, consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey, tenant string) *pb.CreateDependenciesResponse {
	if err := vilidateMicroServiceKey(consumerInfo, false); err != nil {
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
		if err := vilidateMicroServiceKey(providerInfo, true); err != nil {
			return BadParamsResponse(err.Error())
		}
		if _, ok := flag[tostring(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider.")
		} else {
			flag[tostring(providerInfo)] = true
		}
	}
	return nil
}

func tostring(in *pb.MicroServiceKey) string {
	return strings.Join([]string{
		in.Tenant,
		in.AppId,
		in.Stage,
		in.ServiceName,
		in.Version,
	}, "")
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

type MircroServiceDependency struct {
	Dependency []*pb.MicroServiceKey
}

// fuzzyMatch: 是否使用模糊规则
func vilidateMicroServiceKey(in *pb.MicroServiceKey, fuzzyMatch bool) error {
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

func (s *ServiceController) addDependencyRuleOfProvider(ctx context.Context, consumer *pb.MicroServiceKey, providers []*pb.MicroServiceKey, tenant string) error {
	opts := []*registry.PluginOp{}
	//查看provider的consumer是否包含该consumer，不包含则加入
	proProkey := ""
	for _, provider := range providers {
		proProkey = apt.GenerateProviderDependencyRuleKey(tenant, provider)
		err, tmpValue := transferToMircroServiceDependency(ctx, proProkey)
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

func transferToMircroServiceDependency(ctx context.Context, key string) (error, *MircroServiceDependency) {
	mircroServiceDependency := &MircroServiceDependency{
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
		err = json.Unmarshal(res.Kvs[0].Value, mircroServiceDependency)
		if err != nil {
			util.LOGGER.Errorf(nil, "Unmarshal res failed.")
			return err, nil
		}
	} else {
		util.LOGGER.Errorf(nil, "Can not get mircroServiceDependency")
	}
	return nil, mircroServiceDependency
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

func compareVersion(version1 string, version2 string) bool {
	version1Array := strings.Split(version1, ".")
	version2Array := version1Array
	util.LOGGER.Debugf("version1 is %s", version1)
	util.LOGGER.Debugf("version2 is %s", version2)
	flagPlus := false
	if version2[len(version2)-1:] == "+" {
		flagPlus = true
		tmpVersion2 := version2[:len(version2)-1]
		version2Array = strings.Split(tmpVersion2, ".")

	} else {
		version2Array = strings.Split(version2, ".")
	}
	util.LOGGER.Debugf("version1Array is %s", version1Array)
	util.LOGGER.Debugf("version2Array is %s", version2Array)
	for key, value := range version1Array {
		if value == version2Array[key] {
			continue
		} else if flagPlus && value > version2Array[key] {
			return true
		} else {
			return false
		}
	}
	return true
}

func equalServiceDependency(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := tostring(serviceA)
	stringB := tostring(serviceB)
	if stringA == stringB {
		return true
	}
	return false
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

func GetDependencies(ctx context.Context, dependencyKey string, tenant string) ([]*pb.MicroService, error) {
	util.LOGGER.Errorf(nil, "GetDependencies start.")
	opt := &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(dependencyKey),

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
	} else {
		util.LOGGER.Debugf("data.Kvs[0].Value is %s.", data.Kvs[0].Value)
	}
	key := ""
	keySplilt := []string{}
	providerId := ""
	microServices := []*pb.MicroService{}
	for _, kv := range data.Kvs {
		key = string(kv.Key)
		util.LOGGER.Errorf(nil, "key is %s", key)
		keySplilt = strings.Split(key, "/")
		providerId = keySplilt[len(keySplilt)-1]

		provider, err := getServiceByServiceId(ctx, tenant, providerId)
		if err != nil {
			util.LOGGER.Errorf(nil, "Get service failed,%s", err.Error())
			return nil, err
		}
		if provider == nil {
			util.LOGGER.Errorf(nil, "Get service is empty,serviceId is %s", providerId)
			return nil, errors.New("Get service is empty.")
		}
		microServices = append(microServices, provider)
	}
	return microServices, nil
}

func (s *ServiceController) GetProviderDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "GetProConDependencies failed for validating parameters failed.")
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	providerId := in.ServiceId
	tenant := util.ParaseTenantProject(ctx)
	if !ms.ServiceExist(ctx, tenant, providerId) {
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This provider does not exist."),
		}, nil
	}
	keyProDependency := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	services, err := GetDependencies(ctx, keyProDependency, tenant)
	if err != nil {
		return &pb.GetProDependenciesResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, err.Error()),
			Consumers: nil,
		}, err
	}
	return &pb.GetProDependenciesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get all consumers successful."),
		Consumers: services,
	}, nil
}

func (s *ServiceController) GetConsumerDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "GetConsumerDependencies failed for validating parameters failed.")
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	consumerId := in.ServiceId
	tenant := util.ParaseTenantProject(ctx)
	if !ms.ServiceExist(ctx, tenant, consumerId) {
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This consumer does not exist."),
		}, nil
	}
	keyConDependency := apt.GenerateConsumerDependencyKey(tenant, consumerId, "")
	services, err := GetDependencies(ctx, keyConDependency, tenant)
	if err != nil {
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	return &pb.GetConDependenciesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get all providers successful."),
		Providers: services,
	}, nil
}

func batchCommit(ctx context.Context, opts []*registry.PluginOp) error {
	lenOpts := len(opts)
	tmpLen := lenOpts
	tmpOpts := []*registry.PluginOp{}
	var err error
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MAX_TXN_NUMBER_ONE_TIME
		if tmpLen > 0 {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : (i+1)*MAX_TXN_NUMBER_ONE_TIME]
		} else {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : lenOpts]
		}
		_, err = registry.GetRegisterCenter().Txn(ctx, tmpOpts)
		if err != nil {
			return err
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
	err, consumers := transferToMircroServiceDependency(ctx, providerKey)
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

func toMicroServiceKey(tenant string, in *pb.MicroService) *pb.MicroServiceKey {
	return &pb.MicroServiceKey{
		Tenant:      tenant,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Version:     in.Version,
	}
}

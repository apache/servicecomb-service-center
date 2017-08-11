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
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/plugins/dynamic"
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
	"strconv"
	"strings"
	"time"
)

type ServiceController struct {
}

func (s *ServiceController) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	if in == nil || in.Service == nil {
		util.LOGGER.Errorf(nil, "create microservice failed: param empty.operator:%s", remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	service := in.Service
	err := apt.Validate(service)

	serviceFlag := strings.Join([]string{service.AppId, service.ServiceName, service.Version}, "/")
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: invalid parameters.operator:%s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	consumer := &pb.MicroServiceKey{
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
		Tenant:      tenant,
	}

	lockMutex := mux.MuxType(apt.GenerateServiceIndexKey(consumer))
	lock, err := mux.Lock(lockMutex)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s:internal err,create lock failed.operator:%s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Beginning transaction failed."),
		}, err
	}

	serviceId := in.Service.ServiceId
	serviceIdInner, err := ms.GetServiceId(ctx, consumer)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s:internal err,query service failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query service key failed."),
		}, err
	}
	if len(serviceIdInner) > 0 {
		util.LOGGER.Errorf(nil, "create microservice failed, %s:service already exist.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, "Register service already exists."),
			ServiceId: serviceId,
		}, nil
	}

	ok, err := quota.QuotaPlugins[beego.AppConfig.DefaultString("quota_plugin", "buildin")]().Apply4Quotas(ctx, quota.MicroServiceQuotaType, 0)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: check apply quota.operator:%s", serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if !ok {
		util.LOGGER.Errorf(err, "create microservice failed, %s: no quota to apply.operator:%s", serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No quota to create service."),
		}, nil
	}

	// 产生全局service id
	if len(serviceId) == 0 {
		serviceId = dynamic.GetServiceId()
	}
	service.ServiceId = serviceId
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(service)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: json marshal service failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Body error "+err.Error()),
		}, nil
	}
	key := apt.GenerateServiceKey(tenant, serviceId)
	index := apt.GenerateServiceIndexKey(consumer)
	util.LOGGER.Debugf("start register service: %s %v", key, service)
	util.LOGGER.Debugf("start register service index: %s %v", index, serviceId)
	opts := []*registry.PluginOp{
		// Set key file
		{
			Action: registry.PUT,
			Key:    []byte(key),
			Value:  data,
		},
		{
			Action: registry.PUT,
			Key:    []byte(index),
			Value:  []byte(serviceId),
		},
	}

	if len(consumer.Alias) > 0 {
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(apt.GenerateServiceAliasKey(consumer)),
			Value:  []byte(serviceId),
		})
	}

	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed, %s: commit data into etcd failed.operator:%s",
			serviceFlag, remoteIP)
		lock.Unlock()
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}
	lock.Unlock()
	//创建服务间的依赖
	util.LOGGER.Infof("create microservice: add dependency for %s.operator:%s", serviceFlag, remoteIP)
	err = dependency.UpdateAsProviderDependency(ctx, serviceId, consumer)
	if err != nil {
		util.LOGGER.Errorf(err, "create microservice failed: dependency update,as provider failed.%s.operator:%s", serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, err.Error()),
			ServiceId: "",
		}, err
	}

	util.LOGGER.Infof("create microservice successful, %s, serviceId: %s. from remote %s",
		serviceFlag, service.ServiceId, util.GetIPFromContext(ctx))
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Register service successfully."),
		ServiceId: serviceId,
	}, nil
}

func (s *ServiceController) DeleteServicePri(ctx context.Context, ServiceId string, force bool) (*pb.Response, error) {
	tenant := util.ParseTenantProject(ctx)

	service, err := ms.GetServiceByServiceId(ctx, tenant, ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: get service failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}

	if service == nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: service not exist.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, "Service does not exist."), nil
	}

	util.LOGGER.Infof("start delete service %s", ServiceId)

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		keyConDependency := apt.GenerateProviderDependencyKey(tenant, ServiceId, "")
		services, err := dependency.GetDependencies(ctx, keyConDependency, tenant)
		if err != nil {
			util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err, get service dependency failed.", ServiceId)
			return  pb.CreateResponse(pb.Response_FAIL, "Get dependency info failed."), err
		}
		if len(services) > 1 || (len(services) == 1 && services[0].ServiceId != ServiceId) {
			util.LOGGER.Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, other services rely it.", ServiceId)
			return  pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(tenant, ServiceId, "")
		rsp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
			Action:     registry.GET,
			Key:        []byte(instancesKey),
			WithPrefix: true,
			CountOnly:  true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err,get instances failed.", ServiceId)
			return  pb.CreateResponse(pb.Response_FAIL, "Get instance failed."), err
		}

		if rsp.Count > 0 {
			util.LOGGER.Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, exist instance.", ServiceId)
			return  pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, exist instance."), err
		}
	}

	consumer := &pb.MicroServiceKey{
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
		Alias:       service.Alias,
		Tenant:      tenant,
	}

	//refresh msCache consumerCache, ensure that watch can notify consumers when no cache.
	err = dependency.RefreshDependencyCache(tenant, ServiceId, service)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: inner err, refresh service dependency cache failed.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, "Refresh dependency cache failed."), err
	}

	opts := []*registry.PluginOp{
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceIndexKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceAliasKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    []byte(apt.GenerateServiceKey(tenant, ServiceId)),
		},
		{
			Action: registry.DELETE,
			Key: []byte(strings.Join([]string{
				apt.GetServiceRuleRootKey(tenant),
				ServiceId,
				"",
			}, "/")),
		},
	}

	//删除依赖规则
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: inner err, create lock failed.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	optsTmp, err := dependency.DeleteDependencyForService(ctx, consumer, ServiceId)
	lock.Unlock()
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: inner err, delete dependency failed.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	opts = append(opts, optsTmp...)

	//删除黑白名单
	rulekey := apt.GenerateServiceRuleKey(tenant, ServiceId, "")
	opt := &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        []byte(rulekey),
		WithPrefix: true,
	}
	opts = append(opts, opt)
	indexKey := apt.GenerateRuleIndexKey(tenant, ServiceId, "", "")
	opts = append(opts, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(indexKey),
	})
	opts = append(opts, opt)

	//删除shemas
	schemaKey := apt.GenerateServiceSchemaKey(tenant, ServiceId, "")
	opt = &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        []byte(schemaKey),
		WithPrefix: true,
	}
	opts = append(opts, opt)

	//删除tags
	tagsKey := apt.GenerateServiceTagKey(tenant, ServiceId)
	opt = &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(tagsKey),
	}
	opts = append(opts, opt)

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: delete all instances failed.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, "Delete all instances failed for service."), err
	}

	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: commit data into etcd failed.", ServiceId)
		return  pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."), nil
	}

	util.LOGGER.Infof("delete microservice successful: serviceid is %s,operator is %s.", ServiceId, util.GetIPFromContext(ctx))
	return pb.CreateResponse(pb.Response_SUCCESS, "Unregister service successfully."), nil
}

func (s *ServiceController) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.ServiceId == apt.Service.ServiceId {
		util.LOGGER.Errorf(nil, "delete microservice failed: service empty.")
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	resp, err := s.DeleteServicePri(ctx, in.ServiceId,in.Force)

	return &pb.DeleteServiceResponse{
		Response:resp,
	},err
}

func (s *ServiceController)DeleteServices(ctx context.Context, request *pb.DelServicesRequest)(*pb.DelServicesResponse, error){
	// 合法性检查
	if request == nil || request.ServiceIds == nil || len(request.ServiceIds) == 0 {
		return &pb.DelServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request param."),
			Services:nil,
		},nil
	}

	existFlag := map[string]bool{}
	nuoMultilCount := 0
	// 批量删除服务
	serviceRespChan := make(chan *pb.DelServicesRspInfo , len(request.ServiceIds))
	for _, serviceId := range request.ServiceIds{
		//ServiceId重复性检查
		if _ ,ok := existFlag[serviceId]; ok {
			util.LOGGER.Warnf(nil, "delete microservice %s , multiple.", serviceId)
			continue
		} else {
			existFlag[serviceId] = true
			nuoMultilCount++
		}

	        serviceRst := &pb.DelServicesRspInfo{
			ServiceId : serviceId,
			ErrMessage:"",
		}

		//检查服务ID合法性
		in := &pb.DeleteServiceRequest{
			ServiceId : serviceId,
			Force : request.Force,
		}
		err := apt.Validate(in)
		if err != nil {
			util.LOGGER.Errorf(err, "delete microservice failed, serviceId is %s: invalid parameters.", in.ServiceId)
			serviceRst.ErrMessage = err.Error()
			serviceRespChan <- serviceRst
			continue
		}

		//执行删除服务操作
		go func(serviceItem string){
			 resp , err := s.DeleteServicePri(ctx, serviceItem, request.Force)
			if err != nil {
				serviceRst.ErrMessage = err.Error()
			}else if resp.Code != pb.Response_SUCCESS{
				serviceRst.ErrMessage  = resp.Message
			}

			serviceRespChan <- serviceRst
		}(serviceId)
	}

	//获取批量删除服务的结果
	count := 0
	responseCode  := pb.Response_SUCCESS
        delServiceRspInfo := []*pb.DelServicesRspInfo{}
	for serviceRespItem := range serviceRespChan{
		count++
		if len(serviceRespItem.ErrMessage) != 0{
			responseCode = pb.Response_FAIL
		}
		delServiceRspInfo = append(delServiceRspInfo, serviceRespItem)
		//结果收集over，关闭通道
		if count == nuoMultilCount{
			close(serviceRespChan)
		}
	}

	util.LOGGER.Infof("Batch DeleteServices serviceId = %v , result = %d, ", request.ServiceIds, responseCode)
        return &pb.DelServicesResponse{
		Response: pb.CreateResponse(responseCode, "Delete services successfully."),
		Services: delServiceRspInfo,
	},nil
}

func (s *ServiceController) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "get microservice failed, serviceId is %s: invalid parameters.",
			in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)
	service, err := ms.GetServiceByServiceId(ctx, tenant, in.ServiceId)

	if err != nil {
		util.LOGGER.Errorf(err, "get microservice failed, serviceId is %s: inner err,get service failed.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service file failed."),
		}, err
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "get microservice failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service successfully."),
		Service:  service,
	}, nil
}

func (s *ServiceController) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	if in == nil {
		util.LOGGER.Errorf(nil, "get services failed: invalid params.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	services, err := ms.GetAllServiceUtil(ctx)
	if err != nil {
		util.LOGGER.Errorf(err, "get services failed: inner err.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get all service failed."),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get services successfully."),
		Services: services,
	}, nil
}

func (s *ServiceController) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.Properties == nil {
		util.LOGGER.Errorf(nil, "update service properties failed: invalid params.")
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	key := apt.GenerateServiceKey(tenant, in.ServiceId)
	service, err := ms.GetServiceByServiceId(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: query service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query service file failed."),
		}, err
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "update service properties failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	service.Properties = make(map[string]string)
	for propertyKey := range in.Properties {
		service.Properties[propertyKey] = in.Properties[propertyKey]
	}

	data, err := json.Marshal(service)
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: json marshal service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service file marshal error."),
		}, err
	}

	// Set key file
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "update service properties failed, serviceId is %s: commit data into etcd failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("update service properties successful: serviceId is %s.", in.ServiceId)
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Update service successfully."),
	}, nil
}

func (s *ServiceController) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetRules()) == 0 {
		util.LOGGER.Errorf(nil, "add rule failed: invalid parameters.")
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "add rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	opts := []*registry.PluginOp{}
	ruleType, _, err := serviceUtil.GetServiceRuleType(ctx, tenant, in.ServiceId)
	util.LOGGER.Debugf("ruleType is %s", ruleType)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	ruleIds := []string{}
	for _, rule := range in.Rules {
		err := apt.Validate(rule)
		if err != nil {
			util.LOGGER.Errorf(err, "add rule failed, serviceId is %s: invalid rule.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		//黑白名单只能存在一种，黑名单 or 白名单
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else {
			if ruleType != rule.RuleType {
				util.LOGGER.Errorf(nil, "add rule failed, serviceId is %s:can only exist one type, BLACK or WHITE.", in.ServiceId)
				return &pb.AddServiceRulesResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "Service can only contain one rule type, BLACK or WHITE."),
				}, nil
			}
		}

		//同一服务，attribute和pattern确定一个rule
		if serviceUtil.RuleExist(ctx, tenant, in.ServiceId, rule.Attribute, rule.Pattern) {
			util.LOGGER.Infof("This rule more exists, %s ", in.ServiceId)
			continue
		}

		// 产生全局rule id
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &pb.ServiceRule{
			RuleId:      dynamic.GenerateUuid(),
			RuleType:    rule.RuleType,
			Attribute:   rule.Attribute,
			Pattern:     rule.Pattern,
			Description: rule.Description,
			Timestamp:   timestamp,
		}

		key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleAdd.RuleId)
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, ruleAdd.Attribute, ruleAdd.Pattern)
		ruleIds = append(ruleIds, ruleAdd.RuleId)

		util.LOGGER.Debugf("indexKey is : %s", indexKey)
		util.LOGGER.Debugf("start add service rule file: %s", key)
		data, err := json.Marshal(ruleAdd)
		if err != nil {
			util.LOGGER.Errorf(err, "add rule failed, serviceId is %s: marshal rule failed.", in.ServiceId)
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service rule file marshal error."),
			}, err
		}

		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(key),
			Value:  data,
		})
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(indexKey),
			Value:  []byte(ruleAdd.RuleId),
		})
	}
	if len(opts) <= 0 {
		util.LOGGER.Infof("add rule successful, serviceId is %s: rule more exists,no rules to add.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Service rules has been added."),
		}, nil
	}
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "add rule failed, serviceId is %s:commit date into etcd failed.", in.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("add rule successful, serviceId  %s.", in.ServiceId)
	return &pb.AddServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service rules successfully."),
		RuleIds:  ruleIds,
	}, nil
}

func (s *ServiceController) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	if in == nil || in.GetRule() == nil || len(in.ServiceId) == 0 || len(in.RuleId) == 0 {
		util.LOGGER.Errorf(nil, "update rule failed: invalid parameters.")
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "update rule failed, serviceId is %s, ruleId is %s: service not exist.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	err := apt.Validate(in.Rule)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: invalid service rule.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	//是否能改变ruleType
	ruleType, ruleNum, err := serviceUtil.GetServiceRuleType(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: get rule type failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	if ruleNum >= 1 && ruleType != in.Rule.RuleType {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: rule type can exist one type, BLACK or WHITE.rule type is %s", in.ServiceId, in.RuleId, in.Rule.RuleType)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Exist multiple rules,can not change rule type. Rule type is "+ruleType),
		}, nil
	}

	rule, err := serviceUtil.GetOneRule(ctx, tenant, in.ServiceId, in.RuleId)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: query service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service rule file failed."),
		}, err
	}
	if rule == nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s:this rule does not exist,can't update.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist."),
		}, nil
	}

	oldRulePatten := rule.Pattern
	oldRuleAttr := rule.Attribute
	isChangeIndex := false
	if rule.Attribute != in.GetRule().Attribute {
		isChangeIndex = true
		rule.Attribute = in.GetRule().Attribute
	}
	if rule.Pattern != in.GetRule().Pattern {
		isChangeIndex = true
		rule.Pattern = in.GetRule().Pattern
	}
	rule.RuleType = in.GetRule().RuleType
	rule.Description = in.GetRule().Description
	rule.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)

	key := apt.GenerateServiceRuleKey(tenant, in.ServiceId, in.RuleId)
	util.LOGGER.Debugf("start update service rule file: %s", key)
	data, err := json.Marshal(rule)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: marshal service rule failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service rule file marshal error."),
		}, err
	}
	opts := []*registry.PluginOp{}
	if isChangeIndex {
		//加入新的rule index
		indexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, rule.Attribute, rule.Pattern)
		opt := &registry.PluginOp{
			Action: registry.PUT,
			Key:    []byte(indexKey),
			Value:  []byte(rule.RuleId),
		}
		opts = append(opts, opt)

		//删除旧的rule index
		oldIndexKey := apt.GenerateRuleIndexKey(tenant, in.ServiceId, oldRuleAttr, oldRulePatten)
		opt = &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(oldIndexKey),
		}

		opts = append(opts, opt)
	}
	opt := &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	}
	opts = append(opts, opt)
	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "update rule failed, serviceId is %s, ruleId is %s: commit date into etcd failed.", in.ServiceId, in.RuleId)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("update rule successful: servieId is %s, ruleId is %s.", in.ServiceId, in.RuleId)
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
	}, nil
}

func (s *ServiceController) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: invalid params.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	rules, err := serviceUtil.GetRulesUtil(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(nil, "get service rule failed, serviceId is %s: get rule failed.", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service rules failed."),
		}, nil
	}

	return &pb.GetServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (s *ServiceController) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "delete service rule failed: invalid parameters.")
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: service not exist.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	opts := []*registry.PluginOp{}
	key := ""
	indexKey := ""
	for _, ruleId := range in.RuleIds {
		key = apt.GenerateServiceRuleKey(tenant, in.ServiceId, ruleId)
		util.LOGGER.Debugf("start delete service rule file: %s", key)
		data, err := serviceUtil.GetOneRule(ctx, tenant, in.ServiceId, ruleId)
		if err != nil {
			util.LOGGER.Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: get rule of ruleId %s failed.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if data == nil {
			util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: ruleId %s not exist.", in.ServiceId, in.RuleIds, ruleId)
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "This rule does not exist."),
			}, nil
		}
		indexKey = apt.GenerateRuleIndexKey(tenant, in.ServiceId, data.Attribute, data.Pattern)
		opts = append(opts, &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(key),
		})
		opts = append(opts, &registry.PluginOp{
			Action: registry.DELETE,
			Key:    []byte(indexKey),
		})
	}
	if len(opts) <= 0 {
		util.LOGGER.Errorf(nil, "delete service rule failed, serviceId is %s, rule is %v: rule has been deleted.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "No service rule has been deleted."),
		}, nil
	}
	_, err := registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service rule failed, serviceId is %s, rule is %v: commit data into etcd failed.", in.ServiceId, in.RuleIds)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("delete rule successful: serviceId %s, ruleIds %v", in.ServiceId, in.RuleIds)
	return &pb.DeleteServiceRulesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service rules successfully."),
	}, nil
}

func (s *ServiceController) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	if in == nil {
		util.LOGGER.Errorf(nil, "exist failed: invalid params.")
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	switch in.Type {
	case "microservice":
		if len(in.AppId) == 0 || len(in.ServiceName) == 0 || len(in.Version) == 0 {
			util.LOGGER.Errorf(nil, "microservice exist failed: invalid params.")
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}
		err := apt.GetMSExistsReqValidator.Validate(in)
		serviceFlag := strings.Join([]string{in.AppId, in.ServiceName, in.Version}, "/")
		if err != nil {
			util.LOGGER.Errorf(err, "microservice exist failed, service %s: invalid params.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}

		ids, err := ms.FindServiceIds(ctx, in.Version, &pb.MicroServiceKey{
			AppId:       in.AppId,
			ServiceName: in.ServiceName,
			Alias:       in.ServiceName,
			Version:     in.Version,
			Tenant:      tenant,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "microservice exist failed, service %s: find serviceIds failed.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get service file failed."),
			}, err
		}
		if len(ids) <= 0 {
			util.LOGGER.Infof("microservice exist failed, service %s: service not exist.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get service id successfully."),
			ServiceId: ids[0], // 约定多个时，取较新版本
		}, nil
	case "schema":
		if len(in.SchemaId) == 0 || len(in.ServiceId) == 0 {
			util.LOGGER.Errorf(nil, "schema exist failed, serviceId %s, schemaId : invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}

		err := apt.GetSchemaExistsReqValidator.Validate(in)
		if err != nil {
			util.LOGGER.Errorf(err, "schema exist failed, serviceId %s, schemaId : invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
			util.LOGGER.Warnf(nil, "schema exist failed, serviceId %s, schemaId : service not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, in.SchemaId)
		err, exist := serviceUtil.CheckSchemaInfoExist(ctx, key)
		if err != nil {
			util.LOGGER.Errorf(err, "schema exist failed, serviceId %s, schemaId : get schema failed.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if !exist {
			util.LOGGER.Infof("schema exist failed, serviceId %s, schemaId %s: schema not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Schema does not exist."),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Schema exist."),
			SchemaId: in.SchemaId,
		}, nil
	default:
		util.LOGGER.Warnf(nil, "unexpected type '%s' for query.", in.Type)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Only microservice and schema can be used as type."),
		}, nil
	}
}

func (s *ServiceController) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetTags()) == 0 {
		util.LOGGER.Errorf(nil, "add service tags failed: invalid parameters.")
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: invalid parameters.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "add service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	addTags := in.GetTags()
	dataTags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: get existed tag failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags failed."),
		}, err
	}
	if len(dataTags) > 0 {
		for key, value := range addTags {
			dataTags[key] = value
		}
	} else {
		dataTags = addTags
	}

	err = serviceUtil.AddTagIntoETCD(ctx, tenant, in.ServiceId, dataTags)
	if err != nil {
		util.LOGGER.Errorf(err, "add service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("add service tags successful, serviceId %s, tags %v.", in.ServiceId, in.Tags)
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (s *ServiceController) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Key) == 0 || len(in.Value) == 0 {
		util.LOGGER.Errorf(nil, "update service tag failed: invalid parameters.")
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tagFlag := strings.Join([]string{in.Key, in.Value}, "/")
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: invalid params.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: service not exist.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: get tag failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags for service failed."),
		}, err
	}
	//check tag 是否存在
	if _, ok := tags[in.Key]; !ok {
		util.LOGGER.Errorf(nil, "update service tag failed, serviceId %s, tag %s: tag not exist,please add first.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update tag for service failed for update tags not exist, please add first."),
		}, err
	}
	tags[in.Key] = in.Value

	err = serviceUtil.AddTagIntoETCD(ctx, tenant, in.ServiceId, tags)

	if err != nil {
		util.LOGGER.Errorf(err, "update service tag failed, serviceId %s, tag %s: adding service tags failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit into etcd failed."),
		}, err
	}

	util.LOGGER.Infof("update tag successful, serviceId %s, tag %s.", in.ServiceId, tagFlag)
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Update service tag success."),
	}, nil
}

func (s *ServiceController) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Keys) == 0 {
		util.LOGGER.Errorf(nil, "delete service tags failed: invalid parameters.")
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: invalid params.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: query service failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service tags file failed."),
		}, err
	}
	for _, key := range in.Keys {
		if _, ok := tags[key]; !ok {
			util.LOGGER.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: tag %s not exist.", in.ServiceId, in.Keys, key)
			return &pb.DeleteServiceTagsResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(tags, key)
	}

	// tags 可能size == 0
	data, err := json.Marshal(tags)
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: marshall service tag failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Marshal service tags file failed."),
		}, err
	}

	key := apt.GenerateServiceTagKey(tenant, in.ServiceId)

	util.LOGGER.Debugf("start delete service tags file: %s %v", key, in.Keys)
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "delete service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.LOGGER.Infof("delete service tags successful: serviceId %s, tag %v.", in.ServiceId, in.Keys)
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service tags successfully."),
	}, nil
}

func (s *ServiceController) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "get service tags failed: invalid params.")
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: invalid parameters.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: service not exist.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "get service tags failed, serviceId %s: get tag failed.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags for service failed."),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service tags successfully."),
		Tags:     tags,
	}, nil
}

func (s *ServiceController) GetSchemaInfo(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.LOGGER.Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}

	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(nil, "get schema failed, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, request.ServiceId) {
		util.LOGGER.Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	resp, errDo := store.Store().Schema().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: get schema info failed.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get schema info failed."),
		}, errDo
	}
	if resp.Count == 0 {
		util.LOGGER.Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: schema not exists.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Do not have this schema info."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully."),
		Schema:   string(resp.Kvs[0].Value),
	}, nil
}

func (s *ServiceController) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.LOGGER.Errorf(nil, "delete schema failded: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(err, "delete schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, request.ServiceId) {
		util.LOGGER.Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	err, exist := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.LOGGER.Errorf(err, "delete schema failded, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.LOGGER.Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, nil
	}
	_, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    []byte(key),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "delete schema failded, serviceId %s, schemaId %s: delete schema from etcd faild.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Delete schema info failed."),
		}, errDo
	}
	util.LOGGER.Infof("delete schema info successfully.%s", request.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *ServiceController) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	err, rst := s.canModifySchema(ctx, request)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed."),
		}, err
	}
	if !rst {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed, service or schemaId not exist."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	_, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  []byte(request.Schema),
	})
	if errDo != nil {
		util.LOGGER.Errorf(errDo, "update schema failded, serviceId %s, schemaId %s: commit schema into etcd failed.", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed."),
		}, errDo
	}
	util.LOGGER.Infof("update schema success: serviceId %s, schemaId %s.", request.ServiceId, request.SchemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil

}

func (s *ServiceController) canModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (error, bool) {
	if request == nil || len(request.Schema) == 0 || len(request.SchemaId) == 0 || len(request.ServiceId) == 0 {
		util.LOGGER.Errorf(nil, "update schema failded: invalid params.")
		return nil, false
	}
	err := apt.Validate(request)
	if err != nil {
		util.LOGGER.Errorf(err, "update schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return err, false
	}
	tenant := util.ParseTenantProject(ctx)
	service, err := ms.GetServiceByServiceId(ctx, tenant, request.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "update schema failded, serviceId %s, schemaId %s: get service failed.", request.ServiceId, request.SchemaId)
		return err, false
	}
	if service == nil {
		util.LOGGER.Errorf(nil, "update schema failded, serviceId %s, schemaId %s: service not exist,%s", request.ServiceId, request.SchemaId)
		return nil, false
	}
	schemas := service.Schemas
	if !containsValueInSlice(schemas, request.SchemaId) {
		message := "Do not contain " + request.SchemaId + "Schema in service file."
		util.LOGGER.Errorf(nil, "update schema failded, serviceId %s, schemaId %s:%s", request.ServiceId, request.SchemaId, message)
		return nil, false
	}
	return nil, true
}

func containsValueInSlice(in []string, value string) bool {
	if in == nil || len(value) == 0 {
		return false
	}
	for _, i := range in {
		if i == value {
			return true
		}
	}
	return false
}

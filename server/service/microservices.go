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
	"fmt"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/plugins/dynamic"
	rs "github.com/ServiceComb/service-center/server/rest"
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
	"github.com/astaxie/beego"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type ServiceController struct {
}

func (s *ServiceController) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if in == nil || in.Service == nil {
		util.Logger().Errorf(nil, "create microservice failed : param empty.")
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "request format invalid"),
		}, nil
	}

	//create service
	rsp, err := s.CreateServicePri(ctx, in)
	if err != nil || rsp.GetResponse().Code != pb.Response_SUCCESS {
		return rsp, err
	}

	if s.isCreateServiceEx(in) == false {
		return rsp, err
	}

	//create tag,rule,instances
	rsp, err = s.CreateServiceEx(ctx, in, rsp.ServiceId)
	return rsp, err
}

func (s *ServiceController) CreateServicePri(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	if in == nil || in.Service == nil {
		util.Logger().Errorf(nil, "create microservice failed: param empty. operator: %s", remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	service := in.Service
	err := apt.Validate(service)
	serviceFlag := util.StringJoin([]string{service.AppId, service.ServiceName, service.Version}, "/")
	if err != nil {
		util.Logger().Errorf(err, "create microservice failed, %s: invalid parameters. operator: %s",
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
	err = checkBeforeCreate(ctx, consumer)
	if err != nil {
		util.Logger().Errorf(err, "create microservice failed, %s: check service failed before create. operator: %s",
			serviceFlag, remoteIP)
		resp := &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}
		switch err.(type) {
		case errorsEx.InternalError:
			return resp, err
		default:
			return resp, nil
		}
	}

	// 产生全局service id
	serviceId := in.Service.ServiceId
	if len(serviceId) == 0 {
		serviceId = dynamic.GetServiceId()
	}
	service.ServiceId = serviceId
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	service.ModTimestamp = service.Timestamp

	data, err := json.Marshal(service)
	if err != nil {
		util.Logger().Errorf(err, "create microservice failed, %s: json marshal service failed. operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Body error "+err.Error()),
		}, nil
	}
	key := apt.GenerateServiceKey(tenant, serviceId)
	index := apt.GenerateServiceIndexKey(consumer)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(consumer))
	util.Logger().Debugf("start register service: %s %v", key, service)
	util.Logger().Debugf("start register service index: %s %v", index, serviceId)
	opts := []*registry.PluginOp{
		{
			Action: registry.PUT,
			Key:    util.StringToBytesWithNoCopy(key),
			Value:  data,
		},
		{
			Action: registry.PUT,
			Key:    indexBytes,
			Value:  util.StringToBytesWithNoCopy(serviceId),
		},
	}
	uniqueCmpOpts := []*registry.CompareOp{
		{
			Key:    indexBytes,
			Type:   registry.CMP_VERSION,
			Result: registry.CMP_EQUAL,
			Value:  0,
		},
	}

	if len(consumer.Alias) > 0 {
		opts = append(opts, &registry.PluginOp{
			Action: registry.PUT,
			Key:    aliasBytes,
			Value:  util.StringToBytesWithNoCopy(serviceId),
		})
		uniqueCmpOpts = append(uniqueCmpOpts, &registry.CompareOp{
			Key:    aliasBytes,
			Type:   registry.CMP_VERSION,
			Result: registry.CMP_EQUAL,
			Value:  0,
		})
	}

	resp, err := registry.GetRegisterCenter().TxnWithCmp(ctx, opts, uniqueCmpOpts, nil)
	if err != nil {
		util.Logger().Errorf(err, "create microservice failed, %s: commit data into etcd failed. operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}
	if !resp.Succeeded {
		if s.isCreateServiceEx(in) == true {
			serviceIdInner, _ := ms.GetServiceId(ctx, consumer)
			util.Logger().Warnf(nil, "create microservice failed, serviceid = %s , flag = %s: service already exists. operator: %s",
				serviceIdInner, serviceFlag, remoteIP)

			return &pb.CreateServiceResponse{
				Response:  pb.CreateResponse(pb.Response_SUCCESS, "register service successfully"),
				ServiceId: serviceIdInner,
			}, nil
		}

		util.Logger().Warnf(nil, "create microservice failed, %s: service already exists. operator: %s",
			serviceFlag, remoteIP)

		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service already exists."),
		}, nil
	}

	//创建服务间的依赖
	err = dependency.UpdateAsProviderDependency(ctx, serviceId, consumer)
	if err != nil {
		util.Logger().Errorf(err, "create microservice: update dependency as provider %s(%s) failed. operator: %s",
			serviceId, serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}

	util.Logger().Infof("create microservice successful, %s, serviceId: %s. operator: %s",
		serviceFlag, service.ServiceId, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Register service successfully."),
		ServiceId: serviceId,
	}, nil
}

func checkBeforeCreate(ctx context.Context, serviceKey *pb.MicroServiceKey) error {
	ok, err := quota.QuotaPlugins[beego.AppConfig.DefaultString("quota_plugin", "buildin")]().Apply4Quotas(ctx, quota.MicroServiceQuotaType, 0)
	if err != nil {
		return errorsEx.InternalError(err.Error())
	}
	if !ok {
		return fmt.Errorf("No quota to create service.")
	}
	return nil
}

func (s *ServiceController) DeleteServicePri(ctx context.Context, ServiceId string, force bool) (*pb.Response, error) {
	tenant := util.ParseTenantProject(ctx)

	service, err := ms.GetService(ctx, tenant, ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: get service failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}

	if service == nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: service not exist.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, "Service does not exist."), nil
	}

	util.Logger().Infof("start delete service %s", ServiceId)

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		keyConDependency := apt.GenerateProviderDependencyKey(tenant, ServiceId, "")
		services, err := dependency.GetDependencies(ctx, keyConDependency, tenant)
		if err != nil {
			util.Logger().Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err, get service dependency failed.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Get dependency info failed."), err
		}
		if len(services) > 1 || (len(services) == 1 && services[0].ServiceId != ServiceId) {
			util.Logger().Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, other services rely it.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(tenant, ServiceId, "")
		rsp, err := store.Store().Instance().Search(ctx, &registry.PluginOp{
			Action:     registry.GET,
			Key:        util.StringToBytesWithNoCopy(instancesKey),
			WithPrefix: true,
			CountOnly:  true,
		})
		if err != nil {
			util.Logger().Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err,get instances failed.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Get instance failed."), err
		}

		if rsp.Count > 0 {
			util.Logger().Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, exist instance.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, exist instance."), err
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
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, refresh service dependency cache failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, "Refresh dependency cache failed."), err
	}

	opts := []*registry.PluginOp{
		{
			Action: registry.DELETE,
			Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceIndexKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(consumer)),
		},
		{
			Action: registry.DELETE,
			Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceKey(tenant, ServiceId)),
		},
		{
			Action: registry.DELETE,
			Key: util.StringToBytesWithNoCopy(util.StringJoin([]string{
				apt.GetServiceRuleRootKey(tenant),
				ServiceId,
				"",
			}, "/")),
		},
	}

	//删除依赖规则
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, create lock failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	optsTmp, err := dependency.DeleteDependencyForService(ctx, consumer, ServiceId)
	lock.Unlock()
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, delete dependency failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	opts = append(opts, optsTmp...)

	//删除黑白名单
	rulekey := apt.GenerateServiceRuleKey(tenant, ServiceId, "")
	opt := &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        util.StringToBytesWithNoCopy(rulekey),
		WithPrefix: true,
	}
	opts = append(opts, opt)
	indexKey := apt.GenerateRuleIndexKey(tenant, ServiceId, "", "")
	opts = append(opts, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(indexKey),
	})
	opts = append(opts, opt)

	//删除shemas
	schemaKey := apt.GenerateServiceSchemaKey(tenant, ServiceId, "")
	opt = &registry.PluginOp{
		Action:     registry.DELETE,
		Key:        util.StringToBytesWithNoCopy(schemaKey),
		WithPrefix: true,
	}
	opts = append(opts, opt)

	//删除tags
	tagsKey := apt.GenerateServiceTagKey(tenant, ServiceId)
	opt = &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(tagsKey),
	}
	opts = append(opts, opt)

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: delete all instances failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, "Delete all instances failed for service."), err
	}

	_, err = registry.GetRegisterCenter().Txn(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: commit data into etcd failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."), nil
	}

	util.Logger().Infof("delete microservice successful: serviceid is %s,operator is %s.", ServiceId, util.GetIPFromContext(ctx))
	return pb.CreateResponse(pb.Response_SUCCESS, "Unregister service successfully."), nil
}

func (s *ServiceController) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.ServiceId == apt.Service.ServiceId {
		util.Logger().Errorf(nil, "delete microservice failed: service empty.")
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	resp, err := s.DeleteServicePri(ctx, in.ServiceId, in.Force)

	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

func (s *ServiceController) DeleteServices(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	// 合法性检查
	if request == nil || request.ServiceIds == nil || len(request.ServiceIds) == 0 {
		return &pb.DelServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request param."),
			Services: nil,
		}, nil
	}

	existFlag := map[string]bool{}
	nuoMultilCount := 0
	// 批量删除服务
	serviceRespChan := make(chan *pb.DelServicesRspInfo, len(request.ServiceIds))
	for _, serviceId := range request.ServiceIds {
		//ServiceId重复性检查
		if _, ok := existFlag[serviceId]; ok {
			util.Logger().Warnf(nil, "delete microservice %s , multiple.", serviceId)
			continue
		} else {
			existFlag[serviceId] = true
			nuoMultilCount++
		}

		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceId,
			ErrMessage: "",
		}

		//检查服务ID合法性
		in := &pb.DeleteServiceRequest{
			ServiceId: serviceId,
			Force:     request.Force,
		}
		err := apt.Validate(in)
		if err != nil {
			util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: invalid parameters.", in.ServiceId)
			serviceRst.ErrMessage = err.Error()
			serviceRespChan <- serviceRst
			continue
		}

		//执行删除服务操作
		go func(serviceItem string) {
			resp, err := s.DeleteServicePri(ctx, serviceItem, request.Force)
			if err != nil {
				serviceRst.ErrMessage = err.Error()
			} else if resp.Code != pb.Response_SUCCESS {
				serviceRst.ErrMessage = resp.Message
			}

			serviceRespChan <- serviceRst
		}(serviceId)
	}

	//获取批量删除服务的结果
	count := 0
	responseCode := pb.Response_SUCCESS
	delServiceRspInfo := []*pb.DelServicesRspInfo{}
	for serviceRespItem := range serviceRespChan {
		count++
		if len(serviceRespItem.ErrMessage) != 0 {
			responseCode = pb.Response_FAIL
		}
		delServiceRspInfo = append(delServiceRspInfo, serviceRespItem)
		//结果收集over，关闭通道
		if count == nuoMultilCount {
			close(serviceRespChan)
		}
	}

	util.Logger().Infof("Batch DeleteServices serviceId = %v , result = %d, ", request.ServiceIds, responseCode)
	return &pb.DelServicesResponse{
		Response: pb.CreateResponse(responseCode, "Delete services successfully."),
		Services: delServiceRspInfo,
	}, nil
}

func (s *ServiceController) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "get microservice failed, serviceId is %s: invalid parameters.",
			in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)
	service, err := ms.GetService(ctx, tenant, in.ServiceId)

	if err != nil {
		util.Logger().Errorf(err, "get microservice failed, serviceId is %s: inner err,get service failed.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service file failed."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "get microservice failed, serviceId is %s: service not exist.", in.ServiceId)
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
		util.Logger().Errorf(nil, "get services failed: invalid params.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	services, err := ms.GetAllServiceUtil(ctx)
	if err != nil {
		util.Logger().Errorf(err, "get services failed: inner err.")
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
		util.Logger().Errorf(nil, "update service properties failed: invalid params.")
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	key := apt.GenerateServiceKey(tenant, in.ServiceId)
	service, err := ms.GetService(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: query service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query service file failed."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "update service properties failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	service.Properties = make(map[string]string)
	for propertyKey := range in.Properties {
		service.Properties[propertyKey] = in.Properties[propertyKey]
	}
	service.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(service)
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: json marshal service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service file marshal error."),
		}, err
	}

	// Set key file
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(key),
		Value:  data,
	})
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: commit data into etcd failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("update service properties successful: serviceId is %s.", in.ServiceId)
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Update service successfully."),
	}, nil
}

func (s *ServiceController) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	if in == nil {
		util.Logger().Errorf(nil, "exist failed: invalid params.")
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	switch in.Type {
	case "microservice":
		if len(in.AppId) == 0 || len(in.ServiceName) == 0 || len(in.Version) == 0 {
			util.Logger().Errorf(nil, "microservice exist failed: invalid params.")
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}
		err := apt.GetMSExistsReqValidator.Validate(in)
		serviceFlag := util.StringJoin([]string{in.AppId, in.ServiceName, in.Version}, "/")
		if err != nil {
			util.Logger().Errorf(err, "microservice exist failed, service %s: invalid params.", serviceFlag)
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
			util.Logger().Errorf(err, "microservice exist failed, service %s: find serviceIds failed.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get service file failed."),
			}, err
		}
		if len(ids) <= 0 {
			util.Logger().Infof("microservice exist failed, service %s: service not exist.", serviceFlag)
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
			util.Logger().Errorf(nil, "schema exist failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request."),
			}, nil
		}

		err := apt.GetSchemaExistsReqValidator.Validate(in)
		if err != nil {
			util.Logger().Errorf(err, "schema exist failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, nil
		}
		if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
			util.Logger().Warnf(nil, "schema exist failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, in.SchemaId)
		err, exist := serviceUtil.CheckSchemaInfoExist(ctx, key)
		if err != nil {
			util.Logger().Errorf(err, "schema exist failed, serviceId %s, schemaId %s: get schema failed.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		if !exist {
			util.Logger().Infof("schema exist failed, serviceId %s, schemaId %s: schema not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Schema does not exist."),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Schema exist."),
			SchemaId: in.SchemaId,
		}, nil
	default:
		util.Logger().Warnf(nil, "unexpected type '%s' for query.", in.Type)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Only microservice and schema can be used as type."),
		}, nil
	}
}

func (s *ServiceController) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.GetTags()) == 0 {
		util.Logger().Errorf(nil, "add service tags failed: invalid parameters.")
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "add service tags failed, serviceId %s, tags %v: invalid parameters.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)
	// service id存在性校验
	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "add service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	addTags := in.GetTags()
	dataTags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "add service tags failed, serviceId %s, tags %v: get existed tag failed.", in.ServiceId, in.Tags)
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
		util.Logger().Errorf(err, "add service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("add service tags successful, serviceId %s, tags %v.", in.ServiceId, in.Tags)
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (s *ServiceController) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Key) == 0 || len(in.Value) == 0 {
		util.Logger().Errorf(nil, "update service tag failed: invalid parameters.")
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tagFlag := util.StringJoin([]string{in.Key, in.Value}, "/")
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "update service tag failed, serviceId %s, tag %s: invalid params.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(err, "update service tag failed, serviceId %s, tag %s: service not exist.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update service tag failed, serviceId %s, tag %s: get tag failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get tags for service failed."),
		}, err
	}
	//check tag 是否存在
	if _, ok := tags[in.Key]; !ok {
		util.Logger().Errorf(nil, "update service tag failed, serviceId %s, tag %s: tag not exist,please add first.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Update tag for service failed for update tags not exist, please add first."),
		}, err
	}
	tags[in.Key] = in.Value

	err = serviceUtil.AddTagIntoETCD(ctx, tenant, in.ServiceId, tags)

	if err != nil {
		util.Logger().Errorf(err, "update service tag failed, serviceId %s, tag %s: adding service tags failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit into etcd failed."),
		}, err
	}

	util.Logger().Infof("update tag successful, serviceId %s, tag %s.", in.ServiceId, tagFlag)
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Update service tag success."),
	}, nil
}

func (s *ServiceController) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.Keys) == 0 {
		util.Logger().Errorf(nil, "delete service tags failed: invalid parameters.")
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "delete service tags failed, serviceId %s, tags %v: invalid params.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(nil, "delete service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "delete service tags failed, serviceId %s, tags %v: query service failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service tags file failed."),
		}, err
	}
	for _, key := range in.Keys {
		if _, ok := tags[key]; !ok {
			util.Logger().Errorf(nil, "delete service tags failed, serviceId %s, tags %v: tag %s not exist.", in.ServiceId, in.Keys, key)
			return &pb.DeleteServiceTagsResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(tags, key)
	}

	// tags 可能size == 0
	data, err := json.Marshal(tags)
	if err != nil {
		util.Logger().Errorf(err, "delete service tags failed, serviceId %s, tags %v: marshall service tag failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Marshal service tags file failed."),
		}, err
	}

	key := apt.GenerateServiceTagKey(tenant, in.ServiceId)

	util.Logger().Debugf("start delete service tags file: %s %v", key, in.Keys)
	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    util.StringToBytesWithNoCopy(key),
		Value:  data,
	})
	if err != nil {
		util.Logger().Errorf(err, "delete service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Commit operations failed."),
		}, err
	}

	util.Logger().Infof("delete service tags successful: serviceId %s, tag %v.", in.ServiceId, in.Keys)
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service tags successfully."),
	}, nil
}

func (s *ServiceController) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.Logger().Errorf(nil, "get service tags failed: invalid params.")
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "get service tags failed, serviceId %s: invalid parameters.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, in.ServiceId) {
		util.Logger().Errorf(err, "get service tags failed, serviceId %s: service not exist.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, tenant, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "get service tags failed, serviceId %s: get tag failed.", in.ServiceId)
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
		util.Logger().Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}

	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, request.ServiceId) {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	resp, errDo := store.Store().Schema().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(key),
	})
	if errDo != nil {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: get schema info failed.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get schema info failed."),
		}, errDo
	}
	if resp.Count == 0 {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: schema not exists.", request.ServiceId, request.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Do not have this schema info."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully."),
		Schema:   util.BytesToStringWithNoCopy(resp.Kvs[0].Value),
	}, nil
}

func (s *ServiceController) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.Logger().Errorf(nil, "delete schema failded: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)

	if !ms.ServiceExist(ctx, tenant, request.ServiceId) {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	err, exist := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, nil
	}
	_, errDo := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.DELETE,
		Key:    util.StringToBytesWithNoCopy(key),
	})
	if errDo != nil {
		util.Logger().Errorf(errDo, "delete schema failded, serviceId %s, schemaId %s: delete schema from etcd faild.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Delete schema info failed."),
		}, errDo
	}
	util.Logger().Infof("delete schema info successfully.%s", request.SchemaId)
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
		Key:    util.StringToBytesWithNoCopy(key),
		Value:  util.StringToBytesWithNoCopy(request.Schema),
	})
	if errDo != nil {
		util.Logger().Errorf(errDo, "update schema failded, serviceId %s, schemaId %s: commit schema into etcd failed.", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed."),
		}, errDo
	}
	util.Logger().Infof("update schema success: serviceId %s, schemaId %s.", request.ServiceId, request.SchemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil

}

func (s *ServiceController) canModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (error, bool) {
	if request == nil || len(request.Schema) == 0 || len(request.SchemaId) == 0 || len(request.ServiceId) == 0 {
		util.Logger().Errorf(nil, "update schema failded: invalid params.")
		return nil, false
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return err, false
	}
	tenant := util.ParseTenantProject(ctx)
	service, err := ms.GetService(ctx, tenant, request.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: get service failed.", request.ServiceId, request.SchemaId)
		return err, false
	}
	if service == nil {
		util.Logger().Errorf(nil, "update schema failded, serviceId %s, schemaId %s: service not exist,%s", request.ServiceId, request.SchemaId)
		return nil, false
	}
	schemas := service.Schemas
	if !containsValueInSlice(schemas, request.SchemaId) {
		message := "Do not contain " + request.SchemaId + "Schema in service file."
		util.Logger().Errorf(nil, "update schema failded, serviceId %s, schemaId %s:%s", request.ServiceId, request.SchemaId, message)
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

func (s *ServiceController) CreateServiceEx(ctx context.Context, in *pb.CreateServiceRequest, serviceId string) (*pb.CreateServiceResponse, error) {
	result := &pb.CreateServiceResponse{
		ServiceId: serviceId,
		Response:  &pb.Response{},
	}
	var chanLen int = 0
	createRespChan := make(chan *pb.Response, 10)
	//create rules
	if in.Rules != nil && len(in.Rules) != 0 {
		chanLen++
		go func() {
			req := &pb.AddServiceRulesRequest{
				ServiceId: serviceId,
				Rules:     in.Rules,
			}
			chanRsp := &pb.Response{}
			rsp, err := s.AddRule(ctx, req)
			if err != nil {
				chanRsp.Message = err.Error()
			}

			if rsp.Response.Code != pb.Response_SUCCESS {
				chanRsp.Message = rsp.Response.Message
			}
			createRespChan <- chanRsp
		}()
	}
	//create tags
	if in.Tags != nil && len(in.Tags) != 0 {
		chanLen++
		go func() {
			req := &pb.AddServiceTagsRequest{
				ServiceId: serviceId,
				Tags:      in.Tags,
			}
			chanRsp := &pb.Response{}
			rsp, err := s.AddTags(ctx, req)
			if err != nil {
				chanRsp.Message = err.Error()
			}

			if rsp.Response.Code != pb.Response_SUCCESS {
				chanRsp.Message = rsp.Response.Message
			}
			createRespChan <- chanRsp
		}()
	}
	// create instance
	if in.Instances != nil && len(in.Instances) != 0 {
		chanLen++
		go func() {
			chanRsp := &pb.Response{}
			for _, ins := range in.Instances {
				req := &pb.RegisterInstanceRequest{
					Instance: ins,
				}
				req.Instance.ServiceId = serviceId
				rsp, err := rs.InstanceAPI.Register(ctx, req)
				if err != nil {
					chanRsp.Message += fmt.Sprintf("{instance:%v,result:%s}", ins.Endpoints, err.Error())
				}
				if rsp.Response.Code != pb.Response_SUCCESS {
					chanRsp.Message += fmt.Sprintf("{instance:%v,result:%s}", ins.Endpoints, rsp.Response.Message)
				}
				createRespChan <- chanRsp
			}
		}()
	}
	if chanLen == 0 {
		close(createRespChan)
		return result, nil
	}

	// handle result
	var errMessages []string
	var count int = 0
	for createResp := range createRespChan {
		count++
		if len(createResp.Message) != 0 {
			errMessages = append(errMessages, createResp.Message)
		}

		if count == chanLen {
			close(createRespChan)
		}
	}

	if len(errMessages) != 0 {
		result.Response.Code = pb.Response_FAIL
		errMessage, err := json.Marshal(errMessages)
		if err != nil {
			result.Response.Message = "marshal errMessages error"
			util.Logger().Error("marshal errMessages error", err)
			return result, nil
		}
		result.Response.Message = fmt.Sprintf("ErrMessage : %s", errMessage)
	} else {
		result.Response.Code = pb.Response_SUCCESS
	}

	util.Logger().Infof("CreateServiceEx, serviceid = %s, result = %s ", result.ServiceId, result.Response.Message)
	return result, nil
}

func (s *ServiceController) isCreateServiceEx(in *pb.CreateServiceRequest) bool {
	if (in.Rules != nil && len(in.Rules) != 0) || (in.Tags != nil && len(in.Tags) != 0) || (in.Instances != nil && len(in.Instances) != 0) {
		return true
	}
	return false
}

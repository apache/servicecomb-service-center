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
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
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
	return s.CreateServiceEx(ctx, in, rsp.ServiceId)
}

func (s *ServiceController) CreateServicePri(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	service := in.Service
	serviceFlag := util.StringJoin([]string{service.AppId, service.ServiceName, service.Version}, "/")

	err := apt.Validate(service)
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
	err = checkBeforeCreate(ctx, tenant)
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
	opts := []registry.PluginOp{
		registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)),
		registry.OpPut(registry.WithKey(indexBytes), registry.WithStrValue(serviceId)),
	}
	uniqueCmpOpts := []registry.CompareOp{
		registry.OpCmp(registry.CmpVer(indexBytes), registry.CMP_EQUAL, 0),
	}

	if len(consumer.Alias) > 0 {
		opts = append(opts, registry.OpPut(registry.WithKey(aliasBytes), registry.WithStrValue(serviceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			registry.OpCmp(registry.CmpVer(aliasBytes), registry.CMP_EQUAL, 0))
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
			serviceIdInner, _ := serviceUtil.GetServiceId(ctx, consumer)
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

	util.Logger().Infof("create microservice successful, %s, serviceId: %s. operator: %s",
		serviceFlag, service.ServiceId, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Register service successfully."),
		ServiceId: serviceId,
	}, nil
}

func checkBeforeCreate(ctx context.Context, tenant string) error {
	ok, err := quota.QuotaPlugins[quota.QuataType]().Apply4Quotas(ctx, quota.MicroServiceQuotaType, tenant, "", 1)
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

	service, err := serviceUtil.GetService(ctx, tenant, ServiceId)
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
		dr := serviceUtil.NewConsumerDependencyRelation(ctx, tenant, ServiceId, service)
		services, err := dr.GetDependencyProviderIds()
		if err != nil {
			util.Logger().Errorf(err, "delete microservice failed, serviceId is %s:(unforce) inner err, get service dependency failed.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Get dependency info failed."), err
		}
		if len(services) > 1 || (len(services) == 1 && services[0] != ServiceId) {
			util.Logger().Errorf(nil, "delete microservice failed, serviceId is %s:(unforce) can't delete, other services rely it.", ServiceId)
			return pb.CreateResponse(pb.Response_FAIL, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(tenant, ServiceId, "")
		rsp, err := store.Store().Instance().Search(ctx,
			registry.WithStrKey(instancesKey),
			registry.WithPrefix(),
			registry.WithCountOnly())
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
	err = serviceUtil.RefreshDependencyCache(ctx, tenant, ServiceId, service)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, refresh service dependency cache failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, "Refresh dependency cache failed."), err
	}

	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceIndexKey(consumer))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceAliasKey(consumer))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceKey(tenant, ServiceId))),
		registry.OpDel(registry.WithStrKey(
			util.StringJoin([]string{apt.GetServiceRuleRootKey(tenant), ServiceId, ""}, "/"))),
	}

	//删除依赖规则
	lock, err := mux.Lock(mux.GLOBAL_LOCK)
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, create lock failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	optsTmp, err := serviceUtil.DeleteDependencyForService(ctx, consumer, ServiceId)
	lock.Unlock()
	if err != nil {
		util.Logger().Errorf(err, "delete microservice failed, serviceId is %s: inner err, delete dependency failed.", ServiceId)
		return pb.CreateResponse(pb.Response_FAIL, err.Error()), err
	}
	opts = append(opts, optsTmp...)

	//删除黑白名单
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceRuleKey(tenant, ServiceId, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateRuleIndexKey(tenant, ServiceId, "", ""))))

	//删除shemas
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaKey(tenant, ServiceId, "")),
		registry.WithPrefix()))

	//删除tags
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceTagKey(tenant, ServiceId))))

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
	service, err := serviceUtil.GetService(ctx, tenant, in.ServiceId,
		serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))...)

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
	services, err := serviceUtil.GetAllServiceUtil(ctx,
		serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))...)
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
	service, err := serviceUtil.GetService(ctx, tenant, in.ServiceId)
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
	_, err = registry.GetRegisterCenter().Do(ctx,
		registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue(data))
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

		ids, err := serviceUtil.FindServiceIds(ctx, in.Version, &pb.MicroServiceKey{
			AppId:       in.AppId,
			ServiceName: in.ServiceName,
			Alias:       in.ServiceName,
			Version:     in.Version,
			Tenant:      tenant,
		}, serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))...)
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

		opts := serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))

		if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId, opts...) {
			util.Logger().Warnf(nil, "schema exist failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, in.SchemaId)
		exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key, opts...)
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
				rsp, err := apt.InstanceAPI.Register(ctx, req)
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

	// handle result
	var errMessages []string
	for createResp := range createRespChan {
		chanLen--
		if len(createResp.Message) != 0 {
			errMessages = append(errMessages, createResp.Message)
		}

		if 0 == chanLen {
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
		result.Response.Message = fmt.Sprintf("ErrMessage : %s", string(errMessage))
	} else {
		result.Response.Code = pb.Response_SUCCESS
	}

	util.Logger().Infof("CreateServiceEx, serviceid = %s, result = %s ", result.ServiceId, result.Response.Message)
	return result, nil
}

func (s *ServiceController) isCreateServiceEx(in *pb.CreateServiceRequest) bool {
	if len(in.Rules) == 0 && len(in.Tags) == 0 && len(in.Instances) == 0 {
		return false
	}
	return true
}

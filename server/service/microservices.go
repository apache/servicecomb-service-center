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
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type MicroServiceService struct {
	instanceService pb.SerivceInstanceCtrlServerEx
}

const (
	EXIST_TYPE_MICROSERVICE = "microservice"
	EXIST_TYPE_SCHEMA       = "schema"
)

func (s *MicroServiceService) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if in == nil || in.Service == nil {
		util.Logger().Errorf(nil, "create micro-service failed : param empty.")
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "request format invalid"),
		}, nil
	}

	//create service
	rsp, err := s.CreateServicePri(ctx, in)
	if err != nil || rsp.Response.Code != pb.Response_SUCCESS {
		return rsp, err
	}

	if s.isCreateServiceEx(in) == false {
		return rsp, err
	}

	//create tag,rule,instances
	return s.CreateServiceEx(ctx, in, rsp.ServiceId)
}

func (s *MicroServiceService) CreateServicePri(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	service := in.Service
	serviceFlag := util.StringJoin([]string{service.AppId, service.ServiceName, service.Version}, "/")

	serviceUtil.SetServiceDefaultValue(service)

	err := apt.Validate(service)
	if err != nil {
		util.Logger().Errorf(err, "create micro-service failed, %s: invalid parameters. operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
	}
	reporter, quotaErr := checkQuota(ctx, domainProject)
	if reporter != nil {
		defer reporter.Close()
	}
	if quotaErr != nil {
		util.Logger().Errorf(quotaErr, "create micro-service failed, %s: check service failed before create. operator: %s",
			serviceFlag, remoteIP)
		resp := &pb.CreateServiceResponse{
			Response: pb.CreateResponseWithSCErr(quotaErr),
		}
		if quotaErr.InternalError() {
			return resp, quotaErr
		}
		return resp, nil
	}

	// 产生全局service id
	serviceId := in.Service.ServiceId
	if len(serviceId) == 0 {
		serviceId = plugin.Plugins().UUID().GetServiceId()
	}
	service.ServiceId = serviceId
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	service.ModTimestamp = service.Timestamp

	data, err := json.Marshal(service)
	if err != nil {
		util.Logger().Errorf(err, "create micro-service failed, %s: json marshal service failed. operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Body error "+err.Error()),
		}, err
	}
	key := apt.GenerateServiceKey(domainProject, serviceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	index := apt.GenerateServiceIndexKey(serviceKey)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(serviceKey))

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithKey(keyBytes), registry.WithValue(data)),
		registry.OpPut(registry.WithKey(indexBytes), registry.WithStrValue(serviceId)),
	}
	uniqueCmpOpts := []registry.CompareOp{
		registry.OpCmp(registry.CmpVer(indexBytes), registry.CMP_EQUAL, 0),
		registry.OpCmp(registry.CmpVer(keyBytes), registry.CMP_EQUAL, 0),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, registry.OpPut(registry.WithKey(aliasBytes), registry.WithStrValue(serviceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			registry.OpCmp(registry.CmpVer(aliasBytes), registry.CMP_EQUAL, 0))
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts, uniqueCmpOpts, nil)
	if err != nil {
		util.Logger().Errorf(err, "create micro-service failed, %s: commit data into etcd failed. operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "commit operations failed."),
		}, err
	}
	if !resp.Succeeded {
		if s.isCreateServiceEx(in) == true {
			serviceIdInner, _ := serviceUtil.GetServiceId(ctx, serviceKey)
			util.Logger().Warnf(nil, "create micro-service failed, serviceId = %s , flag = %s: service already exists. operator: %s",
				serviceIdInner, serviceFlag, remoteIP)

			return &pb.CreateServiceResponse{
				Response:  pb.CreateResponse(pb.Response_SUCCESS, "register service successfully"),
				ServiceId: serviceIdInner,
			}, nil
		}

		util.Logger().Warnf(nil, "create micro-service failed, %s: service already exists. operator: %s",
			serviceFlag, remoteIP)

		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceAlreadyExists, "Service already exists."),
		}, nil
	}

	if reporter != nil {
		if err := reporter.ReportUsedQuota(ctx); err != nil {
			util.Logger().Errorf(err, "report used quota failed.")
		}
	}
	util.Logger().Infof("create micro-service successful, %s, serviceId: %s. operator: %s",
		serviceFlag, service.ServiceId, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Register service successfully."),
		ServiceId: serviceId,
	}, nil
}

func checkQuota(ctx context.Context, domainProject string) (quota.QuotaReporter, *scerr.Error) {
	if core.IsSCInstance(ctx) {
		util.Logger().Infof("it is service-center")
		return nil, nil
	}
	res := quota.NewApplyQuotaResource(quota.MicroServiceQuotaType, domainProject, "", 1)
	rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
	return rst.Reporter, rst.Err
}

func (s *MicroServiceService) DeleteServicePri(ctx context.Context, serviceId string, force bool) (*pb.Response, error) {
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceId == apt.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s", title, serviceId)
		return pb.CreateResponse(scerr.ErrInvalidParams, err.Error()), nil
	}

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s: get service failed.", title, serviceId)
		return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	if service == nil {
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s: service not exist.", title, serviceId)
		return pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, service)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			util.Logger().Errorf(err, "delete micro-service failed, serviceId is %s: inner err, get service dependency failed.", serviceId)
			return pb.CreateResponse(scerr.ErrInternal, "Get dependency info failed."), err
		}
		if len(services) > 1 || (len(services) == 1 && services[0] != serviceId) {
			util.Logger().Errorf(nil, "delete micro-service failed, serviceId is %s: can't delete, other services rely it.", serviceId)
			return pb.CreateResponse(scerr.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(domainProject, serviceId, "")
		rsp, err := store.Store().Instance().Search(ctx,
			registry.WithStrKey(instancesKey),
			registry.WithPrefix(),
			registry.WithCountOnly())
		if err != nil {
			util.Logger().Errorf(err, "delete micro-service failed, serviceId is %s: inner err,get instances failed.", serviceId)
			return pb.CreateResponse(scerr.ErrInternal, "Get instance failed."), err
		}

		if rsp.Count > 0 {
			util.Logger().Errorf(nil, "delete micro-service failed, serviceId is %s: can't delete, exist instance.", serviceId)
			return pb.CreateResponse(scerr.ErrDeployedInstance, "Can not delete this service, exist instance."), err
		}
	}

	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
		Alias:       service.Alias,
	}
	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceIndexKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceAliasKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceKey(domainProject, serviceId))),
		registry.OpDel(registry.WithStrKey(
			util.StringJoin([]string{apt.GetServiceRuleRootKey(domainProject), serviceId, ""}, "/"))),
	}

	//删除依赖规则
	optDeleteDep, err := serviceUtil.DeleteDependencyForDeleteService(domainProject, serviceId, serviceKey)
	if err != nil {
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s: inner err, delete dependency failed.", title, serviceId)
		return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除黑白名单
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceId, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateRuleIndexKey(domainProject, serviceId, "", ""))))

	//删除schemas
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaKey(domainProject, serviceId, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, "")),
		registry.WithPrefix()))

	//删除tags
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceTagKey(domainProject, serviceId))))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s: delete all instances failed.", title, serviceId)
		return pb.CreateResponse(scerr.ErrInternal, "Delete all instances failed for service."), err
	}

	err = backend.BatchCommit(ctx, opts)
	if err != nil {
		util.Logger().Errorf(err, "%s micro-service failed, serviceId is %s: commit data into etcd failed.", title, serviceId)
		return pb.CreateResponse(scerr.ErrUnavailableBackend, "Commit operations failed."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	util.Logger().Infof("%s micro-service successful: serviceId is %s, operator is %s.", title, serviceId, util.GetIPFromContext(ctx))
	return pb.CreateResponse(pb.Response_SUCCESS, "Unregister service successfully."), nil
}

func (s *MicroServiceService) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.Logger().Errorf(nil, "delete micro-service failed: service empty.")
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "delete micro-service failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	resp, err := s.DeleteServicePri(ctx, in.ServiceId, in.Force)

	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) DeleteServices(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	// 合法性检查
	if request == nil || request.ServiceIds == nil || len(request.ServiceIds) == 0 {
		return &pb.DelServicesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request param."),
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
			util.Logger().Warnf(nil, "delete micro-service %s , multiple.", serviceId)
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
			util.Logger().Errorf(err, "delete micro-service failed, serviceId is %s: invalid parameters.", in.ServiceId)
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
	delServiceRspInfo := make([]*pb.DelServicesRspInfo, 0, len(serviceRespChan))
	for serviceRespItem := range serviceRespChan {
		count++
		if len(serviceRespItem.ErrMessage) != 0 {
			responseCode = scerr.ErrInvalidParams
		}
		delServiceRspInfo = append(delServiceRspInfo, serviceRespItem)
		//结果收集over，关闭通道
		if count == nuoMultilCount {
			close(serviceRespChan)
		}
	}

	util.Logger().Infof("Batch DeleteServices, count is %s, serviceId = %v , result = %d, ", len(request.ServiceIds), request.ServiceIds, responseCode)

	resp := &pb.DelServicesResponse{
		Services: delServiceRspInfo,
	}
	if responseCode != pb.Response_SUCCESS {
		resp.Response = pb.CreateResponse(responseCode, "Delete services failed.")
	} else {
		resp.Response = pb.CreateResponse(responseCode, "Delete services successfully.")
	}
	return resp, nil
}

func (s *MicroServiceService) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "get micro-service failed, serviceId is %s: invalid parameters.",
			in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)
	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)

	if err != nil {
		util.Logger().Errorf(err, "get micro-service failed, serviceId is %s: inner err,get service failed.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get service file failed."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "get micro-service failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get service successfully."),
		Service:  service,
	}, nil
}

func (s *MicroServiceService) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	if in == nil {
		util.Logger().Errorf(nil, "get services failed: invalid params.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "request format invalid."),
		}, nil
	}
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		util.Logger().Errorf(err, "get services failed: inner err.")
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "get services successfully."),
		Services: services,
	}, nil
}

func (s *MicroServiceService) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || in.Properties == nil {
		util.Logger().Errorf(nil, "update service properties failed: invalid params.")
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "request format invalid."),
		}, nil
	}
	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: invalid parameters.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	key := apt.GenerateServiceKey(domainProject, in.ServiceId)
	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: query service failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "update service properties failed, serviceId is %s: service not exist.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
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
			Response: pb.CreateResponse(scerr.ErrInternal, "service file marshal error."),
		}, err
	}

	// Set key file
	_, err = backend.Registry().Do(ctx,
		registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(err, "update service properties failed, serviceId is %s: commit data into etcd failed.", in.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	util.Logger().Infof("update service properties successful: serviceId is %s.", in.ServiceId)
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "update service successfully."),
	}, nil
}

func (s *MicroServiceService) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	if in == nil {
		util.Logger().Errorf(nil, "exist failed: invalid params")
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "request format invalid"),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	switch in.Type {
	case EXIST_TYPE_MICROSERVICE:
		if len(in.AppId) == 0 || len(in.ServiceName) == 0 || len(in.Version) == 0 {
			util.Logger().Errorf(nil, "micro-service exist failed: invalid params.")
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "invalid request."),
			}, nil
		}
		err := apt.GetMSExistsReqValidator.Validate(in)
		serviceFlag := util.StringJoin([]string{in.AppId, in.ServiceName, in.Version}, "/")
		if err != nil {
			util.Logger().Errorf(err, "micro-service exist failed, service %s: invalid params.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
			}, nil
		}

		ids, err := serviceUtil.FindServiceIds(ctx, in.Version, &pb.MicroServiceKey{
			Environment: in.Environment,
			AppId:       in.AppId,
			ServiceName: in.ServiceName,
			Alias:       in.ServiceName,
			Version:     in.Version,
			Tenant:      domainProject,
		})
		if err != nil {
			util.Logger().Errorf(err, "micro-service exist failed, service %s: find serviceIds failed.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "get service file failed."),
			}, err
		}
		if len(ids) <= 0 {
			util.Logger().Infof("micro-service exist failed, service %s: service not exist.", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.Response_SUCCESS, "get service id successfully."),
			ServiceId: ids[0], // 约定多个时，取较新版本
		}, nil
	case EXIST_TYPE_SCHEMA:
		if len(in.SchemaId) == 0 || len(in.ServiceId) == 0 {
			util.Logger().Errorf(nil, "schema exist failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "invalid request."),
			}, nil
		}

		err := apt.GetSchemaExistsReqValidator.Validate(in)
		if err != nil {
			util.Logger().Errorf(err, "schema exist failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
			}, nil
		}

		if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
			util.Logger().Warnf(nil, "schema exist failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, in.SchemaId)
		exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
		if err != nil {
			util.Logger().Errorf(err, "schema exist failed, serviceId %s, schemaId %s: get schema failed.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if !exist {
			util.Logger().Infof("schema exist failed, serviceId %s, schemaId %s: schema not exist.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "schema does not exist."),
			}, nil
		}
		schemaSummary, err := getSchemaSummary(ctx, domainProject, in.ServiceId, in.SchemaId)
		if err != nil {
			util.Logger().Errorf(err, "schema exist failed, serviceId %s, schemaId %s: get schema summary failed.", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Schema exist."),
			SchemaId: in.SchemaId,
			Summary:  schemaSummary,
		}, nil
	default:
		util.Logger().Warnf(nil, "unexpected type '%s' for query.", in.Type)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Only micro-service and schema can be used as type."),
		}, nil
	}
}

func (s *MicroServiceService) CreateServiceEx(ctx context.Context, in *pb.CreateServiceRequest, serviceId string) (*pb.CreateServiceResponse, error) {
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
				rsp, err := s.instanceService.Register(ctx, req)
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
		result.Response.Code = scerr.ErrInvalidParams
		errMessage, err := json.Marshal(errMessages)
		if err != nil {
			result.Response.Message = "marshal errMessages error"
			util.Logger().Error("marshal errMessages error", err)
			return result, nil
		}
		result.Response.Message = fmt.Sprintf("errMessage : %s", string(errMessage))
	} else {
		result.Response.Code = pb.Response_SUCCESS
	}

	util.Logger().Infof("createServiceEx, serviceId = %s, result = %s ", result.ServiceId, result.Response.Message)
	return result, nil
}

func (s *MicroServiceService) isCreateServiceEx(in *pb.CreateServiceRequest) bool {
	if len(in.Rules) == 0 && len(in.Tags) == 0 && len(in.Instances) == 0 {
		return false
	}
	return true
}

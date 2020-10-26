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
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"

	"context"
)

type MicroServiceService struct {
	// schemaEditable determines whether schema modification is allowed for
	// services in production environment
	schemaEditable  bool
	instanceService proto.ServiceInstanceCtrlServerEx
}

const (
	ExistTypeMicroservice = "microservice"
	ExistTypeSchema       = "schema"
)

func (s *MicroServiceService) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if in == nil || in.Service == nil {
		log.Errorf(nil, "create micro-service failed: request body is empty")
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Request body is empty"),
		}, nil
	}

	//create service
	rsp, err := s.CreateServicePri(ctx, in)
	if err != nil || rsp.Response.GetCode() != proto.Response_SUCCESS {
		return rsp, err
	}

	if !s.isCreateServiceEx(in) {
		return rsp, err
	}

	//create tag,rule,instances
	return s.CreateServiceEx(ctx, in, rsp.ServiceId)
}

func (s *MicroServiceService) CreateServicePri(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	service := in.Service
	serviceFlag := util.StringJoin([]string{
		service.Environment, service.AppId, service.ServiceName, service.Version}, "/")

	serviceUtil.SetServiceDefaultValue(service)

	err := Validate(in)
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
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
	reporter := checkQuota(ctx, domainProject)
	defer reporter.Close(ctx)

	if reporter != nil && reporter.Err != nil {
		log.Errorf(reporter.Err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		resp := &pb.CreateServiceResponse{
			Response: proto.CreateResponseWithSCErr(reporter.Err),
		}
		if reporter.Err.InternalError() {
			return resp, reporter.Err
		}
		return resp, nil
	}

	index := apt.GenerateServiceIndexKey(serviceKey)

	// 产生全局service id
	requestServiceID := service.ServiceId
	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, index)
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	service.ModTimestamp = service.Timestamp

	data, err := json.Marshal(service)
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, json marshal service failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	key := apt.GenerateServiceKey(domainProject, service.ServiceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(serviceKey))

	opts := []client.PluginOp{
		client.OpPut(client.WithKey(keyBytes), client.WithValue(data)),
		client.OpPut(client.WithKey(indexBytes), client.WithStrValue(service.ServiceId)),
	}
	uniqueCmpOpts := []client.CompareOp{
		client.OpCmp(client.CmpVer(indexBytes), client.CmpEqual, 0),
		client.OpCmp(client.CmpVer(keyBytes), client.CmpEqual, 0),
	}
	failOpts := []client.PluginOp{
		client.OpGet(client.WithKey(indexBytes)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, client.OpPut(client.WithKey(aliasBytes), client.WithStrValue(service.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			client.OpCmp(client.CmpVer(aliasBytes), client.CmpEqual, 0))
		failOpts = append(failOpts, client.OpGet(client.WithKey(aliasBytes)))
	}

	resp, err := client.Instance().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		if len(requestServiceID) != 0 {
			if len(resp.Kvs) == 0 ||
				requestServiceID != util.BytesToStringWithNoCopy(resp.Kvs[0].Value) {
				log.Warnf("create micro-service[%s] failed, service already exists, operator: %s",
					serviceFlag, remoteIP)
				return &pb.CreateServiceResponse{
					Response: proto.CreateResponse(scerr.ErrServiceAlreadyExists,
						"ServiceID conflict or found the same service with different id."),
				}, nil
			}
		}

		if len(resp.Kvs) == 0 {
			// internal error?
			log.Errorf(nil, "create micro-service[%s] failed, unexpected txn response, operator: %s",
				serviceFlag, remoteIP)
			return &pb.CreateServiceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, "Unexpected txn response."),
			}, nil
		}

		serviceIDInner := util.BytesToStringWithNoCopy(resp.Kvs[0].Value)
		log.Warnf("create micro-service[%s][%s] failed, service already exists, operator: %s",
			serviceIDInner, serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response:  proto.CreateResponse(proto.Response_SUCCESS, "register service successfully"),
			ServiceId: serviceIDInner,
		}, nil
	}

	if err := reporter.ReportUsedQuota(ctx); err != nil {
		log.Errorf(err, "report the used quota failed")
	}

	log.Infof("create micro-service[%s][%s] successfully, operator: %s",
		service.ServiceId, serviceFlag, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "Register service successfully."),
		ServiceId: service.ServiceId,
	}, nil
}

func checkQuota(ctx context.Context, domainProject string) *quota.ApplyQuotaResult {
	if core.IsSCInstance(ctx) {
		log.Debugf("register service-center, skip quota check")
		return nil
	}
	res := quota.NewApplyQuotaResource(quota.MicroServiceQuotaType, domainProject, "", 1)
	rst := quota.Apply(ctx, res)
	return rst
}

func (s *MicroServiceService) DeleteServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInvalidParams, err.Error()), nil
	}

	service, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	if service == nil {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, service)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Errorf(nil, "delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP)
			return proto.CreateResponse(scerr.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := kv.Store().Instance().Search(ctx,
			client.WithStrKey(instancesKey),
			client.WithPrefix(),
			client.WithCountOnly())
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
		}

		if rsp.Count > 0 {
			log.Errorf(nil, "delete micro-service[%s] failed, service deployed instances[%s], operator: %s",
				serviceID, rsp.Count, remoteIP)
			return proto.CreateResponse(scerr.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}
	}

	serviceIDKey := apt.GenerateServiceKey(domainProject, serviceID)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
		Alias:       service.Alias,
	}
	opts := []client.PluginOp{
		client.OpDel(client.WithStrKey(apt.GenerateServiceIndexKey(serviceKey))),
		client.OpDel(client.WithStrKey(apt.GenerateServiceAliasKey(serviceKey))),
		client.OpDel(client.WithStrKey(serviceIDKey)),
	}

	//删除依赖规则
	optDeleteDep, err := serviceUtil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除黑白名单
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(client.WithStrKey(
		util.StringJoin([]string{apt.GetServiceRuleIndexRootKey(domainProject), serviceID, ""}, "/")),
		client.WithPrefix()))

	//删除schemas
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除tags
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateInstanceKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(apt.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(serviceIDKey)),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}
	if !resp.Succeeded {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	log.Infof("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP)
	return proto.CreateResponse(proto.Response_SUCCESS, "Unregister service successfully."), nil
}

func (s *MicroServiceService) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "delete micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.DeleteServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	resp, err := s.DeleteServicePri(ctx, in.ServiceId, in.Force)

	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) DeleteServices(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	// 合法性检查
	if len(request.ServiceIds) == 0 {
		log.Errorf(nil, "delete all micro-services failed, 'serviceIDs' is empty, operator: %s", remoteIP)
		return &pb.DelServicesResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "'serviceIDs' is empty"),
			Services: nil,
		}, nil
	}

	existFlag := map[string]bool{}
	nuoMultilCount := 0
	// 批量删除服务
	serviceRespChan := make(chan *pb.DelServicesRspInfo, len(request.ServiceIds))
	for _, serviceID := range request.ServiceIds {
		//ServiceId重复性检查
		if _, ok := existFlag[serviceID]; ok {
			log.Warnf("duplicate micro-service[%s] serviceID, operator: %s", serviceID, remoteIP)
			continue
		} else {
			existFlag[serviceID] = true
			nuoMultilCount++
		}

		//检查服务ID合法性
		in := &pb.DeleteServiceRequest{
			ServiceId: serviceID,
			Force:     request.Force,
		}
		err := Validate(in)
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP)
			serviceRespChan <- &pb.DelServicesRspInfo{
				ServiceId:  serviceID,
				ErrMessage: err.Error(),
			}
			continue
		}

		//执行删除服务操作
		gopool.Go(s.getDeleteServiceFunc(ctx, serviceID, request.Force, serviceRespChan))
	}

	//获取批量删除服务的结果
	count := 0
	responseCode := proto.Response_SUCCESS
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

	log.Infof("Batch delete micro-services by serviceIDs[%d]: %v, result code: %d, operator: %s",
		len(request.ServiceIds), request.ServiceIds, responseCode, remoteIP)

	resp := &pb.DelServicesResponse{
		Services: delServiceRspInfo,
	}
	if responseCode != proto.Response_SUCCESS {
		resp.Response = proto.CreateResponse(responseCode, "Delete services failed.")
	} else {
		resp.Response = proto.CreateResponse(responseCode, "Delete services successfully.")
	}
	return resp, nil
}

func (s *MicroServiceService) getDeleteServiceFunc(ctx context.Context, serviceID string, force bool, serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		resp, err := s.DeleteServicePri(ctx, serviceID, force)
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.GetCode() != proto.Response_SUCCESS {
			serviceRst.ErrMessage = resp.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}

func (s *MicroServiceService) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)
	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)

	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed, get service file failed", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if service == nil {
		log.Errorf(nil, "get micro-service[%s] failed, service does not exist", in.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service successfully."),
		Service:  service,
	}, nil
}

func (s *MicroServiceService) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all services successfully."),
		Services: services,
	}, nil
}

func (s *MicroServiceService) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	key := apt.GenerateServiceKey(domainProject, in.ServiceId)
	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, get service file failed, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if service == nil {
		log.Errorf(nil, "update service[%s] properties failed, service does not exist, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	copyServiceRef := *service
	copyServiceRef.Properties = in.Properties
	copyServiceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(copyServiceRef)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, json marshal service failed, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	// Set key file
	resp, err := client.Instance().TxnWithCmp(ctx,
		[]client.PluginOp{client.OpPut(client.WithStrKey(key), client.WithValue(data))},
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(key)),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "update service[%s] properties failed, service does not exist, operator: %s",
			in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("update service[%s] properties successfully, operator: %s", in.ServiceId, remoteIP)
	return &pb.UpdateServicePropsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "update service successfully."),
	}, nil
}

func (s *MicroServiceService) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	switch in.Type {
	case ExistTypeMicroservice:
		err := ExistenceReqValidator().Validate(in)
		serviceFlag := util.StringJoin([]string{in.Environment, in.AppId, in.ServiceName, in.Version}, "/")
		if err != nil {
			log.Errorf(err, "micro-service[%s] exist failed", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
			}, nil
		}

		ids, exist, err := serviceUtil.FindServiceIds(ctx, in.Version, &pb.MicroServiceKey{
			Environment: in.Environment,
			AppId:       in.AppId,
			ServiceName: in.ServiceName,
			Alias:       in.ServiceName,
			Version:     in.Version,
			Tenant:      domainProject,
		})
		if err != nil {
			log.Errorf(err, "micro-service[%s] exist failed, find serviceIDs failed", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if !exist {
			log.Infof("micro-service[%s] exist failed, service does not exist", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists, serviceFlag+" does not exist."),
			}, nil
		}
		if len(ids) == 0 {
			log.Infof("micro-service[%s] exist failed, version mismatch", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrServiceVersionNotExists, serviceFlag+" version mismatch."),
			}, nil
		}
		return &pb.GetExistenceResponse{
			Response:  proto.CreateResponse(proto.Response_SUCCESS, "get service id successfully."),
			ServiceId: ids[0], // 约定多个时，取较新版本
		}, nil
	case ExistTypeSchema:
		err := GetSchemaReqValidator().Validate(in)
		if err != nil {
			log.Errorf(err, "schema[%s/%s] exist failed", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
			}, nil
		}

		if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
			log.Warnf("schema[%s/%s] exist failed, service does not exist", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
			}, nil
		}

		key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, in.SchemaId)
		exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
		if err != nil {
			log.Errorf(err, "schema[%s/%s] exist failed, get schema failed", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		if !exist {
			log.Infof("schema[%s/%s] exist failed, schema does not exist", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrSchemaNotExists, "schema does not exist."),
			}, nil
		}
		schemaSummary, err := getSchemaSummary(ctx, domainProject, in.ServiceId, in.SchemaId)
		if err != nil {
			log.Errorf(err, "schema[%s/%s] exist failed, get schema summary failed", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(proto.Response_SUCCESS, "Schema exist."),
			SchemaId: in.SchemaId,
			Summary:  schemaSummary,
		}, nil
	default:
		log.Warnf("unexpected type '%s' for existence query.", in.Type)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Only micro-service and schema can be used as type."),
		}, nil
	}
}

func (s *MicroServiceService) CreateServiceEx(ctx context.Context, in *pb.CreateServiceRequest, serviceID string) (*pb.CreateServiceResponse, error) {
	result := &pb.CreateServiceResponse{
		ServiceId: serviceID,
		Response:  &pb.Response{},
	}
	var chanLen int = 0
	createRespChan := make(chan *pb.Response, 10)
	//create rules
	if in.Rules != nil && len(in.Rules) != 0 {
		chanLen++
		gopool.Go(func(_ context.Context) {
			req := &pb.AddServiceRulesRequest{
				ServiceId: serviceID,
				Rules:     in.Rules,
			}
			chanRsp := &pb.Response{}
			rsp, err := s.AddRule(ctx, req)
			if err != nil {
				chanRsp.Message = err.Error()
			}

			if rsp.Response.GetCode() != proto.Response_SUCCESS {
				chanRsp.Message = rsp.Response.GetMessage()
			}
			createRespChan <- chanRsp
		})
	}
	//create tags
	if in.Tags != nil && len(in.Tags) != 0 {
		chanLen++
		gopool.Go(func(_ context.Context) {
			req := &pb.AddServiceTagsRequest{
				ServiceId: serviceID,
				Tags:      in.Tags,
			}
			chanRsp := &pb.Response{}
			rsp, err := s.AddTags(ctx, req)
			if err != nil {
				chanRsp.Message = err.Error()
			}

			if rsp.Response.GetCode() != proto.Response_SUCCESS {
				chanRsp.Message = rsp.Response.GetMessage()
			}
			createRespChan <- chanRsp
		})
	}
	// create instance
	if in.Instances != nil && len(in.Instances) != 0 {
		chanLen++
		gopool.Go(func(_ context.Context) {
			chanRsp := &pb.Response{}
			for _, ins := range in.Instances {
				req := &pb.RegisterInstanceRequest{
					Instance: ins,
				}
				req.Instance.ServiceId = serviceID
				rsp, err := s.instanceService.Register(ctx, req)
				if err != nil {
					chanRsp.Message += fmt.Sprintf("{instance:%v,result:%s}", ins.Endpoints, err.Error())
				}
				if rsp.Response.GetCode() != proto.Response_SUCCESS {
					chanRsp.Message += fmt.Sprintf("{instance:%v,result:%s}", ins.Endpoints, rsp.Response.GetMessage())
				}
				createRespChan <- chanRsp
			}
		})
	}

	// handle result
	var errMessages []string
	for createResp := range createRespChan {
		chanLen--
		if len(createResp.GetMessage()) != 0 {
			errMessages = append(errMessages, createResp.GetMessage())
		}

		if 0 == chanLen {
			close(createRespChan)
		}
	}

	if len(errMessages) != 0 {
		result.Response.Code = scerr.ErrInvalidParams
		result.Response.Message = fmt.Sprintf("errMessages: %v", errMessages)
	} else {
		result.Response.Code = proto.Response_SUCCESS
	}

	log.Infof("createServiceEx, serviceID: %s, result code: %s, operator: %s",
		result.ServiceId, result.Response.GetMessage(), util.GetIPFromContext(ctx))
	return result, nil
}

func (s *MicroServiceService) isCreateServiceEx(in *pb.CreateServiceRequest) bool {
	if len(in.Rules) == 0 && len(in.Tags) == 0 && len(in.Instances) == 0 {
		return false
	}
	return true
}

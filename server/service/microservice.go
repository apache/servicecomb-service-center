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
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/proto"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"

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
			Response: pb.CreateResponse(pb.ErrInvalidParams, "Request body is empty"),
		}, nil
	}

	//create service
	rsp, err := s.CreateServicePri(ctx, in)
	if err != nil || rsp.Response.GetCode() != pb.ResponseSuccess {
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
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}
	return datasource.Instance().RegisterService(ctx, in)
}

func (s *MicroServiceService) DeleteServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == core.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrInvalidParams, err.Error()), nil
	}

	service, err := serviceUtil.GetService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}

	if service == nil {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."), nil
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, service)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Errorf(nil, "delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP)
			return pb.CreateResponse(pb.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := path.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := kv.Store().Instance().Search(ctx,
			client.WithStrKey(instancesKey),
			client.WithPrefix(),
			client.WithCountOnly())
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP)
			return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
		}

		if rsp.Count > 0 {
			log.Errorf(nil, "delete micro-service[%s] failed, service deployed instances[%s], operator: %s",
				serviceID, rsp.Count, remoteIP)
			return pb.CreateResponse(pb.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}
	}

	serviceIDKey := path.GenerateServiceKey(domainProject, serviceID)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
		Alias:       service.Alias,
	}
	opts := []client.PluginOp{
		client.OpDel(client.WithStrKey(path.GenerateServiceIndexKey(serviceKey))),
		client.OpDel(client.WithStrKey(path.GenerateServiceAliasKey(serviceKey))),
		client.OpDel(client.WithStrKey(serviceIDKey)),
	}

	//删除依赖规则
	optDeleteDep, err := serviceUtil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除黑白名单
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateServiceRuleKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(client.WithStrKey(
		util.StringJoin([]string{path.GetServiceRuleIndexRootKey(domainProject), serviceID, ""}, "/")),
		client.WithPrefix()))

	//删除schemas
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除tags
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateInstanceKey(domainProject, serviceID, "")),
		client.WithPrefix()))
	opts = append(opts, client.OpDel(
		client.WithStrKey(path.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		client.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := client.Instance().TxnWithCmp(ctx, opts,
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(serviceIDKey)),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()), err
	}
	if !resp.Succeeded {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	log.Infof("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP)
	return pb.CreateResponse(pb.ResponseSuccess, "Unregister service successfully."), nil
}

func (s *MicroServiceService) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "delete micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UnregisterService(ctx, in)
}

func (s *MicroServiceService) DeleteServices(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	// 合法性检查
	if len(request.ServiceIds) == 0 {
		log.Errorf(nil, "delete all micro-services failed, 'serviceIDs' is empty, operator: %s", remoteIP)
		return &pb.DelServicesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, "'serviceIDs' is empty"),
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
	responseCode := pb.ResponseSuccess
	delServiceRspInfo := make([]*pb.DelServicesRspInfo, 0, len(serviceRespChan))
	for serviceRespItem := range serviceRespChan {
		count++
		if len(serviceRespItem.ErrMessage) != 0 {
			responseCode = pb.ErrInvalidParams
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
	if responseCode != pb.ResponseSuccess {
		resp.Response = pb.CreateResponse(responseCode, "Delete services failed.")
	} else {
		resp.Response = pb.CreateResponse(responseCode, "Delete services successfully.")
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
		} else if resp.GetCode() != pb.ResponseSuccess {
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
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetService(ctx, in)
}

func (s *MicroServiceService) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	return datasource.Instance().GetServices(ctx, in)
}

func (s *MicroServiceService) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	err := Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "update service[%s] properties failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UpdateService(ctx, in)
}

func (s *MicroServiceService) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	switch in.Type {
	case ExistTypeMicroservice:
		err := ExistenceReqValidator().Validate(in)
		if err != nil {
			serviceFlag := util.StringJoin([]string{in.Environment, in.AppId, in.ServiceName, in.Version}, "/")
			log.Errorf(err, "micro-service[%s] exist failed", serviceFlag)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
			}, nil
		}

		return datasource.Instance().ExistService(ctx, in)
	case ExistTypeSchema:
		err := GetSchemaReqValidator().Validate(in)
		if err != nil {
			log.Errorf(err, "schema[%s/%s] exist failed", in.ServiceId, in.SchemaId)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
			}, nil
		}

		return datasource.Instance().ExistSchema(ctx, in)
	default:
		log.Warnf("unexpected type '%s' for existence query.", in.Type)
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, "Only micro-service and schema can be used as type."),
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

			if rsp.Response.GetCode() != pb.ResponseSuccess {
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

			if rsp.Response.GetCode() != pb.ResponseSuccess {
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
				if rsp.Response.GetCode() != pb.ResponseSuccess {
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
		result.Response.Code = pb.ErrInvalidParams
		result.Response.Message = fmt.Sprintf("errMessages: %v", errMessages)
	} else {
		result.Response.Code = pb.ResponseSuccess
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

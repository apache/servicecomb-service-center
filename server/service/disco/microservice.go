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

package disco

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
)

type MicroServiceService struct {
}

func (s *MicroServiceService) Create(ctx context.Context, in *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if in == nil || in.Service == nil {
		log.Error("create micro-service failed: request body is empty", nil)
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
	datasource.SetServiceDefaultValue(service)
	if err := validator.Validate(in); err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP), err)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}
	if quotaErr := checkServiceQuota(ctx); quotaErr != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP), quotaErr)
		response, err := datasource.WrapErrResponse(quotaErr)
		return &pb.CreateServiceResponse{
			Response: response,
		}, err
	}

	return RegisterService(ctx, in)
}

func (s *MicroServiceService) Delete(ctx context.Context, in *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("delete micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP), err)
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return UnregisterService(ctx, in)
}

func (s *MicroServiceService) DeleteServices(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	// 合法性检查
	if len(request.ServiceIds) == 0 {
		log.Error(fmt.Sprintf("delete all micro-services failed, 'serviceIDs' is empty, operator: %s", remoteIP), nil)
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
			log.Warn(fmt.Sprintf("duplicate micro-service[%s] serviceID, operator: %s", serviceID, remoteIP))
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
		err := validator.Validate(in)
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP), err)
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

	log.Info(fmt.Sprintf("Batch delete micro-services by serviceIDs[%d]: %v, result code: %d, operator: %s",
		len(request.ServiceIds), request.ServiceIds, responseCode, remoteIP))

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
		resp, err := UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceID,
			Force:     force,
		})
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.Response.GetCode() != pb.ResponseSuccess {
			serviceRst.ErrMessage = resp.Response.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}

func (s *MicroServiceService) GetOne(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	service, err := GetService(ctx, in)
	if err != nil {
		log.Error(fmt.Sprintf("get micro-service[%s] failed", in.ServiceId), err)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, err
	}
	return &pb.GetServiceResponse{Response: pb.CreateResponse(pb.ResponseSuccess, ""), Service: service}, nil
}

func (s *MicroServiceService) GetServices(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	resp, err := datasource.GetMetadataManager().GetServices(ctx, in)
	if err == nil && len(resp.Services) > 0 {
		resp.Services = datasource.RemoveGlobalServices(in.WithShared, util.ParseDomainProject(ctx), resp.Services)
	}
	return resp, err
}

func (s *MicroServiceService) UpdateProperties(ctx context.Context, in *pb.UpdateServicePropsRequest) (*pb.UpdateServicePropsResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("update service[%s] properties failed, operator: %s", in.ServiceId, remoteIP), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().UpdateService(ctx, in)
}

func (s *MicroServiceService) Exist(ctx context.Context, in *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	switch in.Type {
	case datasource.ExistTypeMicroservice:
		err := validator.ExistenceReqValidator().Validate(in)
		if err != nil {
			serviceFlag := util.StringJoin([]string{in.Environment, in.AppId, in.ServiceName, in.Version}, "/")
			log.Error(fmt.Sprintf("micro-service[%s] exist failed", serviceFlag), err)
			return &pb.GetExistenceResponse{
				Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
			}, nil
		}

		return datasource.GetMetadataManager().ExistService(ctx, in)
	case datasource.ExistTypeSchema:
		err := validator.GetSchemaReqValidator().Validate(in)
		if err != nil {
			log.Error(fmt.Sprintf("schema[%s/%s] exist failed", in.ServiceId, in.SchemaId), err)
			return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
		}
		schema, err := ExistSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: in.ServiceId,
			SchemaId:  in.SchemaId,
		})
		if err != nil {
			return nil, err
		}
		return &pb.GetExistenceResponse{
			ServiceId: in.ServiceId,
			SchemaId:  schema.SchemaId,
			Summary:   schema.Summary,
		}, nil
	default:
		log.Warn(fmt.Sprintf("unexpected type '%s' for existence query.", in.Type))
		return nil, pb.NewError(pb.ErrInvalidParams, "Only micro-service and schema can be used as type.")
	}
}

func (s *MicroServiceService) CreateServiceEx(ctx context.Context, in *pb.CreateServiceRequest, serviceID string) (*pb.CreateServiceResponse, error) {
	result := &pb.CreateServiceResponse{
		ServiceId: serviceID,
		Response:  &pb.Response{},
	}
	var chanLen = 0
	createRespChan := make(chan *pb.Response, 10)
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
				rsp, err := RegisterInstance(ctx, req)
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

	log.Info(fmt.Sprintf("createServiceEx, serviceID: %s, result code: %s, operator: %s",
		result.ServiceId, result.Response.GetMessage(), util.GetIPFromContext(ctx)))
	return result, nil
}

func (s *MicroServiceService) isCreateServiceEx(in *pb.CreateServiceRequest) bool {
	if len(in.Rules) == 0 && len(in.Tags) == 0 && len(in.Instances) == 0 {
		return false
	}
	return true
}

func checkServiceQuota(ctx context.Context) error {
	if core.IsSCInstance(ctx) {
		log.Debug("skip quota check")
		return nil
	}
	return quotasvc.ApplyService(ctx, 1)
}

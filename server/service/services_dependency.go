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
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/mux"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strings"
)

func (s *ServiceController) CreateDependenciesForMircServices(ctx context.Context, in *pb.CreateDependenciesRequest) (*pb.CreateDependenciesResponse, error) {
	dependencyInfos := in.Dependencies
	if dependencyInfos == nil {
		return dependency.BadParamsResponse("Invalid request body."), nil
	}
	tenant := util.ParseTenantProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerInfo := pb.TransferToMicroServiceKeys([]*pb.DependencyMircroService{dependencyInfo.Consumer}, tenant)[0]
		providersInfo := pb.TransferToMicroServiceKeys(dependencyInfo.Providers, tenant)
		consumerFlag := strings.Join([]string{consumerInfo.AppId, consumerInfo.ServiceName, consumerInfo.Version}, "/")
		rsp := dependency.ParamsChecker(ctx, consumerInfo, providersInfo, tenant)
		if rsp != nil {
			util.LOGGER.Errorf(nil, "create dependency faild, conusmer %s: invalid params.%s", consumerFlag, rsp.Response.Message)
			return rsp, nil
		}

		consumerId, err := ms.GetServiceId(ctx, consumerInfo)
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
		err = dependency.UpdateServiceForAddDependency(ctx, consumerId, dependencyInfo.Providers, tenant)
		if err != nil {
			util.LOGGER.Errorf(err, "create dependency faild, conusmer %s: Update service failed.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}

		//建立依赖规则，用于维护依赖关系
		lock, err := mux.Lock(mux.GLOBAL_LOCK)
		if err != nil {
			util.LOGGER.Errorf(err, "create dependency faild, conusmer %s: create lock failed.", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}

		err = dependency.CreateDependencyRule(ctx, consumerId, consumerInfo, providersInfo)
		lock.Unlock()

		if err != nil {
			util.LOGGER.Errorf(err, "create dependency rule failed: consumer %s", consumerFlag)
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		//更新consumer作为provider的依赖关系
		err = dependency.UpdateAsProviderDependency(ctx, consumerId, consumerInfo)
		if err != nil {
			util.LOGGER.Errorf(nil, "Dependency update,as provider,update it's consumer list failed. %s", err.Error())
			return &pb.CreateDependenciesResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
			}, err
		}
		//更新consumer作为consumer的依赖关系
		err = dependency.UpdateAsConsumerDependency(ctx, consumerId, providersInfo, tenant)
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

func (s *ServiceController) GetProviderDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	err := apt.Validate(in)
	if err != nil {
		util.LOGGER.Errorf(err, "GetProviderDependencies failed for validating parameters failed.")
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	providerId := in.ServiceId
	tenant := util.ParseTenantProject(ctx)
	if !ms.ServiceExist(ctx, tenant, providerId) {
		util.LOGGER.Errorf(nil, "GetProviderDependencies failed, providerId is %s: service not exist.",
			providerId)
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This provider does not exist."),
		}, nil
	}
	keyProDependency := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	services, err := dependency.GetDependencies(ctx, keyProDependency, tenant)
	if err != nil {
		util.LOGGER.Errorf(err, "GetProviderDependencies failed, providerId is %s.", providerId)
		return &pb.GetProDependenciesResponse{
			Response:  pb.CreateResponse(pb.Response_FAIL, err.Error()),
			Consumers: nil,
		}, err
	}
	util.LOGGER.Infof("GetProviderDependencies successfully, providerId is %s.", providerId)
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
	tenant := util.ParseTenantProject(ctx)
	if !ms.ServiceExist(ctx, tenant, consumerId) {
		util.LOGGER.Errorf(nil, "GetConsumerDependencies failed, consumerId is %s: service not exist.",
			consumerId)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "This consumer does not exist."),
		}, nil
	}
	keyConDependency := apt.GenerateConsumerDependencyKey(tenant, consumerId, "")
	services, err := dependency.GetDependencies(ctx, keyConDependency, tenant)
	if err != nil {
		util.LOGGER.Errorf(err, "GetConsumerDependencies failed, consumerId is %s.", consumerId)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}
	util.LOGGER.Infof("GetConsumerDependencies successfully, consumerId is %s.", consumerId)
	return &pb.GetConDependenciesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get all providers successful."),
		Providers: services,
	}, nil
}

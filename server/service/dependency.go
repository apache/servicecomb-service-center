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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

func (s *MicroServiceService) AddDependenciesForMicroServices(ctx context.Context, in *pb.AddDependenciesRequest) (*pb.AddDependenciesResponse, error) {
	if err := Validate(in); err != nil {
		return &pb.AddDependenciesResponse{
			Response: serviceUtil.BadParamsResponse(err.Error()).Response,
		}, nil
	}

	resp, err := s.AddOrUpdateDependencies(ctx, in.Dependencies, false)
	return &pb.AddDependenciesResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) CreateDependenciesForMicroServices(ctx context.Context, in *pb.CreateDependenciesRequest) (*pb.CreateDependenciesResponse, error) {
	if err := Validate(in); err != nil {
		return &pb.CreateDependenciesResponse{
			Response: serviceUtil.BadParamsResponse(err.Error()).Response,
		}, nil
	}

	resp, err := s.AddOrUpdateDependencies(ctx, in.Dependencies, true)
	return &pb.CreateDependenciesResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) AddOrUpdateDependencies(ctx context.Context, dependencyInfos []*pb.ConsumerDependency, override bool) (*pb.Response, error) {
	opts := make([]registry.PluginOp, 0, len(dependencyInfos))
	domainProject := util.ParseDomainProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerFlag := util.StringJoin([]string{dependencyInfo.Consumer.AppId, dependencyInfo.Consumer.ServiceName, dependencyInfo.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		rsp := serviceUtil.ParamsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer is %s, %s",
				override, consumerFlag, rsp.Response.Message)
			return rsp.Response, nil
		}

		consumerId, err := serviceUtil.GetServiceId(ctx, consumerInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, get consumer %s id failed",
				override, consumerFlag)
			return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
		if len(consumerId) == 0 {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer %s does not exist.",
				override, consumerFlag)
			return pb.CreateResponse(scerr.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag)), nil
		}

		dependencyInfo.Override = override
		data, err := json.Marshal(dependencyInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, marshal consumer %s dependency failed",
				override, consumerFlag)
			return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
		}

		id := apt.DEPS_QUEUE_UUID
		if !override {
			id = util.GenerateUuid()
		}
		key := apt.GenerateConsumerDependencyQueueKey(domainProject, consumerId, id)
		opts = append(opts, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)))
	}

	err := backend.BatchCommit(ctx, opts)
	if err != nil {
		log.Errorf(err, "put request into dependency queue failed, override: %t, %v", override, dependencyInfos)
		return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	log.Infof("put request into dependency queue successfully, override: %t, %v, from remote %s",
		override, dependencyInfos, util.GetIPFromContext(ctx))
	return pb.CreateResponse(pb.Response_SUCCESS, "Create dependency successfully."), nil
}

func (s *MicroServiceService) GetProviderDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "GetProviderDependencies failed for validating parameters failed.")
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)
	providerServiceId := in.ServiceId

	provider, err := serviceUtil.GetService(ctx, domainProject, providerServiceId)
	if err != nil {
		log.Errorf(err, "GetProviderDependencies failed, %s.", providerServiceId)
		return nil, err
	}
	if provider == nil {
		log.Errorf(err, "GetProviderDependencies failed for provider does not exist, %s.", providerServiceId)
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Provider does not exist"),
		}, nil
	}

	dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, provider)
	services, err := dr.GetDependencyConsumers(toDependencyFilterOptions(in)...)
	if err != nil {
		log.Errorf(err, "GetProviderDependencies failed.")
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetProDependenciesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get all consumers successful."),
		Consumers: services,
	}, nil
}

func (s *MicroServiceService) GetConsumerDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "GetConsumerDependencies failed for validating parameters failed.")
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	consumerId := in.ServiceId
	domainProject := util.ParseDomainProject(ctx)

	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerId)
	if err != nil {
		log.Errorf(err, "GetConsumerDependencies failed for get consumer failed.")
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if consumer == nil {
		log.Errorf(err, "GetConsumerDependencies failed for consumer does not exist, %s.", consumerId)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Consumer does not exist"),
		}, nil
	}

	dr := serviceUtil.NewConsumerDependencyRelation(ctx, domainProject, consumer)
	services, err := dr.GetDependencyProviders(toDependencyFilterOptions(in)...)
	if err != nil {
		log.Errorf(err, "GetConsumerDependencies failed for get providers failed.")
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetConDependenciesResponse{
		Response:  pb.CreateResponse(pb.Response_SUCCESS, "Get all providers successfully."),
		Providers: services,
	}, nil
}

func toDependencyFilterOptions(in *pb.GetDependenciesRequest) (opts []serviceUtil.DependencyRelationFilterOption) {
	if in.SameDomain {
		opts = append(opts, serviceUtil.WithSameDomainProject())
	}
	if in.NoSelf {
		opts = append(opts, serviceUtil.WithoutSelfDependency())
	}
	return opts
}

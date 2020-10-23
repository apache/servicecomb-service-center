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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

func (ds *DataSource) SearchProviderDependency(ctx context.Context, request *pb.GetDependenciesRequest) ([]*pb.MicroService, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	provider, err := serviceUtil.GetService(ctx, domainProject, providerServiceID)

	if err != nil {
		log.Errorf(err, "GetProviderDependencies failed, provider is %s", providerServiceID)
		return nil, err
	}
	if provider == nil {
		log.Errorf(err, "GetProviderDependencies failed for provider[%s] does not exist", providerServiceID)
		return nil, err
	}

	dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, provider)
	services, err := dr.GetDependencyConsumers(toDependencyFilterOptions(request)...)
	return services, err
}

func (ds *DataSource) SearchConsumerDependency(ctx context.Context, request *pb.GetDependenciesRequest) ([]*pb.MicroService, error) {
	consumerID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)
	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerID)

	if err != nil {
		log.Errorf(err, "GetConsumerDependencies failed, consumer is %s", consumerID)
		return nil, err
	}
	if consumer == nil {
		log.Errorf(err, "GetConsumerDependencies failed for consumer[%s] does not exist", consumerID)
		return nil, err
	}

	dr := serviceUtil.NewConsumerDependencyRelation(ctx, domainProject, consumer)
	services, err := dr.GetDependencyProviders(toDependencyFilterOptions(request)...)
	return services, err
}

func (ds *DataSource) DeleteDependency() {
	panic("implement me")
}

func (ds *DataSource) AddOrUpdateDependencies(ctx context.Context, dependencyInfos []*pb.ConsumerDependency, override bool) (*pb.Response, error) {
	opts := make([]client.PluginOp, 0, len(dependencyInfos))
	domainProject := util.ParseDomainProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerFlag := util.StringJoin([]string{dependencyInfo.Consumer.Environment, dependencyInfo.Consumer.AppId, dependencyInfo.Consumer.ServiceName, dependencyInfo.Consumer.Version}, "/")
		consumerInfo := proto.DependenciesToKeys([]*pb.MicroServiceKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := proto.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		rsp := serviceUtil.ParamsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer is %s, %s",
				override, consumerFlag, rsp.Response.GetMessage())
			return rsp.Response, nil
		}

		consumerID, err := serviceUtil.GetServiceID(ctx, consumerInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, get consumer[%s] id failed",
				override, consumerFlag)
			return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
		if len(consumerID) == 0 {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer[%s] does not exist",
				override, consumerFlag)
			return proto.CreateResponse(scerr.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag)), nil
		}

		dependencyInfo.Override = override
		data, err := json.Marshal(dependencyInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, marshal consumer[%s] dependency failed",
				override, consumerFlag)
			return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
		}

		id := apt.DepsQueueUUID
		if !override {
			id = util.GenerateUUID()
		}
		key := GenerateConsumerDependencyQueueKey(domainProject, consumerID, id)
		opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))
	}

	err := client.BatchCommit(ctx, opts)
	if err != nil {
		log.Errorf(err, "put request into dependency queue failed, override: %t, %v", override, dependencyInfos)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	log.Infof("put request into dependency queue successfully, override: %t, %v, from remote %s",
		override, dependencyInfos, util.GetIPFromContext(ctx))
	return proto.CreateResponse(proto.Response_SUCCESS, "Create dependency successfully."), nil
}

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

package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) SearchProviderDependency(ctx context.Context, request *discovery.GetDependenciesRequest) (*discovery.GetProDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	filter := mutil.GeneratorServiceFilter(ctx, providerServiceID)
	provider, err := mutil.GetService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("query provider service failed, there is no provider %s in db", providerServiceID))
			return &discovery.GetProDependenciesResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "Provider does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("query provider from db error, provider is %s", providerServiceID), err)
		return nil, err
	}
	if provider == nil {
		log.Error(fmt.Sprintf("GetProviderDependencies failed for provider %s", providerServiceID), err)
		return &discovery.GetProDependenciesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "Provider does not exist"),
		}, nil
	}

	dr := mutil.NewProviderDependencyRelation(ctx, domainProject, provider.Service)
	services, err := dr.GetDependencyConsumers(mutil.ToDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("GetProviderDependencies failed, provider is %s/%s/%s/%s",
			provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version), err)
		return &discovery.GetProDependenciesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	return &discovery.GetProDependenciesResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "Get all consumers successful."),
		Consumers: services,
	}, nil
}

func (ds *DataSource) SearchConsumerDependency(ctx context.Context, request *discovery.GetDependenciesRequest) (*discovery.GetConDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	consumerID := request.ServiceId

	filter := mutil.GeneratorServiceFilter(ctx, consumerID)
	consumer, err := mutil.GetService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("query consumer service failed, there is no consumer %s in db", consumerID))
			return &discovery.GetConDependenciesResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "Consumer does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("query consumer from db error, consumer is %s", consumerID), err)
		return nil, err
	}
	if consumer == nil {
		log.Error(fmt.Sprintf("GetConsumerDependencies failed for consumer %s does not exist", consumerID), err)
		return &discovery.GetConDependenciesResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "Consumer does not exist"),
		}, nil
	}

	dr := mutil.NewConsumerDependencyRelation(ctx, domainProject, consumer.Service)
	services, err := dr.GetDependencyProviders(mutil.ToDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("query consumer failed, consumer is %s/%s/%s/%s",
			consumer.Service.Environment, consumer.Service.AppId, consumer.Service.ServiceName, consumer.Service.Version), err)
		return &discovery.GetConDependenciesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	return &discovery.GetConDependenciesResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "Get all providers successfully."),
		Providers: services,
	}, nil
}

func (ds *DataSource) AddOrUpdateDependencies(ctx context.Context, dependencys []*discovery.ConsumerDependency, override bool) (*discovery.Response, error) {
	domainProject := util.ParseDomainProject(ctx)
	for _, dependency := range dependencys {
		consumerFlag := util.StringJoin([]string{
			dependency.Consumer.Environment,
			dependency.Consumer.AppId,
			dependency.Consumer.ServiceName,
			dependency.Consumer.Version}, "/")
		consumerInfo := discovery.DependenciesToKeys([]*discovery.MicroServiceKey{dependency.Consumer}, domainProject)[0]
		providersInfo := discovery.DependenciesToKeys(dependency.Providers, domainProject)

		rsp := datasource.ParamsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t consumer is %s %s",
				override, consumerFlag, rsp.Response.GetMessage()), nil)
			return rsp.Response, nil
		}

		consumerID, err := mutil.GetServiceID(ctx, consumerInfo)
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, get consumer %s id failed",
				override, consumerFlag), err)
			return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
		}
		if len(consumerID) == 0 {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t consumer %s does not exist",
				override, consumerFlag), err)
			return discovery.CreateResponse(discovery.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag)), nil
		}

		dependency.Override = override
		if !override {
			id := util.GenerateUUID()

			domain := util.ParseDomain(ctx)
			project := util.ParseProject(ctx)
			data := &db.ConsumerDep{
				Domain:      domain,
				Project:     project,
				ConsumerID:  consumerID,
				UUID:        id,
				ConsumerDep: dependency,
			}
			insertRes, err := client.GetMongoClient().Insert(ctx, db.CollectionDep, data)
			if err != nil {
				log.Error("failed to insert dep to mongodb", err)
				return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
			}
			log.Info(fmt.Sprintf("insert dep to mongodb success %s", insertRes.InsertedID))
		}
		err = mutil.SyncDependencyRule(ctx, domainProject, dependency)
		if err != nil {
			return nil, err
		}
	}
	return discovery.CreateResponse(discovery.ResponseSuccess, "Create dependency successfully."), nil
}

func (ds *DataSource) DeleteDependency() {
	panic("implement me")
}

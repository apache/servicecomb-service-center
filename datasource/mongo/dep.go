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

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"

	"fmt"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) SearchProviderDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	providerServiceID := request.ServiceId
	filter := GeneratorServiceFilter(ctx, providerServiceID)
	provider, err := GetService(ctx, filter)
	if err != nil {
		log.Error("GetProviderDependencies failed, provider is "+providerServiceID, err)
		return nil, err
	}
	if provider == nil {
		log.Error(fmt.Sprintf("GetProviderDependencies failed for provider %s", providerServiceID), err)
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Provider does not exist"),
		}, nil
	}

	services, err := GetDependencyProviders(ctx, provider.ServiceInfo, request)
	if err != nil {
		log.Error(fmt.Sprintf("GetProviderDependencies failed, provider is %s/%s/%s/%s",
			provider.ServiceInfo.Environment, provider.ServiceInfo.AppId, provider.ServiceInfo.ServiceName, provider.ServiceInfo.Version), err)
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetProDependenciesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Get all consumers successful."),
		Consumers: services,
	}, nil
}

func (ds *DataSource) SearchConsumerDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	consumerID := request.ServiceId

	filter := GeneratorServiceFilter(ctx, consumerID)
	consumer, err := GetService(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("GetConsumerDependencies failed, consumer is %s", consumerID), err)
		return nil, err
	}
	if consumer == nil {
		log.Error(fmt.Sprintf("GetConsumerDependencies failed for consumer %s does not exist", consumerID), err)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Consumer does not exist"),
		}, nil
	}

	services, err := GetDependencyProviders(ctx, consumer.ServiceInfo, request)
	if err != nil {
		log.Error(fmt.Sprintf("GetConsumerDependencies failed, consumer is %s/%s/%s/%s",
			consumer.ServiceInfo.Environment, consumer.ServiceInfo.AppId, consumer.ServiceInfo.ServiceName, consumer.ServiceInfo.Version), err)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetConDependenciesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Get all providers successfully."),
		Providers: services,
	}, nil
}

func (ds *DataSource) AddOrUpdateDependencies(ctx context.Context, dependencyInfos []*pb.ConsumerDependency, override bool) (*pb.Response, error) {
	domainProject := util.ParseDomainProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerFlag := util.StringJoin([]string{
			dependencyInfo.Consumer.Environment,
			dependencyInfo.Consumer.AppId,
			dependencyInfo.Consumer.ServiceName,
			dependencyInfo.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		rsp := datasource.ParamsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t consumer is %s %s",
				override, consumerFlag, rsp.Response.GetMessage()), nil)
			return rsp.Response, nil
		}

		consumerID, err := GetServiceID(ctx, consumerInfo)
		if err != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, get consumer %s id failed",
				override, consumerFlag), err)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
		if len(consumerID) == 0 {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t consumer %s does not exist",
				override, consumerFlag), err)
			return pb.CreateResponse(pb.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag)), nil
		}

		dependencyInfo.Override = override
		id := DepsQueueUUID
		if !override {
			id = util.GenerateUUID()
		}

		domain := util.ParseDomain(ctx)
		project := util.ParseProject(ctx)
		data := &Dependency{
			Domain:         domain,
			Project:        project,
			ConsumerID:     consumerID,
			UUID:           id,
			DependencyInfo: dependencyInfo,
		}
		insertRes, err := client.GetMongoClient().Insert(ctx, CollectionDep, data)
		if err != nil {
			log.Error("failed to insert dep to mongodb", err)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
		log.Error(fmt.Sprintf("insert dep to mongodb success %s", insertRes.InsertedID), err)
	}
	return pb.CreateResponse(pb.ResponseSuccess, "Create dependency successfully."), nil
}

func (ds *DataSource) DeleteDependency() {
	panic("implement me")
}

func GetDependencyProviders(ctx context.Context, consumer *pb.MicroService, request *pb.GetDependenciesRequest) ([]*pb.MicroService, error) {
	keys, err := GetProviderKeys(ctx, consumer)
	if err != nil {
		return nil, err
	}

	services := make([]*pb.MicroService, 0, len(keys))

	for _, key := range keys {
		domainProject := util.ParseDomainProject(ctx)
		if request.SameDomain && key.Tenant != domainProject {
			continue
		}

		providerIDs, err := ParseDependencyRule(ctx, key)
		if err != nil {
			return nil, err
		}

		if key.ServiceName == "*" {
			services = services[:0]
		}

		for _, providerID := range providerIDs {
			filter := GeneratorServiceFilter(ctx, providerID)
			provider, err := GetService(ctx, filter)
			if err != nil {
				log.Warn(fmt.Sprintf("get provider[%s/%s/%s/%s] failed",
					key.Environment, key.AppId, key.ServiceName, key.Version))
				continue
			}
			if provider == nil {
				log.Warn(fmt.Sprintf("provider[%s/%s/%s/%s] does not exist",
					key.Environment, key.AppId, key.ServiceName, key.Version))
				continue
			}
			if request.NoSelf && providerID == consumer.ServiceId {
				continue
			}
			services = append(services, provider.ServiceInfo)
		}

		if key.ServiceName == "*" {
			break
		}
	}

	return services, nil
}

func GetProviderKeys(ctx context.Context, consumer *pb.MicroService) ([]*pb.MicroServiceKey, error) {
	if consumer == nil {
		return nil, ErrInvalidConsumer
	}
	domainProject := util.ParseDomainProject(ctx)
	consumerMicroServiceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: consumer.Environment,
		AppId:       consumer.AppId,
		ServiceName: consumer.ServiceName,
		Alias:       consumer.Alias,
		Version:     consumer.Version,
	}

	filter := GenerateConsumerDependencyRuleKey(ctx, consumerMicroServiceKey)

	findRes, err := client.GetMongoClient().Find(ctx, CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	var services []*pb.MicroServiceKey
	for findRes.Next(ctx) {
		var tempMongoDep Dependency
		err := findRes.Decode(&tempMongoDep)
		if err != nil {
			return nil, err
		}
		providers := tempMongoDep.DependencyInfo.Providers
		services = append(services, providers...)
	}
	return services, nil
}

func GenerateConsumerDependencyRuleKey(ctx context.Context, in *pb.MicroServiceKey) bson.M {

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	if in == nil {
		return bson.M{
			ColumnDomain:  domain,
			ColumnProject: project,
		}
	}
	if in.ServiceName == "*" {
		return bson.M{
			ColumnDomain:  domain,
			ColumnProject: project,
			StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnEnv}): in.Environment,
		}
	}
	return bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnEnv}):     in.Environment,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnAppID}):   in.AppId,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnVersion}): in.Version,
	}
}

func ParseDependencyRule(ctx context.Context, dependencyRule *pb.MicroServiceKey) (serviceIDs []string, err error) {
	switch {
	case dependencyRule.ServiceName == "*":
		splited := strings.Split(dependencyRule.Tenant, "/")
		filter := bson.M{
			ColumnDomain:  splited[0],
			ColumnProject: splited[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnEnv}): dependencyRule.Environment}
		findRes, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
		if err != nil {
			return nil, err
		}
		for findRes.Next(ctx) {
			var service Service
			err = findRes.Decode(&service)
			if err != nil {
				return nil, err
			}
			serviceIDs = append(serviceIDs, service.ServiceInfo.ServiceId)
		}
	default:
		serviceIDs, err = FindServiceIds(ctx, dependencyRule)
	}
	return
}

func FindServiceIds(ctx context.Context, key *pb.MicroServiceKey) ([]string, error) {
	versionRule := key.Version
	splited := strings.Split(key.Tenant, "/")
	if len(versionRule) == 0 {
		return nil, nil
	}
	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		filter := bson.M{
			ColumnDomain:  splited[0],
			ColumnProject: splited[1]}
		return GetFilterVersionService(ctx, filter)
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter := bson.M{
			ColumnDomain:  splited[0],
			ColumnProject: splited[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): bson.M{"$gte": start}}
		return GetFilterVersionService(ctx, filter)
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter := bson.M{
			ColumnDomain:  splited[0],
			ColumnProject: splited[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): bson.M{"$gts": start, "$lt": end}}
		return GetFilterVersionService(ctx, filter)
	default:
		filter := bson.M{
			ColumnDomain:  splited[0],
			ColumnProject: splited[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):         key.Environment,
			StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):       key.AppId,
			StringBuilder([]string{ColumnServiceInfo, ColumnServiceName}): key.ServiceName,
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}):     key.Version}
		return GetFilterVersionService(ctx, filter)
	}
}

func GetFilterVersionService(ctx context.Context, m bson.M) (serviceIDs []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, CollectionService, m)
	if err != nil {
		return nil, err
	}
	for findRes.Next(ctx) {
		var service Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		serviceIDs = append(serviceIDs, service.ServiceInfo.ServiceId)
	}
	return
}

func GetServiceID(ctx context.Context, key *pb.MicroServiceKey) (serviceID string, err error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):         key.Environment,
		StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):       key.AppId,
		StringBuilder([]string{ColumnServiceInfo, ColumnServiceName}): key.ServiceName,
		StringBuilder([]string{ColumnServiceInfo, ColumnVersion}):     key.Version}

	findRes, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return "", nil
	}
	var service []*Service
	for findRes.Next(ctx) {
		var temp *Service
		err := findRes.Decode(&temp)
		if err != nil {
			return "", nil
		}
		service = append(service, temp)
	}
	if service == nil {
		return "", nil
	}
	return service[0].ServiceInfo.ServiceId, nil
}

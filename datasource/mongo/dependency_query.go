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
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DependencyRelation struct {
	ctx           context.Context
	domainProject string
	consumer      *pb.MicroService
	provider      *pb.MicroService
}

type DependencyRelationFilterOpt struct {
	SameDomainProject bool
	NonSelf           bool
}

type DependencyRelationFilterOption func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt

func NewConsumerDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, consumer, nil)
}

func NewProviderDependencyRelation(ctx context.Context, domainProject string, provider *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, nil, provider)
}

func NewDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroService) *DependencyRelation {
	return &DependencyRelation{
		ctx:           ctx,
		domainProject: domainProject,
		consumer:      consumer,
		provider:      provider,
	}
}

func (dr *DependencyRelation) GetDependencyProviders(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	services := make([]*pb.MicroService, 0, len(keys))
	op := ToDependencyRelationFilterOpt(opts...)

	for _, key := range keys {
		if op.SameDomainProject && key.Tenant != dr.domainProject {
			continue
		}
		providerIDs, err := dr.parseDependencyRule(key)

		if err != nil {
			return nil, err
		}

		if key.ServiceName == "*" {
			services = services[:0]
		}

		for _, providerID := range providerIDs {
			filter := GeneratorServiceFilter(dr.ctx, providerID)
			provider, err := GetService(dr.ctx, filter)
			if errors.Is(err, datasource.ErrNoDocuments) {
				log.Warn(fmt.Sprintf("provider[%s/%s/%s/%s] does not exist",
					key.Environment, key.AppId, key.ServiceName, key.Version))
				continue
			}
			if err != nil {
				log.Warn(fmt.Sprintf("get provider[%s/%s/%s/%s] failed",
					key.Environment, key.AppId, key.ServiceName, key.Version))
				continue
			}
			if op.NonSelf && providerID == dr.consumer.ServiceId {
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

func (dr *DependencyRelation) GetDependencyConsumers(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	consumerDependAllList, err := dr.GetDependencyConsumersOfProvider()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s consumers failed", dr.provider.ServiceId), err)
		return nil, err
	}
	consumers := make([]*pb.MicroService, 0)
	op := ToDependencyRelationFilterOpt(opts...)
	for _, consumer := range consumerDependAllList {
		if op.SameDomainProject && consumer.Tenant != dr.domainProject {
			continue
		}
		service, err := dr.GetServiceByMicroServiceKey(consumer)
		if err != nil {
			return nil, err
		}
		if service == nil {
			log.Warn(fmt.Sprintf("consumer[%s/%s/%s/%s] does not exist",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version))
			continue
		}
		if op.NonSelf && service.ServiceId == dr.provider.ServiceId {
			continue
		}
		consumers = append(consumers, service)
	}
	return consumers, nil
}

func (dr *DependencyRelation) GetDependencyConsumersOfProvider() ([]*pb.MicroServiceKey, error) {
	if dr.provider == nil {
		return nil, ErrInvalidConsumer
	}
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		log.Error(fmt.Sprintf("get consumers that depend on all services failed, %s", dr.provider.ServiceId), err)
		return nil, err
	}
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) GetServiceByMicroServiceKey(service *pb.MicroServiceKey) (*pb.MicroService, error) {
	filter, err := MicroServiceKeyFilter(service)
	if err != nil {
		log.Error("get serivce failed", err)
		return nil, err
	}
	findRes, err := client.GetMongoClient().Find(dr.ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}

	for findRes.Next(dr.ctx) {
		var service Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		if service.ServiceInfo != nil {
			return service.ServiceInfo, nil
		}
	}
	return nil, nil
}

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*pb.MicroServiceKey, error) {
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	filter := GenerateProviderDependencyRuleKey(dr.domainProject, dr.provider.ServiceId, providerService)
	findRes, err := client.GetMongoClient().Find(dr.ctx, CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}

	var msKeys []*pb.MicroServiceKey
	for findRes.Next(dr.ctx) {
		var dep *Dependency
		err = findRes.Decode(&dep)
		if err != nil {
			return nil, err
		}
		msKeys = append(msKeys, dep.DependencyInfo.Consumer)
	}
	return msKeys, nil
}

func (dr *DependencyRelation) getProviderKeys() ([]*pb.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, ErrInvalidConsumer
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)
	filter := GenerateConsumerDependencyRuleKey(dr.domainProject, dr.consumer.ServiceId, consumerMicroServiceKey)
	findRes, err := client.GetMongoClient().Find(dr.ctx, CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}
	var dep *Dependency
	for findRes.Next(dr.ctx) {
		err = findRes.Decode(&dep)
		if err != nil {
			return nil, err
		}
		if dep != nil {
			break
		}
	}
	if dep != nil {
		return dep.DependencyInfo.Providers, nil
	}
	return nil, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *pb.MicroServiceKey) (serviceIDs []string, err error) {
	switch {
	case dependencyRule.ServiceName == "*":
		log.Info(fmt.Sprintf("service[%s/%s/%s/%s] rely all service",
			dr.consumer.Environment, dr.consumer.AppId, dr.consumer.ServiceName, dr.consumer.Version))
		filter, err := MicroServiceKeyFilter(dependencyRule)
		if err != nil {
			log.Error("get serivce failed", err)
			return nil, err
		}
		findRes, err := client.GetMongoClient().Find(dr.ctx, CollectionService, filter)
		if err != nil {
			return nil, err
		}
		if findRes.Err() != nil {
			return nil, findRes.Err()
		}
		for findRes.Next(dr.ctx) {
			var service Service
			err = findRes.Decode(&service)
			if err != nil {
				return nil, err
			}
			serviceIDs = append(serviceIDs, service.ServiceInfo.ServiceId)
		}
	default:
		serviceIDs, _, err = FindServiceIds(dr.ctx, dependencyRule.Version, dependencyRule)
	}
	return
}

func (dr *DependencyRelation) GetDependencyConsumerIds() ([]string, error) {
	consumerDependAllList, err := dr.GetDependencyConsumersOfProvider()
	if err != nil {
		return nil, err
	}
	consumerIDs := make([]string, 0, len(consumerDependAllList))
	for _, consumer := range consumerDependAllList {
		consumerID, err := GetServiceID(dr.ctx, consumer)
		if err != nil {
			log.Error(fmt.Sprintf("get consumer[%s/%s/%s/%s] failed",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version), err)
			return nil, err
		}
		if len(consumerID) == 0 {
			log.Warn(fmt.Sprintf("get consumer[%s/%s/%s/%s] not exist",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version))
			continue
		}
		consumerIDs = append(consumerIDs, consumerID)
	}
	return consumerIDs, nil
}

func MicroServiceKeyFilter(key *pb.MicroServiceKey) (bson.M, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, ErrInvalidDomainProject
	}
	return bson.M{
		ColumnDomain:  tenant[0],
		ColumnProject: tenant[1],
		StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):     key.Environment,
		StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):   key.AppId,
		StringBuilder([]string{ColumnServiceInfo, ColumnAlias}):   key.Alias,
		StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): key.Version}, nil
}

func FindServiceIds(ctx context.Context, versionRule string, key *pb.MicroServiceKey) (serviceIDs []string, b bool, err error) {
	match, err := ParseVersionRule(ctx, versionRule, key)
	if err != nil {
		return nil, false, ErrInvalidDomainProject
	}
	if match == nil {
		tenant := strings.Split(key.Tenant, "/")
		if len(tenant) != 2 {
			return nil, false, ErrInvalidDomainProject
		}
		filter := bson.M{
			ColumnDomain:  tenant[0],
			ColumnProject: tenant[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):         key.Environment,
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}):     versionRule,
			StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):       key.AppId,
			StringBuilder([]string{ColumnServiceInfo, ColumnServiceName}): key.ServiceName,
			StringBuilder([]string{ColumnServiceInfo, ColumnAlias}):       key.Alias}
		findRes, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
		if err != nil {
			return nil, false, nil
		}
		if findRes.Err() != nil {
			return nil, false, findRes.Err()
		}
		for findRes.Next(ctx) {
			var service Service
			err = findRes.Decode(&service)
			if err != nil {
				return nil, false, err
			}
			serviceIDs = append(serviceIDs, service.ServiceInfo.ServiceId)
		}
		return serviceIDs, false, nil
	}
	return match, false, nil
}

func ParseVersionRule(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, ErrInvalidDomainProject
	}
	if len(versionRule) == 0 {
		return nil, nil
	}

	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		filter := bson.M{
			ColumnDomain:  tenant[0],
			ColumnProject: tenant[1]}
		return GetFilterVersionServiceLatest(ctx, filter)
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter := bson.M{
			ColumnDomain:  tenant[0],
			ColumnProject: tenant[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): bson.M{"$gte": start}}
		return GetFilterVersionService(ctx, filter)
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter := bson.M{
			ColumnDomain:  tenant[0],
			ColumnProject: tenant[1],
			StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): bson.M{"$gte": start, "$lte": end}}
		return GetFilterVersionService(ctx, filter)
	default:
		return nil, nil
	}
}

func GetFilterVersionService(ctx context.Context, m bson.M) (serviceIDs []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, CollectionService, m)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
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

func GetFilterVersionServiceLatest(ctx context.Context, m bson.M) (serviceIDs []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, CollectionService, m,
		&options.FindOptions{
			Sort: bson.M{StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): -1}})
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}
	for findRes.Next(ctx) {
		var service Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		serviceIDs = append(serviceIDs, service.ServiceInfo.ServiceId)
		if serviceIDs != nil {
			return serviceIDs, nil
		}
	}
	return
}

func WithSameDomainProject() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.SameDomainProject = true
		return opt
	}
}
func WithoutSelfDependency() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.NonSelf = true
		return opt
	}
}

func ToDependencyFilterOptions(in *pb.GetDependenciesRequest) (opts []DependencyRelationFilterOption) {
	if in.SameDomain {
		opts = append(opts, WithSameDomainProject())
	}
	if in.NoSelf {
		opts = append(opts, WithoutSelfDependency())
	}
	return opts
}

func ToDependencyRelationFilterOpt(opts ...DependencyRelationFilterOption) (op DependencyRelationFilterOpt) {
	for _, opt := range opts {
		op = opt(op)
	}
	return
}

func GenerateConsumerDependencyRuleKey(domainProject string, serviceID string, in *pb.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(domainProject, serviceID, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, serviceID string, in *pb.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(domainProject, serviceID, in)
}

func GenerateServiceDependencyRuleKey(domainProject string, serviceID string, in *pb.MicroServiceKey) bson.M {
	if in == nil {
		return bson.M{
			ColumnConsumerID: serviceID,
			StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnTenant}): domainProject}
	}
	if in.ServiceName == "*" {
		return bson.M{
			ColumnConsumerID: serviceID,
			StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnTenant}):      domainProject,
			StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnEnv}):         in.Environment,
			StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnServiceName}): in.ServiceName}
	}
	return bson.M{
		ColumnConsumerID: serviceID,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnTenant}):      domainProject,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnEnv}):         in.Environment,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnAppID}):       in.AppId,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnVersion}):     in.Version,
		StringBuilder([]string{ColumnDependencyInfo, ColumnConsumer, ColumnServiceName}): in.ServiceName}
}

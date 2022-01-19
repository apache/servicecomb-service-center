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

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type DependencyRelation struct {
	ctx           context.Context
	domainProject string
	consumer      *discovery.MicroService
	provider      *discovery.MicroService
}

type DependencyRelationFilterOpt struct {
	SameDomainProject bool
	NonSelf           bool
}

type DependencyRelationFilterOption func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt

func NewConsumerDependencyRelation(ctx context.Context, domainProject string, consumer *discovery.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, consumer, nil)
}

func NewProviderDependencyRelation(ctx context.Context, domainProject string, provider *discovery.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, nil, provider)
}

func NewDependencyRelation(ctx context.Context, domainProject string, consumer *discovery.MicroService, provider *discovery.MicroService) *DependencyRelation {
	return &DependencyRelation{
		ctx:           ctx,
		domainProject: domainProject,
		consumer:      consumer,
		provider:      provider,
	}
}

func (dr *DependencyRelation) GetDependencyProviders(opts ...DependencyRelationFilterOption) ([]*discovery.MicroService, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	services := make([]*discovery.MicroService, 0, len(keys))
	op := ToDependencyRelationFilterOpt(opts...)

	for _, key := range keys {
		if op.SameDomainProject && key.Tenant != dr.domainProject {
			continue
		}
		providerIDs, err := dr.parseDependencyRule(key)

		if err != nil {
			return nil, err
		}

		for _, providerID := range providerIDs {
			provider, err := GetServiceByID(dr.ctx, providerID)
			if err != nil {
				if errors.Is(err, datasource.ErrNoData) {
					log.Warn(fmt.Sprintf("provider[%s/%s/%s/%s] does not exist",
						key.Environment, key.AppId, key.ServiceName, key.Version))
				} else {
					log.Warn(fmt.Sprintf("get provider[%s/%s/%s/%s] failed",
						key.Environment, key.AppId, key.ServiceName, key.Version))
				}
				continue
			}
			if op.NonSelf && providerID == dr.consumer.ServiceId {
				continue
			}
			services = append(services, provider.Service)
		}
	}
	return services, nil
}

func (dr *DependencyRelation) GetDependencyConsumers(opts ...DependencyRelationFilterOption) ([]*discovery.MicroService, error) {
	consumerDependAllList, err := dr.GetDependencyConsumersOfProvider()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s consumers failed", dr.provider.ServiceId), err)
		return nil, err
	}
	consumers := make([]*discovery.MicroService, 0)
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

func (dr *DependencyRelation) GetDependencyConsumersOfProvider() ([]*discovery.MicroServiceKey, error) {
	if dr.provider == nil {
		return nil, util.ErrInvalidConsumer
	}
	providerService := discovery.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependList, err := dr.GetConsumerOfSameServiceNameAndAppID(providerService)
	if err != nil {
		log.Error(fmt.Sprintf("get consumers that depend on rule[%s/%s/%s/%s] failed",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version), err)
		return nil, err
	}
	return consumerDependList, nil
}

func (dr *DependencyRelation) GetConsumerOfSameServiceNameAndAppID(provider *discovery.MicroServiceKey) ([]*discovery.MicroServiceKey, error) {
	filter := GenerateRuleKeyWithSameServiceNameAndAppID(path.DepsProvider, dr.domainProject, provider)
	depRules, err := getServiceKeysInDep(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	var allConsumers []*discovery.MicroServiceKey
	for _, depRule := range depRules {
		allConsumers = append(allConsumers, depRule.Dep.Dependency...)
	}
	return allConsumers, nil
}

func (dr *DependencyRelation) GetServiceByMicroServiceKey(service *discovery.MicroServiceKey) (*discovery.MicroService, error) {
	filter, err := MicroServiceKeyFilter(service)
	if err != nil {
		log.Error("get serivce failed", err)
		return nil, err
	}
	findRes, err := mongo.GetClient().GetDB().Collection(model.CollectionService).Find(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}

	for findRes.Next(dr.ctx) {
		var service model.Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		if service.Service != nil {
			return service.Service, nil
		}
	}
	return nil, nil
}

func getServiceKeysInDep(ctx context.Context, filter interface{}) ([]*model.DependencyRule, error) {
	findRes, err := mongo.GetClient().GetDB().Collection(model.CollectionDep).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer findRes.Close(ctx)
	var depRules []*model.DependencyRule
	for findRes.Next(ctx) {
		var tmp *model.DependencyRule
		err := findRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, tmp)
	}
	return depRules, nil
}

func (dr *DependencyRelation) getProviderKeys() ([]*discovery.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, util.ErrInvalidConsumer
	}
	consumerMicroServiceKey := discovery.MicroServiceToKey(dr.domainProject, dr.consumer)
	filter := GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)

	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *discovery.MicroServiceKey) (serviceIDs []string, err error) {
	serviceIDs, _, err = FindServiceIds(dr.ctx, dependencyRule, false)
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
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
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

func MicroServiceKeyFilter(key *discovery.MicroServiceKey) (bson.M, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, util.ErrInvalidDomainProject
	}
	filter := util.NewDomainProjectFilter(tenant[0], tenant[1],
		util.ServiceEnv(key.Environment),
		util.ServiceAppID(key.AppId),
		util.ServiceServiceName(key.ServiceName),
		util.ServiceVersion(key.Version),
	)
	return filter, nil
}

func FindServiceIds(ctx context.Context, key *discovery.MicroServiceKey, matchVersion bool) ([]string, bool, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, false, util.ErrInvalidDomainProject
	}

	baseFilter := bson.D{
		{Key: model.ColumnDomain, Value: tenant[0]},
		{Key: model.ColumnProject, Value: tenant[1]},
		{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnEnv}), Value: key.Environment},
		{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnAppID}), Value: key.AppId}}

	serviceIds, exist, err := findServiceKeysByServiceName(ctx, key, baseFilter, matchVersion)
	if err != nil {
		return nil, false, err
	}
	if len(serviceIds) == 0 {
		if exist {
			// service exist but version not matched
			return nil, true, nil
		}
		if len(key.Alias) == 0 {
			return nil, false, nil
		}
		serviceIds, exist, err = findServiceKeysByAlias(ctx, key, baseFilter, matchVersion)
		if err != nil {
			return nil, false, err
		}
		return serviceIds, exist, nil
	}
	return serviceIds, exist, nil
}

func serviceVersionFilter(ctx context.Context, version string, filter bson.D, matchVersion bool) ([]string, bool, error) {
	num, err := mongo.GetClient().GetDB().Collection(model.CollectionService).CountDocuments(ctx, filter)
	if err != nil || num == 0 {
		return nil, false, err
	}
	newFilter := filter
	if matchVersion {
		newFilter = findServiceKeys(ctx, version, filter)
	}
	ids, err := GetVersionService(ctx, newFilter)
	if err != nil {
		return nil, false, err
	}
	return ids, true, nil
}

func findServiceKeysByServiceName(ctx context.Context, key *discovery.MicroServiceKey, baseFilter bson.D, matchVersion bool) ([]string, bool, error) {
	filter := append(baseFilter,
		bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceName}), Value: key.ServiceName})
	return serviceVersionFilter(ctx, key.Version, filter, matchVersion)
}

func findServiceKeysByAlias(ctx context.Context, key *discovery.MicroServiceKey, baseFilter bson.D, matchVersion bool) ([]string, bool, error) {
	filter := append(baseFilter,
		bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnAlias}), Value: key.Alias})
	return serviceVersionFilter(ctx, key.Version, filter, matchVersion)
}

type ServiceVersionFilter func(ctx context.Context, filter bson.D) ([]string, error)

func findServiceKeys(ctx context.Context, version string, filter bson.D) (newFilter bson.D) {
	filter = append(filter, bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: version})
	return filter
}

func GetVersionService(ctx context.Context, m bson.D) (serviceIds []string, err error) {
	findRes, err := mongo.GetClient().GetDB().Collection(model.CollectionService).Find(ctx, m, &options.FindOptions{
		Sort: bson.M{util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}): -1}})
	if err != nil {
		return
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}
	for findRes.Next(ctx) {
		var service *model.Service
		err = findRes.Decode(&service)
		if err != nil {
			return
		}
		serviceIds = append(serviceIds, service.Service.ServiceId)
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

func ToDependencyFilterOptions(in *discovery.GetDependenciesRequest) (opts []DependencyRelationFilterOption) {
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

func GenerateConsumerDependencyRuleKey(domainProject string, in *discovery.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(path.DepsConsumer, domainProject, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, in *discovery.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(path.DepsProvider, domainProject, in)
}

func GenerateRuleKeyWithSameServiceNameAndAppID(serviceType string, domainProject string, in *discovery.MicroServiceKey) bson.M {
	return util.NewFilter(
		util.ServiceType(serviceType),
		util.ServiceKeyTenant(domainProject),
		util.ServiceKeyAppID(in.AppId),
		util.ServiceKeyServiceName(in.ServiceName),
	)
}

func GenerateServiceDependencyRuleKey(serviceType string, domainProject string, in *discovery.MicroServiceKey) bson.M {
	if in == nil {
		return util.NewFilter(
			util.ServiceType(serviceType),
			util.ServiceKeyTenant(domainProject),
		)
	}
	return util.NewFilter(
		util.ServiceType(serviceType),
		util.ServiceKeyTenant(domainProject),
		util.ServiceKeyServiceEnv(in.Environment),
		util.ServiceKeyAppID(in.AppId),
		util.ServiceKeyServiceVersion(in.Version),
		util.ServiceKeyServiceName(in.ServiceName),
	)
}

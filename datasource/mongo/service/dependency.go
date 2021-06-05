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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/datasource/util/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
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
	op := dependencyRelationFilterOption(opts...)

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
			filter := mutil.NewBasicFilter(dr.ctx, mutil.ServiceServiceID(providerID))
			provider, err := findService(dr.ctx, filter)
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
		if key.ServiceName == "*" {
			break
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
	op := dependencyRelationFilterOption(opts...)
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

func (dr *DependencyRelation) GetServiceByMicroServiceKey(service *discovery.MicroServiceKey) (*discovery.MicroService, error) {
	tenant := strings.Split(service.Tenant, "/")
	if len(tenant) != 2 {
		return nil, mutil.ErrInvalidDomainProject
	}
	filter := mutil.NewDomainProjectFilter(tenant[0], tenant[1],
		mutil.ServiceEnv(service.Environment),
		mutil.ServiceAppID(service.AppId),
		mutil.ServiceAlias(service.Alias),
		mutil.ServiceVersion(service.Version),
	)
	svc, err := findMicroService(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (dr *DependencyRelation) GetDependencyConsumersOfProvider() ([]*discovery.MicroServiceKey, error) {
	if dr.provider == nil {
		return nil, mutil.ErrInvalidConsumer
	}
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		log.Error(fmt.Sprintf("get consumers that depend on all services failed, %s", dr.provider.ServiceId), err)
		return nil, err
	}
	providerService := discovery.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependList, err := dr.GetConsumerOfSameServiceNameAndAppID(providerService)
	if err != nil {
		log.Errorf(err, "get consumers that depend on rule[%s/%s/%s/%s] failed",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return nil, err
	}
	consumerDependAllList = append(consumerDependAllList, consumerDependList...)
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) GetConsumerOfSameServiceNameAndAppID(provider *discovery.MicroServiceKey) ([]*discovery.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	filter := mutil.NewFilter(
		mutil.DependencyRuleType(path.DepsProvider),
		mutil.ServiceKeyTenant(dr.domainProject),
		mutil.ServiceKeyAppID(provider.AppId),
		mutil.ServiceKeyServiceName(provider.ServiceName))
	provider.Version = providerVersion
	depRules, err := findDeps(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	var allConsumers []*discovery.MicroServiceKey
	var latestServiceID []string

	for _, depRule := range depRules {
		providerVersionRule := depRule.ServiceKey.Version
		if providerVersionRule == "latest" {
			if latestServiceID == nil {
				latestServiceID, _, err = findServiceIDs(dr.ctx, providerVersionRule, provider)
				if err != nil {
					log.Error(fmt.Sprintf("get service[%s/%s/%s/%s]'s serviceID failed",
						provider.Environment, provider.AppId, provider.ServiceName, providerVersionRule), err)
					return nil, err
				}
			}
			if len(latestServiceID) == 0 {
				log.Info(fmt.Sprintf("service[%s/%s/%s/%s] does not exist",
					provider.Environment, provider.AppId, provider.ServiceName, providerVersionRule))
				continue
			}
			if dr.provider.ServiceId != latestServiceID[0] {
				continue
			}
		} else {
			if !versionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}
		if len(depRule.Dep.Dependency) > 0 {
			allConsumers = append(allConsumers, depRule.Dep.Dependency...)
		}
	}
	return allConsumers, nil
}

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*discovery.MicroServiceKey, error) {
	providerService := discovery.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	filter := generateProviderDependencyRuleKey(dr.domainProject, providerService)
	return findMicroServiceDependencies(dr.ctx, filter)
}

func (dr *DependencyRelation) getProviderKeys() ([]*discovery.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, mutil.ErrInvalidConsumer
	}
	consumerMicroServiceKey := discovery.MicroServiceToKey(dr.domainProject, dr.consumer)
	filter := generateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)

	consumerDependency, err := findMicroServiceDependency(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *discovery.MicroServiceKey) (serviceIDs []string, err error) {
	switch {
	case dependencyRule.ServiceName == "*":
		log.Info(fmt.Sprintf("service[%s/%s/%s/%s] rely all service",
			dr.consumer.Environment, dr.consumer.AppId, dr.consumer.ServiceName, dr.consumer.Version))
		tenant := strings.Split(dependencyRule.Tenant, "/")
		if len(tenant) != 2 {
			return nil, mutil.ErrInvalidDomainProject
		}
		filter := mutil.NewDomainProjectFilter(tenant[0], tenant[1], mutil.ServiceEnv(dependencyRule.Environment))
		return findServicesIDs(dr.ctx, filter)
	default:
		serviceIDs, _, err = findServiceIDs(dr.ctx, dependencyRule.Version, dependencyRule)
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
		consumerID, err := getServiceID(dr.ctx, consumer)
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

func dependencyRelationFilterOption(opts ...DependencyRelationFilterOption) (op DependencyRelationFilterOpt) {
	for _, opt := range opts {
		op = opt(op)
	}
	return
}

func (ds *DataSource) SearchProviderDependency(ctx context.Context, request *discovery.GetDependenciesRequest) (*discovery.GetProDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(providerServiceID))
	provider, err := findService(ctx, filter)
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

	dr := NewProviderDependencyRelation(ctx, domainProject, provider.Service)
	services, err := dr.GetDependencyConsumers(dependencyFilterOptions(request)...)
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
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(consumerID))
	consumer, err := findService(ctx, filter)
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

	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer.Service)
	services, err := dr.GetDependencyProviders(dependencyFilterOptions(request)...)
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

		rsp := paramsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t consumer is %s %s",
				override, consumerFlag, rsp.Response.GetMessage()), nil)
			return rsp.Response, nil
		}

		consumerID, err := getServiceID(ctx, consumerInfo)
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
			data := &model.ConsumerDep{
				Domain:      domain,
				Project:     project,
				ConsumerID:  consumerID,
				UUID:        id,
				ConsumerDep: dependency,
			}
			err := insertDep(ctx, data)
			if err != nil {
				log.Error("failed to insert dep to mongodb", err)
				return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
			}
			log.Info("succeed to insert dep to mongodb")
		}
		err = syncDependencyRule(ctx, domainProject, dependency)
		if err != nil {
			return nil, err
		}
	}
	return discovery.CreateResponse(discovery.ResponseSuccess, "Create dependency successfully."), nil
}

func (ds *DataSource) DeleteDependency() {
	panic("implement me")
}

func (ds *DataSource) DependencyHandle(ctx context.Context) (err error) {
	return nil
}

func syncDependencyRule(ctx context.Context, domainProject string, r *discovery.ConsumerDependency) error {

	consumerInfo := discovery.DependenciesToKeys([]*discovery.MicroServiceKey{r.Consumer}, domainProject)[0]
	providersInfo := discovery.DependenciesToKeys(r.Providers, domainProject)

	var dep datasource.Dependency
	//var err error
	dep.DomainProject = domainProject
	dep.Consumer = consumerInfo
	dep.ProvidersRule = providersInfo
	// add mongo get dep here

	oldProviderRules, err := getOldProviderRule(ctx, &dep)
	if err != nil {
		return err
	}

	if r.Override {
		datasource.ParseOverrideRules(ctx, &dep, oldProviderRules)
	} else {
		datasource.ParseAddOrUpdateRules(ctx, &dep, oldProviderRules)
	}
	return updateDeps(ctx, domainProject, &dep)
}

func getOldProviderRule(ctx context.Context, dep *datasource.Dependency) (*discovery.MicroServiceDependency, error) {
	microServiceDependency := &discovery.MicroServiceDependency{
		Dependency: []*discovery.MicroServiceKey{},
	}
	filter := generateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	depRule, err := findDep(ctx, filter)
	if err != nil && err != datasource.ErrNoData {
		return nil, err
	}
	if depRule == nil {
		return microServiceDependency, nil
	}
	return depRule.Dep, nil
}

func updateDeps(ctx context.Context, domainProject string, dep *datasource.Dependency) error {
	var upsert = true
	for _, r := range dep.DeleteDependencyRuleList {
		filter := generateProviderDependencyRuleKey(domainProject, r)
		pullValue := mutil.NewFilter(mutil.DepDependency(dep.Consumer))
		updateFilter := mutil.NewFilter(mutil.Pull(pullValue))
		err := updateServiceDependencies(ctx, filter, updateFilter)
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	for _, r := range dep.CreateDependencyRuleList {
		filter := generateProviderDependencyRuleKey(domainProject, r)
		addToSetValue := mutil.NewFilter(mutil.DepDependency(dep.Consumer))
		updateFilter := mutil.NewFilter(mutil.AddToSet(addToSetValue))
		err := updateServiceDependencies(ctx, filter, updateFilter, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	filter := generateConsumerDependencyRuleKey(domainProject, dep.Consumer)
	if len(dep.ProvidersRule) == 0 {
		err := deleteServiceDependencies(ctx, filter)
		if err != nil {
			return err
		}
	} else {
		updateValue := mutil.NewFilter(mutil.DepDependency(dep.ProvidersRule))
		updateFilter := mutil.NewFilter(mutil.Set(updateValue))
		err := updateServiceDependencies(ctx, filter, updateFilter, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
	}

	err := CleanUpDepRules(ctx, domainProject)
	if err != nil {
		return err
	}

	return nil
}

func CleanUpDepRules(ctx context.Context, domainProject string) error {
	if len(domainProject) == 0 {
		return mutil.ErrInvalidDomainProject
	}

	cache := make(map[*model.DelDepCacheKey]bool)
	err := removeProviderRuleOfConsumer(ctx, domainProject, cache)

	if err != nil {
		return err
	}

	return removeProviderRuleKeys(ctx, domainProject, cache)
}

func removeProviderRuleOfConsumer(ctx context.Context, domainProject string, cache map[*model.DelDepCacheKey]bool) error {
	filter := generateConsumerDependencyRuleKey(domainProject, nil)
	depRules, err := getDepRules(ctx, filter)
	if err != nil {
		return err
	}
	for _, depRule := range depRules {
		err := removeConsumerDeps(ctx, depRule, cache)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeProviderRuleKeys(ctx context.Context, domainProject string, cache map[*model.DelDepCacheKey]bool) error {
	filter := generateProviderDependencyRuleKey(domainProject, nil)
	depRules, err := getDepRules(ctx, filter)
	if err != nil {
		return err
	}
	for _, depRule := range depRules {
		err := removeProviderDeps(ctx, depRule, cache)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeProviderDeps(ctx context.Context, depRule *model.DependencyRule, cache map[*model.DelDepCacheKey]bool) (err error) {
	id := &model.DelDepCacheKey{
		Key:  depRule.ServiceKey,
		Type: path.DepsConsumer,
	}
	exist, ok := cache[id]
	if !ok {
		if depRule.ServiceKey.ServiceName == "*" {
			return nil
		}

		_, exist, err = findServiceIDs(ctx, depRule.ServiceKey.Version, depRule.ServiceKey)
		if err != nil {
			return err
		}

		cache[id] = exist
	}

	filter := bson.M{
		model.ColumnServiceKey: depRule.ServiceKey,
	}
	if !exist {
		err = deleteServiceDependencies(ctx, filter)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeConsumerDeps(ctx context.Context, depRule *model.DependencyRule, cache map[*model.DelDepCacheKey]bool) (err error) {
	var left []*discovery.MicroServiceKey
	for _, key := range depRule.Dep.Dependency {
		if key.ServiceName == "*" {
			left = append(left, key)
			continue
		}

		id := &model.DelDepCacheKey{
			Key:  key,
			Type: path.DepsProvider,
		}
		exist, ok := cache[id]
		if !ok {
			_, exist, err = findServiceIDs(ctx, key.Version, key)
			if err != nil {
				return err
			}
			cache[id] = exist
		}

		if exist {
			left = append(left, key)
		}
	}
	if len(depRule.Dep.Dependency) == len(left) {
		return nil
	}
	filter := mutil.NewFilter(mutil.ServiceKey(depRule.ServiceKey))

	if len(left) == 0 {
		err = deleteServiceDependencies(ctx, filter)
	} else {
		updateValue := mutil.NewFilter(mutil.DepDependency(left))
		updateFilter := mutil.NewFilter(mutil.Set(updateValue))
		err = updateServiceDependencies(ctx, filter, updateFilter)
	}
	if err != nil {
		return err
	}
	return nil
}

func getServiceID(ctx context.Context, key *discovery.MicroServiceKey) (id string, err error) {
	filter := mutil.NewBasicFilter(
		ctx,
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		mutil.ServiceServiceName(key.ServiceName),
		mutil.ServiceVersion(key.Version),
	)
	service, err := findService(ctx, filter)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return "", err
	}
	if service != nil {
		id = service.Service.ServiceId
	}
	if len(id) == 0 && len(key.Alias) != 0 {
		filter = mutil.NewBasicFilter(
			ctx,
			mutil.ServiceEnv(key.Environment),
			mutil.ServiceAppID(key.AppId),
			mutil.ServiceAlias(key.Alias),
			mutil.ServiceVersion(key.Version),
		)
		service, err = findService(ctx, filter)
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			return "", err
		}
		if service != nil {
			id = service.Service.ServiceId
		}
	}
	return id, err
}

func paramsChecker(consumerInfo *discovery.MicroServiceKey, providersInfo []*discovery.MicroServiceKey) *discovery.CreateDependenciesResponse {
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			break
		}
		if len(providerInfo.AppId) == 0 {
			providerInfo.AppId = consumerInfo.AppId
		}

		version := providerInfo.Version
		if len(version) == 0 {
			return badParamsDependenciesResponse("Required provider version")
		}

		providerInfo.Version = ""
		if _, ok := flag[path.GenerateProviderDependencyRuleKey(providerInfo.Tenant, providerInfo)]; ok {
			return badParamsDependenciesResponse("Invalid request body for provider info.Duplicate provider or (serviceName and appId is same).")
		}
		flag[path.GenerateProviderDependencyRuleKey(providerInfo.Tenant, providerInfo)] = true
		providerInfo.Version = version
	}
	return nil
}

func badParamsDependenciesResponse(detailErr string) *discovery.CreateDependenciesResponse {
	log.Error(fmt.Sprintf("request params is invalid. %s", detailErr), nil)
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &discovery.CreateDependenciesResponse{Response: discovery.CreateResponse(discovery.ErrInvalidParams, detailErr)}
}

func existDependencyRule(ctx context.Context, filter interface{}, target *discovery.MicroServiceKey) (bool, error) {
	microServiceDependency, err := findMicroServiceDependency(ctx, filter)
	if err != nil {
		return false, err
	}

	if len(microServiceDependency.Dependency) != 0 {
		isEqual, err := containServiceDependency(microServiceDependency.Dependency, target)
		if err != nil {
			return false, err
		}
		if isEqual {
			return true, nil
		}
	}
	return false, nil
}

func containServiceDependency(services []*discovery.MicroServiceKey, service *discovery.MicroServiceKey) (bool, error) {
	if services == nil || service == nil {
		return false, errors.New("invalid params input")
	}
	for _, value := range services {
		rst := equalServiceDependency(service, value)
		if rst {
			return true, nil
		}
	}
	return false, nil
}

func equalServiceDependency(serviceA *discovery.MicroServiceKey, serviceB *discovery.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	return stringA == stringB
}

func toString(in *discovery.MicroServiceKey) string {
	return path.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func generateProviderDependencyRuleKey(domainProject string, in *discovery.MicroServiceKey) bson.M {
	return generateServiceDependencyRuleKey(path.DepsProvider, domainProject, in)
}

func generateServiceDependencyRuleKey(dependencyRuleType string, domainProject string, in *discovery.MicroServiceKey) bson.M {
	if in == nil {
		return mutil.NewFilter(
			mutil.DependencyRuleType(dependencyRuleType),
			mutil.ServiceKeyTenant(domainProject),
		)
	}
	if in.ServiceName == "*" {
		return mutil.NewFilter(
			mutil.DependencyRuleType(dependencyRuleType),
			mutil.ServiceKeyTenant(domainProject),
			mutil.ServiceKeyServiceEnv(in.Environment),
			mutil.ServiceKeyServiceName(in.ServiceName),
		)
	}
	return mutil.NewFilter(
		mutil.DependencyRuleType(dependencyRuleType),
		mutil.ServiceKeyTenant(domainProject),
		mutil.ServiceKeyServiceEnv(in.Environment),
		mutil.ServiceKeyAppID(in.AppId),
		mutil.ServiceKeyServiceVersion(in.Version),
		mutil.ServiceKeyServiceName(in.ServiceName),
	)
}

func generateConsumerDependencyRuleKey(domainProject string, in *discovery.MicroServiceKey) bson.M {
	return generateServiceDependencyRuleKey(path.DepsConsumer, domainProject, in)
}

func existDependencyRuleByProviderConsumer(ctx context.Context, provider *discovery.MicroServiceKey, consumer *discovery.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}
	consumerKey := generateConsumerDependencyRuleKey(consumer.Tenant, consumer)
	existed, err := existDependencyRule(ctx, consumerKey, provider)
	if err != nil || existed {
		return existed, err
	}
	providerKey := generateProviderDependencyRuleKey(targetDomainProject, provider)
	return existDependencyRule(ctx, providerKey, consumer)
}

// not prepare for latest scene, should merge it with find serviceids func.
func versionMatchRule(version, versionRule string) bool {
	if len(versionRule) == 0 {
		return false
	}
	rangeIdx := strings.Index(versionRule, "-")
	versionInt, _ := validate.VersionToInt64(version)
	switch {
	case versionRule[len(versionRule)-1:] == "+":
		start, _ := validate.VersionToInt64(versionRule[:len(versionRule)-1])
		return versionInt >= start
	case rangeIdx > 0:
		start, _ := validate.VersionToInt64(versionRule[:rangeIdx])
		end, _ := validate.VersionToInt64(versionRule[rangeIdx+1:])
		return versionInt >= start && versionInt < end
	default:
		return version == versionRule
	}
}

func dependencyFilterOptions(in *discovery.GetDependenciesRequest) (opts []DependencyRelationFilterOption) {
	if in.SameDomain {
		opts = append(opts, withSameDomainProject())
	}
	if in.NoSelf {
		opts = append(opts, withoutSelfDependency())
	}
	return opts
}

func withSameDomainProject() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.SameDomainProject = true
		return opt
	}
}

func withoutSelfDependency() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.NonSelf = true
		return opt
	}
}

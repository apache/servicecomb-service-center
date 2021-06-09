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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type DepManager struct {
}

func (ds *DepManager) SearchProviderDependency(ctx context.Context, request *discovery.GetDependenciesRequest) (*discovery.GetProDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(providerServiceID))
	provider, err := dao.GetService(ctx, filter)
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
	services, err := dr.GetDependencyConsumers(ToDependencyFilterOptions(request)...)
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

func (ds *DepManager) SearchConsumerDependency(ctx context.Context, request *discovery.GetDependenciesRequest) (*discovery.GetConDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	consumerID := request.ServiceId
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(consumerID))
	consumer, err := dao.GetService(ctx, filter)
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
	services, err := dr.GetDependencyProviders(ToDependencyFilterOptions(request)...)
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

func (ds *DepManager) AddOrUpdateDependencies(ctx context.Context, dependencys []*discovery.ConsumerDependency, override bool) (*discovery.Response, error) {
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

		consumerID, err := GetServiceID(ctx, consumerInfo)
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
			insertRes, err := client.GetMongoClient().Insert(ctx, model.CollectionDep, data)
			if err != nil {
				log.Error("failed to insert dep to mongodb", err)
				return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
			}
			log.Info(fmt.Sprintf("insert dep to mongodb success %s", insertRes.InsertedID))
		}
		err = syncDependencyRule(ctx, domainProject, dependency)
		if err != nil {
			return nil, err
		}
	}
	return discovery.CreateResponse(discovery.ResponseSuccess, "Create dependency successfully."), nil
}

func (ds *DepManager) DeleteDependency() {
	panic("implement me")
}

func (ds *DepManager) DependencyHandle(ctx context.Context) (err error) {
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

	oldProviderRules, err := GetOldProviderRules(&dep)
	if err != nil {
		return err
	}

	if r.Override {
		datasource.ParseOverrideRules(ctx, &dep, oldProviderRules)
	} else {
		datasource.ParseAddOrUpdateRules(ctx, &dep, oldProviderRules)
	}
	return updateDeps(domainProject, &dep)
}

func GetOldProviderRules(dep *datasource.Dependency) (*discovery.MicroServiceDependency, error) {
	microServiceDependency := &discovery.MicroServiceDependency{
		Dependency: []*discovery.MicroServiceKey{},
	}
	filter := GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	findRes, err := client.GetMongoClient().FindOne(context.TODO(), model.CollectionDep, filter)
	if err != nil {
		log.Error(fmt.Sprintf("get dependency rule [%v] failed", filter), err)
		return nil, err
	}

	if findRes.Err() != nil {
		return microServiceDependency, nil
	}

	var depRule *model.DependencyRule
	err = findRes.Decode(&depRule)
	if err != nil {
		return nil, err
	}
	return depRule.Dep, nil
}

func updateDeps(domainProject string, dep *datasource.Dependency) error {
	var upsert = true
	for _, r := range dep.DeleteDependencyRuleList {
		filter := GenerateProviderDependencyRuleKey(domainProject, r)
		_, err := client.GetMongoClient().Update(context.TODO(), model.CollectionDep, filter, bson.M{"$pull": bson.M{mutil.ConnectWithDot([]string{model.ColumnDep, model.ColumnDependency}): dep.Consumer}})
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	for _, r := range dep.CreateDependencyRuleList {
		filter := GenerateProviderDependencyRuleKey(domainProject, r)
		data := bson.M{
			"$addToSet": bson.M{mutil.ConnectWithDot([]string{model.ColumnDep, model.ColumnDependency}): dep.Consumer},
		}
		_, err := client.GetMongoClient().Update(context.TODO(), model.CollectionDep, filter, data, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	filter := GenerateConsumerDependencyRuleKey(domainProject, dep.Consumer)
	if len(dep.ProvidersRule) == 0 {
		_, err := client.GetMongoClient().Delete(context.TODO(), model.CollectionDep, filter)
		if err != nil {
			return err
		}
	} else {
		updateData := bson.M{
			"$set": bson.M{mutil.ConnectWithDot([]string{model.ColumnDep, model.ColumnDependency}): dep.ProvidersRule},
		}
		_, err := client.GetMongoClient().Update(context.TODO(), model.CollectionDep, filter, updateData, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
	}

	err := CleanUpDepRules(context.TODO(), domainProject)
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
	filter := GenerateConsumerDependencyRuleKey(domainProject, nil)
	depRules, err := GetDepRules(ctx, filter)
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
	filter := GenerateProviderDependencyRuleKey(domainProject, nil)
	depRules, err := GetDepRules(ctx, filter)
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

func GetDepRules(ctx context.Context, filter bson.M) ([]*model.DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter)
	if err != nil {
		return nil, err
	}

	var depRules []*model.DependencyRule
	for findRes.Next(ctx) {
		var depRule *model.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, depRule)
	}
	return depRules, nil
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

		_, exist, err = FindServiceIds(ctx, depRule.ServiceKey.Version, depRule.ServiceKey)
		if err != nil {
			return err
		}

		cache[id] = exist
	}

	filter := bson.M{
		model.ColumnServiceKey: depRule.ServiceKey,
	}
	if !exist {
		_, err = client.GetMongoClient().DocDelete(ctx, model.CollectionDep, filter)
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
			_, exist, err = FindServiceIds(ctx, key.Version, key)
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

	filter := bson.M{
		model.ColumnServiceKey: depRule.ServiceKey,
	}
	if len(left) == 0 {
		_, err = client.GetMongoClient().DocDelete(ctx, model.CollectionDep, filter)
	} else {
		updateData := bson.M{
			"$set": bson.M{mutil.ConnectWithDot([]string{model.ColumnDep, model.ColumnDependency}): left},
		}
		_, err = client.GetMongoClient().Update(ctx, model.CollectionDep, filter, updateData)
	}
	if err != nil {
		return err
	}
	return nil
}

func TransferToMicroServiceDependency(ctx context.Context, filter bson.M) (*discovery.MicroServiceDependency, error) {
	microServiceDependency := &discovery.MicroServiceDependency{
		Dependency: []*discovery.MicroServiceKey{},
	}
	findRes, err := client.GetMongoClient().FindOne(context.TODO(), model.CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() == nil {
		var depRule *model.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		microServiceDependency.Dependency = append(microServiceDependency.Dependency, depRule.Dep.Dependency...)
		return microServiceDependency, nil
	}
	return microServiceDependency, nil
}

func GetConsumerDepInfo(ctx context.Context, filter bson.M) ([]*discovery.ConsumerDependency, error) {
	var ConsumerDeps []*discovery.ConsumerDependency

	findRes, err := client.GetMongoClient().Find(context.TODO(), model.CollectionDep, filter)
	if err != nil {
		return nil, err
	}

	for findRes.Next(ctx) {
		var dep *model.ConsumerDep
		err = findRes.Decode(&dep)
		if err != nil {
			return nil, err
		}
		ConsumerDeps = append(ConsumerDeps, dep.ConsumerDep)
	}
	return ConsumerDeps, nil
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
	filter := mutil.NewBasicFilter(
		ctx,
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		mutil.ServiceServiceName(key.ServiceName),
		mutil.ServiceVersion(key.Version),
	)
	id, err := getServiceID(ctx, filter)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return "", err
	}
	if len(id) == 0 && len(key.Alias) != 0 {
		filter = mutil.NewBasicFilter(
			ctx,
			mutil.ServiceEnv(key.Environment),
			mutil.ServiceAppID(key.AppId),
			mutil.ServiceAlias(key.Alias),
			mutil.ServiceVersion(key.Version),
		)
		return getServiceID(ctx, filter)
	}
	return id, nil
}

func getServiceID(ctx context.Context, filter bson.M) (serviceID string, err error) {
	svc, err := dao.GetService(ctx, filter)
	if err != nil {
		return
	}
	if svc != nil {
		serviceID = svc.Service.ServiceId
		return
	}
	return
}

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

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/validate"
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
			filter := util.NewBasicFilter(dr.ctx, util.ServiceServiceID(providerID))
			provider, err := dao.GetService(dr.ctx, filter)
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
		return nil, util.ErrInvalidConsumer
	}
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		log.Error(fmt.Sprintf("get consumers that depend on all services failed, %s", dr.provider.ServiceId), err)
		return nil, err
	}
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependList, err := dr.GetConsumerOfSameServiceNameAndAppID(providerService)
	if err != nil {
		log.Errorf(err, "get consumers that depend on rule[%s/%s/%s/%s] failed",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return nil, err
	}
	consumerDependAllList = append(consumerDependAllList, consumerDependList...)
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) GetConsumerOfSameServiceNameAndAppID(provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	filter := GenerateRuleKeyWithSameServiceNameAndAppID(path.DepsProvider, dr.domainProject, provider)
	provider.Version = providerVersion
	depRules, err := getServiceKeysInDep(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	var allConsumers []*pb.MicroServiceKey
	var latestServiceID []string

	for _, depRule := range depRules {
		providerVersionRule := depRule.ServiceKey.Version
		if providerVersionRule == "latest" {
			if latestServiceID == nil {
				latestServiceID, _, err = FindServiceIds(dr.ctx, providerVersionRule, provider)
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
			if !VersionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}
		if len(depRule.Dep.Dependency) > 0 {
			allConsumers = append(allConsumers, depRule.Dep.Dependency...)
		}
	}
	return allConsumers, nil
}

// not prepare for latest scene, should merge it with find serviceids func.
func VersionMatchRule(version, versionRule string) bool {
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

func (dr *DependencyRelation) GetServiceByMicroServiceKey(service *pb.MicroServiceKey) (*pb.MicroService, error) {
	filter, err := MicroServiceKeyFilter(service)
	if err != nil {
		log.Error("get serivce failed", err)
		return nil, err
	}
	findRes, err := client.GetMongoClient().Find(dr.ctx, model.CollectionService, filter)
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

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*pb.MicroServiceKey, error) {
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	filter := GenerateProviderDependencyRuleKey(dr.domainProject, providerService)
	findRes, err := client.GetMongoClient().Find(dr.ctx, model.CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}

	var msKeys []*pb.MicroServiceKey
	for findRes.Next(dr.ctx) {
		var depRule *model.DependencyRule
		err = findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		msKeys = append(msKeys, depRule.ServiceKey)
	}
	return msKeys, nil
}

func getServiceKeysInDep(ctx context.Context, filter interface{}) ([]*model.DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionDep, filter)
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

func (dr *DependencyRelation) getProviderKeys() ([]*pb.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, util.ErrInvalidConsumer
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)
	filter := GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)

	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, filter)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *pb.MicroServiceKey) (serviceIDs []string, err error) {
	switch {
	case dependencyRule.ServiceName == "*":
		log.Info(fmt.Sprintf("service[%s/%s/%s/%s] rely all service",
			dr.consumer.Environment, dr.consumer.AppId, dr.consumer.ServiceName, dr.consumer.Version))
		filter, err := RelyAllServiceKey(dependencyRule)
		if err != nil {
			log.Error("get serivce failed", err)
			return nil, err
		}
		findRes, err := client.GetMongoClient().Find(dr.ctx, model.CollectionService, filter)
		if err != nil {
			return nil, err
		}
		for findRes.Next(dr.ctx) {
			var service model.Service
			err = findRes.Decode(&service)
			if err != nil {
				return nil, err
			}
			serviceIDs = append(serviceIDs, service.Service.ServiceId)
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

func MicroServiceKeyFilter(key *pb.MicroServiceKey) (bson.M, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, util.ErrInvalidDomainProject
	}
	filter := util.NewDomainProjectFilter(tenant[0], tenant[1],
		util.ServiceEnv(key.Environment),
		util.ServiceAppID(key.AppId),
		util.ServiceAlias(key.Alias),
		util.ServiceVersion(key.Version),
	)
	return filter, nil
}

func RelyAllServiceKey(key *pb.MicroServiceKey) (bson.M, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, util.ErrInvalidDomainProject
	}
	filter := util.NewDomainProjectFilter(tenant[0], tenant[1], util.ServiceEnv(key.Environment))
	return filter, nil
}

func FindServiceIds(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, bool, error) {
	if len(versionRule) == 0 {
		return nil, false, nil
	}

	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, false, util.ErrInvalidDomainProject
	}

	baseFilter := bson.D{
		{Key: model.ColumnDomain, Value: tenant[0]},
		{Key: model.ColumnProject, Value: tenant[1]},
		{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnEnv}), Value: key.Environment},
		{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnAppID}), Value: key.AppId}}

	serviceIds, exist, err := findServiceKeysByServiceName(ctx, versionRule, key, baseFilter)
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
		serviceIds, exist, err = findServiceKeysByAlias(ctx, versionRule, key, baseFilter)
		if err != nil {
			return nil, false, err
		}
		return serviceIds, exist, nil
	}
	return serviceIds, exist, nil
}

func serviceVersionFilter(ctx context.Context, versionRule string, filter bson.D) ([]string, bool, error) {
	baseExist, err := client.GetMongoClient().DocExist(ctx, model.CollectionService, filter)
	if err != nil || !baseExist {
		return nil, false, err
	}
	filterFunc, newFilter := findServiceKeys(ctx, versionRule, filter)
	if filterFunc == nil {
		//精确匹配,无version返回服务不存在而不是verison匹配错误
		ids, err := GetVersionService(ctx, newFilter)
		if err != nil || len(ids) == 0 {
			return nil, false, err
		}
		return ids, true, nil
	}

	ids, err := filterFunc(ctx, newFilter)
	if err != nil {
		return nil, false, err
	}
	return ids, true, nil
}

func findServiceKeysByServiceName(ctx context.Context, versionRule string, key *pb.MicroServiceKey, baseFilter bson.D) ([]string, bool, error) {
	filter := append(baseFilter,
		bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceName}), Value: key.ServiceName})
	return serviceVersionFilter(ctx, versionRule, filter)
}

func findServiceKeysByAlias(ctx context.Context, versionRule string, key *pb.MicroServiceKey, baseFilter bson.D) ([]string, bool, error) {
	filter := append(baseFilter,
		bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnAlias}), Value: key.Alias})
	return serviceVersionFilter(ctx, versionRule, filter)
}

type ServiceVersionFilter func(ctx context.Context, filter bson.D) ([]string, error)

func findServiceKeys(ctx context.Context, versionRule string, filter bson.D) (filterFunc ServiceVersionFilter, newFilter bson.D) {
	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		return GetVersionServiceLatest, filter
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter = append(filter, bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: bson.M{"$gte": start}})
		return GetVersionService, filter
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter = append(filter, bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: bson.M{"$gte": start, "$lt": end}})
		return GetVersionService, filter
	default:
		filter = append(filter, bson.E{Key: util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: versionRule})
		return nil, filter
	}
}

func GetVersionServiceLatest(ctx context.Context, m bson.D) (serviceIds []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, m,
		&options.FindOptions{
			Sort: bson.M{util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}): -1}})
	if err != nil {
		return nil, err
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
		if serviceIds != nil {
			return
		}
	}
	return
}

func GetVersionService(ctx context.Context, m bson.D) (serviceIds []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, m, &options.FindOptions{
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

func ParseVersionRule(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, util.ErrInvalidDomainProject
	}
	if len(versionRule) == 0 {
		return nil, nil
	}

	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		filter := util.NewDomainProjectFilter(tenant[0], tenant[1])
		return GetFilterVersionServiceLatest(ctx, filter)
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter := util.NewDomainProjectFilter(tenant[0], tenant[1], util.ServiceVersion(bson.M{"$gte": start}))
		return GetFilterVersionService(ctx, filter)
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter := util.NewDomainProjectFilter(tenant[0], tenant[1], util.ServiceVersion(bson.M{"$gte": start, "$lte": end}))
		return GetFilterVersionService(ctx, filter)
	default:
		return nil, nil
	}
}

func GetFilterVersionService(ctx context.Context, m bson.M) (serviceIDs []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, m)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}
	for findRes.Next(ctx) {
		var service model.Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return
}

func GetFilterVersionServiceLatest(ctx context.Context, m bson.M) (serviceIDs []string, err error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, m,
		&options.FindOptions{
			Sort: bson.M{util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}): -1}})
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		return nil, findRes.Err()
	}
	for findRes.Next(ctx) {
		var service model.Service
		err = findRes.Decode(&service)
		if err != nil {
			return nil, err
		}
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
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

func GenerateConsumerDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(path.DepsConsumer, domainProject, in)
}

func GenerateProviderDependencyRuleKey(domainProject string, in *pb.MicroServiceKey) bson.M {
	return GenerateServiceDependencyRuleKey(path.DepsProvider, domainProject, in)
}

func GenerateRuleKeyWithSameServiceNameAndAppID(serviceType string, domainProject string, in *pb.MicroServiceKey) bson.M {
	return util.NewFilter(
		util.ServiceType(serviceType),
		util.ServiceKeyTenant(domainProject),
		util.ServiceKeyAppID(in.AppId),
		util.ServiceKeyServiceName(in.ServiceName),
	)
}

func GenerateServiceDependencyRuleKey(serviceType string, domainProject string, in *pb.MicroServiceKey) bson.M {
	if in == nil {
		return util.NewFilter(
			util.ServiceType(serviceType),
			util.ServiceKeyTenant(domainProject),
		)
	}
	if in.ServiceName == "*" {
		return util.NewFilter(
			util.ServiceType(serviceType),
			util.ServiceKeyTenant(domainProject),
			util.ServiceKeyServiceEnv(in.Environment),
			util.ServiceKeyServiceName(in.ServiceName),
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

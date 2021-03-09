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

package util

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// servicesBasicFilter query services with domain, project, env, appID, serviceName, alias
func ServicesBasicFilter(ctx context.Context, key *discovery.MicroServiceKey) ([]*db.Service, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, errors.New("invalid 'domain' or 'project'")
	}
	filter := bson.M{
		db.ColumnDomain:  tenant[0],
		db.ColumnProject: tenant[1],
		StringBuilder([]string{db.ColumnService, db.ColumnEnv}):         key.Environment,
		StringBuilder([]string{db.ColumnService, db.ColumnAppID}):       key.AppId,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceName}): key.ServiceName,
		StringBuilder([]string{db.ColumnService, db.ColumnAlias}):       key.Alias,
	}
	rangeIdx := strings.Index(key.Version, "-")
	// if the version number is clear, need to add the version number to query
	switch {
	case key.Version == "latest":
		return GetServices(ctx, filter)
	case len(key.Version) > 0 && key.Version[len(key.Version)-1:] == "+":
		return GetServices(ctx, filter)
	case rangeIdx > 0:
		return GetServices(ctx, filter)
	default:
		filter[StringBuilder([]string{db.ColumnService, db.ColumnVersion})] = key.Version
		return GetServices(ctx, filter)
	}
}

func FindServices(ctx context.Context, key *discovery.MicroServiceKey) ([]*db.Service, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, errors.New("invalid 'domain' or 'project'")
	}
	rangeIdx := strings.Index(key.Version, "-")
	filter := bson.M{
		db.ColumnDomain:  tenant[0],
		db.ColumnProject: tenant[1],
		StringBuilder([]string{db.ColumnService, db.ColumnEnv}):         key.Environment,
		StringBuilder([]string{db.ColumnService, db.ColumnAppID}):       key.AppId,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceName}): key.ServiceName,
		StringBuilder([]string{db.ColumnService, db.ColumnAlias}):       key.Alias,
	}
	switch {
	case key.Version == "latest":
		return LatestServicesFilter(ctx, filter)
	case len(key.Version) > 0 && key.Version[len(key.Version)-1:] == "+":
		start := key.Version[:len(key.Version)-1]
		filter[StringBuilder([]string{db.ColumnService, db.ColumnVersion})] = bson.M{"$gte": start}
		return GetServices(ctx, filter)
	case rangeIdx > 0:
		start := key.Version[:rangeIdx]
		end := key.Version[rangeIdx+1:]
		filter[StringBuilder([]string{db.ColumnService, db.ColumnVersion})] = bson.M{"$gte": start, "$lte": end}
		return GetServices(ctx, filter)
	default:
		filter[StringBuilder([]string{db.ColumnService, db.ColumnVersion})] = key.Version
		return GetServices(ctx, filter)
	}
}

func GetService(ctx context.Context, filter bson.M) (*db.Service, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var svc *db.Service
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, datasource.ErrNoData
	}
	err = findRes.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func GetMicroServices(ctx context.Context, filter bson.M) ([]*discovery.MicroService, error) {
	res, err := client.GetMongoClient().Find(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var services []*discovery.MicroService
	for res.Next(ctx) {
		var tmp db.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.Service)
	}
	return services, nil
}

func GetServices(ctx context.Context, filter bson.M) ([]*db.Service, error) {
	resp, err := client.GetMongoClient().Find(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, datasource.ErrNoData
	}
	var services []*db.Service
	for resp.Next(ctx) {
		var service db.Service
		err := resp.Decode(&service)
		if err != nil {
			log.Error("type conversion error", err)
			return nil, err
		}
		services = append(services, &service)
	}
	return services, nil
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
	id, err := getServiceID(ctx, GeneratorServiceNameFilter(ctx, key))
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return "", err
	}
	if len(id) == 0 && len(key.Alias) != 0 {
		return getServiceID(ctx, GeneratorServiceAliasFilter(ctx, key))
	}
	return id, nil
}

func UpdateService(ctx context.Context, filter interface{}, m bson.M) error {
	return client.GetMongoClient().DocUpdate(ctx, db.CollectionService, filter, m)
}

func getServiceID(ctx context.Context, filter bson.M) (serviceID string, err error) {
	svc, err := GetService(ctx, filter)
	if err != nil {
		return
	}
	if svc != nil {
		serviceID = svc.Service.ServiceId
		return
	}
	return
}

func FilterServiceIDs(ctx context.Context, consumerID string, tags []string, services []*db.Service) []string {
	var filterService []*db.Service
	var serviceIDs []string
	if len(services) == 0 {
		return serviceIDs
	}
	filterService = FilterServicesByTags(services, tags)
	filterService = AccessibleFilter(ctx, consumerID, filterService)
	for _, service := range filterService {
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return serviceIDs
}

func GeneratorServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceID}): serviceID}
}

func GeneratorTargetServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseTargetDomain(ctx)
	project := util.ParseTargetProject(ctx)

	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceID}): serviceID}
}

func GeneratorServiceNameFilter(ctx context.Context, service *discovery.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnEnv}):         service.Environment,
		StringBuilder([]string{db.ColumnService, db.ColumnAppID}):       service.AppId,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceName}): service.ServiceName,
		StringBuilder([]string{db.ColumnService, db.ColumnVersion}):     service.Version}
}

func GeneratorServiceAliasFilter(ctx context.Context, service *discovery.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnEnv}):     service.Environment,
		StringBuilder([]string{db.ColumnService, db.ColumnAppID}):   service.AppId,
		StringBuilder([]string{db.ColumnService, db.ColumnAlias}):   service.Alias,
		StringBuilder([]string{db.ColumnService, db.ColumnVersion}): service.Version}
}

func GetServiceByID(ctx context.Context, serviceID string) (*db.Service, error) {
	cacheService, ok := sd.Store().Service().Cache().Get(serviceID).(db.Service)
	if !ok {
		//no service in cache,get it from mongodb
		return GetService(ctx, GeneratorServiceFilter(ctx, serviceID))
	}
	return cacheToService(cacheService), nil
}

func cacheToService(service db.Service) *db.Service {
	return &db.Service{
		Domain:  service.Domain,
		Project: service.Project,
		Tags:    service.Tags,
		Service: service.Service,
	}
}

func ShouldClearService(ctx context.Context, timeLimitStamp string, svc *discovery.MicroService) (bool, error) {
	if svc.Timestamp > timeLimitStamp {
		return false, nil
	}

	getInstsReq := &discovery.GetInstancesRequest{
		ConsumerServiceId: svc.ServiceId,
		ProviderServiceId: svc.ServiceId,
	}

	getInstsResp, err := datasource.Instance().GetInstances(ctx, getInstsReq)
	if err != nil {
		return false, err
	}
	if getInstsResp.Response.GetCode() != discovery.ResponseSuccess {
		return false, NewError("get instance failed: ", getInstsResp.Response.GetMessage())
	}
	//ignore a service if it has instances
	if len(getInstsResp.Instances) > 0 {
		return false, nil
	}
	return true, nil
}

func GeneratorServiceVersionsFilter(ctx context.Context, service *discovery.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnEnv}):         service.Environment,
		StringBuilder([]string{db.ColumnService, db.ColumnAppID}):       service.AppId,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceName}): service.ServiceName}
}

func LatestServicesFilter(ctx context.Context, filter bson.M) ([]*db.Service, error) {
	resp, err := client.GetMongoClient().Find(ctx, db.CollectionService, filter, &options.FindOptions{
		Sort: bson.M{StringBuilder([]string{db.ColumnService, db.ColumnVersion}): -1}})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("no related services were found")
	}
	var services []*db.Service
	for resp.Next(ctx) {
		var service db.Service
		err := resp.Decode(&service)
		if err != nil {
			log.Error("type conversion error", err)
			return nil, err
		}
		services = append(services, &service)
		if services != nil {
			return services, nil
		}
	}
	return services, nil
}

func GetServiceDetailUtil(ctx context.Context, mgs *db.Service, countOnly bool, options []string) (*discovery.ServiceDetail, error) {
	serviceDetail := new(discovery.ServiceDetail)
	serviceID := mgs.Service.ServiceId
	domainProject := util.ParseDomainProject(ctx)
	if countOnly {
		serviceDetail.Statics = new(discovery.Statistics)
	}
	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			serviceDetail.Tags = mgs.Tags
		case "rules":
			rules, err := GetRules(ctx, mgs.Service.ServiceId)
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all rules failed", mgs.Service.ServiceId), err)
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			if countOnly {
				instanceCount, err := GetInstanceCountOfOneService(ctx, serviceID)
				if err != nil {
					log.Error(fmt.Sprintf("get number of service [%s]'s instances failed", serviceID), err)
					return nil, err
				}
				serviceDetail.Statics.Instances = &discovery.StInstance{
					Count: instanceCount,
				}
				continue
			}
			instances, err := GetAllInstancesOfOneService(ctx, serviceID)
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all instances failed", serviceID), err)
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			schemas, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, mgs.Service.ServiceId))
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all schemas failed", mgs.Service.ServiceId), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			service := mgs.Service
			dr := NewDependencyRelation(ctx, domainProject, service, service)
			consumers, err := dr.GetDependencyConsumers(WithoutSelfDependency(), WithSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all consumers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
			}
			providers, err := dr.GetDependencyProviders(WithoutSelfDependency(), WithSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all providers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
				return nil, err
			}
			serviceDetail.Consumers = consumers
			serviceDetail.Providers = providers
		case "":
			continue
		default:
			log.Info(fmt.Sprintf("request option %s is invalid", opt))
		}
	}
	return serviceDetail, nil
}

func GetServicesVersions(ctx context.Context, filter interface{}) ([]string, error) {
	res, err := client.GetMongoClient().Find(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, nil
	}
	var versions []string
	for res.Next(ctx) {
		var tmp db.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		versions = append(versions, tmp.Service.Version)
	}
	return versions, nil
}

func ServiceExist(ctx context.Context, service *discovery.MicroServiceKey) (bool, error) {
	filter := GeneratorServiceNameFilter(ctx, service)
	return client.GetMongoClient().DocExist(ctx, db.CollectionService, filter)
}

func ServiceExistID(ctx context.Context, serviceID string) (bool, error) {
	filter := GeneratorServiceFilter(ctx, serviceID)
	return client.GetMongoClient().DocExist(ctx, db.CollectionService, filter)
}

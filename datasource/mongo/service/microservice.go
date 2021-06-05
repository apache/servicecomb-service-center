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
	"strconv"
	"strings"
	"time"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
)

const baseTen = 10

func (ds *DataSource) RegisterService(ctx context.Context, request *discovery.CreateServiceRequest) (*discovery.CreateServiceResponse, error) {
	service := request.Service
	remoteIP := util.GetIPFromContext(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	serviceFlag := util.StringJoin([]string{service.Environment, service.AppId, service.ServiceName, service.Version}, "/")
	requestServiceID := service.ServiceId

	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, util.StringJoin([]string{domain, project, service.Environment, service.AppId, service.ServiceName, service.Alias, service.Version}, "/"))
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), baseTen)
	service.ModTimestamp = service.Timestamp
	// the service unique index in table is (serviceId/serviceEnv,serviceAppid,servicename,serviceVersion)
	if len(service.Alias) != 0 {
		serviceID, err := getServiceID(ctx, &discovery.MicroServiceKey{
			Environment: service.Environment,
			AppId:       service.AppId,
			ServiceName: service.ServiceName,
			Version:     service.Version,
			Alias:       service.Alias,
		})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			return &discovery.CreateServiceResponse{
				Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()),
			}, err
		}
		if len(serviceID) != 0 {
			if len(requestServiceID) != 0 && requestServiceID != serviceID {
				log.Warn(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s", serviceFlag, remoteIP))
				return &discovery.CreateServiceResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceAlreadyExists, "ServiceID conflict or found the same service with different id."),
				}, nil
			}
			return &discovery.CreateServiceResponse{
				Response:  discovery.CreateResponse(discovery.ResponseSuccess, "register service successfully"),
				ServiceId: serviceID,
			}, nil
		}
	}
	err := insertService(ctx, &model.Service{Domain: domain, Project: project, Service: service})
	if err != nil {
		if mutil.IsDuplicateKey(err) {
			serviceIDInner, err := getServiceID(ctx, &discovery.MicroServiceKey{
				Environment: service.Environment,
				AppId:       service.AppId,
				ServiceName: service.ServiceName,
				Version:     service.Version,
			})
			if err != nil && !errors.Is(err, datasource.ErrNoData) {
				return &discovery.CreateServiceResponse{
					Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()),
				}, err
			}
			// serviceid conflict with the service in the database
			if len(requestServiceID) != 0 && serviceIDInner != requestServiceID {
				return &discovery.CreateServiceResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceAlreadyExists,
						"ServiceID conflict or found the same service with different id."),
				}, nil
			}
			return &discovery.CreateServiceResponse{
				Response:  discovery.CreateResponse(discovery.ResponseSuccess, "register service successfully"),
				ServiceId: serviceIDInner,
			}, nil
		}
		log.Error(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s",
			serviceFlag, remoteIP), err)
		return &discovery.CreateServiceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}

	log.Info(fmt.Sprintf("create micro-service[%s] successfully,operator: %s", service.ServiceId, remoteIP))

	return &discovery.CreateServiceResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "Register service successfully"),
		ServiceId: service.ServiceId,
	}, nil
}

func (ds *DataSource) GetServices(ctx context.Context, request *discovery.GetServicesRequest) (*discovery.GetServicesResponse, error) {
	filter := mutil.NewBasicFilter(ctx)
	services, err := findMicroServices(ctx, filter)
	if err != nil {
		return &discovery.GetServicesResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "get services data failed."),
		}, nil
	}

	return &discovery.GetServicesResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *discovery.GetServiceRequest) (*discovery.GetServiceResponse, error) {
	svc, ok := cache.GetServiceByID(request.ServiceId)
	if !ok {
		var err error
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
		svc, err = findService(ctx, filter)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
				return &discovery.GetServiceResponse{
					Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service not exist."),
				}, nil
			}
			log.Error(fmt.Sprintf("failed to get single service %s from mongo", request.ServiceId), err)
			return &discovery.GetServiceResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, "get service data from mongodb failed."),
			}, err
		}
	}
	return &discovery.GetServiceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get service successfully."),
		Service:  svc.Service,
	}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *discovery.GetExistenceByIDRequest) (*discovery.GetExistenceByIDResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.GetExistenceByIDResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "Check service exist failed."),
			Exist:    false,
		}, err
	}

	return &discovery.GetExistenceByIDResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Check ExistService successfully."),
		Exist:    exist,
	}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *discovery.GetExistenceRequest) (*discovery.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, "/")

	ids, exist, err := findServiceIDs(ctx, request.Version, &discovery.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	})
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] exist failed, find serviceIDs failed", serviceFlag), err)
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, service does not exist", serviceFlag))
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, serviceFlag+" does not exist."),
		}, nil
	}
	if len(ids) == 0 {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, version mismatch", serviceFlag))
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceVersionNotExists, serviceFlag+" version mismatch."),
		}, nil
	}
	return &discovery.GetExistenceResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "get service id successfully."),
		ServiceId: ids[0], // 约定多个时，取较新版本
	}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *discovery.DeleteServiceRequest) (*discovery.DeleteServiceResponse, error) {
	res, err := deleteService(ctx, request.ServiceId, request.Force)
	if err != nil {
		return &discovery.DeleteServiceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "delete service failed"),
		}, err
	}
	return &discovery.DeleteServiceResponse{
		Response: res,
	}, nil
}

func deleteService(ctx context.Context, serviceID string, force bool) (*discovery.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		log.Error(fmt.Sprintf("%s micro-service %s failed, operator: %s", title, serviceID, remoteIP), mutil.ErrNotAllowDeleteSC)
		return discovery.CreateResponse(discovery.ErrInvalidParams, mutil.ErrNotAllowDeleteSC.Error()), nil
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(serviceID))
	microservice, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("%s micro-service %s failed, service does not exist, operator: %s",
				title, serviceID, remoteIP))
			return discovery.CreateResponse(discovery.ErrServiceNotExists, "service does not exist."), nil
		}
		log.Error(fmt.Sprintf("%s micro-service %s failed, get service file failed, operator: %s",
			title, serviceID, remoteIP), err)
		return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
	}
	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := NewProviderDependencyRelation(ctx, util.ParseDomainProject(ctx), microservice.Service)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP), err)
			return discovery.CreateResponse(discovery.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP), err)
			return discovery.CreateResponse(discovery.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}
		//todo wait for dep interface
		instancesExist, err := client.GetMongoClient().DocExist(ctx, model.CollectionInstance, bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}): serviceID})
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get instances number failed, operator: %s",
				serviceID, remoteIP), err)
			return discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()), err
		}
		if instancesExist {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, service deployed instances, operator: %s",
				serviceID, remoteIP), nil)
			return discovery.CreateResponse(discovery.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}

	}

	schemaOps := client.MongoOperation{Table: model.CollectionSchema, Models: []mongo.WriteModel{mongo.NewDeleteManyModel().SetFilter(bson.M{model.ColumnServiceID: serviceID})}}
	rulesOps := client.MongoOperation{Table: model.CollectionRule, Models: []mongo.WriteModel{mongo.NewDeleteManyModel().SetFilter(bson.M{model.ColumnServiceID: serviceID})}}
	instanceOps := client.MongoOperation{Table: model.CollectionInstance, Models: []mongo.WriteModel{mongo.NewDeleteManyModel().SetFilter(bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}): serviceID})}}
	serviceOps := client.MongoOperation{Table: model.CollectionService, Models: []mongo.WriteModel{mongo.NewDeleteOneModel().SetFilter(bson.M{mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceID}): serviceID})}}

	err = client.GetMongoClient().MultiTableBatchUpdate(ctx, []client.MongoOperation{schemaOps, rulesOps, instanceOps, serviceOps})
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()), err
	}

	domainProject := util.ToDomainProject(microservice.Domain, microservice.Project)
	serviceKey := &discovery.MicroServiceKey{
		Tenant:      domainProject,
		Environment: microservice.Service.Environment,
		AppId:       microservice.Service.AppId,
		ServiceName: microservice.Service.ServiceName,
		Version:     microservice.Service.Version,
		Alias:       microservice.Service.Alias,
	}

	err = deleteDependencyByService(ctx, domainProject, serviceKey)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.CreateResponse(discovery.ErrUnavailableBackend, err.Error()), err
	}

	return discovery.CreateResponse(discovery.ResponseSuccess, "Unregister service successfully."), nil
}

func (ds *DataSource) UpdateService(ctx context.Context, request *discovery.UpdateServicePropsRequest) (*discovery.UpdateServicePropsResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	setValue := mutil.NewFilter(
		mutil.ServiceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.ServiceProperty(request.Properties),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setValue),
	)
	err := updateService(ctx, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("update service %s properties failed, update mongo failed", request.ServiceId), err)
		return &discovery.UpdateServicePropsResponse{
			Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, "update doc in mongo failed."),
		}, nil
	}
	return &discovery.UpdateServicePropsResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service successfully."),
	}, nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool, serviceRespChan chan<- *discovery.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &discovery.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		resp, err := deleteService(ctx, serviceID, force)
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.GetCode() != discovery.ResponseSuccess {
			serviceRst.ErrMessage = resp.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *discovery.GetServiceRequest) (*discovery.GetServiceDetailResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	mgSvc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return &discovery.GetServiceDetailResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "Service does not exist."),
			}, nil
		}
		return &discovery.GetServiceDetailResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	svc := mgSvc.Service
	key := &discovery.MicroServiceKey{
		Environment: svc.Environment,
		AppId:       svc.AppId,
		ServiceName: svc.ServiceName,
	}
	filter = mutil.NewBasicFilter(ctx, mutil.ServiceEnv(key.Environment), mutil.ServiceAppID(key.AppId), mutil.ServiceServiceName(key.ServiceName))
	versions, err := getServicesVersions(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("get service %s %s %s all versions failed", svc.Environment, svc.AppId, svc.ServiceName), err)
		return &discovery.GetServiceDetailResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}
	serviceInfo, err := getServiceDetail(ctx, mgSvc, false, options)
	if err != nil {
		return &discovery.GetServiceDetailResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	serviceInfo.MicroService = svc
	serviceInfo.MicroServiceVersions = versions
	return &discovery.GetServiceDetailResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Get service successfully"),
		Service:  serviceInfo,
	}, nil

}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *discovery.GetServicesInfoRequest) (*discovery.GetServicesInfoResponse, error) {
	optionMap := make(map[string]struct{}, len(request.Options))
	for _, opt := range request.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = []string{"tags", "rules", "instances", "schemas", "dependencies"}
	} else {
		for opt := range optionMap {
			options = append(options, opt)
		}
	}
	var st *discovery.Statistics
	if _, ok := optionMap["statistics"]; ok {
		var err error
		st, err = statistics(ctx, request.WithShared)
		if err != nil {
			return &discovery.GetServicesInfoResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
		if len(optionMap) == 1 {
			return &discovery.GetServicesInfoResponse{
				Response:   discovery.CreateResponse(discovery.ResponseSuccess, "Statistics successfully."),
				Statistics: st,
			}, nil
		}
	}
	filters := serverFilter(ctx, request)
	services, err := findServices(ctx, filters)
	if err != nil {
		log.Error("get all services by domain failed", err)
		return &discovery.GetServicesInfoResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	allServiceDetails := make([]*discovery.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, mgSvc := range services {
		if !request.WithShared && apt.IsGlobal(discovery.MicroServiceToKey(domainProject, mgSvc.Service)) {
			continue
		}

		serviceDetail, err := getServiceDetail(ctx, mgSvc, request.CountOnly, options)
		if err != nil {
			return &discovery.GetServicesInfoResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
		serviceDetail.MicroService = mgSvc.Service
		tmpServiceDetail := &discovery.ServiceDetail{}
		err = copier.CopyWithOption(tmpServiceDetail, serviceDetail, copier.Option{DeepCopy: true})
		if err != nil {
			return &discovery.GetServicesInfoResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
		}
		tmpServiceDetail.MicroService.Properties = nil
		tmpServiceDetail.MicroService.Schemas = nil
		instances := tmpServiceDetail.Instances
		for _, instance := range instances {
			instance.Properties = nil
		}
		allServiceDetails = append(allServiceDetails, tmpServiceDetail)
	}

	return &discovery.GetServicesInfoResponse{
		Response:          discovery.CreateResponse(discovery.ResponseSuccess, "Get services info successfully."),
		AllServicesDetail: allServiceDetails,
		Statistics:        nil,
	}, nil
}

func (ds *DataSource) GetApplications(ctx context.Context, request *discovery.GetAppsRequest) (*discovery.GetAppsResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceEnv(request.Environment))
	services, err := findMicroServices(ctx, filter)
	if err != nil {
		return &discovery.GetAppsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "get services data failed."),
		}, nil
	}
	l := len(services)
	if l == 0 {
		return &discovery.GetAppsResponse{
			Response: discovery.CreateResponse(discovery.ResponseSuccess, "get all applications successfully."),
		}, nil
	}
	apps := make([]string, 0, l)
	hash := make(map[string]struct{}, l)
	for _, svc := range services {
		if !request.WithShared && apt.IsGlobal(discovery.MicroServiceToKey(util.ParseDomainProject(ctx), svc)) {
			continue
		}
		if _, ok := hash[svc.AppId]; ok {
			continue
		}
		hash[svc.AppId] = struct{}{}
		apps = append(apps, svc.AppId)
	}
	return &discovery.GetAppsResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get all applications successfully."),
		AppIds:   apps,
	}, nil
}

func (ds *DataSource) GetServicesStatistics(ctx context.Context, request *discovery.GetServicesRequest) (*discovery.GetServicesInfoStatisticsResponse, error) {
	ctx = util.WithCacheOnly(ctx)
	var st *discovery.Statistics
	var err error
	st, err = statistics(ctx, false)
	if err != nil {
		return &discovery.GetServicesInfoStatisticsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	return &discovery.GetServicesInfoStatisticsResponse{
		Response:   discovery.CreateResponse(discovery.ResponseSuccess, "Get services statistics successfully."),
		Statistics: st,
	}, nil
}

func getServiceDetail(ctx context.Context, mgs *model.Service, countOnly bool, options []string) (*discovery.ServiceDetail, error) {
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
			filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(mgs.Service.ServiceId))
			rules, err := getServiceRules(ctx, filter)
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
				filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID))
				instanceCount, err := countInstance(ctx, filter)
				if err != nil {
					log.Error(fmt.Sprintf("get number of service [%s]'s instances failed", serviceID), err)
					return nil, err
				}
				serviceDetail.Statics.Instances = &discovery.StInstance{
					Count: instanceCount,
				}
				continue
			}
			filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID))
			instances, err := findMicroServiceInstances(ctx, filter)
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all instances failed", serviceID), err)
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(serviceID))
			schemas, err := findSchemas(ctx, filter)
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all schemas failed", mgs.Service.ServiceId), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			service := mgs.Service
			dr := NewDependencyRelation(ctx, domainProject, service, service)
			consumers, err := dr.GetDependencyConsumers(withoutSelfDependency(), withSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all consumers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
			}
			providers, err := dr.GetDependencyProviders(withoutSelfDependency(), withSameDomainProject())
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

func serverFilter(ctx context.Context, request *discovery.GetServicesInfoRequest) bson.M {
	var opts []func(filter bson.M)

	if len(request.Environment) > 0 {
		opts = append(opts, mutil.ServiceEnv(request.Environment))
	}
	if len(request.AppId) > 0 {
		opts = append(opts, mutil.ServiceAppID(request.AppId))
	}
	if len(request.ServiceName) > 0 {
		opts = append(opts, mutil.ServiceAppID(request.ServiceName))
	}
	return mutil.NewBasicFilter(ctx, opts...)
}

func statistics(ctx context.Context, withShared bool) (*discovery.Statistics, error) {
	result := &discovery.Statistics{
		Services:  &discovery.StService{},
		Instances: &discovery.StInstance{},
		Apps:      &discovery.StApp{},
	}
	filter := mutil.NewBasicFilter(ctx)

	services, err := findMicroServices(ctx, filter)
	if err != nil {
		return nil, err
	}

	var svcIDs []string
	var svcKeys []*discovery.MicroServiceKey
	for _, svc := range services {
		svcIDs = append(svcIDs, svc.ServiceId)
		svcKeys = append(svcKeys, datasource.TransServiceToKey(util.ParseDomainProject(ctx), svc))
	}
	svcIDToNonVerKey := datasource.SetStaticServices(result, svcKeys, svcIDs, withShared)
	respGetInstanceCountByDomain := make(chan datasource.GetInstanceCountByDomainResponse, 1)
	gopool.Go(func(_ context.Context) {
		getInstanceCountByDomain(ctx, svcIDToNonVerKey, respGetInstanceCountByDomain)
	})

	instances, err := findInstances(ctx, filter)
	if err != nil {
		return nil, err
	}
	var instIDs []string
	for _, inst := range instances {
		instIDs = append(instIDs, inst.Instance.ServiceId)
	}
	datasource.SetStaticInstances(result, svcIDToNonVerKey, instIDs)
	data := <-respGetInstanceCountByDomain
	close(respGetInstanceCountByDomain)
	if data.Err != nil {
		return nil, data.Err
	}
	result.Instances.CountByDomain = data.CountByDomain
	return result, nil
}

func getInstanceCountByDomain(ctx context.Context, svcIDToNonVerKey map[string]string, resp chan datasource.GetInstanceCountByDomainResponse) {
	ret := datasource.GetInstanceCountByDomainResponse{}
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	for sid := range svcIDToNonVerKey {
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(sid))
		num, err := countInstance(ctx, filter)
		if err != nil {
			ret.Err = err
			return
		}
		ret.CountByDomain = ret.CountByDomain + num
	}
	resp <- ret
}

func filterServices(ctx context.Context, key *discovery.MicroServiceKey) ([]*model.Service, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, errors.New("invalid 'domain' or 'project'")
	}
	filter := mutil.NewDomainProjectFilter(tenant[0], tenant[1],
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		mutil.ServiceServiceName(key.ServiceName),
		mutil.ServiceAlias(key.Alias),
	)
	filter = microserviceVersionFilter(key.Version, filter)
	if key.Version == "latest" {
		findOption := &options.FindOptions{Sort: bson.D{{Key: mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: -1}}}
		return findServices(ctx, filter, findOption)
	}
	return findServices(ctx, filter)
}

func findServiceIDs(ctx context.Context, versionRule string, key *discovery.MicroServiceKey) ([]string, bool, error) {
	if len(versionRule) == 0 {
		return nil, false, nil
	}

	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, false, mutil.ErrInvalidDomainProject
	}
	serviceIDs, exist, err := findServiceIDsByServiceName(ctx, versionRule, key)
	if err != nil {
		return nil, false, err
	}
	if len(serviceIDs) == 0 {
		if exist {
			// service exist but version not matched
			return nil, true, nil
		}
		if len(key.Alias) == 0 {
			return nil, false, nil
		}
		serviceIDs, exist, err = findServiceIDsByAlias(ctx, versionRule, key)
		if err != nil {
			return nil, false, err
		}
	}
	return serviceIDs, exist, nil
}

func findServiceIDsByServiceName(ctx context.Context, versionRule string, key *discovery.MicroServiceKey) (serviceIDs []string, exist bool, err error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceEnv(key.Environment), mutil.ServiceAppID(key.AppId), mutil.ServiceServiceName(key.ServiceName))
	baseExist, err := exitService(ctx, filter)
	if err != nil || !baseExist {
		return nil, false, err
	}
	return findServiceIDsByVersionRule(ctx, versionRule, filter)
}

func findServiceIDsByAlias(ctx context.Context, versionRule string, key *discovery.MicroServiceKey) (serviceIDs []string, exist bool, err error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceEnv(key.Environment), mutil.ServiceAppID(key.AppId), mutil.ServiceAlias(key.Alias))
	baseExist, err := exitService(ctx, filter)
	if err != nil || !baseExist {
		return nil, false, err
	}
	return findServiceIDsByVersionRule(ctx, versionRule, filter)
}

func findServiceIDsByVersionRule(ctx context.Context, versionRule string, filter bson.M) (serviceIDs []string, exist bool, err error) {
	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		serviceIDs, err = findLatestServiceID(ctx, filter, &options.FindOptions{Sort: bson.D{{Key: mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}), Value: -1}}})
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = bson.M{"$gte": start}
		serviceIDs, err = findServicesIDs(ctx, filter)
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = bson.M{"$gte": start, "$lt": end}
		serviceIDs, err = findServicesIDs(ctx, filter)
	default:
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = versionRule
		serviceIDs, err = findServicesIDs(ctx, filter)
		// precise search
		if len(serviceIDs) == 0 {
			return nil, false, err
		}
	}
	if err != nil {
		return nil, false, err
	}
	return serviceIDs, true, nil
}

func microserviceVersionFilter(versionRule string, filter bson.M) bson.M {
	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		return filter
	case versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = bson.M{"$gte": start}
		return filter
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = bson.M{"$gte": start, "$lt": end}
		return filter
	default:
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = versionRule
		return filter
	}
}

func filterServiceIDs(ctx context.Context, consumerID string, tags []string, services []*model.Service) []string {
	var filterService []*model.Service
	serviceIDs := make([]string, 0)
	if len(services) == 0 {
		return serviceIDs
	}
	filterService = filterServicesByTags(services, tags)
	filterService = filterServicesByConsumerID(ctx, consumerID, filterService)
	for _, service := range filterService {
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return serviceIDs
}

func filterServicesByTags(services []*model.Service, tags []string) []*model.Service {
	if len(tags) == 0 {
		return services
	}
	var newServices []*model.Service
	for _, service := range services {
		index := 0
		for ; index < len(tags); index++ {
			if _, ok := service.Tags[tags[index]]; !ok {
				break
			}
		}
		if index == len(tags) {
			newServices = append(newServices, service)
		}
	}
	return newServices
}

func filterServicesByConsumerID(ctx context.Context, consumerID string, services []*model.Service) []*model.Service {
	newServices := make([]*model.Service, 0)
	for _, service := range services {
		if err := accessible(ctx, consumerID, service.Service.ServiceId); err != nil {
			findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumerID,
				service.Service.AppId, service.Service.ServiceName, service.Service.Version)
			log.Error(fmt.Sprintf("accessible filter failed, %s", findFlag), err)
			continue
		}
		newServices = append(newServices, service)
	}
	return newServices
}

// filterServicesByBasicConditions query services with domain, project, env, appID, serviceName, alias
func filterServicesByBasicConditions(ctx context.Context, key *discovery.MicroServiceKey) ([]*model.Service, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, errors.New("invalid 'domain' or 'project'")
	}
	filter := mutil.NewDomainProjectFilter(tenant[0], tenant[1],
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		mutil.ServiceServiceName(key.ServiceName),
		mutil.ServiceAlias(key.Alias),
	)
	rangeIdx := strings.Index(key.Version, "-")
	// if the version number is clear, need to add the version number to query
	switch {
	case key.Version == "latest":
		return findServices(ctx, filter)
	case len(key.Version) > 0 && key.Version[len(key.Version)-1:] == "+":
		return findServices(ctx, filter)
	case rangeIdx > 0:
		return findServices(ctx, filter)
	default:
		filter[mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion})] = key.Version
		return findServices(ctx, filter)
	}
}

func accessible(ctx context.Context, consumerID string, providerID string) *errsvc.Error {
	if len(consumerID) == 0 {
		return nil
	}

	consumerDomain, consumerProject := util.ParseDomain(ctx), util.ParseProject(ctx)
	providerDomain, providerProject := util.ParseTargetDomain(ctx), util.ParseTargetProject(ctx)

	filter := mutil.NewDomainProjectFilter(consumerDomain, consumerProject, mutil.ServiceServiceID(consumerID))
	consumerService, err := findService(ctx, filter)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("an error occurred in query consumer(%s)", err.Error()))
	}
	if consumerService == nil {
		return discovery.NewError(discovery.ErrServiceNotExists, "consumer serviceID is invalid")
	}

	filter = mutil.NewDomainProjectFilter(providerDomain, providerProject, mutil.ServiceServiceID(providerID))
	// 跨应用权限
	providerService, err := findService(ctx, filter)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("an error occurred in query provider(%s)", err.Error()))
	}
	if providerService == nil {
		return discovery.NewError(discovery.ErrServiceNotExists, "provider serviceID is invalid")
	}
	err = allowAcrossDimension(ctx, providerService, consumerService)
	if err != nil {
		return discovery.NewError(discovery.ErrPermissionDeny, err.Error())
	}

	// 黑白名单
	filter = mutil.NewDomainProjectFilter(providerDomain, providerProject, mutil.ServiceID(providerID))
	rules, err := findRules(ctx, filter)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("an error occurred in query provider rules(%s)", err.Error()))
	}

	if len(rules) == 0 {
		return nil
	}
	return mutil.MatchRules(rules, consumerService.Service, consumerService.Tags)
}

func allowAcrossDimension(ctx context.Context, providerService *model.Service, consumerService *model.Service) error {
	if providerService.Service.AppId != consumerService.Service.AppId {
		if len(providerService.Service.Properties) == 0 {
			return fmt.Errorf("not allow across app access")
		}

		if allowCrossApp, ok := providerService.Service.Properties[discovery.PropAllowCrossApp]; !ok || strings.ToLower(allowCrossApp) != "true" {
			return fmt.Errorf("not allow across app access")
		}
	}
	if !apt.IsGlobal(discovery.MicroServiceToKey(util.ParseTargetDomainProject(ctx), providerService.Service)) &&
		providerService.Service.Environment != consumerService.Service.Environment {
		return fmt.Errorf("not allow across environment access")
	}
	return nil
}

func getServiceByServiceID(ctx context.Context, serviceID string) (*model.Service, error) {
	return findService(ctx, mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(serviceID)))
}

func deleteDependencyByService(ctx context.Context, domainProject string, service *discovery.MicroServiceKey) error {
	conDep := new(discovery.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*discovery.MicroServiceKey{}
	conDep.Override = true
	err := syncDependencyRule(ctx, domainProject, conDep)
	if err != nil {
		return err
	}
	return nil
}

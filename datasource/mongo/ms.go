/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	dmongo "github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	csync "github.com/go-chassis/cari/sync"
	"github.com/go-chassis/foundation/gopool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sync"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
)

const baseTen = 10

var (
	ErrUndefinedSchemaID    = discovery.NewError(discovery.ErrUndefinedSchemaID, datasource.ErrUndefinedSchemaID.Error())
	ErrModifySchemaNotAllow = discovery.NewError(discovery.ErrModifySchemaNotAllow, datasource.ErrModifySchemaNotAllow.Error())
)

type MetadataManager struct {
	// SchemaNotEditable determines whether schema modification is not allowed
	SchemaNotEditable bool
	// InstanceTTL options
	InstanceTTL        int64
	InstanceProperties map[string]string
}

func (ds *MetadataManager) RegisterService(ctx context.Context, request *discovery.CreateServiceRequest) (*discovery.CreateServiceResponse, error) {
	service := request.Service
	remoteIP := util.GetIPFromContext(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	serviceFlag := util.StringJoin([]string{
		service.Environment, service.AppId, service.ServiceName, service.Version}, "/")
	requestServiceID := service.ServiceId

	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, util.StringJoin([]string{domain, project, service.Environment, service.AppId, service.ServiceName, service.Alias, service.Version}, "/"))
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	service.ModTimestamp = service.Timestamp
	// the service unique index in table is (serviceId/serviceEnv,serviceAppid,servicename,serviceVersion)
	if len(service.Alias) != 0 {
		serviceID, err := GetServiceID(ctx, &discovery.MicroServiceKey{
			Environment: service.Environment,
			AppId:       service.AppId,
			ServiceName: service.ServiceName,
			Version:     service.Version,
			Alias:       service.Alias,
		})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			return nil, discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
		}
		if len(serviceID) != 0 {
			if len(requestServiceID) != 0 && requestServiceID != serviceID {
				log.Warn(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s",
					serviceFlag, remoteIP))
				return nil, discovery.NewError(discovery.ErrServiceAlreadyExists,
					"ServiceID conflict or found the same service with different id.")
			}
			return &discovery.CreateServiceResponse{
				ServiceId: serviceID,
			}, nil
		}
	}
	err := createServiceTxn(ctx, request, domain, project, service)
	if err != nil {
		if dao.IsDuplicateKey(err) {
			serviceIDInner, err := GetServiceID(ctx, &discovery.MicroServiceKey{
				Environment: service.Environment,
				AppId:       service.AppId,
				ServiceName: service.ServiceName,
				Version:     service.Version,
			})
			if err != nil && !errors.Is(err, datasource.ErrNoData) {
				return nil, discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
			}
			// serviceid conflict with the service in the database
			if len(requestServiceID) != 0 && serviceIDInner != requestServiceID {
				return nil, discovery.NewError(discovery.ErrServiceAlreadyExists,
					"ServiceID conflict or found the same service with different id.")
			}
			return &discovery.CreateServiceResponse{
				ServiceId: serviceIDInner,
			}, nil
		}
		log.Error(fmt.Sprintf("create micro-service[%s] failed, service already exists, operator: %s",
			serviceFlag, remoteIP), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return &discovery.CreateServiceResponse{
		ServiceId: service.ServiceId,
	}, nil
}

func createServiceTxn(ctx context.Context, request *discovery.CreateServiceRequest, domain string, project string,
	service *discovery.MicroService) error {
	return dmongo.GetClient().ExecTxn(ctx, func(sessionContext mongo.SessionContext) error {
		_, err := dmongo.GetClient().GetDB().Collection(model.CollectionService).InsertOne(ctx,
			&model.Service{Domain: domain, Project: project, Tags: request.Tags, Service: service})
		if err != nil {
			return err
		}
		return sync.DoCreateOpts(sessionContext, datasource.ResourceService, request)
	})
}

func (ds *MetadataManager) ListService(ctx context.Context, request *discovery.GetServicesRequest) (*discovery.GetServicesResponse, error) {
	services, err := GetAllMicroServicesByDomainProject(ctx)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, "get services data failed.")
	}

	return &discovery.GetServicesResponse{
		Services: services,
	}, nil
}

func (ds *MetadataManager) ListApp(ctx context.Context, request *discovery.GetAppsRequest) (*discovery.GetAppsResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{
		model.ColumnDomain:  domain,
		model.ColumnProject: project,
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnEnv}): request.Environment}

	services, err := dao.GetMicroServices(ctx, filter)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, "get services data failed.")
	}
	l := len(services)
	if l == 0 {
		return &discovery.GetAppsResponse{}, nil
	}
	apps := make([]string, 0, l)
	hash := make(map[string]struct{}, l)
	for _, svc := range services {
		if !request.WithShared && datasource.IsGlobal(discovery.MicroServiceToKey(util.ParseDomainProject(ctx), svc)) {
			continue
		}
		if _, ok := hash[svc.AppId]; ok {
			continue
		}
		hash[svc.AppId] = struct{}{}
		apps = append(apps, svc.AppId)
	}
	return &discovery.GetAppsResponse{
		AppIds: apps,
	}, nil
}

func (ds *MetadataManager) GetService(ctx context.Context, request *discovery.GetServiceRequest) (*discovery.MicroService, error) {
	svc, err := GetServiceByID(ctx, request.ServiceId)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return nil, discovery.NewError(discovery.ErrServiceNotExists, "Service not exist.")
		}
		log.Error(fmt.Sprintf("failed to get single service %s from mongo", request.ServiceId), err)
		return nil, discovery.NewError(discovery.ErrInternal, "get service data from mongodb failed.")
	}
	return svc.Service, nil
}

func (ds *MetadataManager) ExistServiceByID(ctx context.Context, request *discovery.GetExistenceByIDRequest) (*discovery.GetExistenceByIDResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return &discovery.GetExistenceByIDResponse{
		Exist: exist,
	}, nil
}

func (ds *MetadataManager) ExistService(ctx context.Context, request *discovery.GetExistenceRequest) (string, error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, "/")

	ids, exist, err := FindServiceIds(ctx, &discovery.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	}, true)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] exist failed, find serviceIDs failed", serviceFlag), err)
		return "", discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if !exist {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, service does not exist", serviceFlag))
		return "", discovery.NewError(discovery.ErrServiceNotExists, serviceFlag+" does not exist.")
	}
	if len(ids) == 0 {
		log.Info(fmt.Sprintf("micro-service[%s] exist failed, version mismatch", serviceFlag))
		return "", discovery.NewError(discovery.ErrServiceVersionNotExists, serviceFlag+" version mismatch.")
	}
	// 约定多个时，取较新版本
	return ids[0], nil
}

func (ds *MetadataManager) FindService(ctx context.Context, request *discovery.MicroServiceKey) (*discovery.GetServicesResponse, error) {
	filter := mutil.NewBasicFilter(ctx,
		mutil.ServiceEnv(request.Environment),
		mutil.ServiceAppID(request.AppId),
		mutil.ServiceServiceName(request.ServiceName),
	)
	res, err := dmongo.GetClient().GetDB().Collection(model.CollectionService).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := &discovery.GetServicesResponse{}
	for res.Next(ctx) {
		var tmp model.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		result.Services = append(result.Services, tmp.Service)
	}
	return result, nil
}

func (ds *MetadataManager) UnregisterService(ctx context.Context, request *discovery.DeleteServiceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	force := request.Force
	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		log.Error(fmt.Sprintf("%s micro-service %s failed, operator: %s", title, serviceID, remoteIP), mutil.ErrNotAllowDeleteSC)
		return discovery.NewError(discovery.ErrInvalidParams, mutil.ErrNotAllowDeleteSC.Error())
	}
	microservice, err := GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("%s micro-service %s failed, service does not exist, operator: %s",
				title, serviceID, remoteIP))
			return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist.")
		}
		log.Error(fmt.Sprintf("%s micro-service %s failed, get service file failed, operator: %s",
			title, serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := NewProviderDependencyRelation(ctx, util.ParseDomainProject(ctx), microservice.Service)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP), err)
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP), err)
			return discovery.NewError(discovery.ErrDependedOnConsumer, "Can not delete this service, other service rely it.")
		}
		//todo wait for dep interface
		num, err := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).CountDocuments(ctx, bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}): serviceID})
		if err != nil {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, get instances number failed, operator: %s",
				serviceID, remoteIP), err)
			return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
		}
		if num != 0 {
			log.Error(fmt.Sprintf("delete micro-service[%s] failed, service deployed instances, operator: %s",
				serviceID, remoteIP), nil)
			return discovery.NewError(discovery.ErrDeployedInstance, "Can not delete the service deployed instance(s).")
		}
	}

	_, err = dmongo.GetClient().GetDB().Collection(model.CollectionSchema).BulkWrite(ctx,
		[]mongo.WriteModel{mongo.NewDeleteManyModel().SetFilter(bson.M{model.ColumnServiceID: serviceID})})
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	_, err = dmongo.GetClient().GetDB().Collection(model.CollectionInstance).BulkWrite(ctx,
		[]mongo.WriteModel{mongo.NewDeleteManyModel().SetFilter(bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}): serviceID})})
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	err = deleteServiceTxn(ctx, err, serviceID, force)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
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

	err = DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Error(fmt.Sprintf("micro-service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func deleteServiceTxn(ctx context.Context, err error, serviceID string, force bool) error {
	return dmongo.GetClient().ExecTxn(ctx, func(sessionContext mongo.SessionContext) error {
		_, err = dmongo.GetClient().GetDB().Collection(model.CollectionService).BulkWrite(ctx,
			[]mongo.WriteModel{mongo.NewDeleteOneModel().SetFilter(
				bson.M{mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceID}): serviceID})})
		if err != nil {
			return err
		}
		return sync.DoDeleteOpts(sessionContext, datasource.ResourceService, serviceID,
			&discovery.DeleteServiceRequest{ServiceId: serviceID, Force: force})
	})
}

func (ds *MetadataManager) PutServiceProperties(ctx context.Context, request *discovery.UpdateServicePropsRequest) error {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	setFilter := mutil.NewFilter(
		mutil.ServiceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.ServiceProperty(request.Properties),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setFilter),
	)
	err := updateServiceTxn(ctx, request, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("update service %s properties failed, update mongo failed", request.ServiceId), err)
		if err == dao.ErrNoDocuments {
			return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist.")
		}
		return discovery.NewError(discovery.ErrUnavailableBackend, "Update doc in mongo failed.")
	}
	return nil
}

func updateServiceTxn(ctx context.Context, request *discovery.UpdateServicePropsRequest, filter bson.M, updateFilter bson.M) error {
	return dmongo.GetClient().ExecTxn(ctx, func(sessionContext mongo.SessionContext) error {
		err := dao.UpdateService(ctx, filter, updateFilter)
		if err != nil {
			return err
		}
		return sync.DoUpdateOpts(sessionContext, datasource.ResourceService, request)
	})
}

func (ds *MetadataManager) ListServiceDetail(ctx context.Context, request *discovery.GetServicesInfoRequest) (*discovery.GetServicesInfoResponse, error) {
	optionMap := make(map[string]struct{}, len(request.Options))
	for _, opt := range request.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = []string{"tags", "instances", "schemas", "dependencies"}
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
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
		if len(optionMap) == 1 {
			return &discovery.GetServicesInfoResponse{
				Statistics: st,
			}, nil
		}
	}
	filters := ds.filterServices(ctx, request)
	services, err := dao.GetServices(ctx, filters)
	if err != nil {
		log.Error("get all services by domain failed", err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	allServiceDetails := make([]*discovery.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, mgSvc := range services {
		if !request.WithShared && datasource.IsGlobal(discovery.MicroServiceToKey(domainProject, mgSvc.Service)) {
			continue
		}

		serviceDetail, err := getServiceDetailUtil(ctx, mgSvc, request.CountOnly, options)
		if err != nil {
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
		serviceDetail.MicroService = mgSvc.Service
		tmpServiceDetail, err := datasource.NewServiceOverview(serviceDetail, ds.InstanceProperties)
		if err != nil {
			return nil, err
		}
		allServiceDetails = append(allServiceDetails, tmpServiceDetail)
	}

	return &discovery.GetServicesInfoResponse{
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func (ds *MetadataManager) filterServices(ctx context.Context, request *discovery.GetServicesInfoRequest) bson.M {
	var opts []func(filter bson.M)

	if len(request.Environment) > 0 {
		opts = append(opts, mutil.ServiceEnv(request.Environment))
	}
	if len(request.AppId) > 0 {
		opts = append(opts, mutil.ServiceAppID(request.AppId))
	}
	if len(request.ServiceName) > 0 {
		opts = append(opts, mutil.ServiceServiceName(request.ServiceName))
	}
	return mutil.NewBasicFilter(ctx, opts...)
}

func (ds *MetadataManager) GetOverview(ctx context.Context, request *discovery.GetServicesRequest) (
	*discovery.Statistics, error) {
	ctx = util.WithCacheOnly(ctx)
	st, err := statistics(ctx, false)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return st, nil
}

func (ds *MetadataManager) PutManyTags(ctx context.Context, request *discovery.AddServiceTagsRequest) error {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	setFilter := mutil.NewFilter(
		mutil.Tags(request.Tags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setFilter),
	)
	err := dao.UpdateService(ctx, filter, updateFilter)
	if err == nil {
		return nil
	}
	log.Error(fmt.Sprintf("update service %s tags failed.", request.ServiceId), err)
	if err == dao.ErrNoDocuments {
		return discovery.NewError(discovery.ErrServiceNotExists, err.Error())
	}
	return discovery.NewError(discovery.ErrInternal, err.Error())
}

func (ds *MetadataManager) ListTag(ctx context.Context, request *discovery.GetServiceTagsRequest) (*discovery.GetServiceTagsResponse, error) {
	svc, err := GetServiceByID(ctx, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return nil, discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist")
		}
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return &discovery.GetServiceTagsResponse{
		Tags: svc.Tags,
	}, nil
}

func (ds *MetadataManager) PutTag(ctx context.Context, request *discovery.UpdateServiceTagRequest) error {
	svc, err := GetServiceByID(ctx, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist")
		}
		log.Error(fmt.Sprintf("failed to get %s tags", request.ServiceId), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	dataTags := svc.Tags
	if len(dataTags) > 0 {
		if _, ok := dataTags[request.Key]; !ok {
			return discovery.NewError(discovery.ErrTagNotExists, "Tag does not exist")
		}
	}
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	newTags[request.Key] = request.Value
	setFilter := mutil.NewFilter(
		mutil.Tags(newTags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setFilter),
	)
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	err = dao.UpdateService(ctx, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("update service %s tags failed", request.ServiceId), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return nil
}

func (ds *MetadataManager) DeleteManyTags(ctx context.Context, request *discovery.DeleteServiceTagsRequest) error {
	svc, err := GetServiceByID(ctx, request.ServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist")
		}
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	dataTags := svc.Tags
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	if len(dataTags) > 0 {
		for _, key := range request.Keys {
			if _, ok := dataTags[key]; !ok {
				return discovery.NewError(discovery.ErrTagNotExists, "Tag does not exist")
			}
			delete(newTags, key)
		}
	}
	setFilter := mutil.NewFilter(
		mutil.Tags(newTags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setFilter),
	)
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	err = dao.UpdateService(ctx, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("delete service %s tags failed", request.ServiceId), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return nil
}

func (ds *MetadataManager) GetSchema(ctx context.Context, request *discovery.GetSchemaRequest) (*discovery.GetSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, "GetSchema failed to check service exist.")
	}
	if !exist {
		return nil, discovery.NewError(discovery.ErrServiceNotExists, "GetSchema service does not exist.")
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	resp, err := dao.GetSchema(ctx, filter)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, "GetSchema failed from mongodb.")
	}
	if resp == nil {
		return nil, schema.ErrSchemaNotFound
	}
	return &discovery.GetSchemaResponse{
		Schema:        resp.Schema,
		SchemaSummary: resp.SchemaSummary,
	}, nil
}

func (ds *MetadataManager) GetAllSchemas(ctx context.Context, request *discovery.GetAllSchemaRequest) (*discovery.GetAllSchemaResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	svc, err := dao.GetService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return nil, discovery.NewError(discovery.ErrServiceNotExists, "ListSchema failed for service not exist")
		}
		log.Error(fmt.Sprintf("get service[%s] all schemas failed, get service failed", request.ServiceId), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	schemasList := svc.Service.Schemas
	if len(schemasList) == 0 {
		return &discovery.GetAllSchemaResponse{
			Schemas: []*discovery.Schema{},
		}, nil
	}
	schemas := make([]*discovery.Schema, 0, len(schemasList))
	for _, schemaID := range schemasList {
		tempSchema := &discovery.Schema{}
		tempSchema.SchemaId = schemaID
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.SchemaID(schemaID))
		schema, err := dao.GetSchema(ctx, filter)
		if err != nil {
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
		if schema == nil {
			schemas = append(schemas, tempSchema)
			continue
		}
		tempSchema.Summary = schema.SchemaSummary
		if request.WithSchema {
			tempSchema.Schema = schema.Schema
		}
		schemas = append(schemas, tempSchema)
	}
	return &discovery.GetAllSchemaResponse{
		Schemas: schemas,
	}, nil
}

func (ds *MetadataManager) ExistSchema(ctx context.Context, request *discovery.GetExistenceRequest) (*discovery.GetExistenceResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrServiceNotExists, "ExistSchema failed for get service failed")
	}
	if !exist {
		return nil, discovery.NewError(discovery.ErrServiceNotExists, "ExistSchema failed for service not exist")
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	Schema, err := dao.GetSchema(ctx, filter)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, "ExistSchema failed for get schema failed.")
	}
	if Schema == nil {
		return nil, schema.ErrSchemaNotFound
	}
	return &discovery.GetExistenceResponse{
		Summary:   Schema.SchemaSummary,
		SchemaId:  Schema.SchemaID,
		ServiceId: Schema.ServiceID,
	}, nil
}

func (ds *MetadataManager) DeleteSchema(ctx context.Context, request *discovery.DeleteSchemaRequest) error {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return discovery.NewError(discovery.ErrUnavailableBackend, "DeleteSchema failed for get service failed.")
	}
	if !exist {
		return discovery.NewError(discovery.ErrSchemaNotExists, "DeleteSchema failed for service not exist.")
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	result, err := dmongo.GetClient().GetDB().Collection(model.CollectionSchema).DeleteOne(ctx, filter)
	if err != nil {
		return discovery.NewError(discovery.ErrUnavailableBackend, "DeleteSchema failed for delete schema failed.")
	}
	if result != nil && result.DeletedCount == 0 {
		return schema.ErrSchemaNotFound
	}
	return nil
}

func (ds *MetadataManager) ModifySchema(ctx context.Context, request *discovery.ModifySchemaRequest) (*discovery.ModifySchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId
	schema := discovery.Schema{
		SchemaId: request.SchemaId,
		Summary:  request.Summary,
		Schema:   request.Schema,
	}
	err := ds.modifySchema(ctx, request.ServiceId, &schema)
	if err != nil {
		log.Error(fmt.Sprintf("modify schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP), err)
		return nil, err
	}
	log.Info(fmt.Sprintf("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP))
	return &discovery.ModifySchemaResponse{}, nil
}

func (ds *MetadataManager) ModifySchemas(ctx context.Context, request *discovery.ModifySchemasRequest) (*discovery.ModifySchemasResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	svc, err := dao.GetService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return nil, discovery.NewError(discovery.ErrServiceNotExists, "Service not exist")
		}
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if respErr := ds.modifySchemas(ctx, svc.Service, request.Schemas); respErr != nil {
		return nil, respErr
	}
	return &discovery.ModifySchemasResponse{}, nil

}

func (ds *MetadataManager) modifySchemas(ctx context.Context, service *discovery.MicroService, schemas []*discovery.Schema) error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	filter := mutil.NewFilter(mutil.ServiceID(serviceID))
	schemasFromDatabase, err := dao.GetSchemas(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("modify service %s schemas failed, get schemas failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		datasource.SchemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	var schemasOps []mongo.WriteModel
	var serviceOps []mongo.WriteModel
	if !ds.isSchemaEditable() {
		if len(service.Schemas) == 0 {
			errQuota := quotasvc.ApplySchema(ctx, serviceID, int64(len(nonExistSchemaIds)))
			if errQuota != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
				return errQuota
			}
			filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
			setFilter := mutil.NewFilter(mutil.ServiceSchemas(nonExistSchemaIds))
			updateFilter := mutil.NewFilter(mutil.Set(setFilter))
			serviceOps = append(serviceOps, mongo.NewUpdateOneModel().SetUpdate(updateFilter).SetFilter(filter))
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Error(fmt.Sprintf("modify service %s schemas failed, operator: %s", serviceID, remoteIP), err)
				return discovery.NewError(discovery.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := dao.SchemaSummaryExist(ctx, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return discovery.NewError(discovery.ErrInternal, err.Error())
				}
				if !exist {
					filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(needUpdateSchema.SchemaId))
					setFilter := mutil.NewFilter(
						mutil.Schema(needUpdateSchema.Schema),
						mutil.SchemaSummary(needUpdateSchema.Summary),
					)
					updateFilter := mutil.NewFilter(
						mutil.Set(setFilter),
					)
					schemasOps = append(schemasOps, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateFilter))
				} else {
					log.Warn(fmt.Sprintf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP))
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			schemasOps = append(schemasOps, mongo.NewInsertOneModel().SetDocument(&model.Schema{
				Domain:        domain,
				Project:       project,
				ServiceID:     serviceID,
				SchemaID:      schema.SchemaId,
				Schema:        schema.Schema,
				SchemaSummary: schema.Summary,
			}))
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			errQuota := quotasvc.ApplySchema(ctx, serviceID, int64(quotaSize))
			if errQuota != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
				return errQuota
			}
		}
		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			schemasOps = append(schemasOps, mongo.NewInsertOneModel().SetDocument(&model.Schema{
				Domain:        domain,
				Project:       project,
				ServiceID:     serviceID,
				SchemaID:      schema.SchemaId,
				Schema:        schema.Schema,
				SchemaSummary: schema.Summary,
			}))
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Info(fmt.Sprintf("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
			setFilter := mutil.NewFilter(
				mutil.Schema(schema.Schema),
				mutil.SchemaSummary(schema.Summary),
			)
			updateFilter := mutil.NewFilter(
				mutil.Set(setFilter),
			)
			schemasOps = append(schemasOps, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateFilter))
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Info(fmt.Sprintf("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
			schemasOps = append(schemasOps, mongo.NewDeleteOneModel().SetFilter(filter))
		}
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
		setFilter := mutil.NewFilter(mutil.ServiceSchemas(schemaIDs))
		updateFilter := mutil.NewFilter(mutil.Set(setFilter))
		serviceOps = append(serviceOps, mongo.NewUpdateOneModel().SetUpdate(updateFilter).SetFilter(filter))
	}
	if len(schemasOps) > 0 {
		_, err = dmongo.GetClient().GetDB().Collection(model.CollectionSchema).BulkWrite(ctx, schemasOps)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	if len(serviceOps) > 0 {
		_, err = dmongo.GetClient().GetDB().Collection(model.CollectionService).BulkWrite(ctx, serviceOps)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	return nil
}

// modifySchema will be modified in the following cases
// 1.service have no relation --> update the schema && update the service
// 2.service is editable && service have relation with the schema --> update the schema
// 3.service is editable && service have no relation with the schema --> update the schema && update the service
// 4.service can't edit && service have relation with the schema && schema summary not exist --> update the schema
func (ds *MetadataManager) modifySchema(ctx context.Context, serviceID string, schema *discovery.Schema) *errsvc.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	svc, err := GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist.")
		}
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	microservice := svc.Service
	var isExist bool
	for _, sid := range microservice.Schemas {
		if sid == schema.SchemaId {
			isExist = true
			break
		}
	}
	var newSchemas []string
	if !ds.isSchemaEditable() {
		if len(microservice.Schemas) != 0 && !isExist {
			return ErrUndefinedSchemaID
		}
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
		respSchema, err := dao.GetSchema(ctx, filter)
		if err != nil {
			return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
		}
		if respSchema != nil {
			if len(schema.Summary) == 0 {
				log.Error(fmt.Sprintf("modify schema %s %s failed, get schema summary failed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return ErrModifySchemaNotAllow
			}
			if len(respSchema.SchemaSummary) != 0 {
				log.Error(fmt.Sprintf("mode, schema %s %s already exist, can not be changed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return ErrModifySchemaNotAllow
			}
		}
		if len(microservice.Schemas) == 0 {
			copy(newSchemas, microservice.Schemas)
			newSchemas = append(newSchemas, schema.SchemaId)
		}
	} else {
		if !isExist {
			newSchemas = append(microservice.Schemas, schema.SchemaId)
		}
	}
	if len(newSchemas) != 0 {
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
		setFilter := mutil.NewFilter(
			mutil.ServiceSchemas(newSchemas),
		)
		updateFilter := mutil.NewFilter(
			mutil.Set(setFilter),
		)
		err = dao.UpdateService(ctx, filter, updateFilter)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
	setFilter := mutil.NewFilter(
		mutil.Schema(schema.Schema),
		mutil.SchemaSummary(schema.Summary),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	err = dao.UpdateSchema(ctx, filter, updateFilter, options.Update().SetUpsert(true))
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return nil
}

func (ds *MetadataManager) isSchemaEditable() bool {
	return !ds.SchemaNotEditable
}

func getServiceDetailUtil(ctx context.Context, mgs *model.Service, countOnly bool, options []string) (*discovery.ServiceDetail, error) {
	serviceDetail := new(discovery.ServiceDetail)
	serviceID := mgs.Service.ServiceId
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	if countOnly {
		serviceDetail.Statics = new(discovery.Statistics)
	}
	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			serviceDetail.Tags = mgs.Tags
		case "instances":
			if countOnly {
				instanceCount, err := CountInstance(ctx, serviceID)
				if err != nil {
					log.Error(fmt.Sprintf("get number of service [%s]'s instances failed", serviceID), err)
					return nil, err
				}
				serviceDetail.Statics.Instances = &discovery.StInstance{
					Count: instanceCount,
				}
				continue
			}
			filter := mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(serviceID))
			instances, err := dao.GetMicroServiceInstances(ctx, filter)
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all instances failed", serviceID), err)
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID))
			schemas, err := dao.GetSchemas(ctx, filter)
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all schemas failed", mgs.Service.ServiceId), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			service := mgs.Service
			consumers, err := GetConsumers(ctx, domainProject, service,
				WithoutSelfDependency(), WithSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all consumers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
			}
			providers, err := GetProviders(ctx, domainProject, service,
				WithoutSelfDependency(), WithSameDomainProject())
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

// Instance management
func (ds *MetadataManager) RegisterInstance(ctx context.Context,
	request *discovery.RegisterInstanceRequest) (*discovery.RegisterInstanceResponse, error) {

	isCustomID := true

	instance := request.Instance
	if len(instance.InstanceId) == 0 {
		isCustomID = false
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	// if queueSize is more than 0 and channel is not full, then do fast register instance
	if fastRegConfig.QueueSize > 0 && len(GetFastRegisterInstanceService().InstEventCh) < fastRegConfig.QueueSize {
		// fast register, just add instance to channel and batch register them later
		event := &InstanceRegisterEvent{ctx, request, isCustomID, 0}
		GetFastRegisterInstanceService().AddEvent(event)

		return &discovery.RegisterInstanceResponse{
			InstanceId: instance.InstanceId,
		}, nil
	}

	return RegisterInstanceSingle(ctx, request, isCustomID)
}

func sendEvent(ctx context.Context, action string, resourceType string, resource interface{}) {
	if !util.EnableSync(ctx) {
		return
	}
	event.Publish(ctx, action, resourceType, resource)
}

func RegisterInstanceSingle(ctx context.Context, request *discovery.RegisterInstanceRequest,
	isUserDefinedID bool) (*discovery.RegisterInstanceResponse, error) {

	resp, needRegister, err := preProcessRegister(ctx, request.Instance, isUserDefinedID)
	if err != nil || !needRegister {
		log.Error("pre process instance err, or instance already existed", err)
		return resp, err
	}

	return registryInstance(ctx, request)
}

func RegisterInstanceBatch(ctx context.Context, events []*InstanceRegisterEvent) error {
	instances := make([]interface{}, len(events))

	for i, event := range events {
		eventCtx := event.Ctx
		instance := event.Request.Instance

		_, needRegister, err := preProcessRegister(eventCtx, instance, event.isCustomID)
		if err != nil || !needRegister {
			log.Error("pre process instance err, or instance existed", err)
			return err
		}

		domain := util.ParseDomain(eventCtx)
		project := util.ParseProject(eventCtx)

		data := model.Instance{
			Domain:      domain,
			Project:     project,
			RefreshTime: time.Now(),
			Instance:    instance,
		}
		instances[i] = data
	}

	return registryInstances(ctx, instances)
}

func preProcessRegister(ctx context.Context, instance *discovery.MicroServiceInstance,
	isUserDefinedID bool) (*discovery.RegisterInstanceResponse, bool, error) {
	// 允许自定义 id
	if isUserDefinedID {
		needRegister, err := sendHeartbeatInstead(ctx, instance)
		if err != nil {
			return nil, false, err
		}
		if !needRegister {
			return &discovery.RegisterInstanceResponse{
				InstanceId: instance.InstanceId,
			}, false, nil
		}
	}
	if instance.HealthCheck == nil {
		instance.HealthCheck = &discovery.HealthCheck{
			Mode:     discovery.CHECK_BY_HEARTBEAT,
			Interval: datasource.DefaultLeaseRenewalInterval,
			Times:    datasource.DefaultLeaseRetryTimes,
		}
	}
	return &discovery.RegisterInstanceResponse{
		InstanceId: instance.InstanceId,
	}, true, nil
}

func sendHeartbeatInstead(ctx context.Context, instance *discovery.MicroServiceInstance) (bool, error) {
	remoteIP := util.GetIPFromContext(ctx)

	err := datasource.GetMetadataManager().SendHeartbeat(ctx, &discovery.HeartbeatRequest{
		InstanceId: instance.InstanceId,
		ServiceId:  instance.ServiceId,
	})
	if err == nil {
		log.Info(fmt.Sprintf("register instance successful, reuse instance[%s/%s], operator %s",
			instance.ServiceId, instance.InstanceId, remoteIP))
		return false, nil
	}

	if errsvc.IsErrEqualCode(err, discovery.ErrInstanceNotExists) {
		// register a new one
		return true, nil
	}

	log.Error(fmt.Sprintf("register instance failed, reuse instance %s %s, operator %s",
		instance.ServiceId, instance.InstanceId, remoteIP), err)
	return false, err
}

func (ds *MetadataManager) ExistInstance(ctx context.Context, request *discovery.MicroServiceInstanceKey) (*discovery.GetExistenceByIDResponse, error) {
	exist, err := ExistInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return &discovery.GetExistenceByIDResponse{
		Exist: exist,
	}, nil
}

// GetInstance returns instance under the current domain
func (ds *MetadataManager) GetInstance(ctx context.Context, request *discovery.GetOneInstanceRequest) (*discovery.GetOneInstanceResponse, error) {
	service := &model.Service{}
	service.Service = &discovery.MicroService{}
	var err error
	var serviceIDs []string
	if len(request.ConsumerServiceId) > 0 {
		service, err = GetServiceByID(ctx, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider instance %s %s",
					request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
				return nil, discovery.NewError(discovery.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf(" get consumer failed, consumer %s find provider instance %s",
				request.ConsumerServiceId, request.ProviderInstanceId), err)
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	provider, err := GetServiceByID(ctx, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist, consumer %s find provider instance %s %s",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId))
			return nil, discovery.NewError(discovery.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId))
		}
		log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instance %s %s",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	findFlag := func() string {
		return fmt.Sprintf("consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version,
			request.ProviderInstanceId)
	}

	domainProject := util.ParseDomainProject(ctx)
	services, err := filterServices(ctx, discovery.MicroServiceToKey(domainProject, provider.Service))
	if err != nil {
		log.Error(fmt.Sprintf("get instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	serviceIDs = filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	if len(serviceIDs) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("query service failed", mes)
		return nil, discovery.NewError(discovery.ErrInstanceNotExists, mes.Error())
	}

	instances, err := GetAllInstancesOfOneService(ctx, request.ProviderServiceId)

	if err != nil {
		log.Error(fmt.Sprintf("get instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if len(instances) == 0 {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("find Instances by ProviderID failed", mes)
		return nil, discovery.NewError(discovery.ErrInstanceNotExists, mes.Error())
	}

	instance := instances[0]
	// use explicit instanceId to query
	if len(request.ProviderInstanceId) != 0 {
		isExist := false
		for _, ins := range instances {
			if ins.InstanceId == request.ProviderInstanceId {
				instance = ins
				isExist = true
				break
			}
		}
		if !isExist {
			mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
			log.Error("get instance failed", mes)
			return nil, discovery.NewError(discovery.ErrInstanceNotExists, mes.Error())
		}
	}

	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instance = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, newRev)

	return &discovery.GetOneInstanceResponse{
		Instance: instance,
	}, nil
}

func (ds *MetadataManager) ListInstance(ctx context.Context, request *discovery.GetInstancesRequest) (*discovery.GetInstancesResponse, error) {
	service := &model.Service{}
	var err error

	if len(request.ConsumerServiceId) > 0 {
		service, err = GetServiceByID(ctx, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider %s instances",
					request.ConsumerServiceId, request.ProviderServiceId))
				return nil, discovery.NewError(discovery.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %s instances",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}

	provider, err := GetServiceByID(ctx, request.ProviderServiceId)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider does not exist, consumer %s find provider %s  instances",
				request.ConsumerServiceId, request.ProviderServiceId))
			return nil, discovery.NewError(discovery.ErrServiceNotExists,
				fmt.Sprintf("provider[%s] does not exist.", request.ProviderServiceId))
		}
		log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instances %s",
			request.ConsumerServiceId, request.ProviderServiceId), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	findFlag := func() string {
		return fmt.Sprintf("consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Service.ServiceId, provider.Service.Environment, provider.Service.AppId, provider.Service.ServiceName, provider.Service.Version)
	}

	rev, _ := ctx.Value(util.CtxRequestRevision).(string)
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, []*model.Service{provider})
	if len(serviceIDs) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("query service failed", mes)
		return nil, discovery.NewError(discovery.ErrServiceNotExists, mes.Error())
	}

	instances, err := GetAllInstancesOfOneService(ctx, request.ProviderServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("get instances failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.GetInstancesResponse{
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) ListManyInstances(ctx context.Context, request *discovery.GetAllInstancesRequest) (*discovery.GetAllInstancesResponse, error) {
	findRes, err := GetInstances(ctx)
	if err != nil {
		return nil, err
	}
	resp := &discovery.GetAllInstancesResponse{}

	for _, inst := range findRes {
		resp.Instances = append(resp.Instances, inst.Instance)
	}

	return resp, nil
}

// FindInstances returns instances under the specified domain
func (ds *MetadataManager) FindInstances(ctx context.Context, request *discovery.FindInstancesRequest) (*discovery.FindInstancesResponse, error) {
	provider := &discovery.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.Alias,
	}
	rev, ok := ctx.Value(util.CtxRequestRevision).(string)
	if !ok {
		err := errors.New("rev request context is not type string")
		log.Error("", err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if datasource.IsGlobal(provider) {
		return ds.findSharedServiceInstance(ctx, request, provider, rev)
	}

	return ds.findInstance(ctx, request, provider, rev)
}

func (ds *MetadataManager) PutInstance(ctx context.Context, request *discovery.RegisterInstanceRequest) error {
	instance := request.Instance
	serviceID := instance.ServiceId
	instanceID := instance.InstanceId
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID), mutil.InstanceInstanceID(instanceID))
	exist, err := ExistInstance(ctx, serviceID, instanceID)
	if err != nil {
		log.Error(fmt.Sprintf("update instance %s/%s failed", serviceID, instanceID), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if !exist {
		log.Error(fmt.Sprintf("update instance %s/%s failed, instance does not exist",
			serviceID, instanceID), err)
		return discovery.NewError(discovery.ErrInstanceNotExists, "Service instance does not exist.")
	}

	setFilter := mutil.NewFilter(
		mutil.Instance(instance),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	if err := dao.UpdateInstance(ctx, filter, updateFilter); err != nil {
		log.Error(fmt.Sprintf("update instance %s/%s failed", serviceID, instanceID), err)
		return err
	}
	sendEvent(ctx, csync.UpdateAction, datasource.ResourceInstance, instance)
	log.Info(fmt.Sprintf("update instance[%s/%s] successfully", serviceID, instanceID))
	return nil
}

func (ds *MetadataManager) PutInstanceStatus(ctx context.Context, request *discovery.UpdateInstanceStatusRequest) error {
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, "/")

	// todo finish get instance
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	exist, err := ExistInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if !exist {
		log.Error(fmt.Sprintf("update instance %s status failed, instance does not exist", updateStatusFlag), err)
		return discovery.NewError(discovery.ErrInstanceNotExists, "Service instance does not exist.")
	}

	setFilter := mutil.NewFilter(
		mutil.InstanceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.InstanceStatus(request.Status),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	if err := dao.UpdateInstance(ctx, filter, updateFilter); err != nil {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		return err
	}

	log.Info(fmt.Sprintf("update instance[%s] status successfully", updateStatusFlag))
	return nil
}

func (ds *MetadataManager) PutInstanceProperties(ctx context.Context, request *discovery.UpdateInstancePropsRequest) error {
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	exist, err := ExistInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if !exist {
		log.Error(fmt.Sprintf("update instance %s properties failed, instance does not exist", instanceFlag), err)
		return discovery.NewError(discovery.ErrInstanceNotExists, "Service instance does not exist.")
	}

	filter := mutil.NewDomainProjectFilter(domain, project, mutil.InstanceServiceID(request.ServiceId), mutil.InstanceInstanceID(request.InstanceId))
	setFilter := mutil.NewFilter(
		mutil.InstanceModTime(strconv.FormatInt(time.Now().Unix(), baseTen)),
		mutil.InstanceProperties(request.Properties),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setFilter))
	if err := dao.UpdateInstance(ctx, filter, updateFilter); err != nil {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		return err
	}

	log.Info(fmt.Sprintf("update instance[%s] properties successfully", instanceFlag))
	return nil
}

func (ds *MetadataManager) UnregisterInstance(ctx context.Context, request *discovery.UnregisterInstanceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID), mutil.InstanceInstanceID(instanceID))
	result, err := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteMany(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("unregister instance failed, instance %s, operator %s revoke instance failed",
			instanceFlag, remoteIP), err)
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if result.DeletedCount == 0 {
		return discovery.NewError(discovery.ErrInstanceNotExists, "Instance not found")
	}
	sendEvent(ctx, csync.DeleteAction, datasource.ResourceInstance, request)
	log.Info(fmt.Sprintf("unregister instance[%s], operator %s", instanceFlag, remoteIP))
	return nil
}

func (ds *MetadataManager) SendHeartbeat(ctx context.Context, request *discovery.HeartbeatRequest) error {
	remoteIP := util.GetIPFromContext(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	err := KeepAliveLease(ctx, request)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance %s operator %s", instanceFlag, remoteIP), err)
		return err
	}
	sendEvent(ctx, csync.UpdateAction, datasource.ResourceHeartbeat, request)
	return nil
}

func (ds *MetadataManager) SendManyHeartbeat(ctx context.Context, request *discovery.HeartbeatSetRequest) (*discovery.HeartbeatSetResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(request.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *discovery.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0

	for _, heartbeatElement := range request.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warn(fmt.Sprintf("instance[%s/%s] is duplicate request heartbeat set",
				heartbeatElement.ServiceId, heartbeatElement.InstanceId))
			continue
		} else {
			existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId] = true
			noMultiCounter++
		}
		gopool.Go(getHeartbeatFunc(ctx, domainProject, instancesHbRst, heartbeatElement))
	}

	count := 0
	instanceHbRstArr := make([]*discovery.InstanceHbRst, 0, heartBeatCount)
	for hbRst := range instancesHbRst {
		count++
		instanceHbRstArr = append(instanceHbRstArr, hbRst)
		sendEvent(ctx, csync.UpdateAction, datasource.ResourceHeartbeat,
			&discovery.HeartbeatRequest{ServiceId: hbRst.ServiceId, InstanceId: hbRst.InstanceId})
		if count == noMultiCounter {
			close(instancesHbRst)
		}
	}

	log.Info(fmt.Sprintf("batch update heartbeats failed %v", instanceHbRstArr))
	return &discovery.HeartbeatSetResponse{
		Instances: instanceHbRstArr,
	}, nil
}

func registryInstance(ctx context.Context, request *discovery.RegisterInstanceRequest) (*discovery.RegisterInstanceResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance
	ttl := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)

	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	instanceID := instance.InstanceId
	data := &model.Instance{
		Domain:      domain,
		Project:     project,
		RefreshTime: time.Now(),
		Instance:    instance,
	}

	insertRes, err := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(ctx, data)
	if err != nil {
		if dao.IsDuplicateKey(err) {
			return &discovery.RegisterInstanceResponse{
				InstanceId: instanceID,
			}, nil
		}
		log.Error(fmt.Sprintf("register instance failed %s instanceID %s operator %s", instanceFlag, instanceID, remoteIP), err)
		return nil, discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}

	// need to complete the instance offline function in time, so you need to check the heartbeat after registering the instance
	err = heartbeat.Instance().CheckInstance(ctx, instance)
	if err != nil {
		log.Error(fmt.Sprintf("fail to check instance, instance[%s]. operator %s", instance.InstanceId, remoteIP), err)
	}

	sendEvent(ctx, csync.CreateAction, datasource.ResourceInstance, request)
	log.Info(fmt.Sprintf("register instance %s, instanceID %s, operator %s",
		instanceFlag, insertRes.InsertedID, remoteIP))
	return &discovery.RegisterInstanceResponse{
		InstanceId: instanceID,
	}, nil
}

func registryInstances(ctx context.Context, instances []interface{}) error {
	opts := options.InsertManyOptions{}
	opts.SetOrdered(false)
	opts.SetBypassDocumentValidation(true)
	_, err := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertMany(ctx, instances, &opts)
	if err != nil {
		log.Error("Batch register instance failed", err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}

	return nil
}

func (ds *MetadataManager) findSharedServiceInstance(ctx context.Context, request *discovery.FindInstancesRequest, provider *discovery.MicroServiceKey, rev string) (*discovery.FindInstancesResponse, error) {
	var err error
	// it means the shared micro-services must be the same env with SC.
	provider.Environment = apt.Service.Environment
	findFlag := func() string {
		return fmt.Sprintf("find shared provider[%s/%s/%s/%s]", provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}
	services, err := filterServices(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find shared service instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if len(services) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("find shared service instance failed", mes)
		return nil, discovery.NewError(discovery.ErrServiceNotExists, mes.Error())
	}
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	inFilter := mutil.NewFilter(mutil.In(serviceIDs))
	filter := mutil.NewFilter(mutil.InstanceServiceID(inFilter))
	option := &options.FindOptions{Sort: bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}): -1}}
	instances, err := dao.GetMicroServiceInstances(ctx, filter, option)
	if err != nil {
		log.Error(fmt.Sprintf("find shared service instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	// TODO support gRPC output context
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.FindInstancesResponse{
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) findInstance(ctx context.Context, request *discovery.FindInstancesRequest, provider *discovery.MicroServiceKey, rev string) (*discovery.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &model.Service{Service: &discovery.MicroService{Environment: request.Environment}}
	if len(request.ConsumerServiceId) > 0 {
		service, err = GetServiceByID(ctx, request.ConsumerServiceId)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Debug(fmt.Sprintf("consumer does not exist, consumer %s find provider %s/%s/%s",
					request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName))
				return nil, discovery.NewError(discovery.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId))
			}
			log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %s/%s/%s",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName), err)
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
		provider.Environment = service.Service.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
			request.ConsumerServiceId, service.Service.Environment, service.Service.AppId, service.Service.ServiceName, service.Service.Version,
			provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
	}
	services, err := filterServices(ctx, provider)
	if err != nil {
		log.Error(fmt.Sprintf("find instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	if len(services) == 0 {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
		log.Error("find instance failed", mes)
		return nil, discovery.NewError(discovery.ErrServiceNotExists, mes.Error())
	}
	serviceIDs := filterServiceIDs(ctx, request.ConsumerServiceId, request.Tags, services)
	inFilter := mutil.NewFilter(mutil.In(serviceIDs))
	filter := mutil.NewFilter(mutil.InstanceServiceID(inFilter))
	option := &options.FindOptions{Sort: bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}): -1}}
	instances, err := dao.GetMicroServiceInstances(ctx, filter, option)
	if err != nil {
		log.Error(fmt.Sprintf("find instance failed %s", findFlag()), err)
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	// add dependency queue
	if len(request.ConsumerServiceId) > 0 &&
		len(serviceIDs) > 0 {
		provider, err = ds.reshapeProviderKey(ctx, provider, serviceIDs[0])
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = AddServiceVersionRule(ctx, domainProject, service.Service, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag())
			log.Error("add service version rule failed", mes)
			return nil, discovery.NewError(discovery.ErrServiceNotExists, mes.Error())
		}
		if err != nil {
			log.Error(fmt.Sprintf("add service version rule failed %s", findFlag()), err)
			return nil, discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	newRev, _ := formatRevision(request.ConsumerServiceId, instances)
	if rev == newRev {
		instances = nil // for gRPC
	}
	_ = util.WithResponseRev(ctx, newRev)
	return &discovery.FindInstancesResponse{
		Instances: instances,
	}, nil
}

func (ds *MetadataManager) reshapeProviderKey(ctx context.Context, provider *discovery.MicroServiceKey, providerID string) (*discovery.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	providerService, err := GetServiceByID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = discovery.MicroServiceToKey(provider.Tenant, providerService.Service)
	provider.Version = versionRule
	return provider, nil
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *discovery.MicroService, provider *discovery.MicroServiceKey) error {
	consumerKey := discovery.MicroServiceToKey(domainProject, consumer)
	exist, err := DependencyRuleExist(ctx, provider, consumerKey)
	if exist || err != nil {
		return err
	}

	r := &discovery.ConsumerDependency{
		Consumer:  consumerKey,
		Providers: []*discovery.MicroServiceKey{provider},
		Override:  false,
	}
	err = syncDependencyRule(ctx, domainProject, r)

	if err != nil {
		return err
	}

	return nil
}

func DependencyRuleExist(ctx context.Context, provider *discovery.MicroServiceKey, consumer *discovery.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}
	consumerKey := GenerateConsumerDependencyRuleKey(consumer.Tenant, consumer)
	existed, err := DependencyRuleExistUtil(ctx, consumerKey, provider)
	if err != nil || existed {
		return existed, err
	}
	providerKey := GenerateProviderDependencyRuleKey(targetDomainProject, provider)
	return DependencyRuleExistUtil(ctx, providerKey, consumer)
}

func DependencyRuleExistUtil(ctx context.Context, key bson.M, target *discovery.MicroServiceKey) (bool, error) {
	compareData, err := TransferToMicroServiceDependency(ctx, key)
	if err != nil {
		return false, err
	}

	if len(compareData.Dependency) != 0 {
		isEqual, err := datasource.ContainServiceDependency(compareData.Dependency, target)
		if err != nil {
			return false, err
		}
		if isEqual {
			return true, nil
		}
	}
	return false, nil
}

func KeepAliveLease(ctx context.Context, request *discovery.HeartbeatRequest) error {
	_, err := heartbeat.Instance().Heartbeat(ctx, request)
	if err != nil {
		return discovery.NewError(discovery.ErrInstanceNotExists, err.Error())
	}
	return nil
}

func getHeartbeatFunc(ctx context.Context, domainProject string, instancesHbRst chan<- *discovery.InstanceHbRst, element *discovery.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &discovery.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}

		req := &discovery.HeartbeatRequest{
			InstanceId: element.InstanceId,
			ServiceId:  element.ServiceId,
		}

		err := KeepAliveLease(ctx, req)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Error(fmt.Sprintf("heartbeat set failed %s %s", element.ServiceId, element.InstanceId), err)
		}
		instancesHbRst <- hbRst
	}
}

func filterServices(ctx context.Context, key *discovery.MicroServiceKey) ([]*model.Service, error) {
	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, errors.New("invalid 'domain' or 'project'")
	}
	services, ok := cache.GetServiceByName(ctx, key)
	if ok {
		return services, nil
	}
	serviceNameOption := mutil.ServiceServiceName(key.ServiceName)
	if len(key.Alias) > 0 {
		serviceNameOption = mutil.Or(serviceNameOption, mutil.ServiceAlias(key.Alias))
	}
	filter := mutil.NewDomainProjectFilter(tenant[0], tenant[1],
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		serviceNameOption,
	)
	return dao.GetServices(ctx, filter)
}

func filterServiceIDs(ctx context.Context, consumerID string, tags []string, services []*model.Service) []string {
	var filterService []*model.Service
	serviceIDs := make([]string, 0)
	if len(services) == 0 {
		return serviceIDs
	}
	filterService = filterTags(services, tags)
	filterService = filterAccess(ctx, consumerID, filterService)
	for _, service := range filterService {
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return serviceIDs
}

func filterTags(services []*model.Service, tags []string) []*model.Service {
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

func filterAccess(ctx context.Context, consumerID string, services []*model.Service) []*model.Service {
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

func accessible(ctx context.Context, consumerID string, providerID string) *errsvc.Error {
	if len(consumerID) == 0 {
		return nil
	}

	consumerService, err := GetServiceByID(ctx, consumerID)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("an error occurred in query consumer(%s)", err.Error()))
	}
	if consumerService == nil {
		return discovery.NewError(discovery.ErrServiceNotExists, "consumer serviceID is invalid")
	}

	// 跨应用权限
	providerService, err := GetServiceByIDAcrossDomain(ctx, providerID)
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
	return nil
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
	if !datasource.IsGlobal(discovery.MicroServiceToKey(util.ParseTargetDomainProject(ctx), providerService.Service)) &&
		providerService.Service.Environment != consumerService.Service.Environment {
		return fmt.Errorf("not allow across environment access")
	}
	return nil
}

func DeleteDependencyForDeleteService(domainProject string, serviceID string, service *discovery.MicroServiceKey) error {
	conDep := new(discovery.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*discovery.MicroServiceKey{}
	conDep.Override = true
	err := syncDependencyRule(context.TODO(), domainProject, conDep)
	if err != nil {
		return err
	}
	return nil
}

func formatRevision(consumerServiceID string, instances []*discovery.MicroServiceInstance) (string, error) {
	if instances == nil {
		return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(consumerServiceID))), nil
	}
	copyInstance := make([]*discovery.MicroServiceInstance, len(instances))
	copy(copyInstance, instances)
	sort.Sort(InstanceSlice(copyInstance))
	data, err := json.Marshal(copyInstance)
	if err != nil {
		log.Error("fail to marshal instance json", err)
		return "", err
	}
	s := fmt.Sprintf("%s.%x", consumerServiceID, sha1.Sum(data))
	return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(s))), nil
}

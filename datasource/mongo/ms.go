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
	"errors"
	"fmt"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	service := request.Service
	serviceUtil.SetServiceDefaultValue(service)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	//todo add quota check
	requestServiceID := service.ServiceId

	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, util.StringJoin([]string{domain, project, service.Environment, service.AppId, service.ServiceName, service.Alias, service.Version}, "/"))
		service.ServiceId = uuid.Generator().GetServiceID(ctx)
	}
	// the service unique index in table is (serviceId,serviceEnv,serviceAppid,servicename,serviceAlias,serviceVersion)
	existID, err := ServiceExistID(ctx, service.ServiceId)
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Check service exist failed"),
		}, err
	}
	exist, err := ServiceExist(ctx, &pb.MicroServiceKey{
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Alias:       service.Alias,
		Version:     service.Version,
	})
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Check service exist failed"),
		}, err
	}
	if existID || exist {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceAlreadyExists, "ServiceID conflict or found the same service."),
		}, nil
	}
	insertRes, err := client.GetMongoClient().Insert(ctx, CollectionService, &MgService{Domain: domain, Project: project, Service: service})
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Register service failed."),
		}, err
	}
	if insertRes == nil {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceAlreadyExists, "ServiceID or ServiceInfo conflict."),
		}, nil
	}

	remoteIP := util.GetIPFromContext(ctx)
	log.Infof("create micro-service[%s][%s] successfully,operator: %s", service.ServiceId, insertRes.InsertedID, remoteIP)

	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Register service successfully"),
		ServiceId: service.ServiceId,
	}, nil
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{Domain: domain, Project: project}

	services, err := GetServices(ctx, filter)
	if err != nil {
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get services data failed."),
		}, nil
	}

	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetApplications(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{Domain: domain, Project: project, ServiceEnv: request.Environment}

	services, err := GetServices(ctx, filter)
	if err != nil {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get services data failed."),
		}, nil
	}
	l := len(services)
	if l == 0 {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get services data failed."),
		}, nil
	}
	apps := make([]string, 0, l)
	hash := make(map[string]struct{}, l)
	for _, svc := range services {
		if !request.WithShared && apt.IsGlobal(pb.MicroServiceToKey(util.ParseDomainProject(ctx), svc)) {
			continue
		}
		if _, ok := hash[svc.AppId]; ok {
			continue
		}
		hash[svc.AppId] = struct{}{}
		apps = append(apps, svc.AppId)
	}
	return &pb.GetAppsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all applications successfully."),
		AppIds:   apps,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to get single service [%s] from mongo", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get service data from mongodb failed."),
		}, err
	}
	if svc != nil {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.ResponseSuccess, "Get service successfully."),
			Service:  svc.Service,
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service not exist."),
	}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {

	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetExistenceByIDResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Check service exist failed."),
			Exist:    false,
		}, err
	}

	return &pb.GetExistenceByIDResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Check ExistService successfully."),
		Exist:    exist,
	}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	serviceKey := &pb.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
	}
	//todo add verison match.
	services, err := GetServices(ctx, GeneratorServiceNameFilter(ctx, serviceKey))
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if len(services) != 0 {
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "get service id successfully."),
			ServiceId: services[0].ServiceId,
		}, nil
	}
	services, err = GetServices(ctx, GeneratorServiceAliasFilter(ctx, serviceKey))
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if len(services) != 0 {
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "get service id successfully."),
			ServiceId: services[0].ServiceId,
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist"),
	}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Delete service failed,failed to get service."),
		}, err
	}
	if !exist {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Delete service failed,service not exist."),
		}, nil
	}
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "DelService failed to create session."),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "DelService failed to start session."),
		}, err
	}
	defer session.EndSession(ctx)
	//todo delete instance,tags,schemas...
	res, err := DelServicePri(ctx, request.ServiceId, request.Force)
	if err != nil {
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Txn delete service abort failed."),
			}, errAbort
		}
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Delete service failed"),
		}, err
	}
	errCommit := session.CommitTransaction(ctx)
	if errCommit != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Txn delete service commit failed."),
		}, errCommit
	}
	return &pb.DeleteServiceResponse{
		Response: res,
	}, nil
}

func DelServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return pb.CreateResponse(scerr.ErrInvalidParams, err.Error()), nil
	}
	microservice, err := GetService(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
	}
	if microservice == nil {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}
	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		log.Infof("force delete,should del instance...")
		//todo wait for dep interface
	}
	filter := GeneratorServiceFilter(ctx, serviceID)
	//todo del instances
	tables := []string{CollectionService, CollectionSchema, CollectionRule}
	for _, col := range tables {
		err := client.GetMongoClient().DocDeleteMany(ctx, col, filter)
		if err != nil {
			return pb.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
	}
	return pb.CreateResponse(pb.ResponseSuccess, "Unregister service successfully."), nil

}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {

	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "UpdateService failed,failed to get service."),
		}, err
	}
	if !exist {
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "UpdateService failed,service not exist."),
		}, nil
	}

	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceModTime: strconv.FormatInt(time.Now().Unix(), 10), ServiceProperty: request.Properties}})
	if err != nil {
		log.Errorf(err, "update service [%s] properties failed, update mongo failed", request.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Update doc in mongo failed."),
		}, nil
	}
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service successfully."),
	}, nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
	serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {}
}

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceDetailResponse, error) {
	mgSvc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if mgSvc == nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	svc := mgSvc.Service
	versions, err := GetServicesVersions(ctx, bson.M{})
	if err != nil {
		log.Errorf(err, "get service[%s/%s/%s] all versions failed",
			svc.Environment, svc.AppId, svc.ServiceName)
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}
	serviceInfo, err := getServiceDetailUtil(ctx, mgSvc, false, options)
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	serviceInfo.MicroService = svc
	serviceInfo.MicroServiceVersions = versions
	return &pb.GetServiceDetailResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get service successfully"),
		Service:  serviceInfo,
	}, nil

}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *pb.GetServicesInfoRequest) (
	*pb.GetServicesInfoResponse, error) {
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
	//todo add get statistics info
	services, err := GetMgServices(ctx, bson.M{})
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesInfoResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	allServiceDetails := make([]*pb.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, mgSvc := range services {
		if !request.WithShared && apt.IsGlobal(pb.MicroServiceToKey(domainProject, mgSvc.Service)) {
			continue
		}
		if len(request.AppId) > 0 {
			if request.AppId != mgSvc.Service.AppId {
				continue
			}
			if len(request.ServiceName) > 0 && request.ServiceName != mgSvc.Service.ServiceName {
				continue
			}
		}

		serviceDetail, err := getServiceDetailUtil(ctx, mgSvc, request.CountOnly, options)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		serviceDetail.MicroService = mgSvc.Service
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}

	return &pb.GetServicesInfoResponse{
		Response:          pb.CreateResponse(pb.ResponseSuccess, "Get services info successfully."),
		AllServicesDetail: allServiceDetails,
		Statistics:        nil,
	}, nil
}

func (ds *DataSource) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	service, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to add tags for service [%s] for get service failed,", request.ServiceId)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if service == nil {
		return &pb.AddServiceTagsResponse{Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service not exist")}, nil
	}
	//todo add quto check
	dataTags := service.Tags
	tags := request.Tags
	for key, value := range dataTags {
		if _, ok := tags[key]; ok {
			continue
		}
		tags[key] = value
	}
	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: tags}})
	if err != nil {
		log.Errorf(err, "update service [%s] tags failed.", request.ServiceId)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Add service tags successfully."),
	}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to get service [%s] tags", request.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist"),
		}, nil
	}
	return &pb.GetServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get service tags successfully."),
		Tags:     svc.Tags,
	}, nil
}

func (ds *DataSource) UpdateTag(ctx context.Context, request *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to get service [%s] tags", request.ServiceId)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist"),
		}, nil
	}
	dataTags := svc.Tags
	if len(dataTags) > 0 {
		if _, ok := dataTags[request.Key]; !ok {
			return &pb.UpdateServiceTagResponse{
				Response: pb.CreateResponse(scerr.ErrTagNotExists, "Tag does not exist"),
			}, nil
		}
	}
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	newTags[request.Key] = request.Value

	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: newTags}})
	if err != nil {
		log.Errorf(err, "update service [%s] tags failed.", request.ServiceId)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service tag success."),
	}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to get service [%s] tags", request.ServiceId)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist"),
		}, nil
	}
	dataTags := svc.Tags
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	if len(dataTags) > 0 {
		for _, key := range request.Keys {
			if _, ok := dataTags[key]; !ok {
				return &pb.DeleteServiceTagsResponse{
					Response: pb.CreateResponse(scerr.ErrTagNotExists, "Tag does not exist"),
				}, nil
			}
			delete(newTags, key)
		}
	}
	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: newTags}})
	if err != nil {
		log.Errorf(err, "delete service [%s] tags failed.", request.ServiceId)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service tag success."),
	}, nil
}

func (ds *DataSource) GetSchema(ctx context.Context, request *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetSchema failed to check service exist."),
		}, nil
	}
	if !exist {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetSchema service does not exist."),
		}, nil
	}
	mgSchema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId))
	if err != nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetSchema failed from mongodb."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response:      pb.CreateResponse(pb.ResponseSuccess, "Get schema info successfully."),
		Schema:        mgSchema.Schema,
		SchemaSummary: mgSchema.SchemaSummary,
	}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetAllSchemas failed for get service failed"),
		}, nil
	}
	if !exist {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetAllSchemas failed for service not exist"),
		}, nil
	}

	schemas, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetAllSchemas failed for get schemas failed"),
		}, nil
	}
	return &pb.GetAllSchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all schema info successfully."),
		Schemas:  schemas,
	}, nil
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "ExistSchema failed for get service failed"),
		}, nil
	}
	if !exist {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "ExistSchema failed for service not exist"),
		}, nil
	}
	mgSchema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId))
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ExistSchema failed for get schema failed."),
		}, nil
	}
	if mgSchema == nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "ExistSchema failed for schema not exist."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Schema exist."),
		Summary:   mgSchema.SchemaSummary,
		SchemaId:  mgSchema.SchemaID,
		ServiceId: mgSchema.ServiceID,
	}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "DeleteSchema failed for get service failed."),
		}, nil
	}
	if !exist {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "DeleteSchema failed for service not exist."),
		}, nil
	}
	filter := GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId)
	_, err = client.GetMongoClient().Delete(ctx, CollectionSchema, filter)
	if err != nil {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "DeleteSchema failed for delete schema failed."),
		}, nil
	}
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Delete schema info successfully."),
	}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId
	schema := pb.Schema{
		SchemaId: request.SchemaId,
		Summary:  request.Summary,
		Schema:   request.Schema,
	}
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ModifySchema failed to create session."),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ModifySchema failed to start session."),
		}, err
	}
	defer session.EndSession(ctx)
	err = ds.modifySchema(ctx, request.ServiceId, &schema)
	if err != nil {
		log.Errorf(err, "modify schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP)
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.ModifySchemaResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Txn ModifySchema Abort failed."),
			}, errAbort
		}
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Txn ModifySchema failed."),
		}, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Txn ModifySchema CommitTransaction failed."),
		}, err
	}
	log.Infof("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schema info success."),
	}, nil
}

func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(scerr.ErrInternal, err.Error())}, err
	}
	if svc == nil {
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service not exist")}, nil
	}
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ModifySchemas failed to start session"),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ModifySchemas failed to start session"),
		}, err
	}
	defer session.EndSession(ctx)
	err = ds.modifySchemas(ctx, svc.Service, request.Schemas)
	if err != nil {
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.ModifySchemasResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Txn ModifySchemas Abort failed."),
			}, errAbort
		}
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(scerr.ErrInternal, err.Error())}, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Txn ModifySchemas CommitTransaction failed."),
		}, err
	}
	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schemas info success"),
	}, nil

}

func (ds *DataSource) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	if svc == nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
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
	if !ds.isSchemaEditable(microservice) {
		if len(microservice.Schemas) != 0 && !isExist {
			return scerr.NewError(scerr.ErrUndefinedSchemaID, "Non-existent schemaID can't be added request "+pb.ENV_PROD)
		}
		respSchema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId))
		if err != nil {
			return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
		}
		if schema != nil {
			if len(schema.Summary) == 0 {
				log.Errorf(err, "modify schema[%s/%s] failed, get schema summary failed, operator: %s",
					serviceID, schema.SchemaId, remoteIP)
				return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
			}
			if len(respSchema.SchemaSummary) != 0 {
				log.Errorf(err, "%s mode, schema[%s/%s] already exist, can not be changed, operator: %s",
					pb.ENV_PROD, serviceID, schema.SchemaId, remoteIP)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not be changed request "+pb.ENV_PROD)
			}
		}
		if len(microservice.Schemas) == 0 {
			copy(newSchemas, microservice.Schemas)
			newSchemas = append(newSchemas, schema.SchemaId)
		}
	} else {
		if !isExist {
			copy(newSchemas, microservice.Schemas)
			newSchemas = append(newSchemas, schema.SchemaId)
		}
	}
	if len(newSchemas) != len(microservice.Schemas) {
		err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: newSchemas}})
		if err != nil {
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}
	newSchema := bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}
	err = UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), newSchema, options.FindOneAndUpdate().SetUpsert(true))
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return nil
}

func (ds *DataSource) modifySchemas(ctx context.Context, service *pb.MicroService, schemas []*pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		log.Errorf(nil, "modify service[%s] schemas failed, get schemas failed, operator: %s",
			serviceID, remoteIP)
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		schemasAnalysis(schemas, schemasFromDatabase, service.Schemas)
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			//todo add quota check
			err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: nonExistSchemaIds}})
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
					serviceID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Errorf(errInfo, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return scerr.NewError(scerr.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := SchemaExist(ctx, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return scerr.NewError(scerr.ErrInternal, err.Error())
				}
				if !exist {
					err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, needUpdateSchema.SchemaId), bson.M{"$set": bson.M{Schema: needUpdateSchema.Schema, SchemaSummary: needUpdateSchema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
					if err != nil {
						return scerr.NewError(scerr.ErrInternal, err.Error())
					}
				} else {
					log.Warnf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP)
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
		}
	} else {

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Infof("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Infof("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			err = DeleteSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
		}

		err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: schemaIDs}})
		if err != nil {
			log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
				serviceID, remoteIP)
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}
	return nil
}

func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "failed to add rules for service [%s] for get service failed,", request.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if !exist {
		return &pb.AddServiceRulesResponse{Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist")}, nil
	}
	//todo add quota check
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	var ruleType string
	if len(rules) != 0 {
		ruleType = rules[0].RuleType
	}
	ruleIDs := make([]string, 0, len(request.Rules))
	for _, rule := range request.Rules {
		if len(ruleType) == 0 {
			ruleType = rule.RuleType
		} else if ruleType != rule.RuleType {
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrBlackAndWhiteRule, "Service can only contain one rule type,Black or white."),
			}, nil
		}
		//the rule unique index is (serviceid,attribute,pattern)
		exist, err := RuleExist(ctx, GeneratorRuleAttFilter(ctx, request.ServiceId, rule.Attribute, rule.Pattern))
		if err != nil {
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Can not check rule if exist."),
			}, nil
		}
		if exist {
			continue
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &MgRule{
			Domain:    util.ParseDomain(ctx),
			Project:   util.ParseProject(ctx),
			ServiceID: request.ServiceId,
			Rule: &pb.ServiceRule{
				RuleId:       util.GenerateUUID(),
				RuleType:     rule.RuleType,
				Attribute:    rule.Attribute,
				Pattern:      rule.Pattern,
				Description:  rule.Description,
				Timestamp:    timestamp,
				ModTimestamp: timestamp,
			},
		}
		ruleIDs = append(ruleIDs, ruleAdd.Rule.RuleId)
		_, err = client.GetMongoClient().Insert(ctx, CollectionRule, ruleAdd)
		if err != nil {
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
	}
	return &pb.AddServiceRulesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Add service rules successfully."),
		RuleIds:  ruleIDs,
	}, nil
}

func (ds *DataSource) GetRules(ctx context.Context, request *pb.GetServiceRulesRequest) (
	*pb.GetServiceRulesResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetRules failed for get service failed."),
		}, nil
	}
	if !exist {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetRules failed for service not exist."),
		}, nil
	}
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.GetServiceRulesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get service rules successfully."),
		Rules:    rules,
	}, nil
}

func (ds *DataSource) DeleteRule(ctx context.Context, request *pb.DeleteServiceRulesRequest) (
	*pb.DeleteServiceRulesResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "failed to add tags for service [%s] for get service failed,", request.ServiceId)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Failed to check service exist"),
		}, err
	}
	if !exist {
		return &pb.DeleteServiceRulesResponse{Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service not exist")}, nil
	}
	for _, ruleID := range request.RuleIds {
		exist, err := RuleExist(ctx, GeneratorRuleFilter(ctx, request.ServiceId, ruleID))
		if err != nil {
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, nil
		}
		if !exist {
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
			}, nil
		}
	}

	return &pb.DeleteServiceRulesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Delete service rules successfully."),
	}, nil
}

func (ds *DataSource) UpdateRule(ctx context.Context, request *pb.UpdateServiceRuleRequest) (
	*pb.UpdateServiceRuleResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "UpdateRule failed for get service failed."),
			//Schemas:  nil,
		}, nil
	}
	if !exist {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "UpdateRule failed for service not exist."),
		}, nil
	}
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "UpdateRule failed for get rule."),
		}, nil
	}
	if len(rules) >= 1 && rules[0].RuleType != request.Rule.RuleType {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrModifyRuleNotAllow, "Exist multiple rules, can not change rule type. Rule type is ."+rules[0].RuleType),
		}, nil
	}
	exist, err = RuleExist(ctx, GeneratorRuleFilter(ctx, request.ServiceId, request.RuleId))
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	if !exist {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrRuleNotExists, "This rule does not exist."),
		}, nil
	}

	newRule := bson.M{"$set": bson.M{RuleRuletype: request.Rule.RuleType,
		RulePattern: request.Rule.Pattern, RuleAttribute: request.Rule.Attribute,
		RuleDescription: request.Rule.Description,
		RuleModTime:     strconv.FormatInt(time.Now().Unix(), 10)}}

	err = UpdateRule(ctx, GeneratorRuleFilter(ctx, request.ServiceId, request.RuleId), newRule)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service rules succesfully."),
	}, nil
}

func (ds *DataSource) isSchemaEditable(service *pb.MicroService) bool {
	return (len(service.Environment) != 0 && service.Environment != pb.ENV_PROD) || ds.SchemaEditable
}

func ServiceExist(ctx context.Context, service *pb.MicroServiceKey) (bool, error) {
	filter := GeneratorServiceNameFilter(ctx, service)
	return client.GetMongoClient().DocExist(ctx, CollectionService, filter)
}

func ServiceExistID(ctx context.Context, serviceID string) (bool, error) {
	filter := GeneratorServiceFilter(ctx, serviceID)
	return client.GetMongoClient().DocExist(ctx, CollectionService, filter)
}

func GetService(ctx context.Context, filter bson.M) (*MgService, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var svc *MgService
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	err = findRes.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func GetServices(ctx context.Context, filter bson.M) ([]*pb.MicroService, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var services []*pb.MicroService
	for res.Next(ctx) {
		var tmp MgService
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.Service)
	}
	return services, nil
}

func GetMgServices(ctx context.Context, filter bson.M) ([]*MgService, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var services []*MgService
	for res.Next(ctx) {
		var tmp *MgService
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp)
	}
	return services, nil
}

func GetServicesVersions(ctx context.Context, filter interface{}) ([]string, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return nil, nil
	}
	var versions []string
	for res.Next(ctx) {
		var tmp string
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		versions = append(versions, tmp)
	}
	return versions, nil
}

func getServiceDetailUtil(ctx context.Context, mgs *MgService, countOnly bool, options []string) (*pb.ServiceDetail, error) {
	serviceDetail := new(pb.ServiceDetail)
	if countOnly {
		serviceDetail.Statics = new(pb.Statistics)
	}
	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			serviceDetail.Tags = mgs.Tags
		case "rules":
			rules, err := GetRules(ctx, mgs.Service.ServiceId)
			if err != nil {
				log.Errorf(err, "get service[%s]'s all rules failed", mgs.Service.ServiceId)
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			//todo wait instance interface
		case "schemas":
			schemas, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, mgs.Service.ServiceId))
			if err != nil {
				log.Errorf(err, "get service[%s]'s all schemas failed", mgs.Service.ServiceId)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			//todo wait dependencied interface
		case "":
			continue
		default:
			log.Errorf(nil, "request option[%s] is invalid", opt)
		}
	}
	return serviceDetail, nil
}

func UpdateService(ctx context.Context, filter interface{}, m bson.M) error {
	return client.GetMongoClient().DocUpdate(ctx, CollectionService, filter, m)
}

func GetRules(ctx context.Context, serviceID string) ([]*pb.ServiceRule, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{Domain: domain, Project: project, RuleServiceID: serviceID}

	ruleRes, err := client.GetMongoClient().Find(ctx, CollectionRule, filter)
	if err != nil {
		return nil, err
	}
	var rules []*pb.ServiceRule
	for ruleRes.Next(ctx) {
		var tmpRule *MgRule
		err := ruleRes.Decode(&tmpRule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, tmpRule.Rule)
	}
	return rules, nil
}

func UpdateRule(ctx context.Context, filter interface{}, m bson.M) error {
	return client.GetMongoClient().DocUpdate(ctx, CollectionRule, filter, m)
}

func UpdateSchema(ctx context.Context, filter interface{}, m bson.M, opts ...*options.FindOneAndUpdateOptions) error {
	return client.GetMongoClient().DocUpdate(ctx, CollectionSchema, filter, m, opts...)
}

func DeleteSchema(ctx context.Context, filter interface{}) error {
	res, err := client.GetMongoClient().DocDelete(ctx, CollectionSchema, filter)
	if err != nil {
		return err
	}
	if !res {
		return errors.New("delete schema failed")
	}
	return nil
}

func RuleExist(ctx context.Context, filter bson.M) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, CollectionRule, filter)
}

func GeneratorServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, ServiceServiceID: serviceID}
}

func GeneratorServiceNameFilter(ctx context.Context, service *pb.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, ServiceEnv: service.Environment, ServiceAppID: service.AppId, ServiceServiceName: service.ServiceName, ServiceVersion: service.Version}
}

func GeneratorServiceAliasFilter(ctx context.Context, service *pb.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, ServiceEnv: service.Environment, ServiceAppID: service.AppId, ServiceAlias: service.Alias, ServiceVersion: service.Version}
}

func GeneratorRuleAttFilter(ctx context.Context, serviceID, attribute, pattern string) bson.M {
	return bson.M{RuleServiceID: serviceID, RuleAttribute: attribute, RulePattern: pattern}
}

func GeneratorSchemaFilter(ctx context.Context, serviceID, schemaID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, SchemaServiceID: serviceID, SchemaID: schemaID}
}

func GeneratorRuleFilter(ctx context.Context, serviceID, ruleID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, RuleServiceID: serviceID, RuleRuleID: ruleID}
}

func GetSchemas(ctx context.Context, filter bson.M) ([]*pb.Schema, error) {
	getRes, err := client.GetMongoClient().Find(ctx, CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schemas []*pb.Schema
	for getRes.Next(ctx) {
		var tmp *MgSchema
		err = getRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, &pb.Schema{
			SchemaId: tmp.SchemaID,
			Summary:  tmp.SchemaSummary,
			Schema:   tmp.Schema,
		})
	}
	return schemas, nil
}

func GetSchema(ctx context.Context, filter bson.M) (*MgSchema, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schema *MgSchema
	err = findRes.Decode(&schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func SchemaExist(ctx context.Context, serviceID, schemaID string) (bool, error) {
	num, err := client.GetMongoClient().Count(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schemaID))
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

// Instance management
func (ds *DataSource) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	return &pb.RegisterInstanceResponse{}, nil
}

// GetInstances returns instances under the current domain
func (ds *DataSource) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	return &pb.GetOneInstanceResponse{}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	return &pb.GetInstancesResponse{}, nil
}

// GetProviderInstances returns instances under the specified domain
func (ds *DataSource) GetProviderInstances(ctx context.Context, request *pb.HeartbeatSetElement) (instances []*pb.MicroServiceInstance, rev string, err error) {
	return nil, "", nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	return nil, "", nil
}

// FindInstances returns instances under the specified domain
func (ds *DataSource) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	return &pb.FindInstancesResponse{}, nil
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	return &pb.UpdateInstanceStatusResponse{}, nil
}

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	return &pb.UpdateInstancePropsResponse{}, nil
}

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	return &pb.UnregisterInstanceResponse{}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return heartbeat.Instance().Heartbeat(ctx, request)
}

func (ds *DataSource) HeartbeatSet(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	return &pb.HeartbeatSetResponse{}, nil
}

func (ds *DataSource) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	return &pb.BatchFindInstancesResponse{}, nil
}

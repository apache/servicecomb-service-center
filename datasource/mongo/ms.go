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
	"fmt"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	service := request.Service

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
			Response: pb.CreateResponse(pb.ErrInternal, "Check service exist failed"),
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
			Response: pb.CreateResponse(pb.ErrInternal, "Check service exist failed"),
		}, err
	}
	if existID || exist {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrServiceAlreadyExists, "ServiceID conflict or found the same service."),
		}, nil
	}
	insertRes, err := client.GetMongoClient().Insert(ctx, CollectionService, &Service{Domain: domain, Project: project, ServiceInfo: service})
	if err != nil {
		if client.IsDuplicateKey(err) {
			return &pb.CreateServiceResponse{
				Response: pb.CreateResponse(pb.ErrServiceAlreadyExists, "ServiceID or ServiceInfo conflict."),
			}, nil
		}
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Register service failed."),
		}, err
	}

	remoteIP := util.GetIPFromContext(ctx)
	log.Info(fmt.Sprintf("create micro-service[%s][%s] successfully,operator: %s",
		service.ServiceId, insertRes.InsertedID, remoteIP))

	return &pb.CreateServiceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Register service successfully"),
		ServiceId: service.ServiceId,
	}, nil
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{ColumnDomain: domain, ColumnProject: project}

	services, err := GetServices(ctx, filter)
	if err != nil {
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "get services data failed."),
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

	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnServiceInfo, ColumnEnv}): request.Environment}

	services, err := GetServices(ctx, filter)
	if err != nil {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "get services data failed."),
		}, nil
	}
	l := len(services)
	if l == 0 {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "get services data failed."),
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
		log.Error(fmt.Sprintf("failed to get single service %s from mongo", request.ServiceId), err)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "get service data from mongodb failed."),
		}, err
	}
	if svc != nil {
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(pb.ResponseSuccess, "Get service successfully."),
			Service:  svc.ServiceInfo,
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service not exist."),
	}, nil
}

func (ds *DataSource) ExistServiceByID(ctx context.Context, request *pb.GetExistenceByIDRequest) (*pb.GetExistenceByIDResponse, error) {

	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetExistenceByIDResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Check service exist failed."),
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
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if len(services) != 0 {
		return &pb.GetExistenceResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "get service id successfully."),
			ServiceId: services[0].ServiceId,
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist"),
	}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Delete service failed,failed to get service."),
		}, err
	}
	if !exist {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Delete service failed,service not exist."),
		}, nil
	}
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "DelService failed to create session."),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "DelService failed to start session."),
		}, err
	}
	defer session.EndSession(ctx)
	//todo delete instance,tags,schemas...
	res, err := DelServicePri(ctx, request.ServiceId, request.Force)
	if err != nil {
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.DeleteServiceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, "Txn delete service abort failed."),
			}, errAbort
		}
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Delete service failed"),
		}, err
	}
	errCommit := session.CommitTransaction(ctx)
	if errCommit != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Txn delete service commit failed."),
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
		log.Error(fmt.Sprintf("%s micro-service %s failed, operator: %s", title, serviceID, remoteIP), ErrNotAllowDeleteSC)
		return pb.CreateResponse(pb.ErrInvalidParams, ErrNotAllowDeleteSC.Error()), nil
	}
	microservice, err := GetService(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		log.Error(fmt.Sprintf("%s micro-service %s failed, get service file failed, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}
	if microservice == nil {
		log.Error(fmt.Sprintf("%s micro-service %s failed, service does not exist, operator: %s",
			title, serviceID, remoteIP), err)
		return pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."), nil
	}
	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		log.Info("force delete,should del instance...")
		//todo wait for dep interface
	}
	filter := GeneratorServiceFilter(ctx, serviceID)
	//todo del instances
	tables := []string{CollectionService, CollectionSchema, CollectionRule}
	for _, col := range tables {
		_, err := client.GetMongoClient().Delete(ctx, col, filter)
		if err != nil {
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
	}
	return pb.CreateResponse(pb.ResponseSuccess, "Unregister service successfully."), nil

}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {

	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "UpdateService failed,failed to get service."),
		}, err
	}
	if !exist {
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "UpdateService failed,service not exist."),
		}, nil
	}

	updateData := bson.M{
		"$set": bson.M{
			StringBuilder([]string{ColumnServiceInfo, ColumnModTime}):  strconv.FormatInt(time.Now().Unix(), 10),
			StringBuilder([]string{ColumnServiceInfo, ColumnProperty}): request.Properties}}
	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), updateData)
	if err != nil {
		log.Error(fmt.Sprintf("update service %s properties failed, update mongo failed", request.ServiceId), err)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, "Update doc in mongo failed."),
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
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if mgSvc == nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	svc := mgSvc.ServiceInfo
	versions, err := GetServicesVersions(ctx, bson.M{})
	if err != nil {
		log.Error(fmt.Sprintf("get service %s %s %s all versions failed", svc.Environment, svc.AppId, svc.ServiceName), err)
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}
	serviceInfo, err := getServiceDetailUtil(ctx, mgSvc, false, options)
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
	services, err := GetMongoServices(ctx, bson.M{})
	if err != nil {
		log.Error("get all services by domain failed", err)
		return &pb.GetServicesInfoResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	allServiceDetails := make([]*pb.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	for _, mgSvc := range services {
		if !request.WithShared && apt.IsGlobal(pb.MicroServiceToKey(domainProject, mgSvc.ServiceInfo)) {
			continue
		}
		if len(request.AppId) > 0 {
			if request.AppId != mgSvc.ServiceInfo.AppId {
				continue
			}
			if len(request.ServiceName) > 0 && request.ServiceName != mgSvc.ServiceInfo.ServiceName {
				continue
			}
		}

		serviceDetail, err := getServiceDetailUtil(ctx, mgSvc, request.CountOnly, options)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		serviceDetail.MicroService = mgSvc.ServiceInfo
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
		log.Error(fmt.Sprintf("failed to add tags for service %s for get service failed", request.ServiceId), err)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if service == nil {
		return &pb.AddServiceTagsResponse{Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service not exist")}, nil
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
	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ColumnTag: tags}})
	if err != nil {
		log.Error(fmt.Sprintf("update service %s tags failed.", request.ServiceId), err)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Add service tags successfully."),
	}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist"),
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
		log.Error(fmt.Sprintf("failed to get %s tags", request.ServiceId), err)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist"),
		}, nil
	}
	dataTags := svc.Tags
	if len(dataTags) > 0 {
		if _, ok := dataTags[request.Key]; !ok {
			return &pb.UpdateServiceTagResponse{
				Response: pb.CreateResponse(pb.ErrTagNotExists, "Tag does not exist"),
			}, nil
		}
	}
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	newTags[request.Key] = request.Value

	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ColumnTag: newTags}})
	if err != nil {
		log.Error(fmt.Sprintf("update service %s tags failed", request.ServiceId), err)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service tag success."),
	}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	if svc == nil {
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist"),
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
					Response: pb.CreateResponse(pb.ErrTagNotExists, "Tag does not exist"),
				}, nil
			}
			delete(newTags, key)
		}
	}
	err = UpdateService(ctx, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ColumnTag: newTags}})
	if err != nil {
		log.Error(fmt.Sprintf("delete service %s tags failed", request.ServiceId), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
			Response: pb.CreateResponse(pb.ErrInternal, "GetSchema failed to check service exist."),
		}, nil
	}
	if !exist {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "GetSchema service does not exist."),
		}, nil
	}
	Schema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId))
	if err != nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "GetSchema failed from mongodb."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response:      pb.CreateResponse(pb.ResponseSuccess, "Get schema info successfully."),
		Schema:        Schema.SchemaInfo,
		SchemaSummary: Schema.SchemaSummary,
	}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "GetAllSchemas failed for get service failed"),
		}, nil
	}
	if !exist {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "GetAllSchemas failed for service not exist"),
		}, nil
	}

	schemas, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "GetAllSchemas failed for get schemas failed"),
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
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "ExistSchema failed for get service failed"),
		}, nil
	}
	if !exist {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "ExistSchema failed for service not exist"),
		}, nil
	}
	Schema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId))
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "ExistSchema failed for get schema failed."),
		}, nil
	}
	if Schema == nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(pb.ErrSchemaNotExists, "ExistSchema failed for schema not exist."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Schema exist."),
		Summary:   Schema.SchemaSummary,
		SchemaId:  Schema.SchemaID,
		ServiceId: Schema.ServiceID,
	}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "DeleteSchema failed for get service failed."),
		}, nil
	}
	if !exist {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "DeleteSchema failed for service not exist."),
		}, nil
	}
	filter := GeneratorSchemaFilter(ctx, request.ServiceId, request.SchemaId)
	_, err = client.GetMongoClient().Delete(ctx, CollectionSchema, filter)
	if err != nil {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, "DeleteSchema failed for delete schema failed."),
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
			Response: pb.CreateResponse(pb.ErrInternal, "ModifySchema failed to create session."),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "ModifySchema failed to start session."),
		}, err
	}
	defer session.EndSession(ctx)
	err = ds.modifySchema(ctx, request.ServiceId, &schema)
	if err != nil {
		log.Error(fmt.Sprintf("modify schema %s %s failed, operator %s", serviceID, schemaID, remoteIP), err)
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.ModifySchemaResponse{
				Response: pb.CreateResponse(pb.ErrInternal, "Txn ModifySchema Abort failed."),
			}, errAbort
		}
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Txn ModifySchema failed."),
		}, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Txn ModifySchema CommitTransaction failed."),
		}, err
	}
	log.Info(fmt.Sprintf("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP))
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schema info success."),
	}, nil
}

func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(pb.ErrInternal, err.Error())}, err
	}
	if svc == nil {
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service not exist")}, nil
	}
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "ModifySchemas failed to start session"),
		}, err
	}
	if err = session.StartTransaction(); err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "ModifySchemas failed to start session"),
		}, err
	}
	defer session.EndSession(ctx)
	err = ds.modifySchemas(ctx, svc.ServiceInfo, request.Schemas)
	if err != nil {
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return &pb.ModifySchemasResponse{
				Response: pb.CreateResponse(pb.ErrInternal, "Txn ModifySchemas Abort failed."),
			}, errAbort
		}
		return &pb.ModifySchemasResponse{Response: pb.CreateResponse(pb.ErrInternal, err.Error())}, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Txn ModifySchemas CommitTransaction failed."),
		}, err
	}
	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "modify schemas info success"),
	}, nil

}

func (ds *DataSource) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *pb.Error {
	remoteIP := util.GetIPFromContext(ctx)
	svc, err := GetService(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if svc == nil {
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}
	microservice := svc.ServiceInfo
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
			return pb.NewError(pb.ErrUndefinedSchemaID, "Non-existent schemaID can't be added request "+pb.ENV_PROD)
		}
		respSchema, err := GetSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId))
		if err != nil {
			return pb.NewError(pb.ErrUnavailableBackend, err.Error())
		}
		if schema != nil {
			if len(schema.Summary) == 0 {
				log.Error(fmt.Sprintf("modify schema %s %s failed, get schema summary failed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return pb.NewError(pb.ErrUnavailableBackend, err.Error())
			}
			if len(respSchema.SchemaSummary) != 0 {
				log.Error(fmt.Sprintf("mode, schema %s %s already exist, can not be changed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return pb.NewError(pb.ErrModifySchemaNotAllow, "schema already exist, can not be changed request "+pb.ENV_PROD)
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

		updateData := bson.M{StringBuilder([]string{ColumnServiceInfo, ColumnSchemas}): newSchemas}
		err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": updateData})
		if err != nil {
			return pb.NewError(pb.ErrInternal, err.Error())
		}
	}
	newSchema := bson.M{"$set": bson.M{ColumnSchemaInfo: schema.Schema, ColumnSchemaSummary: schema.Summary}}
	err = UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), newSchema, options.FindOneAndUpdate().SetUpsert(true))
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	return nil
}

func (ds *DataSource) modifySchemas(ctx context.Context, service *pb.MicroService, schemas []*pb.Schema) *pb.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		log.Error(fmt.Sprintf("modify service %s schemas failed, get schemas failed, operator: %s", serviceID, remoteIP), err)
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		datasource.SchemasAnalysis(schemas, schemasFromDatabase, service.Schemas)
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			//todo add quota check
			updateData := bson.M{StringBuilder([]string{ColumnServiceInfo, ColumnSchemas}): nonExistSchemaIds}
			err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": updateData})
			if err != nil {
				log.Error(fmt.Sprintf("modify service %s schemas failed, update service.Schemas failed, operator: %s",
					serviceID, remoteIP), err)
				return pb.NewError(pb.ErrInternal, err.Error())
			}
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Error(fmt.Sprintf("modify service %s schemas failed, operator: %s", serviceID, remoteIP), err)
				return pb.NewError(pb.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := SchemaExist(ctx, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return pb.NewError(pb.ErrInternal, err.Error())
				}
				if !exist {
					err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, needUpdateSchema.SchemaId), bson.M{"$set": bson.M{ColumnSchemaInfo: needUpdateSchema.Schema, ColumnSchemaSummary: needUpdateSchema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
					if err != nil {
						return pb.NewError(pb.ErrInternal, err.Error())
					}
				} else {
					log.Warn(fmt.Sprintf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP))
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{ColumnSchemaInfo: schema.Schema, ColumnSchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return pb.NewError(pb.ErrInternal, err.Error())
			}
		}
	} else {

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Info(fmt.Sprintf("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{ColumnSchemaInfo: schema.Schema, ColumnSchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Info(fmt.Sprintf("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			err := UpdateSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{ColumnSchemaInfo: schema.Schema, ColumnSchemaSummary: schema.Summary}}, options.FindOneAndUpdate().SetUpsert(true))
			if err != nil {
				return pb.NewError(pb.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Info(fmt.Sprintf("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP))
			err = DeleteSchema(ctx, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId))
			if err != nil {
				return pb.NewError(pb.ErrInternal, err.Error())
			}
		}

		updateData := bson.M{StringBuilder([]string{ColumnServiceInfo, ColumnSchemas}): schemaIDs}
		err := UpdateService(ctx, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": updateData})
		if err != nil {
			log.Error(fmt.Sprintf("modify service %s schemas failed, update service.Schemas failed, operator: %s", serviceID, remoteIP), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
	}
	return nil
}

func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	exist, err := ServiceExistID(ctx, request.ServiceId)
	if err != nil {
		log.Error(fmt.Sprintf("failed to add rules for service %s for get service failed", request.ServiceId), err)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if !exist {
		return &pb.AddServiceRulesResponse{Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service does not exist")}, nil
	}
	//todo add quota check
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
				Response: pb.CreateResponse(pb.ErrBlackAndWhiteRule, "Service can only contain one rule type,Black or white."),
			}, nil
		}
		//the rule unique index is (serviceid,attribute,pattern)
		exist, err := RuleExist(ctx, GeneratorRuleAttFilter(ctx, request.ServiceId, rule.Attribute, rule.Pattern))
		if err != nil {
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.ErrUnavailableBackend, "Can not check rule if exist."),
			}, nil
		}
		if exist {
			continue
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		ruleAdd := &Rule{
			Domain:    util.ParseDomain(ctx),
			Project:   util.ParseProject(ctx),
			ServiceID: request.ServiceId,
			RuleInfo: &pb.ServiceRule{
				RuleId:       util.GenerateUUID(),
				RuleType:     rule.RuleType,
				Attribute:    rule.Attribute,
				Pattern:      rule.Pattern,
				Description:  rule.Description,
				Timestamp:    timestamp,
				ModTimestamp: timestamp,
			},
		}
		ruleIDs = append(ruleIDs, ruleAdd.RuleInfo.RuleId)
		_, err = client.GetMongoClient().Insert(ctx, CollectionRule, ruleAdd)
		if err != nil {
			return &pb.AddServiceRulesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "GetRules failed for get service failed."),
		}, nil
	}
	if !exist {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "GetRules failed for service not exist."),
		}, nil
	}
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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
		log.Error(fmt.Sprintf("failed to add tags for service %s for get service failed", request.ServiceId), err)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "Failed to check service exist"),
		}, err
	}
	if !exist {
		return &pb.DeleteServiceRulesResponse{Response: pb.CreateResponse(pb.ErrServiceNotExists, "Service not exist")}, nil
	}
	for _, ruleID := range request.RuleIds {
		exist, err := RuleExist(ctx, GeneratorRuleFilter(ctx, request.ServiceId, ruleID))
		if err != nil {
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, nil
		}
		if !exist {
			return &pb.DeleteServiceRulesResponse{
				Response: pb.CreateResponse(pb.ErrRuleNotExists, "This rule does not exist."),
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
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "UpdateRule failed for get service failed."),
		}, nil
	}
	if !exist {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, "UpdateRule failed for service not exist."),
		}, nil
	}
	rules, err := GetRules(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, "UpdateRule failed for get rule."),
		}, nil
	}
	if len(rules) >= 1 && rules[0].RuleType != request.Rule.RuleType {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrModifyRuleNotAllow, "Exist multiple rules, can not change rule type. Rule type is ."+rules[0].RuleType),
		}, nil
	}
	exist, err = RuleExist(ctx, GeneratorRuleFilter(ctx, request.ServiceId, request.RuleId))
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, nil
	}
	if !exist {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrRuleNotExists, "This rule does not exist."),
		}, nil
	}

	newRule := bson.M{
		StringBuilder([]string{ColumnRuleInfo, ColumnRuleType}):    request.Rule.RuleType,
		StringBuilder([]string{ColumnRuleInfo, ColumnPattern}):     request.Rule.Pattern,
		StringBuilder([]string{ColumnRuleInfo, ColumnAttribute}):   request.Rule.Attribute,
		StringBuilder([]string{ColumnRuleInfo, ColumnDescription}): request.Rule.Description,
		StringBuilder([]string{ColumnRuleInfo, ColumnModTime}):     strconv.FormatInt(time.Now().Unix(), 10)}

	err = UpdateRule(ctx, GeneratorRuleFilter(ctx, request.ServiceId, request.RuleId), bson.M{"$set": newRule})
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
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

func GetService(ctx context.Context, filter bson.M) (*Service, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var svc *Service
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
		var tmp Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.ServiceInfo)
	}
	return services, nil
}

func GetMongoServices(ctx context.Context, filter bson.M) ([]*Service, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var services []*Service
	for res.Next(ctx) {
		var tmp *Service
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

func getServiceDetailUtil(ctx context.Context, mgs *Service, countOnly bool, options []string) (*pb.ServiceDetail, error) {
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
			rules, err := GetRules(ctx, mgs.ServiceInfo.ServiceId)
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all rules failed", mgs.ServiceInfo.ServiceId), err)
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			//todo wait instance interface
		case "schemas":
			schemas, err := GetSchemas(ctx, GeneratorServiceFilter(ctx, mgs.ServiceInfo.ServiceId))
			if err != nil {
				log.Error(fmt.Sprintf("get service %s's all schemas failed", mgs.ServiceInfo.ServiceId), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			//todo wait dependencied interface
		case "":
			continue
		default:
			log.Info(fmt.Sprintf("request option %s is invalid", opt))
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
	filter := bson.M{ColumnDomain: domain, ColumnProject: project, ColumnServiceID: serviceID}

	ruleRes, err := client.GetMongoClient().Find(ctx, CollectionRule, filter)
	if err != nil {
		return nil, err
	}
	var rules []*pb.ServiceRule
	for ruleRes.Next(ctx) {
		var tmpRule *Rule
		err := ruleRes.Decode(&tmpRule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, tmpRule.RuleInfo)
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
		return ErrDeleteSchemaFailed
	}
	return nil
}

func RuleExist(ctx context.Context, filter bson.M) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, CollectionRule, filter)
}

func GeneratorServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnServiceInfo, ColumnServiceID}): serviceID}
}

func GeneratorServiceNameFilter(ctx context.Context, service *pb.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):         service.Environment,
		StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):       service.AppId,
		StringBuilder([]string{ColumnServiceInfo, ColumnServiceName}): service.ServiceName,
		StringBuilder([]string{ColumnServiceInfo, ColumnVersion}):     service.Version}
}

func GeneratorServiceAliasFilter(ctx context.Context, service *pb.MicroServiceKey) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnServiceInfo, ColumnEnv}):     service.Environment,
		StringBuilder([]string{ColumnServiceInfo, ColumnAppID}):   service.AppId,
		StringBuilder([]string{ColumnServiceInfo, ColumnAlias}):   service.Alias,
		StringBuilder([]string{ColumnServiceInfo, ColumnVersion}): service.Version}
}

func GeneratorRuleAttFilter(ctx context.Context, serviceID, attribute, pattern string) bson.M {
	return bson.M{
		ColumnServiceID: serviceID,
		StringBuilder([]string{ColumnRuleInfo, ColumnAttribute}): attribute,
		StringBuilder([]string{ColumnRuleInfo, ColumnPattern}):   pattern}
}

func GeneratorSchemaFilter(ctx context.Context, serviceID, schemaID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{ColumnDomain: domain, ColumnProject: project, ColumnServiceID: serviceID, ColumnSchemaID: schemaID}
}

func GeneratorRuleFilter(ctx context.Context, serviceID, ruleID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{
		ColumnDomain:    domain,
		ColumnProject:   project,
		ColumnServiceID: serviceID,
		StringBuilder([]string{ColumnRuleInfo, ColumnRuleID}): ruleID}
}

func GetSchemas(ctx context.Context, filter bson.M) ([]*pb.Schema, error) {
	getRes, err := client.GetMongoClient().Find(ctx, CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schemas []*pb.Schema
	for getRes.Next(ctx) {
		var tmp *Schema
		err = getRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, &pb.Schema{
			SchemaId: tmp.SchemaID,
			Summary:  tmp.SchemaSummary,
			Schema:   tmp.SchemaInfo,
		})
	}
	return schemas, nil
}

func GetSchema(ctx context.Context, filter bson.M) (*Schema, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schema *Schema
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
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance

	// 允许自定义 id
	if len(instance.InstanceId) > 0 {
		resp, err := ds.Heartbeat(ctx, &pb.HeartbeatRequest{
			InstanceId: instance.InstanceId,
			ServiceId:  instance.ServiceId,
		})
		if err != nil || resp == nil {
			log.Error(fmt.Sprintf("register service %s's instance failed, endpoints %s, host '%s', operator %s",
				instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
			return &pb.RegisterInstanceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, nil
		}
		switch resp.Response.GetCode() {
		case pb.ResponseSuccess:
			log.Info(fmt.Sprintf("register instance successful, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP))
			return &pb.RegisterInstanceResponse{
				Response:   resp.Response,
				InstanceId: instance.InstanceId,
			}, nil
		case pb.ErrInstanceNotExists:
			// register a new one
			return registryInstance(ctx, request)
		default:
			log.Error(fmt.Sprintf("register instance failed, reuse instance %s %s, operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP), err)
			return &pb.RegisterInstanceResponse{
				Response: resp.Response,
			}, err
		}
	}

	if err := preProcessRegisterInstance(ctx, instance); err != nil {
		log.Error(fmt.Sprintf("register service %s instance failed, endpoints %s, host %s operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP), err)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}, nil
	}
	return registryInstance(ctx, request)
}

// GetInstances returns instances under the current domain
func (ds *DataSource) GetInstance(ctx context.Context, request *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	service := &Service{}
	var err error
	if len(request.ConsumerServiceId) > 0 {
		filter := GeneratorServiceFilter(ctx, request.ConsumerServiceId)
		service, err = GetService(ctx, filter)
		if err != nil {
			log.Error(fmt.Sprintf(" get consumer failed, consumer %s find provider instance %s",
				request.ConsumerServiceId, request.ProviderInstanceId), err)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Error(fmt.Sprintf("consumer does not exist, consumer %s find provider instance %s %s",
				request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
			return &pb.GetOneInstanceResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	filter := GeneratorServiceFilter(ctx, request.ProviderServiceId)
	provider, err := GetService(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instance %s %s",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Error(fmt.Sprintf("provider does not exist, consumer %s find provider instance %s %s",
			request.ConsumerServiceId, request.ProviderServiceId, request.ProviderInstanceId), err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := func() string {
		return fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instance[%s]",
			request.ConsumerServiceId, service.ServiceInfo.Environment, service.ServiceInfo.AppId, service.ServiceInfo.ServiceName, service.ServiceInfo.Version,
			provider.ServiceInfo.ServiceId, provider.ServiceInfo.Environment, provider.ServiceInfo.AppId, provider.ServiceInfo.ServiceName, provider.ServiceInfo.Version,
			request.ProviderInstanceId)
	}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter = bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}): request.ProviderServiceId}
	findOneRes, err := client.GetMongoClient().FindOne(ctx, CollectionInstance, filter)
	if err != nil {
		mes := fmt.Errorf("%s failed, provider instance does not exist", findFlag())
		log.Error("FindInstances.GetWithProviderID failed", err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, mes.Error()),
		}, nil
	}
	var instance Instance
	err = findOneRes.Decode(&instance)
	if err != nil {
		log.Error(fmt.Sprintf("FindInstances.GetWithProviderID failed %s failed", findFlag()), err)
		return &pb.GetOneInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetOneInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get instance successfully."),
		Instance: instance.InstanceInfo,
	}, nil
}

func (ds *DataSource) GetInstances(ctx context.Context, request *pb.GetInstancesRequest) (*pb.GetInstancesResponse, error) {
	service := &Service{}
	var err error

	if len(request.ConsumerServiceId) > 0 {
		filter := GeneratorServiceFilter(ctx, request.ConsumerServiceId)
		service, err = GetService(ctx, filter)
		if err != nil {
			log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %sinstances",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Error(fmt.Sprintf("consumer does not exist, consumer %s find provider %s instances",
				request.ConsumerServiceId, request.ProviderServiceId), err)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
	}

	filter := GeneratorServiceFilter(ctx, request.ProviderServiceId)
	provider, err := GetService(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("get provider failed, consumer %s find provider instances %s",
			request.ConsumerServiceId, request.ProviderServiceId), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if provider == nil {
		log.Error(fmt.Sprintf("provider does not exist, consumer %s find provider %s  instances",
			request.ConsumerServiceId, request.ProviderServiceId), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists,
				fmt.Sprintf("Provider[%s] does not exist.", request.ProviderServiceId)),
		}, nil
	}

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s][%s/%s/%s/%s] instances",
		request.ConsumerServiceId, service.ServiceInfo.Environment, service.ServiceInfo.AppId, service.ServiceInfo.ServiceName, service.ServiceInfo.Version,
		provider.ServiceInfo.ServiceId, provider.ServiceInfo.Environment, provider.ServiceInfo.AppId, provider.ServiceInfo.ServiceName, provider.ServiceInfo.Version)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter = bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}): request.ProviderServiceId}
	resp, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter)
	if err != nil {
		log.Error(fmt.Sprintf("FindInstancesCache.Get failed %s failed", findFlag), err)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if resp == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", mes)
		return &pb.GetInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	var instances []*pb.MicroServiceInstance
	for resp.Next(ctx) {
		var instance Instance
		err := resp.Decode(&instance)
		if err != nil {
			log.Error(fmt.Sprintf("FindInstances.GetWithProviderID failed %s failed", findFlag), err)
			return &pb.GetInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		instances = append(instances, instance.InstanceInfo)
	}

	return &pb.GetInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

// GetProviderInstances returns instances under the specified domain
func (ds *DataSource) GetProviderInstances(ctx context.Context, request *pb.GetProviderInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}): request.ProviderServiceId}

	findRes, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter)
	if err != nil {
		return
	}

	for findRes.Next(ctx) {
		var mongoInstance Instance
		err := findRes.Decode(&mongoInstance)
		if err == nil {
			instances = append(instances, mongoInstance.InstanceInfo)
		}
	}

	return instances, "", nil
}

func (ds *DataSource) GetAllInstances(ctx context.Context, request *pb.GetAllInstancesRequest) (*pb.GetAllInstancesResponse, error) {

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{ColumnDomain: domain, ColumnProject: project}

	findRes, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	resp := &pb.GetAllInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all instances successfully"),
	}

	for findRes.Next(ctx) {
		var instance Instance
		err := findRes.Decode(&instance)
		if err != nil {
			return &pb.GetAllInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		resp.Instances = append(resp.Instances, instance.InstanceInfo)
	}

	return resp, nil
}

func (ds *DataSource) BatchGetProviderInstances(ctx context.Context, request *pb.BatchGetInstancesRequest) (instances []*pb.MicroServiceInstance, rev string, err error) {
	if request == nil || len(request.ServiceIds) == 0 {
		return nil, "", ErrInvalidParamBatchGetInstancesRequest
	}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	for _, providerServiceID := range request.ServiceIds {
		filter := bson.M{
			ColumnDomain:  domain,
			ColumnProject: project,
			StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}): providerServiceID}
		findRes, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter)
		if err != nil {
			return instances, "", nil
		}

		for findRes.Next(ctx) {
			var mongoInstance Instance
			err := findRes.Decode(&mongoInstance)
			if err == nil {
				instances = append(instances, mongoInstance.InstanceInfo)
			}
		}
	}

	return instances, "", nil
}

// FindInstances returns instances under the specified domain
func (ds *DataSource) FindInstances(ctx context.Context, request *pb.FindInstancesRequest) (*pb.FindInstancesResponse, error) {
	provider := &pb.MicroServiceKey{
		Tenant:      util.ParseTargetDomainProject(ctx),
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.VersionRule,
	}

	return ds.findInstance(ctx, request, provider)
}

func (ds *DataSource) UpdateInstanceStatus(ctx context.Context, request *pb.UpdateInstanceStatusRequest) (*pb.UpdateInstanceStatusResponse, error) {
	updateStatusFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId, request.Status}, "/")

	// todo finish get instance
	instance, err := GetInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance %s status failed, instance does not exist", updateStatusFlag), err)
		return &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.InstanceInfo.Status = request.Status

	if err := UpdateInstanceS(ctx, copyInstanceRef.InstanceInfo); err != nil {
		log.Error(fmt.Sprintf("update instance %s status failed", updateStatusFlag), err)
		resp := &pb.UpdateInstanceStatusResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] status successfully", updateStatusFlag)
	return &pb.UpdateInstanceStatusResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance status successfully."),
	}, nil
}

func (ds *DataSource) UpdateInstanceProperties(ctx context.Context, request *pb.UpdateInstancePropsRequest) (*pb.UpdateInstancePropsResponse, error) {
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")

	instance, err := GetInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if instance == nil {
		log.Error(fmt.Sprintf("update instance %s properties failed, instance does not exist", instanceFlag), err)
		return &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponse(pb.ErrInstanceNotExists, "Service instance does not exist."),
		}, nil
	}

	copyInstanceRef := *instance
	copyInstanceRef.InstanceInfo.Properties = request.Properties

	// todo finish update instance
	if err := UpdateInstanceP(ctx, copyInstanceRef.InstanceInfo); err != nil {
		log.Error(fmt.Sprintf("update instance %s properties failed", instanceFlag), err)
		resp := &pb.UpdateInstancePropsResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("update instance[%s] properties successfully", instanceFlag)
	return &pb.UpdateInstancePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service instance properties successfully."),
	}, nil
}

func (ds *DataSource) UnregisterInstance(ctx context.Context, request *pb.UnregisterInstanceRequest) (*pb.UnregisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	instanceID := request.InstanceId

	instanceFlag := util.StringJoin([]string{serviceID, instanceID}, "/")

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}):  serviceID,
		StringBuilder([]string{ColumnInstanceInfo, ColumnInstanceID}): instanceID}
	_, err := client.GetMongoClient().Delete(ctx, CollectionInstance, filter)
	if err != nil {
		log.Error(fmt.Sprintf("unregister instance failed, instance %s, operator %s revoke instance failed", instanceFlag, remoteIP), err)
		return &pb.UnregisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrInternal, "delete instance failed"),
		}, err
	}

	log.Infof("unregister instance[%s], operator %s", instanceFlag, remoteIP)
	return &pb.UnregisterInstanceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Unregister service instance successfully."),
	}, nil
}

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")
	err := KeepAliveLease(ctx, request)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance %s operator %s", instanceFlag, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess,
			"Update service instance heartbeat successfully."),
	}, nil
}

func (ds *DataSource) HeartbeatSet(ctx context.Context, request *pb.HeartbeatSetRequest) (*pb.HeartbeatSetResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	heartBeatCount := len(request.Instances)
	existFlag := make(map[string]bool, heartBeatCount)
	instancesHbRst := make(chan *pb.InstanceHbRst, heartBeatCount)
	noMultiCounter := 0

	for _, heartbeatElement := range request.Instances {
		if _, ok := existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId]; ok {
			log.Warnf("instance[%s/%s] is duplicate request heartbeat set",
				heartbeatElement.ServiceId, heartbeatElement.InstanceId)
			continue
		} else {
			existFlag[heartbeatElement.ServiceId+heartbeatElement.InstanceId] = true
			noMultiCounter++
		}
		gopool.Go(getHeartbeatFunc(ctx, domainProject, instancesHbRst, heartbeatElement))
	}

	count := 0
	successFlag := false
	failFlag := false
	instanceHbRstArr := make([]*pb.InstanceHbRst, 0, heartBeatCount)

	for hbRst := range instancesHbRst {
		count++
		if len(hbRst.ErrMessage) != 0 {
			failFlag = true
		} else {
			successFlag = true
		}
		instanceHbRstArr = append(instanceHbRstArr, hbRst)
		if count == noMultiCounter {
			close(instancesHbRst)
		}
	}

	if !failFlag && successFlag {
		log.Infof("batch update heartbeats[%d] successfully", count)
		return &pb.HeartbeatSetResponse{
			Response:  pb.CreateResponse(pb.ResponseSuccess, "Heartbeat set successfully."),
			Instances: instanceHbRstArr,
		}, nil
	}

	log.Info(fmt.Sprintf("batch update heartbeats failed %v", request.Instances))
	return &pb.HeartbeatSetResponse{
		Response:  pb.CreateResponse(pb.ErrInstanceNotExists, "Heartbeat set failed."),
		Instances: instanceHbRstArr,
	}, nil
}

func (ds *DataSource) BatchFind(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindInstancesResponse, error) {
	response := &pb.BatchFindInstancesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Batch query service instances successfully."),
	}

	var err error

	response.Services, err = ds.batchFindServices(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	response.Instances, err = ds.batchFindInstances(ctx, request)
	if err != nil {
		return &pb.BatchFindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return response, nil
}

func registryInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance
	instanceID := instance.InstanceId
	data := &Instance{
		Domain:       domain,
		Project:      project,
		RefreshTime:  time.Now(),
		InstanceInfo: instance,
	}

	instanceFlag := fmt.Sprintf("endpoints %v, host '%s', serviceID %s",
		instance.Endpoints, instance.HostName, instance.ServiceId)

	insertRes, err := client.GetMongoClient().Insert(ctx, CollectionInstance, data)
	if err != nil {
		log.Error(fmt.Sprintf("register instance failed %s instanceID %s operator %s", instanceFlag, instanceID, remoteIP), err)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}

	log.Infof("register instance %s, instanceID %s, operator %s",
		instanceFlag, insertRes.InsertedID, remoteIP)
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}

func (ds *DataSource) findInstance(ctx context.Context, request *pb.FindInstancesRequest, provider *pb.MicroServiceKey) (*pb.FindInstancesResponse, error) {
	var err error
	domainProject := util.ParseDomainProject(ctx)
	service := &Service{ServiceInfo: &pb.MicroService{Environment: request.Environment}}
	if len(request.ConsumerServiceId) > 0 {
		filter := GeneratorServiceFilter(ctx, request.ConsumerServiceId)
		service, err = GetService(ctx, filter)
		if err != nil {
			log.Error(fmt.Sprintf("get consumer failed, consumer %s find provider %s/%s/%s/%s",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		if service == nil {
			log.Error(fmt.Sprintf("consumer does not exist, consumer %s find provider %s/%s/%s/%s",
				request.ConsumerServiceId, request.Environment, request.AppId, request.ServiceName, request.VersionRule), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists,
					fmt.Sprintf("Consumer[%s] does not exist.", request.ConsumerServiceId)),
			}, nil
		}
		provider.Environment = service.ServiceInfo.Environment
	}

	// provider is not a shared micro-service,
	// only allow shared micro-service instances found request different domains.
	ctx = util.SetTargetDomainProject(ctx, util.ParseDomain(ctx), util.ParseProject(ctx))
	provider.Tenant = util.ParseTargetDomainProject(ctx)

	findFlag := fmt.Sprintf("Consumer[%s][%s/%s/%s/%s] find provider[%s/%s/%s/%s]",
		request.ConsumerServiceId, service.ServiceInfo.Environment, service.ServiceInfo.AppId, service.ServiceInfo.ServiceName, service.ServiceInfo.Version,
		provider.Environment, provider.AppId, provider.ServiceName, provider.Version)

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	resp, err := client.GetMongoClient().Find(ctx, CollectionInstance, bson.M{ColumnDomain: domain, ColumnProject: project})
	if err != nil {
		log.Error(fmt.Sprintf("FindInstancesCache.Get failed %s failed", findFlag), err)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}
	if resp == nil {
		mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
		log.Error("FindInstancesCache.Get failed", mes)
		return &pb.FindInstancesResponse{
			Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
		}, nil
	}

	var instances []*pb.MicroServiceInstance
	for resp.Next(ctx) {
		var instance Instance
		err := resp.Decode(&instance)
		if err != nil {
			log.Error(fmt.Sprintf("FindInstances.GetWithProviderID failed %s failed", findFlag), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
		instances = append(instances, instance.InstanceInfo)
	}

	// add dependency queue
	if len(request.ConsumerServiceId) > 0 &&
		len(instances) > 0 {
		provider, err = ds.reshapeProviderKey(ctx, provider, instances[0].ServiceId)
		if err != nil {
			return nil, err
		}
		if provider != nil {
			err = AddServiceVersionRule(ctx, domainProject, service.ServiceInfo, provider)
		} else {
			mes := fmt.Errorf("%s failed, provider does not exist", findFlag)
			log.Error("AddServiceVersionRule failed", mes)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, mes.Error()),
			}, nil
		}
		if err != nil {
			log.Error(fmt.Sprintf("AddServiceVersionRule failed %s failed", findFlag), err)
			return &pb.FindInstancesResponse{
				Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
			}, err
		}
	}

	return &pb.FindInstancesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Query service instances successfully."),
		Instances: instances,
	}, nil
}

func (ds *DataSource) reshapeProviderKey(ctx context.Context, provider *pb.MicroServiceKey, providerID string) (
	*pb.MicroServiceKey, error) {
	//维护version的规则,service name 可能是别名，所以重新获取
	filter := GeneratorServiceFilter(ctx, providerID)
	providerService, err := GetService(ctx, filter)
	if providerService == nil {
		return nil, err
	}

	versionRule := provider.Version
	provider = pb.MicroServiceToKey(provider.Tenant, providerService.ServiceInfo)
	provider.Version = versionRule
	return provider, nil
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	return nil
}

func GetInstance(ctx context.Context, serviceID string, instanceID string) (*Instance, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}):  serviceID,
		StringBuilder([]string{ColumnInstanceInfo, ColumnInstanceID}): instanceID}
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instance *Instance
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	err = findRes.Decode(&instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func UpdateInstanceS(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}):  instance.ServiceId,
		StringBuilder([]string{ColumnInstanceInfo, ColumnInstanceID}): instance.InstanceId}
	_, err := client.GetMongoClient().Update(ctx, CollectionInstance, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.status": instance.Status}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func UpdateInstanceP(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		ColumnDomain:  domain,
		ColumnProject: project,
		StringBuilder([]string{ColumnInstanceInfo, ColumnServiceID}):  instance.ServiceId,
		StringBuilder([]string{ColumnInstanceInfo, ColumnInstanceID}): instance.InstanceId}
	_, err := client.GetMongoClient().Update(ctx, CollectionInstance, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.properties": instance.Properties}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func KeepAliveLease(ctx context.Context, request *pb.HeartbeatRequest) *pb.Error {
	_, err := heartbeat.Instance().Heartbeat(ctx, request)
	if err != nil {
		return pb.NewError(pb.ErrInstanceNotExists, err.Error())
	}
	return nil
}

func getHeartbeatFunc(ctx context.Context, domainProject string, instancesHbRst chan<- *pb.InstanceHbRst, element *pb.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &pb.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}

		req := &pb.HeartbeatRequest{
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

func (ds *DataSource) batchFindServices(ctx context.Context, request *pb.BatchFindInstancesRequest) (
	*pb.BatchFindResult, error) {
	if len(request.Services) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)

	services := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Services {
		findCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.FindInstances(findCtx, &pb.FindInstancesRequest{
			ConsumerServiceId: request.ConsumerServiceId,
			AppId:             key.Service.AppId,
			ServiceName:       key.Service.ServiceName,
			VersionRule:       key.Service.Version,
			Environment:       key.Service.Environment,
		})
		if err != nil {
			return nil, err
		}
		failed, ok := failedResult[resp.Response.GetCode()]
		AppendFindResponse(findCtx, int64(index), resp.Response, resp.Instances,
			&services.Updated, &services.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[resp.Response.GetCode()] = failed
		}
	}
	for _, result := range failedResult {
		services.Failed = append(services.Failed, result)
	}
	return services, nil
}

func (ds *DataSource) batchFindInstances(ctx context.Context, request *pb.BatchFindInstancesRequest) (*pb.BatchFindResult, error) {
	if len(request.Instances) == 0 {
		return nil, nil
	}
	cloneCtx := util.CloneContext(ctx)
	// can not find the shared provider instances
	cloneCtx = util.SetTargetDomainProject(cloneCtx, util.ParseDomain(ctx), util.ParseProject(ctx))

	instances := &pb.BatchFindResult{}
	failedResult := make(map[int32]*pb.FindFailedResult)
	for index, key := range request.Instances {
		getCtx := util.SetContext(cloneCtx, util.CtxRequestRevision, key.Rev)
		resp, err := ds.GetInstance(getCtx, &pb.GetOneInstanceRequest{
			ConsumerServiceId:  request.ConsumerServiceId,
			ProviderServiceId:  key.Instance.ServiceId,
			ProviderInstanceId: key.Instance.InstanceId,
		})
		if err != nil {
			return nil, err
		}
		failed, ok := failedResult[resp.Response.GetCode()]
		AppendFindResponse(getCtx, int64(index), resp.Response, []*pb.MicroServiceInstance{resp.Instance},
			&instances.Updated, &instances.NotModified, &failed)
		if !ok && failed != nil {
			failedResult[resp.Response.GetCode()] = failed
		}
	}
	for _, result := range failedResult {
		instances.Failed = append(instances.Failed, result)
	}
	return instances, nil
}

func AppendFindResponse(ctx context.Context, index int64, resp *pb.Response, instances []*pb.MicroServiceInstance,
	updatedResult *[]*pb.FindResult, notModifiedResult *[]int64, failedResult **pb.FindFailedResult) {
	if code := resp.GetCode(); code != pb.ResponseSuccess {
		if *failedResult == nil {
			*failedResult = &pb.FindFailedResult{
				Error: pb.NewError(code, resp.GetMessage()),
			}
		}
		(*failedResult).Indexes = append((*failedResult).Indexes, index)
		return
	}
	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	if len(iv) > 0 && iv == ov {
		*notModifiedResult = append(*notModifiedResult, index)
		return
	}
	*updatedResult = append(*updatedResult, &pb.FindResult{
		Index:     index,
		Instances: instances,
		Rev:       ov,
	})
}

func preProcessRegisterInstance(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	if len(instance.Status) == 0 {
		instance.Status = pb.MSI_UP
	}

	if len(instance.InstanceId) == 0 {
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	instance.ModTimestamp = instance.Timestamp

	// 这里应该根据租约计时
	renewalInterval := apt.RegistryDefaultLeaseRenewalinterval
	retryTimes := apt.RegistryDefaultLeaseRetrytimes
	if instance.HealthCheck == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: renewalInterval,
			Times:    retryTimes,
		}
	} else {
		// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
		switch instance.HealthCheck.Mode {
		case pb.CHECK_BY_HEARTBEAT:
			d := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
			if d <= 0 {
				return pb.NewError(pb.ErrInvalidParams, "Invalid 'healthCheck' settings in request body.")
			}
		case pb.CHECK_BY_PLATFORM:
			// 默认120s
			instance.HealthCheck.Interval = renewalInterval
			instance.HealthCheck.Times = retryTimes
		}
	}

	filter := GeneratorServiceFilter(ctx, instance.ServiceId)
	microservice, err := GetService(ctx, filter)
	if microservice == nil || err != nil {
		return pb.NewError(pb.ErrServiceNotExists, "Invalid 'serviceID' in request body.")
	}
	instance.Version = microservice.ServiceInfo.Version
	return nil
}

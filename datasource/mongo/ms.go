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
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	exist, err := ServiceExist(ctx, service.ServiceId)
	if err != nil {
		log.Errorf(err, "check service %s exist failed", service.ServiceId)
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "check service exist failed"),
		}, nil
	}
	if exist {
		return &pb.CreateServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceAlreadyExists, "ServiceID conflict"),
		}, nil
	}
	insertRes, err := client.GetMongoClient().Insert(ctx, CollectionService, &MgService{Domain: domain, Project: project, Service: service})
	if err != nil {
		switch tt := err.(type) {
		case mongo.WriteException:
			if tt.WriteConcernError != nil {
				for _, writeError := range tt.WriteErrors {
					if writeError.Code == ErrorDuplicateKey {
						existSvc, errGet := GetService(ctx, service.ServiceId)
						if errGet != nil {
							return &pb.CreateServiceResponse{
								Response: pb.CreateResponse(scerr.ErrInternal, errGet.Error()),
							}, nil
						}
						if existSvc == nil {
							return &pb.CreateServiceResponse{
								Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
							}, nil
						}
						return &pb.CreateServiceResponse{
							Response:  pb.CreateResponse(pb.ResponseSuccess, "Register service successfully"),
							ServiceId: existSvc.Service.ServiceId,
						}, nil
					}
				}
			}
		default:
			return &pb.CreateServiceResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, nil
		}
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
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		log.Errorf(err, "get all service under domain [%s] project [%s] failed", domain, project)
		return &pb.GetServicesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get services data failed."),
		}, nil
	}
	var services []*pb.MicroService
	for res.Next(ctx) {
		var service MgService
		err := res.Decode(&service)
		if err != nil {
			log.Errorf(err, "failed to decode service data")
			return &pb.GetServicesResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "unmarshal service data failed."),
			}, nil
		}
		services = append(services, service.Service)
	}
	return &pb.GetServicesResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceResponse, error) {
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "failed to get single service [%s] from mongo", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "get service data from mongodb failed."),
		}, nil
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

	res, err := client.GetMongoClient().Count(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		log.Errorf(err, "failed to check if service exist.")
		return &pb.GetExistenceByIDResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "check service exist failed"),
		}, nil
	}
	return &pb.GetExistenceByIDResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Check ExistService successfully."),
		Exist:    res != 0,
	}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	return nil, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {

	service, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "delete service failed,faield to get service."),
		}, nil
	}
	if service == nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "delete service failed,service not exist."),
		}, nil
	}
	//todo delete instance,tags,schemas...
	_, err = client.GetMongoClient().Delete(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId))
	if err != nil {
		return &pb.DeleteServiceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "delete service failed"),
		}, nil
	}
	return &pb.DeleteServiceResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Unregister service successfully."),
	}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {

	service, err := GetService(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "update service [%s] properties failed,get service file failed", request.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, nil
	}
	if service == nil {
		log.Errorf(nil, "update service [%s] properties failed, service does not exist", request.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	_, err = client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceModTime: strconv.FormatInt(time.Now().Unix(), 10), ServiceProperty: request.Properties}})
	if err != nil {
		log.Errorf(err, "update service [%s] properties failed, update mongo failed", request.ServiceId)
		return &pb.UpdateServicePropsResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Update doc in mongo failed."),
		}, nil
	}
	return &pb.UpdateServicePropsResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "update service successfully."),
	}, nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
	serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {}
}

func (ds *DataSource) GetServiceDetail(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceDetailResponse, error) {
	return nil, nil
}

func (ds *DataSource) GetServicesInfo(ctx context.Context, request *pb.GetServicesInfoRequest) (
	*pb.GetServicesInfoResponse, error) {

	return nil, nil
}

func (ds *DataSource) AddTags(ctx context.Context, request *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	service, err := GetService(ctx, request.ServiceId)
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

	_, err = client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: tags}})
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
	svc, err := GetService(ctx, request.ServiceId)
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
	svc, err := GetService(ctx, request.ServiceId)
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

	_, err = client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: newTags}})
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
	svc, err := GetService(ctx, request.ServiceId)
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
	_, err = client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, request.ServiceId), bson.M{"$set": bson.M{ServiceTag: newTags}})
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
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetSchema failed to check service exist."),
		}, nil
	}
	if svc == nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetSchema service does not exist."),
		}, nil
	}
	schema, err := GetSchema(ctx, request.ServiceId, request.SchemaId)
	if err != nil {
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetSchema failed from mongodb."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response:      pb.CreateResponse(pb.ResponseSuccess, "Get schema info successfully."),
		Schema:        schema.Schema,
		SchemaSummary: schema.SchemaSummary,
	}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetAllSchemas failed for get service failed"),
		}, nil
	}
	if svc == nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetAllSchemas failed for service not exist"),
		}, nil
	}

	filter := bson.M{Domain: util.ParseDomain(ctx), Project: util.ParseProject(ctx), SchemaServiceID: request.ServiceId}

	resp, err := client.GetMongoClient().Find(ctx, CollectionSchema, filter)
	if err != nil {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "GetAllSchemas failed for get schemas failed"),
		}, nil
	}
	var schemas []*pb.Schema
	for resp.Next(ctx) {
		var tmpSchema *MgSchema
		err := resp.Decode(&tmpSchema)
		if err != nil {
			return &pb.GetAllSchemaResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "GetAllSchemas failed for decode schema data."),
			}, nil
		}
		schemas = append(schemas, &pb.Schema{
			SchemaId: tmpSchema.SchemaID,
			Summary:  tmpSchema.SchemaSummary,
			Schema:   tmpSchema.Schema,
		})
	}
	return &pb.GetAllSchemaResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get all schema info successfully."),
		Schemas:  schemas,
	}, nil
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "ExistSchema failed for get service failed"),
			//Schemas:  nil,
		}, nil
	}
	if svc == nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "ExistSchema failed for service not exist"),
		}, nil
	}
	schema, err := GetSchema(ctx, request.ServiceId, request.SchemaId)
	if err != nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "ExistSchema failed for get schema failed."),
		}, nil
	}
	if schema == nil {
		return &pb.GetExistenceResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "ExistSchema failed for schema not exist."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Schema exist."),
		Summary:   schema.SchemaSummary,
		SchemaId:  schema.SchemaID,
		ServiceId: schema.ServiceID,
	}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "DeleteSchema failed for get service failed."),
		}, nil
	}
	if svc == nil {
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
			Response: pb.CreateResponse(scerr.ErrInternal, "ModifySchema failed to start session."),
		}, err
	}
	if session != nil {
		defer session.EndSession(ctx)
	}
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
	svc, err := GetService(ctx, request.ServiceId)
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
	if session != nil {
		defer session.EndSession(ctx)
	}
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
	svc, err := GetService(ctx, serviceID)
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
		respSchema, err := GetSchema(ctx, serviceID, schema.SchemaId)
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
	//todo session affair
	if len(newSchemas) != len(microservice.Schemas) {
		_, err := client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: newSchemas}})
		if err != nil {
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}
	newSchema := bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}
	updateOptions := options.Update().SetUpsert(true)
	_, err = client.GetMongoClient().Update(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), newSchema, updateOptions)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return nil
}

func (ds *DataSource) modifySchemas(ctx context.Context, service *pb.MicroService, schemas []*pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := GetSchemas(ctx, serviceID)
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
			_, err := client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: nonExistSchemaIds}})
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
					_, err := client.GetMongoClient().Update(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, needUpdateSchema.SchemaId), bson.M{"$set": bson.M{Schema: needUpdateSchema.Schema, SchemaSummary: needUpdateSchema.Summary}}, options.Update().SetUpsert(true))
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
			_, err = client.GetMongoClient().Update(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.Update().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
		}
	} else {

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			_, err = client.GetMongoClient().Update(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.Update().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Infof("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			_, err = client.GetMongoClient().Update(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId), bson.M{"$set": bson.M{Schema: schema.Schema, SchemaSummary: schema.Summary}}, options.Update().SetUpsert(true))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Infof("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			_, err = client.GetMongoClient().Delete(ctx, CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schema.SchemaId))
			if err != nil {
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
		}

		_, err := client.GetMongoClient().Update(ctx, CollectionService, GeneratorServiceFilter(ctx, serviceID), bson.M{"$set": bson.M{ServiceSchemas: schemaIDs}})
		if err != nil {
			log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
				serviceID, remoteIP)
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}
	return nil
}

func (ds *DataSource) AddRule(ctx context.Context, request *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	service, err := GetService(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "failed to add rules for service [%s] for get service failed,", request.ServiceId)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if service == nil {
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
		exist, err := RuleExist(ctx, request.ServiceId, rule.Attribute, rule.Pattern)
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
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "GetRules failed for get service failed."),
		}, nil
	}
	if svc == nil {
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
	service, err := GetService(ctx, request.ServiceId)
	if err != nil {
		log.Errorf(err, "failed to add tags for service [%s] for get service failed,", request.ServiceId)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Failed to check service exist"),
		}, nil
	}
	if service == nil {
		return &pb.DeleteServiceRulesResponse{Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service not exist")}, nil
	}
	for _, ruleID := range request.RuleIds {
		exist, err := RuleExistByID(ctx, request.ServiceId, ruleID)
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
		_, err = client.GetMongoClient().Delete(ctx, CollectionRule, GeneratorRuleFilter(ctx, request.ServiceId, ruleID))
		if err != nil {
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
	svc, err := GetService(ctx, request.ServiceId)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "UpdateRule failed for get service failed."),
			//Schemas:  nil,
		}, nil
	}
	if svc == nil {
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
	exist, err := RuleExistByID(ctx, request.ServiceId, request.RuleId)
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

	_, err = client.GetMongoClient().Update(ctx, CollectionRule, GeneratorRuleFilter(ctx, request.ServiceId, request.RuleId), newRule)
	if err != nil {
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.UpdateServiceRuleResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Update service rules succesfully."),
	}, nil
}

func RuleExist(ctx context.Context, serviceID, attribute, pattern string) (bool, error) {
	filter := bson.M{RuleServiceID: serviceID, RuleAttribute: attribute, RulePattern: pattern}
	num, err := client.GetMongoClient().Count(ctx, CollectionRule, filter)
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

func (ds *DataSource) isSchemaEditable(service *pb.MicroService) bool {
	return (len(service.Environment) != 0 && service.Environment != pb.ENV_PROD) || ds.SchemaEditable
}

func GetSchemas(ctx context.Context, serviceID string) ([]*pb.Schema, error) {
	filter := GeneratorServiceFilter(ctx, serviceID)
	getRes, err := client.GetMongoClient().Find(ctx, CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schemas []*pb.Schema
	for getRes.Next(ctx) {
		var tmp MgSchema
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

func GetSchema(ctx context.Context, serviceID, schemaID string) (*MgSchema, error) {
	filter := GeneratorSchemaFilter(ctx, serviceID, schemaID)
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

func GetService(ctx context.Context, serviceID string) (*MgService, error) {
	filter := GeneratorServiceFilter(ctx, serviceID)

	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionService, filter)
	if err != nil {
		return nil, err
	}
	var svc *MgService
	err = findRes.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
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

func GetRule(ctx context.Context, serviceID, ruleID string) (*pb.ServiceRule, error) {
	filter := GeneratorRuleFilter(ctx, serviceID, ruleID)
	ruleRes, err := client.GetMongoClient().FindOne(ctx, CollectionRule, filter)
	if err != nil {
		return nil, err
	}
	var rule *pb.ServiceRule
	err = ruleRes.Decode(&rule)
	if err != nil {
		return rule, err
	}
	return nil, nil
}

func RuleExistByID(ctx context.Context, serviceID, ruleID string) (bool, error) {
	num, err := client.GetMongoClient().Count(ctx, CollectionRule, GeneratorRuleFilter(ctx, serviceID, ruleID))
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

func ServiceExist(ctx context.Context, serviceID string) (bool, error) {
	num, err := client.GetMongoClient().Count(ctx, CollectionService, GeneratorServiceFilter(ctx, serviceID))
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

func GeneratorServiceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{Domain: domain, Project: project, ServiceServiceID: serviceID}
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

func (ds *DataSource) GetApplications(ctx context.Context, request *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	return &pb.GetAppsResponse{}, nil
}

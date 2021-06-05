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

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

func (ds *DataSource) GetSchema(ctx context.Context, request *discovery.GetSchemaRequest) (*discovery.GetSchemaResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.GetSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "getSchema failed to check service exist."),
		}, nil
	}
	if !exist {
		return &discovery.GetSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "getSchema service does not exist."),
		}, nil
	}
	filter = mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	schema, err := findSchema(ctx, filter)
	if err != nil && err != datasource.ErrNoData {
		return &discovery.GetSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "getSchema failed from mongodb."),
		}, nil
	}
	if schema == nil {
		return &discovery.GetSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrSchemaNotExists, "do not have this schema info."),
		}, nil
	}
	return &discovery.GetSchemaResponse{
		Response:      discovery.CreateResponse(discovery.ResponseSuccess, "get schema info successfully."),
		Schema:        schema.Schema,
		SchemaSummary: schema.SchemaSummary,
	}, nil
}

func (ds *DataSource) GetAllSchemas(ctx context.Context, request *discovery.GetAllSchemaRequest) (*discovery.GetAllSchemaResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(request.ServiceId))
	svc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return &discovery.GetAllSchemaResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "getAllSchemas failed for service not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("get service[%s] all schemas failed, get service failed", request.ServiceId), err)
		return &discovery.GetAllSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, err
	}
	schemasList := svc.Service.Schemas
	if len(schemasList) == 0 {
		return &discovery.GetAllSchemaResponse{
			Response: discovery.CreateResponse(discovery.ResponseSuccess, "do not have this schema info."),
			Schemas:  []*discovery.Schema{},
		}, nil
	}
	schemas := make([]*discovery.Schema, 0, len(schemasList))
	for _, schemaID := range schemasList {
		tempSchema := &discovery.Schema{}
		tempSchema.SchemaId = schemaID
		filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(request.ServiceId), mutil.SchemaID(schemaID))
		schema, err := findSchema(ctx, filter)
		if err != nil && err != datasource.ErrNoData {
			return &discovery.GetAllSchemaResponse{
				Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
			}, err
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
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get all schema info successfully."),
		Schemas:  schemas,
	}, nil
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *discovery.GetExistenceRequest) (*discovery.GetExistenceResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "existSchema failed for get service failed"),
		}, nil
	}
	if !exist {
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "existSchema failed for service not exist"),
		}, nil
	}
	filter = mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	Schema, err := findSchema(ctx, filter)
	if err != nil && err != datasource.ErrNoData {
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, "existSchema failed for get schema failed."),
		}, nil
	}
	if Schema == nil {
		return &discovery.GetExistenceResponse{
			Response: discovery.CreateResponse(discovery.ErrSchemaNotExists, "existSchema failed for schema not exist."),
		}, nil
	}
	return &discovery.GetExistenceResponse{
		Response:  discovery.CreateResponse(discovery.ResponseSuccess, "Schema exist."),
		Summary:   Schema.SchemaSummary,
		SchemaId:  Schema.SchemaID,
		ServiceId: Schema.ServiceID,
	}, nil
}

func (ds *DataSource) DeleteSchema(ctx context.Context, request *discovery.DeleteSchemaRequest) (*discovery.DeleteSchemaResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	exist, err := exitService(ctx, filter)
	if err != nil {
		return &discovery.DeleteSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "deleteSchema failed for get service failed."),
		}, nil
	}
	if !exist {
		return &discovery.DeleteSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "deleteSchema failed for service not exist."),
		}, nil
	}
	filter = mutil.NewBasicFilter(ctx, mutil.ServiceID(request.ServiceId), mutil.SchemaID(request.SchemaId))
	isDeleted, err := deleteSchema(ctx, filter)
	if err != nil {
		return &discovery.DeleteSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrUnavailableBackend, "deleteSchema failed for delete schema failed."),
		}, err
	}
	if !isDeleted {
		return &discovery.DeleteSchemaResponse{
			Response: discovery.CreateResponse(discovery.ErrSchemaNotExists, "deleteSchema failed for schema not exist."),
		}, nil
	}
	return &discovery.DeleteSchemaResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "delete schema info successfully."),
	}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *discovery.ModifySchemaRequest) (*discovery.ModifySchemaResponse, error) {
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
		resp := &discovery.ModifySchemaResponse{
			Response: discovery.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}
	log.Info(fmt.Sprintf("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP))
	return &discovery.ModifySchemaResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "modify schema info success."),
	}, nil
}

func (ds *DataSource) ModifySchemas(ctx context.Context, request *discovery.ModifySchemasRequest) (*discovery.ModifySchemasResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	svc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return &discovery.ModifySchemasResponse{Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service not exist")}, nil
		}
		return &discovery.ModifySchemasResponse{Response: discovery.CreateResponse(discovery.ErrInternal, err.Error())}, err
	}
	respErr := ds.modifySchemas(ctx, svc.Service, request.Schemas)
	if respErr != nil {
		resp := &discovery.ModifySchemasResponse{
			Response: discovery.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, err
		}
		return resp, nil
	}
	return &discovery.ModifySchemasResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "modify schemas info success"),
	}, nil

}

func (ds *DataSource) modifySchemas(ctx context.Context, service *discovery.MicroService, schemas []*discovery.Schema) *errsvc.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	filter := mutil.NewFilter(mutil.ServiceID(serviceID))
	schemasFromDatabase, err := findSchemas(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("modify service %s schemas failed, get schemas failed, operator: %s", serviceID, remoteIP), err)
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		datasource.SchemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	var schemasOps []mongo.WriteModel
	var serviceOps []mongo.WriteModel
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.TypeSchema, util.ParseDomainProject(ctx), serviceID, int64(len(nonExistSchemaIds)))
			errQuota := quota.Apply(ctx, res)
			if errQuota != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), errQuota)
				return errQuota
			}
			filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
			setValue := mutil.NewFilter(mutil.ServiceSchemas(nonExistSchemaIds))
			updateFilter := mutil.NewFilter(mutil.Set(setValue))
			serviceOps = append(serviceOps, mongo.NewUpdateOneModel().SetUpdate(updateFilter).SetFilter(filter))
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Error(fmt.Sprintf("modify service %s schemas failed, operator: %s", serviceID, remoteIP), err)
				return discovery.NewError(discovery.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := schemaSummaryExist(ctx, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return discovery.NewError(discovery.ErrInternal, err.Error())
				}
				if !exist {
					filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(needUpdateSchema.SchemaId))
					setValue := mutil.NewFilter(
						mutil.ColSchema(needUpdateSchema.Schema),
						mutil.SchemaSummary(needUpdateSchema.Summary),
					)
					updateFilter := mutil.NewFilter(
						mutil.Set(setValue),
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
			res := quota.NewApplyQuotaResource(quota.TypeSchema, util.ParseDomainProject(ctx), serviceID, int64(quotaSize))
			err := quota.Apply(ctx, res)
			if err != nil {
				log.Error(fmt.Sprintf("modify service[%s] schemas failed, operator: %s", serviceID, remoteIP), err)
				return err
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
			setValue := mutil.NewFilter(
				mutil.ColSchema(schema.Schema),
				mutil.SchemaSummary(schema.Summary),
			)
			updateFilter := mutil.NewFilter(
				mutil.Set(setValue),
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
		updateValue := mutil.NewFilter(mutil.Set(setFilter))
		serviceOps = append(serviceOps, mongo.NewUpdateOneModel().SetUpdate(updateValue).SetFilter(filter))
	}
	if len(schemasOps) > 0 {
		err = batchUpdateSchema(ctx, schemasOps)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	if len(serviceOps) > 0 {
		err = batchUpdateServices(ctx, serviceOps)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	return nil
}

// modifySchema will be modified in the following cases
// 1.service have no relation --> update the schema && update the service
// 2.service is editable && service have relation with the schema --> update the shema
// 3.service is editable && service have no relation with the schema --> update the schema && update the service
// 4.service can't edit && service have relation with the schema && schema summary not exist --> update the schema
func (ds *DataSource) modifySchema(ctx context.Context, serviceID string, schema *discovery.Schema) *errsvc.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
	remoteIP := util.GetIPFromContext(ctx)
	svc, err := findService(ctx, filter)
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
	if !ds.isSchemaEditable(microservice) {
		if len(microservice.Schemas) != 0 && !isExist {
			return discovery.NewError(discovery.ErrUndefinedSchemaID, "non-existent schemaID can't be added request "+discovery.ENV_PROD)
		}
		filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
		respSchema, err := findSchema(ctx, filter)
		if err != nil && err != datasource.ErrNoData {
			return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
		}
		if respSchema != nil {
			if len(schema.Summary) == 0 {
				log.Error(fmt.Sprintf("modify schema %s %s failed, get schema summary failed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return discovery.NewError(discovery.ErrModifySchemaNotAllow,
					"schema already exist, can not be changed request "+discovery.ENV_PROD)
			}
			if len(respSchema.SchemaSummary) != 0 {
				log.Error(fmt.Sprintf("mode, schema %s %s already exist, can not be changed, operator: %s",
					serviceID, schema.SchemaId, remoteIP), err)
				return discovery.NewError(discovery.ErrModifySchemaNotAllow, "schema already exist, can not be changed request "+discovery.ENV_PROD)
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
	if len(newSchemas) != 0 {
		filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(serviceID))
		setValue := mutil.NewFilter(
			mutil.ServiceSchemas(newSchemas),
		)
		updateFilter := mutil.NewFilter(
			mutil.Set(setValue),
		)
		err = updateService(ctx, filter, updateFilter)
		if err != nil {
			return discovery.NewError(discovery.ErrInternal, err.Error())
		}
	}
	filter = mutil.NewDomainProjectFilter(domain, project, mutil.ServiceID(serviceID), mutil.SchemaID(schema.SchemaId))
	setValue := mutil.NewFilter(
		mutil.ColSchema(schema.Schema),
		mutil.SchemaSummary(schema.Summary),
	)
	updateFilter := mutil.NewFilter(mutil.Set(setValue))
	err = updateSchema(ctx, filter, updateFilter, options.FindOneAndUpdate().SetUpsert(true))
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return nil
}

func (ds *DataSource) isSchemaEditable(service *discovery.MicroService) bool {
	return (len(service.Environment) != 0 && service.Environment != discovery.ENV_PROD) || ds.SchemaEditable
}

func schemaSummaryExist(ctx context.Context, serviceID, schemaID string) (bool, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(serviceID), mutil.SchemaID(schemaID))
	schema, err := findSchema(ctx, filter)
	if err != nil {
		return false, err
	}
	return len(schema.SchemaSummary) != 0, nil
}

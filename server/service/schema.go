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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

func (s *MicroServiceService) GetSchemaInfo(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.SchemaId) == 0 {
		util.Logger().Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}

	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, in.SchemaId)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key))
	resp, errDo := backend.Store().Schema().Search(ctx, opts...)
	if errDo != nil {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: get schema info failed.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get schema info failed."),
		}, errDo
	}
	if resp.Count == 0 {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: schema not exists.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "Do not have this schema info."),
		}, nil
	}

	schemaSummary, err := getSchemaSummary(ctx, domainProject, in.ServiceId, in.SchemaId)
	if err != nil {
		util.Logger().Errorf(err, "get schema failed, serviceId %s, schemaId %s: get schema summary failed.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetSchemaResponse{
		Response:      pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully."),
		Schema:        util.BytesToStringWithNoCopy(resp.Kvs[0].Value),
		SchemaSummary: schemaSummary,
	}, nil
}

func (s *MicroServiceService) GetAllSchemaInfo(ctx context.Context, in *pb.GetAllSchemaRequest) (*pb.GetAllSchemaResponse, error) {
	if in == nil || len(in.ServiceId) == 0 {
		util.Logger().Errorf(nil, "get all schema failed: invalid params.")
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}

	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s: invalid params.", in.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "get all schemas failed: get service failed. %s", in.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Invalid request."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "get all schemas failed: service does not exist. %s", in.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	schemasList := service.Schemas
	if schemasList == nil || len(schemasList) == 0 {
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Do not have this schema info."),
			Schemas:  []*pb.Schema{},
		}, nil
	}

	key := apt.GenerateServiceSchemaSummaryKey(domainProject, in.ServiceId, "")
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
	resp, errDo := backend.Store().SchemaSummary().Search(ctx, opts...)
	if errDo != nil {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s: get schema info failed.", in.ServiceId)
		return &pb.GetAllSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get schema info failed."),
		}, errDo
	}

	respWithSchema := &registry.PluginResponse{}
	if in.WithSchema {
		key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, "")
		opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
		respWithSchema, errDo = backend.Store().Schema().Search(ctx, opts...)
		if errDo != nil {
			util.Logger().Errorf(errDo, "get schema failed, serviceId %s: get schema info failed.", in.ServiceId)
			return &pb.GetAllSchemaResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Get schema info failed."),
			}, errDo
		}
	}

	schemas := make([]*pb.Schema, 0, len(schemasList))
	for _, schemaId := range schemasList {
		tempSchema := &pb.Schema{}
		tempSchema.SchemaId = schemaId
		for _, summarySchema := range resp.Kvs {
			schemaIdOfSummary, summaryData := pb.GetInfoFromSchemaSummaryKV(summarySchema)
			if schemaId == schemaIdOfSummary {
				tempSchema.Summary = util.BytesToStringWithNoCopy(summaryData)
			}
		}

		for _, contentSchema := range respWithSchema.Kvs {
			schemaIdOfSchema, schemaData := pb.GetInfoFromSchemaKV(contentSchema)
			if schemaId == schemaIdOfSchema {
				tempSchema.Schema = util.BytesToStringWithNoCopy(schemaData)
			}
		}
		schemas = append(schemas, tempSchema)
	}

	return &pb.GetAllSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get all schema info successfully."),
		Schemas:  schemas,
	}, nil

}

func (s *MicroServiceService) DeleteSchema(ctx context.Context, in *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.SchemaId) == 0 {
		util.Logger().Errorf(nil, "delete schema failed: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		util.Logger().Errorf(nil, "delete schema failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, in.SchemaId)
	exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failed, serviceId %s, schemaId %s: get schema failed.", in.ServiceId, in.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.Logger().Errorf(nil, "delete schema failed, serviceId %s, schemaId %s: schema not exist.", in.ServiceId, in.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "Schema info does not exist."),
		}, nil
	}
	epSummaryKey := apt.GenerateServiceSchemaSummaryKey(domainProject, in.ServiceId, in.SchemaId)
	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(epSummaryKey)),
		registry.OpDel(registry.WithStrKey(key)),
	}
	_, errDo := backend.Registry().Txn(ctx, opts)
	if errDo != nil {
		util.Logger().Errorf(errDo, "delete schema failed, serviceId %s, schemaId %s: delete schema from etcd failed.", in.ServiceId, in.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Delete schema info failed."),
		}, errDo
	}
	util.Logger().Infof("delete schema info successfully.%s", in.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *MicroServiceService) ModifySchemas(ctx context.Context, in *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas failed: invalid params.")
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request."),
		}, nil
	}
	serviceId := in.ServiceId

	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas failed: get service failed. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Invalid request."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "modify schemas failed: service does not exist. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	respErr := modifySchemas(ctx, domainProject, service, in.Schemas)
	if respErr != nil {
		util.Logger().Errorf(nil, "modify schemas failed: commit data failed, serviceId %s", serviceId)
		resp := &pb.ModifySchemasResponse{
			Response: pb.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, respErr
		}
		return resp, nil
	}

	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "modify schemas info successfully."),
	}, nil
}

func schemasAnalysis(schemas []*pb.Schema, schemasFromDb []*pb.Schema, schemaIdsInService []string) (
	[]*pb.Schema, []*pb.Schema, []*pb.Schema, []string) {
	needUpdateSchemas := make([]*pb.Schema, 0, len(schemas))
	needAddSchemas := make([]*pb.Schema, 0, len(schemas))
	needDeleteSchemas := make([]*pb.Schema, 0, len(schemasFromDb))
	nonExistSchemaIds := make([]string, 0, len(schemas))

	duplicate := make(map[string]struct{})
	for _, schema := range schemas {
		if _, ok := duplicate[schema.SchemaId]; ok {
			continue
		}
		duplicate[schema.SchemaId] = struct{}{}

		exist := false
		for _, schemaFromDb := range schemasFromDb {
			if schema.SchemaId == schemaFromDb.SchemaId {
				needUpdateSchemas = append(needUpdateSchemas, schema)
				exist = true
				break
			}
		}
		if !exist {
			needAddSchemas = append(needAddSchemas, schema)
		}

		exist = false
		for _, schemaId := range schemaIdsInService {
			if schema.SchemaId == schemaId {
				exist = true
			}
		}
		if !exist {
			nonExistSchemaIds = append(nonExistSchemaIds, schema.SchemaId)
		}
	}

	for _, schemaFromDb := range schemasFromDb {
		exist := false
		for _, schema := range schemas {
			if schema.SchemaId == schemaFromDb.SchemaId {
				exist = true
				break
			}
		}
		if !exist {
			needDeleteSchemas = append(needDeleteSchemas, schemaFromDb)
		}
	}

	return needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds
}

func modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService, schemas []*pb.Schema) *scerr.Error {
	serviceId := service.ServiceId
	schemasFromDatabase, err := GetSchemasFromDatabase(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(nil, "modify schema failed: get schema from database failed, %s", serviceId)
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds := schemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]registry.PluginOp, 0)
	if len(service.Environment) == 0 || service.Environment == pb.ENV_PROD {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceId, int64(len(nonExistSchemaIds)))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			errQuota := rst.Err
			if errQuota != nil {
				util.Logger().Errorf(errQuota, "modify schemas info failed, serviceId is %s", serviceId)
				return errQuota
			}

			service.Schemas = nonExistSchemaIds
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "modify schemas info failed, update service failed , %s", serviceId)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Sprintf("non-exist schemaId %s", util.StringJoin(nonExistSchemaIds, " ,"))
				util.Logger().Errorf(nil, "modify schemas failed, serviceId %s, %s", serviceId, errInfo)
				return scerr.NewError(scerr.ErrUndefinedSchemaId, errInfo)
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := isExistSchemaSummary(ctx, domainProject, serviceId, needUpdateSchema.SchemaId)
				if err != nil {
					return scerr.NewError(scerr.ErrInternal, err.Error())
				}
				if !exist {
					opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, needUpdateSchema)
					pluginOps = append(pluginOps, opts...)
				} else {
					util.Logger().Warnf(nil, "schema and summary already exist, skip to update, serviceId %s, schemaId %s", serviceId, needUpdateSchema.SchemaId)
				}
			}
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceId, int64(quotaSize))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			err := rst.Err
			if err != nil {
				util.Logger().Errorf(err, "modify schemas info failed, check resource num failed, %s", serviceId)
				return err
			}
		}

		for _, schema := range needUpdateSchemas {
			util.Logger().Infof("update schema: serviceId %s, schemaId %s", serviceId, schema.SchemaId)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
		for _, schema := range needDeleteSchemas {
			util.Logger().Infof("delete non-exist schema: serviceId %s, schemaId %s", serviceId, schema.SchemaId)
			opts := schemaWithDatabaseOpera(registry.OpDel, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}

		if len(nonExistSchemaIds) != 0 {
			service.Schemas = append(service.Schemas, nonExistSchemaIds...)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "modify schema info failed, update service failed , %s", serviceId)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	for _, schema := range needAddSchemas {
		util.Logger().Infof("add new schema: serviceId %s, schemaId %s", serviceId, schema.SchemaId)
		opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
		pluginOps = append(pluginOps, opts...)
	}

	if len(pluginOps) != 0 {
		err = backend.BatchCommit(ctx, pluginOps)
		if err != nil {
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}
	util.Logger().Infof("modify schemas info successfully, serviceId %s, schemaIds %s", serviceId, parseSchemaIds(schemas))

	return nil
}

func parseSchemaIds(schemas []*pb.Schema) string {
	schemaIdsArr := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		schemaIdsArr = append(schemaIdsArr, schema.SchemaId)
	}
	return util.StringJoin(schemaIdsArr, " , ")
}

func isExistSchemaId(service *pb.MicroService, schemas []*pb.Schema) bool {
	serviceSchemaIds := service.Schemas
	for _, schema := range schemas {
		if !containsValueInSlice(serviceSchemaIds, schema.SchemaId) {
			util.Logger().Errorf(nil, "schema not exist schemaId: %s, serviceId %s", schema.SchemaId, service.ServiceId)
			return false
		}
	}
	return true
}

func schemaWithDatabaseOpera(invoke registry.Operation, domainProject string, serviceId string, schema *pb.Schema) []registry.PluginOp {
	pluginOps := make([]registry.PluginOp, 0)
	key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schema.SchemaId)
	opt := invoke(registry.WithStrKey(key), registry.WithStrValue(schema.Schema))
	pluginOps = append(pluginOps, opt)
	keySummary := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schema.SchemaId)
	opt = invoke(registry.WithStrKey(keySummary), registry.WithStrValue(schema.Summary))
	pluginOps = append(pluginOps, opt)
	return pluginOps
}

func GetSchemasFromDatabase(ctx context.Context, domainProject string, serviceId string) ([]*pb.Schema, error) {
	key := apt.GenerateServiceSchemaKey(domainProject, serviceId, "")
	util.Logger().Debugf("key is %s", key)
	resp, err := backend.Registry().Do(ctx,
		registry.GET,
		registry.WithPrefix(),
		registry.WithStrKey(key))
	if err != nil {
		util.Logger().Errorf(err, "Get schema of one service failed. %s", serviceId)
		return nil, err
	}
	schemas := make([]*pb.Schema, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		tmp := strings.Split(key, "/")
		schemaId := tmp[len(tmp)-1]
		schema := util.BytesToStringWithNoCopy(kv.Value)
		schemaStruct := &pb.Schema{
			SchemaId: schemaId,
			Schema:   schema,
		}
		schemas = append(schemas, schemaStruct)
	}
	return schemas, nil
}

func (s *MicroServiceService) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	respErr := s.canModifySchema(ctx, domainProject, request)
	if respErr != nil {
		resp := &pb.ModifySchemaResponse{
			Response: pb.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, respErr
		}
		return resp, nil
	}

	serviceId := request.ServiceId
	schemaId := request.SchemaId

	schema := pb.Schema{
		SchemaId: schemaId,
		Summary:  request.Summary,
		Schema:   request.Schema,
	}
	err := s.modifySchema(ctx, serviceId, &schema)
	if err != nil {
		util.Logger().Errorf(err, "modify schema failed, serviceId %s, schemaId %s", serviceId, schemaId)
		resp := &pb.ModifySchemaResponse{
			Response: pb.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	util.Logger().Infof("modify schema successfully: serviceId %s, schemaId %s.", serviceId, schemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "modify schema info success"),
	}, nil
}

func (s *MicroServiceService) canModifySchema(ctx context.Context, domainProject string, in *pb.ModifySchemaRequest) *scerr.Error {
	serviceId := in.ServiceId
	schemaId := in.SchemaId
	if len(schemaId) == 0 || len(serviceId) == 0 {
		util.Logger().Errorf(nil, "update schema failed: invalid params.")
		return scerr.NewError(scerr.ErrInvalidParams, "serviceId or schemaId is nil")
	}
	err := Validate(in)
	if err != nil {
		util.Logger().Errorf(err, "update schema failed, serviceId %s, schemaId %s: invalid params.", serviceId, schemaId)
		return scerr.NewError(scerr.ErrInvalidParams, err.Error())
	}

	res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceId, 1)
	rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		util.Logger().Errorf(errQuota, "modify schema info failed, check resource num failed, %s, %s", serviceId, schemaId)
		return errQuota
	}
	if len(in.Summary) == 0 {
		util.Logger().Warnf(nil, "service %s schema %s summary is empty.", in.ServiceId, schemaId)
	}
	return nil
}

func (s *MicroServiceService) modifySchema(ctx context.Context, serviceId string, schema *pb.Schema) *scerr.Error {
	domainProject := util.ParseDomainProject(ctx)
	schemaId := schema.SchemaId

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "modify schema failed, serviceId %s, schemaId %s: get service failed.", serviceId, schemaId)
		return scerr.NewError(scerr.ErrInternal, "get service failed")
	}
	if service == nil {
		util.Logger().Errorf(nil, "modify schema failed, serviceId %s, schemaId %s: service not exist", serviceId, schemaId)
		return scerr.NewError(scerr.ErrServiceNotExists, "service non-exist")
	}

	util.Logger().Infof("start to modify schema, serviceId  %s, schemaId %s", service.ServiceId, schemaId)
	pluginOps := make([]registry.PluginOp, 0, 10)
	isExist := isExistSchemaId(service, []*pb.Schema{schema})

	if len(service.Environment) == 0 || service.Environment == pb.ENV_PROD {
		if len(service.Schemas) != 0 && !isExist {
			return scerr.NewError(scerr.ErrUndefinedSchemaId, "schemaId non-exist， can't be added, environment is production")
		}

		key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schemaId)
		respSchema, err := backend.Store().Schema().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
		if err != nil {
			util.Logger().Errorf(err, "modify schema failed, get schema summary failed, %s %s", serviceId, schemaId)
			return scerr.NewError(scerr.ErrInternal, "get schema summary failed")
		}

		if respSchema.Count != 0 {
			if len(schema.Summary) == 0 {
				util.Logger().Errorf(err, "prod mode, schema already exist, can not change, %s %s", serviceId, schemaId)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not change, environment is production")
			}

			exist, err := isExistSchemaSummary(ctx, domainProject, serviceId, schemaId)
			if err != nil {
				util.Logger().Errorf(err, "check schema summary is exist failed, serviceId %s, schemaId %s", serviceId, schemaId)
				return scerr.NewError(scerr.ErrInternal, "check schema summary existence failed")
			}
			if exist {
				util.Logger().Errorf(err, "prod mode, schema already exist, can not change, %s %s", serviceId, schemaId)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not change, environment is production")
			}
		}

		if len(service.Schemas) == 0 {
			service.Schemas = append(service.Schemas, schemaId)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "modify schema failed, update service failed , serviceId %s, schemaId %s", serviceId, schemaId)
				return scerr.NewError(scerr.ErrInternal, "update service failed")
			}
			pluginOps = append(pluginOps, opt)
		}
	} else {
		if !isExist {
			service.Schemas = append(service.Schemas, schemaId)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "modify schema failed, update service failed , serviceId %s, schemaId %s", serviceId, schemaId)
				return scerr.NewError(scerr.ErrInternal, "update service failed")
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	opts := CommitSchemaInfo(domainProject, serviceId, schema)
	pluginOps = append(pluginOps, opts...)

	err = backend.BatchCommit(ctx, pluginOps)
	if err != nil {
		util.Logger().Errorf(err, "commit update schema failed, serviceId %s, schemaId %s", serviceId, schemaId)
		return scerr.NewError(scerr.ErrInternal, "commit update schema failed")
	}
	return nil
}

func isExistSchemaSummary(ctx context.Context, domainProject, serviceId, schemaId string) (bool, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schemaId)
	resp, err := backend.Store().SchemaSummary().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
	if err != nil {
		return true, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func CommitSchemaInfo(domainProject string, serviceId string, schema *pb.Schema) []registry.PluginOp {
	if len(schema.Summary) != 0 {
		return schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
	} else {
		key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schema.SchemaId)
		opt := registry.OpPut(registry.WithStrKey(key), registry.WithStrValue(schema.Schema))
		return []registry.PluginOp{opt}
	}
}

func containsValueInSlice(in []string, value string) bool {
	if in == nil || len(value) == 0 {
		return false
	}
	for _, i := range in {
		if i == value {
			return true
		}
	}
	return false
}

func getSchemaSummary(ctx context.Context, domainProject string, serviceId string, schemaId string) (string, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schemaId)
	resp, err := backend.Store().SchemaSummary().Search(ctx,
		registry.WithStrKey(key),
	)
	if err != nil {
		util.Logger().Errorf(err, "get %s schema %s summary failed", serviceId, schemaId)
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return util.BytesToStringWithNoCopy(resp.Kvs[0].Value), nil
}

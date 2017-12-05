//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package service

import (
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/infra/quota"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/ServiceComb/service-center/server/plugin"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

func (s *MicroServiceService) GetSchemaInfo(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.SchemaId) == 0 {
		util.Logger().Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}

	err := apt.Validate(in)
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
	resp, errDo := store.Store().Schema().Search(ctx, opts...)
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
	return &pb.GetSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully."),
		Schema:   util.BytesToStringWithNoCopy(resp.Kvs[0].Value),
	}, nil
}

func (s *MicroServiceService) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.Logger().Errorf(nil, "delete schema faild: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "delete schema faild, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		util.Logger().Errorf(nil, "delete schema faild, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "delete schema faild, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.Logger().Errorf(nil, "delete schema failed, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "Schema info does not exist."),
		}, nil
	}
	epSummaryKey := apt.GenerateServiceSchemaSummaryKey(domainProject, request.ServiceId, request.SchemaId)
	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(epSummaryKey)),
		registry.OpDel(registry.WithStrKey(key)),
	}
	_, errDo := backend.Registry().Txn(ctx, opts)
	if errDo != nil {
		util.Logger().Errorf(errDo, "delete schema failed, serviceId %s, schemaId %s: delete schema from etcd failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Delete schema info failed."),
		}, errDo
	}
	util.Logger().Infof("delete schema info successfully.%s", request.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *MicroServiceService) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas faild: invalid params.")
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request."),
		}, nil
	}
	serviceId := request.ServiceId

	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas faild: get service failed. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Invalid request."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "modify schemas faild: service does not exist. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	respErr := modifySchemas(ctx, domainProject, service, request.Schemas)
	if respErr != nil {
		resp := &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(respErr.Code, respErr.Detail),
		}
		if respErr.StatusCode() == http.StatusInternalServerError {
			return resp, respErr
		}
		return resp, nil
	}

	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "modify schemas info successfully."),
	}, nil
}

func schemasAnalysis(schemas []*pb.Schema, schemasFromDb []*pb.Schema, schemaIdsInService []string) ([]*pb.Schema, []*pb.Schema, []string) {
	needUpdateSchemas := make([]*pb.Schema, 0, len(schemas))
	needAddSchemas := make([]*pb.Schema, 0, len(schemas))

	for _, schema := range schemas {
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
	}

	nonExistSchemaIds := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		exist := false
		for _, schemaId := range schemaIdsInService {
			if schema.SchemaId == schemaId {
				exist = true
			}
		}
		if !exist {
			nonExistSchemaIds = append(nonExistSchemaIds, schema.SchemaId)
		}
	}
	return needUpdateSchemas, needAddSchemas, nonExistSchemaIds
}

func modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService, schemas []*pb.Schema) *scerr.Error {
	serviceId := service.ServiceId
	schemasFromDatabase, err := GetSchemasFromDatabase(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(nil, "modify schema failed: get schema from database failed, %s", serviceId)
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	needUpdateSchemas, needAddSchemas, nonExistSchemaIds := schemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]registry.PluginOp, 0)
	if service.Environment == pb.ENV_PROD {
		if len(service.Schemas) == 0 {
			_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SchemaQuotaType, domainProject, serviceId, int16(len(schemas)))
			if err != nil {
				util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s", serviceId)
				return scerr.NewError(scerr.ErrUnavailableQuota, err.Error())
			}
			if !ok {
				util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s", serviceId)
				return scerr.NewError(scerr.ErrNotEnoughQuota, "reach the max size of schema")
			}

			service.Schemas = nonExistSchemaIds
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "update service failed , %s", serviceId)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		} else {
			if len(nonExistSchemaIds) != 0 {
				util.Logger().Errorf(nil, "non-exist schemaId %#v", nonExistSchemaIds)
				return scerr.NewError(scerr.ErrInvalidParams, "non-exist schemaId")
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
					util.Logger().Warnf(nil, "schema and summary more existed, skip,service %s, schemaId %s", service.ServiceId, needUpdateSchema.SchemaId)
				}
			}
		}
	} else {
		needDeleteSchemas := make([]*pb.Schema, 0, len(schemasFromDatabase))
		for _, schemaFromDb := range schemasFromDatabase {
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

		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SchemaQuotaType, domainProject, serviceId, int16(quotaSize))
			if err != nil {
				util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s", serviceId)
				return scerr.NewError(scerr.ErrUnavailableQuota, err.Error())
			}
			if !ok {
				util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s", serviceId)
				return scerr.NewError(scerr.ErrNotEnoughQuota, "reach the max size of schema")
			}
		}

		for _, schema := range needUpdateSchemas {
			util.Logger().Infof("update schema %v", schema)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
		for _, schema := range needDeleteSchemas {
			util.Logger().Infof("delete not exist schema %v", schema)
			opts := schemaWithDatabaseOpera(registry.OpDel, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}

		if len(nonExistSchemaIds) != 0 {
			service.Schemas = append(service.Schemas, nonExistSchemaIds...)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "update service failed , %s", serviceId)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	for _, schema := range needAddSchemas {
		util.Logger().Infof("add new schema %v", schema)
		opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
		pluginOps = append(pluginOps, opts...)
	}

	if len(pluginOps) != 0 {
		err = backend.BatchCommit(ctx, pluginOps)
		if err != nil {
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
	}

	return nil
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
			Response: pb.CreateResponse(respErr.Code, respErr.Detail),
		}
		if respErr.StatusCode() == http.StatusInternalServerError {
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
	err, isInnerErr := s.modifySchema(ctx, serviceId, &schema)
	if err != nil {
		util.Logger().Errorf(err, "modify service schema failed")
		if isInnerErr {
			return &pb.ModifySchemaResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Modify schema info failed."),
			}, err
		}
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	util.Logger().Infof("update schema success: serviceId %s, schemaId %s.", serviceId, schemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil
}

func (s *MicroServiceService) canModifySchema(ctx context.Context, domainProject string, request *pb.ModifySchemaRequest) *scerr.Error {
	serviceId := request.ServiceId
	schemaId := request.SchemaId
	if len(schemaId) == 0 || len(serviceId) == 0 {
		util.Logger().Errorf(nil, "update schema failed: invalid params.")
		return scerr.NewError(scerr.ErrInvalidParams, "serviceId or schemaId is nil")
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "update schema failed, serviceId %s, schemaId %s: invalid params.", serviceId, schemaId)
		return scerr.NewError(scerr.ErrInvalidParams, err.Error())
	}

	_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SchemaQuotaType, domainProject, serviceId, 1)
	if err != nil {
		util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s, %s", serviceId, schemaId)
		return scerr.NewError(scerr.ErrUnavailableQuota, err.Error())
	}
	if !ok {
		util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s, %s", serviceId, schemaId)
		return scerr.NewError(scerr.ErrNotEnoughQuota, "add schema failed, reach max size of schema")
	}
	if len(request.Summary) == 0 {
		util.Logger().Warnf(nil, "service %s schema %s is empty.", request.ServiceId, schemaId)
	}
	return nil
}

func (s *MicroServiceService) modifySchema(ctx context.Context, serviceId string, schema *pb.Schema) (error, bool) {
	domainProject := util.ParseDomainProject(ctx)
	schemaId := schema.SchemaId

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "update schema failed, serviceId %s, schemaId %s: get service failed.", serviceId, schemaId)
		return err, true
	}
	if service == nil {
		util.Logger().Errorf(nil, "update schema failed, serviceId %s, schemaId %s: service not exist,%s", serviceId, schemaId)
		return errors.New("service non-exist"), false
	}

	util.Logger().Infof("modify schema, serviceId  %s", service.ServiceId)
	pluginOps := make([]registry.PluginOp, 0)
	isExist := isExistSchemaId(service, []*pb.Schema{schema})

	if service.Environment == pb.ENV_PROD {
		if len(service.Schemas) != 0 && !isExist {
			return errors.New("non-exist schemaId, prod mode"), false
		}

		key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schemaId)
		respSchema, err := store.Store().Schema().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
		if err != nil {
			util.Logger().Errorf(err, "get schema summary failed, %s %s", serviceId, schemaId)
			return err, true
		}

		if respSchema.Count != 0 {
			if len(schema.Summary) == 0 {
				util.Logger().Errorf(err, "prod mode, schema more exist, can not change, %s %s", serviceId, schemaId)
				return errors.New("schema more exist, can not change, prod mode"), false
			}

			exist, err := isExistSchemaSummary(ctx, domainProject, serviceId, schemaId)
			if err != nil {
				util.Logger().Errorf(err, "check schema summary is exist failed")
				return err, false
			}
			if exist {
				util.Logger().Errorf(err, "prod mode, schema more exist, can not change, %s %s", serviceId, schemaId)
				return errors.New("schema more exist, can not change, prod mode"), false
			}
		}

		if len(service.Schemas) == 0 {
			service.Schemas = append(service.Schemas, schemaId)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "update service failed , %s", serviceId)
				return err, true
			}
			pluginOps = append(pluginOps, opt)
		}
	} else {
		if !isExist {
			service.Schemas = append(service.Schemas, schemaId)
			opt, err := serviceUtil.UpdateService(domainProject, serviceId, service)
			if err != nil {
				util.Logger().Errorf(err, "update service failed , %s", serviceId)
				return err, true
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	opts := CommitSchemaInfo(domainProject, serviceId, schema)
	pluginOps = append(pluginOps, opts...)

	_, err = backend.Registry().Txn(ctx, pluginOps)
	if err != nil {
		util.Logger().Errorf(err, "commit update schema failed")
		return err, true
	}
	return nil, false
}

func isExistSchemaSummary(ctx context.Context, domainProject, serviceId, schemaId string) (bool, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schemaId)
	resp, err := store.Store().SchemaSummary().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
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
	resp, err := store.Store().SchemaSummary().Search(ctx,
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

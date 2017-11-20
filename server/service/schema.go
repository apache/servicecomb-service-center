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
	"encoding/json"
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
	"github.com/ServiceComb/service-center/version"
	"golang.org/x/net/context"
	"strings"
)

func (s *ServiceController) GetSchemaInfo(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
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

	opts := serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId, opts...) {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, in.ServiceId, in.SchemaId)
	opts = append(opts, registry.WithStrKey(key))
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

func (s *ServiceController) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.Logger().Errorf(nil, "delete schema failded: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}
	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, request.ServiceId) {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrSchemaNotExists, "Schema info does not exist."),
		}, nil
	}
	_, errDo := backend.Registry().Do(ctx, registry.DEL, registry.WithStrKey(key))
	if errDo != nil {
		util.Logger().Errorf(errDo, "delete schema failded, serviceId %s, schemaId %s: delete schema from etcd faild.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, "Delete schema info failed."),
		}, errDo
	}
	util.Logger().Infof("delete schema info successfully.%s", request.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *ServiceController) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (*pb.ModifySchemasResponse, error) {
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas failded: invalid params.")
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request."),
		}, nil
	}
	serviceId := request.ServiceId

	domainProject := util.ParseDomainProject(ctx)

	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "modify schemas failded: get service failed. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Invalid request."),
		}, err
	}
	if service == nil {
		util.Logger().Errorf(nil, "modify schemas failded: service does not exist. %s", serviceId)
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	err, isInnErr := modifySchemas(ctx, domainProject, service, request.Schemas)
	if err != nil {
		if isInnErr {
			return &pb.ModifySchemasResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
			}, err
		}
		return &pb.ModifySchemasResponse{
			Response: pb.CreateResponse(scerr.ErrModifySchemaNotAllow, err.Error()),
		}, nil
	}

	return &pb.ModifySchemasResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "modify schemas info successfully."),
	}, nil
}

func modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService, schemas []*pb.Schema) (err error, innerErr bool) {

	serviceId := service.ServiceId
	schemasInDataBase, err := GetSchemasFromDataBase(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(nil, "modify schema failed: get schema from database failed, %s", serviceId)
		return errors.New("exist not exist schemaId"), true
	}

	needUpdateSchemaList := make([]*pb.Schema, 0, len(schemas))
	needAddSchemaList := make([]*pb.Schema, 0, len(schemas))
	newSchemaIdList := make([]string, 0, len(schemas))

	for _, schema := range schemas {
		exist := false
		for _, schemaInner := range schemasInDataBase {
			if schema.SchemaId == schemaInner.SchemaId {
				needUpdateSchemaList = append(needUpdateSchemaList, schema)
				exist = true
				break
			}
		}
		if !exist {
			needAddSchemaList = append(needAddSchemaList, schema)
			newSchemaIdList = append(newSchemaIdList, schema.SchemaId)
		}
	}

	pluginOps := []registry.PluginOp{}
	switch version.Ver().RunMode {
	case "dev":
		needDeleteSchemaList := make([]*pb.Schema, 0, len(schemasInDataBase))
		for _, schemasInner := range schemasInDataBase {
			exist := false
			for _, schema := range schemas {
				if schema.SchemaId == schemasInner.SchemaId {
					exist = true
					break
				}
			}
			if !exist {
				needDeleteSchemaList = append(needDeleteSchemaList, schemasInner)
			}
		}

		quotaSize := len(needAddSchemaList) - len(needDeleteSchemaList)
		if quotaSize > 0 {
			_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SCHEMAQuotaType, domainProject, serviceId, int16(quotaSize))
			if err != nil {
				util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s", serviceId)
				return err, true
			}
			if !ok {
				util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s", serviceId)
				return errors.New("reach the max size of schema"), false
			}
		}

		for _, schema := range needUpdateSchemaList {
			util.Logger().Infof("update schema %v", schema)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
		for _, schema := range needDeleteSchemaList {
			util.Logger().Infof("delete not exist schema %v", schema)
			opts := schemaWithDatabaseOpera(registry.OpDel, domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
	case "prod":
		if len(service.Schemas) != 0 && !isExistSchemaId(service, schemas) {
			return errors.New("exist non-exist schemaId"), false
		}
		quotaSize := len(needAddSchemaList)
		if quotaSize > 0 {
			_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SCHEMAQuotaType, domainProject, serviceId, int16(quotaSize))
			if err != nil {
				util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s", serviceId)
				return err, true
			}
			if !ok {
				util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s", serviceId)
				return errors.New("reach the max size of schema"), false
			}
		}
		schemasFromDatabase, err := GetSchemasSummaryFromDataBase(ctx, domainProject, serviceId)
		if err != nil {
			util.Logger().Errorf(err, "get schema summary failed")
			return errors.New("run mode is prod, schema more exist, can't change"), false
		}
		for _, schema := range needUpdateSchemaList {
			exist := false
			for _, schemaDatabase := range schemasFromDatabase {
				if schema.SchemaId == schemaDatabase.SchemaId {
					exist = true
					break
				}
			}
			//key of summary not exist, exist one chance of changing schema
			if !exist {
				opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
				pluginOps = append(pluginOps, opts...)
			}
		}
		if len(service.Schemas) == 0 {
			key := apt.GenerateServiceKey(domainProject, serviceId)
			service.Schemas = newSchemaIdList
			data, err := json.Marshal(service)
			if err != nil {
				util.Logger().Errorf(err, "marshal service failed.")
				return err, true
			}
			opt := registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))
			pluginOps = append(pluginOps, opt)
		}
		if len(needUpdateSchemaList) > 0 && len(pluginOps) == 0 {
			util.Logger().Errorf(nil, "run mode is prod, schema more exist, can't change.%v", needUpdateSchemaList)
			return errors.New("run mode is prod, schema more exist, can't change"), false
		}
	}

	for _, schema := range needAddSchemaList {
		util.Logger().Infof("add new schema %v", schema)
		opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
		pluginOps = append(pluginOps, opts...)
	}

	if len(pluginOps) != 0 {
		return backend.BatchCommit(ctx, pluginOps), true
	}
	return nil, false
}

func isExistSchemaId(service *pb.MicroService, schemas []*pb.Schema) bool {
	seriviceSchemaIds := service.Schemas
	for _, schema := range schemas {
		if !containsValueInSlice(seriviceSchemaIds, schema.SchemaId) {
			util.Logger().Errorf(nil, "modify schemas failed: exist not exist schemaId, %s", schema.SchemaId)
			return false
		}
	}
	return true
}

func schemaWithDatabaseOpera(invoke registry.Operation, domainProject string, serviceId string, schema *pb.Schema) []registry.PluginOp {
	pluginOps := []registry.PluginOp{}
	key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schema.SchemaId)
	opt := invoke(registry.WithStrKey(key), registry.WithStrValue(schema.Schema))
	pluginOps = append(pluginOps, opt)
	keySummary := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schema.SchemaId)
	opt = invoke(registry.WithStrKey(keySummary), registry.WithStrValue(schema.Summary))
	pluginOps = append(pluginOps, opt)
	return pluginOps
}

func GetSchemasFromDataBase(ctx context.Context, domainProject string, serviceId string) ([]*pb.Schema, error) {
	schemas := []*pb.Schema{}
	key := apt.GenerateServiceSchemaKey(domainProject, serviceId, "")
	util.Logger().Debugf("key is %s", key)
	resp, err := backend.Registry().Do(ctx,
		registry.GET,
		registry.WithPrefix(),
		registry.WithStrKey(key))
	if err != nil {
		util.Logger().Errorf(err, "Get schema of one service failed. %s", serviceId)
		return schemas, err
	}
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

func GetSchemasSummaryFromDataBase(ctx context.Context, domainProject string, serviceId string) ([]*pb.Schema, error) {
	schemas := []*pb.Schema{}
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, "")
	util.Logger().Debugf("key is %s", key)
	resp, err := store.Store().SchemaSummary().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(key))
	if err != nil {
		util.Logger().Errorf(err, "Get schema of one service failed. %s", serviceId)
		return schemas, err
	}
	for _, kv := range resp.Kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		tmp := strings.Split(key, "/")
		schemaId := tmp[len(tmp)-1]
		summary := util.BytesToStringWithNoCopy(kv.Value)
		schemaStruct := &pb.Schema{
			SchemaId: schemaId,
			Summary:  summary,
		}
		schemas = append(schemas, schemaStruct)
	}
	return schemas, nil
}

func (s *ServiceController) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	err, rst := s.canModifySchema(ctx, domainProject, request)
	if err != nil {
		if !rst {
			return &pb.ModifySchemaResponse{
				Response: pb.CreateResponse(scerr.ErrModifySchemaNotAllow, err.Error()),
			}, nil
		}
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Modify schema info failed."),
		}, err
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
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Modify schema info failed."),
		}, nil
	}

	util.Logger().Infof("update schema success: serviceId %s, schemaId %s.", serviceId, schemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil
}

func (s *ServiceController) canModifySchema(ctx context.Context, domainProject string, request *pb.ModifySchemaRequest) (error, bool) {
	serviceId := request.ServiceId
	schemaId := request.SchemaId
	if len(schemaId) == 0 || len(serviceId) == 0 {
		util.Logger().Errorf(nil, "update schema failded: invalid params.")
		return errors.New("invalid request params"), false
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: invalid params.", serviceId, schemaId)
		return err, false
	}

	_, ok, err := plugin.Plugins().Quota().Apply4Quotas(ctx, quota.SCHEMAQuotaType, domainProject, serviceId, 1)
	if err != nil {
		util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s, %s", serviceId, schemaId)
		return err, true
	}
	if !ok {
		util.Logger().Errorf(err, "Add schema info failed, reach the max size of schema, %s, %s", serviceId, schemaId)
		return errors.New("add schema failed, reach max size of schema"), false
	}
	if len(request.Summary) == 0 {
		util.Logger().Infof("%s 's schema %s is empty.", request.ServiceId, schemaId)
	}
	return nil, true
}

func (s *ServiceController) modifySchema(ctx context.Context, serviceId string, schema *pb.Schema) (error, bool) {
	domainProject := util.ParseDomainProject(ctx)
	schemaId := schema.SchemaId
	service, err := serviceUtil.GetService(ctx, domainProject, serviceId)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: get service failed.", serviceId, schemaId)
		return err, true
	}
	if service == nil {
		util.Logger().Errorf(nil, "update schema failded, serviceId %s, schemaId %s: service not exist,%s", serviceId, schemaId)
		return errors.New("service non-exist"), false
	}
	pluginOps := []registry.PluginOp{}
	isExist := isExistSchemaId(service, []*pb.Schema{schema})
	if version.Ver().RunMode == "prod" {
		if !isExist {
			return errors.New("schemaId non-exist"), false
		}
		key := apt.GenerateServiceSchemaKey(domainProject, serviceId, schemaId)
		resp, err := store.Store().Schema().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
		if err != nil {
			util.Logger().Errorf(err, "get schema summary failed, %s %s", serviceId, schemaId)
			return err, true
		}

		if resp.Count == 0 {
			opts := CommitSchemaInfo(domainProject, serviceId, schema)
			pluginOps = append(pluginOps, opts...)
		} else {
			key = apt.GenerateServiceSchemaSummaryKey(domainProject, serviceId, schemaId)
			resp, err = store.Store().SchemaSummary().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
			if err != nil {
				util.Logger().Errorf(err, "get schema summary failed, %s %s", serviceId, schemaId)
				return err, true
			}
			if resp.Count == 0 {
				if len(schema.Summary) != 0 {
					opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceId, schema)
					pluginOps = append(pluginOps, opts...)
				}
			} else {
				util.Logger().Errorf(err, "prod mode, schema more exist, can not change, %s %s", serviceId, schemaId)
				return errors.New("prod mode, schema more exist, can not change"), false
			}
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
		opts := CommitSchemaInfo(domainProject, serviceId, schema)
		pluginOps = append(pluginOps, opts...)
	}
	if len(pluginOps) != 0 {
		_, err = backend.Registry().Txn(ctx, pluginOps)
		if err != nil {
			util.Logger().Errorf(err, "commit update schema failed")
			return err, true
		}
	}
	return nil, false
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

// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service"
	"os"
	"strings"
)

// TODO: define error with names here

var schemaEditable bool

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
	schemaEditableConfig := strings.ToLower(os.Getenv("SCHEMA_EDITABLE"))
	schemaEditable = strings.Compare(schemaEditableConfig, "true") == 0
}

type DataSource struct {
	// schemaEditable determines whether schema modification is allowed for
	SchemaEditable bool
}

func NewDataSource() *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("microservice mgt data source enable etcd mode")

	inst := &DataSource{}
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	ds.SchemaEditable = schemaEditable
	return nil
}

// RegisterService() implement:
// 1. request validation
// 2. capsule request to etcd kv format
// 3. invoke etcd client to store data
// 4. check etcd-client response && construct createServiceResponse
func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	if request == nil || request.Service == nil {
		log.Errorf(nil, "create micro-service failed: request body is empty")
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, "Request body is empty"),
		}, nil
	}

	// start to store service to etcd
	remoteIP := util.GetIPFromContext(ctx)
	serviceBody := request.Service
	serviceFlag := util.StringJoin([]string{
		serviceBody.Environment, serviceBody.AppId, serviceBody.ServiceName, serviceBody.Version}, "/")

	// request validation
	err := service.Validate(request)
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	// construct data to invoke etcd client
	opts, uniqueCmpOpts, failOpts, err := capRegisterData(ctx, request)
	if err != nil {
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	resp, err := backend.Registry().TxnWithCmp(ctx, opts, uniqueCmpOpts, failOpts)

	return newRegisterServiceResp(ctx, serviceBody, resp, err)
}

func (ds *DataSource) GetServices(ctx context.Context, request *pb.GetServicesRequest) (
	*pb.GetServicesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	services, err := getAllServiceUtil(ctx, domainProject)
	if err != nil {
		log.Errorf(err, "get all services by domain failed")
		return &pb.GetServicesResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServicesResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get all services successfully."),
		Services: services,
	}, nil
}

func (ds *DataSource) GetService(ctx context.Context, request *pb.GetServiceRequest) (
	*pb.GetServiceResponse, error) {
	err := service.Validate(request)
	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	singleService, err := getSingleService(ctx, domainProject, request.ServiceId)

	if err != nil {
		log.Errorf(err, "get micro-service[%s] failed, get service file failed", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if singleService == nil {
		log.Errorf(nil, "get micro-service[%s] failed, service does not exist", request.ServiceId)
		return &pb.GetServiceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	return &pb.GetServiceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service successfully."),
		Service:  singleService,
	}, nil
}

func (ds *DataSource) ExistService(ctx context.Context, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse,
	error) {
	domainProject := util.ParseDomainProject(ctx)
	serviceFlag := util.StringJoin([]string{
		request.Environment, request.AppId, request.ServiceName, request.Version}, "/")

	ids, exist, err := findServiceIds(ctx, request.Version, &pb.MicroServiceKey{
		Environment: request.Environment,
		AppId:       request.AppId,
		ServiceName: request.ServiceName,
		Alias:       request.ServiceName,
		Version:     request.Version,
		Tenant:      domainProject,
	})
	if err != nil {
		log.Errorf(err, "micro-service[%s] exist failed, find serviceIDs failed", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Infof("micro-service[%s] exist failed, service does not exist", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, serviceFlag+" does not exist."),
		}, nil
	}
	if len(ids) == 0 {
		log.Infof("micro-service[%s] exist failed, version mismatch", serviceFlag)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceVersionNotExists, serviceFlag+" version mismatch."),
		}, nil
	}
	return &pb.GetExistenceResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "get service id successfully."),
		ServiceId: ids[0], // 约定多个时，取较新版本
	}, nil
}

func (ds *DataSource) UpdateService(ctx context.Context, service *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {
	panic("implement me")
}

func (ds *DataSource) UnregisterService(ctx context.Context, service *pb.DeleteServiceRequest) (
	*pb.DeleteServiceResponse, error) {
	panic("implement me")
}

func (ds *DataSource) RegisterInstance() {
	panic("implement me")
}

func (ds *DataSource) SearchInstance() {
	panic("implement me")
}

func (ds *DataSource) UpdateInstance() {
	panic("implement me")
}

func (ds *DataSource) UnRegisterInstance() {
	panic("implement me")
}

func (ds *DataSource) ModifySchemas(ctx context.Context, request *pb.ModifySchemasRequest) (
	*pb.ModifySchemasResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)

	serviceInfo, err := getService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "modify service[%s] schemas failed, get service failed, operator: %s", serviceID, remoteIP)
		return &pb.ModifySchemasResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if serviceInfo == nil {
		log.Errorf(nil, "modify service[%s] schemas failed, service does not exist, operator: %s",
			serviceID, remoteIP)
		return &pb.ModifySchemasResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	respErr := ds.modifySchemas(ctx, domainProject, serviceInfo, request.Schemas)
	if respErr != nil {
		log.Errorf(nil, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
		resp := &pb.ModifySchemasResponse{
			Response: proto.CreateResponseWithSCErr(respErr),
		}
		if respErr.InternalError() {
			return resp, respErr
		}
		return resp, nil
	}

	return &pb.ModifySchemasResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "modify schemas info successfully."),
	}, nil
}

func (ds *DataSource) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (
	*pb.ModifySchemaResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := request.ServiceId
	schemaID := request.SchemaId

	schema := pb.Schema{
		SchemaId: schemaID,
		Summary:  request.Summary,
		Schema:   request.Schema,
	}
	err := ds.modifySchema(ctx, serviceID, &schema)
	if err != nil {
		log.Errorf(err, "modify schema[%s/%s] failed, operator: %s", serviceID, schemaID, remoteIP)
		resp := &pb.ModifySchemaResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	log.Infof("modify schema[%s/%s] successfully, operator: %s", serviceID, schemaID, remoteIP)
	return &pb.ModifySchemaResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "modify schema info success"),
	}, nil
}

func (ds *DataSource) GetSchema() {
	panic("implement me")
}

func (ds *DataSource) UpdateSchema() {
	panic("implement me")
}

func (ds *DataSource) DeleteSchema() {
	panic("implement me")
}

func (ds *DataSource) ExistSchema(ctx context.Context, request *pb.GetExistenceRequest) (
	*pb.GetExistenceResponse, error) {
	domainProject := util.ParseDomainProject(ctx)

	if !serviceExist(ctx, domainProject, request.ServiceId) {
		log.Warnf("schema[%s/%s] exist failed, service does not exist", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(domainProject, request.ServiceId, request.SchemaId)
	exist, err := checkSchemaInfoExist(ctx, key)
	if err != nil {
		log.Errorf(err, "schema[%s/%s] exist failed, get schema failed", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if !exist {
		log.Infof("schema[%s/%s] exist failed, schema does not exist", request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrSchemaNotExists, "schema does not exist."),
		}, nil
	}
	schemaSummary, err := getSchemaSummary(ctx, domainProject, request.ServiceId, request.SchemaId)
	if err != nil {
		log.Errorf(err, "schema[%s/%s] exist failed, get schema summary failed",
			request.ServiceId, request.SchemaId)
		return &pb.GetExistenceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	return &pb.GetExistenceResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Schema exist."),
		SchemaId: request.SchemaId,
		Summary:  schemaSummary,
	}, nil
}

func (ds *DataSource) AddTag() {
	panic("implement me")
}

func (ds *DataSource) GetTag() {
	panic("implement me")
}

func (ds *DataSource) UpdateTag() {
	panic("implement me")
}

func (ds *DataSource) DeleteTag() {
	panic("implement me")
}

func (ds *DataSource) AddRule() {
	panic("implement me")
}

func (ds *DataSource) GetRule() {
	panic("implement me")
}

func (ds *DataSource) UpdateRule() {
	panic("implement me")
}

func (ds *DataSource) DeleteRule() {
	panic("implement me")
}

func (ds *DataSource) modifySchemas(ctx context.Context, domainProject string, service *pb.MicroService,
	schemas []*pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	serviceID := service.ServiceId
	schemasFromDatabase, err := getSchemasFromDatabase(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(nil, "modify service[%s] schemas failed, get schemas failed, operator: %s",
			serviceID, remoteIP)
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}

	needUpdateSchemas, needAddSchemas, needDeleteSchemas, nonExistSchemaIds :=
		schemasAnalysis(schemas, schemasFromDatabase, service.Schemas)

	pluginOps := make([]registry.PluginOp, 0)
	if !ds.isSchemaEditable(service) {
		if len(service.Schemas) == 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(len(nonExistSchemaIds)))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			errQuota := rst.Err
			if errQuota != nil {
				log.Errorf(errQuota, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return errQuota
			}

			service.Schemas = nonExistSchemaIds
			opt, err := updateService(domainProject, serviceID, service)
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
					serviceID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		} else {
			if len(nonExistSchemaIds) != 0 {
				errInfo := fmt.Errorf("non-existent schemaIDs %v", nonExistSchemaIds)
				log.Errorf(errInfo, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return scerr.NewError(scerr.ErrUndefinedSchemaID, errInfo.Error())
			}
			for _, needUpdateSchema := range needUpdateSchemas {
				exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, needUpdateSchema.SchemaId)
				if err != nil {
					return scerr.NewError(scerr.ErrInternal, err.Error())
				}
				if !exist {
					opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceID, needUpdateSchema)
					pluginOps = append(pluginOps, opts...)
				} else {
					log.Warnf("schema[%s/%s] and it's summary already exist, skip to update, operator: %s",
						serviceID, needUpdateSchema.SchemaId, remoteIP)
				}
			}
		}

		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
		}
	} else {
		quotaSize := len(needAddSchemas) - len(needDeleteSchemas)
		if quotaSize > 0 {
			res := quota.NewApplyQuotaResource(quota.SchemaQuotaType, domainProject, serviceID, int64(quotaSize))
			rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
			err := rst.Err
			if err != nil {
				log.Errorf(err, "modify service[%s] schemas failed, operator: %s", serviceID, remoteIP)
				return err
			}
		}

		var schemaIDs []string
		for _, schema := range needAddSchemas {
			log.Infof("add new schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, service.ServiceId, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needUpdateSchemas {
			log.Infof("update schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
			schemaIDs = append(schemaIDs, schema.SchemaId)
		}

		for _, schema := range needDeleteSchemas {
			log.Infof("delete non-existent schema[%s/%s], operator: %s", serviceID, schema.SchemaId, remoteIP)
			opts := schemaWithDatabaseOpera(registry.OpDel, domainProject, serviceID, schema)
			pluginOps = append(pluginOps, opts...)
		}

		service.Schemas = schemaIDs
		opt, err := updateService(domainProject, serviceID, service)
		if err != nil {
			log.Errorf(err, "modify service[%s] schemas failed, update service.Schemas failed, operator: %s",
				serviceID, remoteIP)
			return scerr.NewError(scerr.ErrInternal, err.Error())
		}
		pluginOps = append(pluginOps, opt)
	}

	if len(pluginOps) != 0 {
		resp, err := backend.BatchCommitWithCmp(ctx, pluginOps,
			[]registry.CompareOp{registry.OpCmp(
				registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
				registry.CmpNotEqual, 0)},
			nil)
		if err != nil {
			return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
		}
		if !resp.Succeeded {
			return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
		}
	}
	return nil
}

func (ds *DataSource) isSchemaEditable(service *pb.MicroService) bool {
	return (len(service.Environment) != 0 && service.Environment != pb.ENV_PROD) || ds.SchemaEditable
}

func (ds *DataSource) modifySchema(ctx context.Context, serviceID string, schema *pb.Schema) *scerr.Error {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	schemaID := schema.SchemaId

	microService, err := getService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "modify schema[%s/%s] failed, get `microService failed, operator: %s",
			serviceID, schemaID, remoteIP)
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	if microService == nil {
		log.Errorf(nil, "modify schema[%s/%s] failed, microService does not exist, operator: %s",
			serviceID, schemaID, remoteIP)
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist")
	}

	var pluginOps []registry.PluginOp
	isExist := isExistSchemaID(microService, []*pb.Schema{schema})

	if !ds.isSchemaEditable(microService) {
		if len(microService.Schemas) != 0 && !isExist {
			return scerr.NewError(scerr.ErrUndefinedSchemaID, "Non-existent schemaID can't be added in "+pb.ENV_PROD)
		}

		key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schemaID)
		respSchema, err := backend.Store().Schema().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
		if err != nil {
			log.Errorf(err, "modify schema[%s/%s] failed, get schema summary failed, operator: %s",
				serviceID, schemaID, remoteIP)
			return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
		}

		if respSchema.Count != 0 {
			if len(schema.Summary) == 0 {
				log.Errorf(err, "%s mode, schema[%s/%s] already exists, can not be changed, operator: %s",
					pb.ENV_PROD, serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow,
					"schema already exist, can not be changed in "+pb.ENV_PROD)
			}

			exist, err := isExistSchemaSummary(ctx, domainProject, serviceID, schemaID)
			if err != nil {
				log.Errorf(err, "check schema[%s/%s] summary existence failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			if exist {
				log.Errorf(err, "%s mode, schema[%s/%s] already exist, can not be changed, operator: %s",
					pb.ENV_PROD, serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrModifySchemaNotAllow, "schema already exist, can not be changed in "+pb.ENV_PROD)
			}
		}

		if len(microService.Schemas) == 0 {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := updateService(domainProject, serviceID, microService)
			if err != nil {
				log.Errorf(err, "modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	} else {
		if !isExist {
			microService.Schemas = append(microService.Schemas, schemaID)
			opt, err := updateService(domainProject, serviceID, microService)
			if err != nil {
				log.Errorf(err, "modify schema[%s/%s] failed, update microService.Schemas failed, operator: %s",
					serviceID, schemaID, remoteIP)
				return scerr.NewError(scerr.ErrInternal, err.Error())
			}
			pluginOps = append(pluginOps, opt)
		}
	}

	opts := commitSchemaInfo(domainProject, serviceID, schema)
	pluginOps = append(pluginOps, opts...)

	resp, err := backend.Registry().TxnWithCmp(ctx, pluginOps,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceID))),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
	}
	return nil
}

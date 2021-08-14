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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

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
	depUtil "github.com/apache/servicecomb-service-center/server/service/dep/etcd"
	"github.com/apache/servicecomb-service-center/server/service/ms"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

// TODO: define error with names here

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
}

type DataSource struct {
	// schemaEditable determines whether schema modification is allowed for
	SchemaEditable bool
	ttlFromEnv     int64
}

func NewDataSource(opts ms.Options) *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("microservice mgt data source enable etcd mode")

	inst := &DataSource{
		SchemaEditable: opts.SchemaEditable,
		ttlFromEnv:     opts.TTL,
	}
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	return nil
}

// RegisterService() implement:
// 1. capsule request to etcd kv format
// 2. invoke etcd client to store data
// 3. check etcd-client response && construct createServiceResponse
func (ds *DataSource) RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (
	*pb.CreateServiceResponse, error) {
	// start to store service to etcd
	serviceBody := request.Service

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

func (ds *DataSource) UpdateService(ctx context.Context, request *pb.UpdateServicePropsRequest) (
	*pb.UpdateServicePropsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	key := apt.GenerateServiceKey(domainProject, request.ServiceId)
	microservice, err := getService(ctx, domainProject, request.ServiceId)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, get service file failed, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	if microservice == nil {
		log.Errorf(nil, "update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	copyServiceRef := *microservice
	copyServiceRef.Properties = request.Properties
	copyServiceRef.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)

	data, err := json.Marshal(copyServiceRef)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, json marshal service failed, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	// Set key file
	resp, err := backend.Registry().TxnWithCmp(ctx,
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))},
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(key)),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "update service[%s] properties failed, operator: %s", request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "update service[%s] properties failed, service does not exist, operator: %s",
			request.ServiceId, remoteIP)
		return &pb.UpdateServicePropsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("update service[%s] properties successfully, operator: %s", request.ServiceId, remoteIP)
	return &pb.UpdateServicePropsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "update service successfully."),
	}, nil
}

func (ds *DataSource) UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (
	*pb.DeleteServiceResponse, error) {
	resp, err := ds.DeleteServicePri(ctx, request.ServiceId, request.Force)
	return &pb.DeleteServiceResponse{
		Response: resp,
	}, err
}

func (ds *DataSource) RegisterInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (
	*pb.RegisterInstanceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance

	//允许自定义id
	if len(instance.InstanceId) > 0 {
		// keep alive the lease ttl
		// there are two reasons for sending a heartbeat here:
		// 1. in the scenario the instance has been removed,
		//    the cast of registration operation can be reduced.
		// 2. in the self-protection scenario, the instance is unhealthy
		//    and needs to be re-registered.
		resp, err := ds.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceId: instance.ServiceId,
			InstanceId: instance.InstanceId})
		if resp == nil {
			log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
				instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
			}, nil
		}
		switch resp.Response.GetCode() {
		case proto.Response_SUCCESS:
			log.Infof("register instance successful, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response:   resp.Response,
				InstanceId: instance.InstanceId,
			}, nil
		case scerr.ErrInstanceNotExists:
			// register a new one
		default:
			log.Errorf(err, "register instance failed, reuse instance[%s/%s], operator %s",
				instance.ServiceId, instance.InstanceId, remoteIP)
			return &pb.RegisterInstanceResponse{
				Response: resp.Response,
			}, err
		}
	}

	if err := preProcessRegisterInstance(ctx, instance); err != nil {
		log.Errorf(err, "register service[%s]'s instance failed, endpoints %v, host '%s', operator %s",
			instance.ServiceId, instance.Endpoints, instance.HostName, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}, nil
	}

	ttl := int64(instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1))
	if ds.ttlFromEnv > 0 {
		ttl = ds.ttlFromEnv
	}
	instanceFlag := fmt.Sprintf("ttl %ds, endpoints %v, host '%s', serviceID %s",
		ttl, instance.Endpoints, instance.HostName, instance.ServiceId)

	//先以domain/project的方式组装
	domainProject := util.ParseDomainProject(ctx)

	var reporter *quota.ApplyQuotaResult
	if !apt.IsSCInstance(ctx) {
		res := quota.NewApplyQuotaResource(quota.MicroServiceInstanceQuotaType,
			domainProject, request.Instance.ServiceId, 1)
		reporter = plugin.Plugins().Quota().Apply4Quotas(ctx, res)
		defer reporter.Close(ctx)

		if reporter.Err != nil {
			log.Errorf(reporter.Err, "register instance failed, %s, operator %s",
				instanceFlag, remoteIP)
			response := &pb.RegisterInstanceResponse{
				Response: proto.CreateResponseWithSCErr(reporter.Err),
			}
			if reporter.Err.InternalError() {
				return response, reporter.Err
			}
			return response, nil
		}
	}

	instanceID := instance.InstanceId
	data, err := json.Marshal(instance)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	leaseID, err := backend.Registry().LeaseGrant(ctx, ttl)
	if err != nil {
		log.Errorf(err, "grant lease failed, %s, operator: %s", instanceFlag, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	// build the request options
	key := apt.GenerateInstanceKey(domainProject, instance.ServiceId, instanceID)
	hbKey := apt.GenerateInstanceLeaseKey(domainProject, instance.ServiceId, instanceID)

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithStrKey(key), registry.WithValue(data),
			registry.WithLease(leaseID)),
		registry.OpPut(registry.WithStrKey(hbKey), registry.WithStrValue(fmt.Sprintf("%d", leaseID)),
			registry.WithLease(leaseID)),
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, instance.ServiceId))),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(nil,
			"register instance failed, %s, instanceID %s, operator %s: service does not exist",
			instanceFlag, instanceID, remoteIP)
		return &pb.RegisterInstanceResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	if err := reporter.ReportUsedQuota(ctx); err != nil {
		log.Errorf(err,
			"register instance failed, %s, instanceID %s, operator %s",
			instanceFlag, instanceID, remoteIP)
	}

	log.Infof("register instance %s, instanceID %s, operator %s",
		instanceFlag, instanceID, remoteIP)
	return &pb.RegisterInstanceResponse{
		Response:   proto.CreateResponse(proto.Response_SUCCESS, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
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

func (ds *DataSource) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)
	instanceFlag := util.StringJoin([]string{request.ServiceId, request.InstanceId}, "/")

	_, ttl, err := HeartbeatUtil(ctx, domainProject, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Errorf(err, "heartbeat failed, instance[%s]. operator %s",
			instanceFlag, remoteIP)
		resp := &pb.HeartbeatResponse{
			Response: proto.CreateResponseWithSCErr(err),
		}
		if err.InternalError() {
			return resp, err
		}
		return resp, nil
	}

	if ttl == 0 {
		log.Errorf(errors.New("connect backend timed out"),
			"heartbeat successful, but renew instance[%s] failed. operator %s", instanceFlag, remoteIP)
	} else {
		log.Infof("heartbeat successful, renew instance[%s] ttl to %d. operator %s", instanceFlag, ttl, remoteIP)
	}
	return &pb.HeartbeatResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Update service instance heartbeat successfully."),
	}, nil
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

func (ds *DataSource) DeleteServicePri(ctx context.Context, serviceID string, force bool) (*pb.Response, error) {
	remoteIP := util.GetIPFromContext(ctx)
	domainProject := util.ParseDomainProject(ctx)

	title := "delete"
	if force {
		title = "force delete"
	}

	if serviceID == apt.Service.ServiceId {
		err := errors.New("not allow to delete service center")
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInvalidParams, err.Error()), nil
	}

	microservice, err := getService(ctx, domainProject, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, get service file failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}

	if microservice == nil {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	// 强制删除，则与该服务相关的信息删除，非强制删除： 如果作为该被依赖（作为provider，提供服务,且不是只存在自依赖）或者存在实例，则不能删除
	if !force {
		dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, microservice)
		services, err := dr.GetDependencyConsumerIds()
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get service dependency failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
		}
		if l := len(services); l > 1 || (l == 1 && services[0] != serviceID) {
			log.Errorf(nil, "delete micro-service[%s] failed, other services[%d] depend on it, operator: %s",
				serviceID, l, remoteIP)
			return proto.CreateResponse(scerr.ErrDependedOnConsumer, "Can not delete this service, other service rely it."), err
		}

		instancesKey := apt.GenerateInstanceKey(domainProject, serviceID, "")
		rsp, err := backend.Store().Instance().Search(ctx,
			registry.WithStrKey(instancesKey),
			registry.WithPrefix(),
			registry.WithCountOnly())
		if err != nil {
			log.Errorf(err, "delete micro-service[%s] failed, get instances failed, operator: %s",
				serviceID, remoteIP)
			return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
		}

		if rsp.Count > 0 {
			log.Errorf(nil, "delete micro-service[%s] failed, service deployed instances[%s], operator: %s",
				serviceID, rsp.Count, remoteIP)
			return proto.CreateResponse(scerr.ErrDeployedInstance, "Can not delete the service deployed instance(s)."), err
		}
	}

	serviceIDKey := apt.GenerateServiceKey(domainProject, serviceID)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: microservice.Environment,
		AppId:       microservice.AppId,
		ServiceName: microservice.ServiceName,
		Version:     microservice.Version,
		Alias:       microservice.Alias,
	}
	opts := []registry.PluginOp{
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceIndexKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(apt.GenerateServiceAliasKey(serviceKey))),
		registry.OpDel(registry.WithStrKey(serviceIDKey)),
	}

	//删除依赖规则
	optDeleteDep, err := depUtil.DeleteDependencyForDeleteService(domainProject, serviceID, serviceKey)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, delete dependency failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrInternal, err.Error()), err
	}
	opts = append(opts, optDeleteDep)

	//删除黑白名单
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceRuleKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(registry.WithStrKey(
		util.StringJoin([]string{apt.GetServiceRuleIndexRootKey(domainProject), serviceID, ""}, "/")),
		registry.WithPrefix()))

	//删除schemas
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, "")),
		registry.WithPrefix()))

	//删除tags
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateServiceTagKey(domainProject, serviceID))))

	//删除instances
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateInstanceKey(domainProject, serviceID, "")),
		registry.WithPrefix()))
	opts = append(opts, registry.OpDel(
		registry.WithStrKey(apt.GenerateInstanceLeaseKey(domainProject, serviceID, "")),
		registry.WithPrefix()))

	//删除实例
	err = serviceUtil.DeleteServiceAllInstances(ctx, serviceID)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, revoke all instances failed, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}

	resp, err := backend.Registry().TxnWithCmp(ctx, opts,
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(serviceIDKey)),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "%s micro-service[%s] failed, operator: %s", title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()), err
	}
	if !resp.Succeeded {
		log.Errorf(err, "%s micro-service[%s] failed, service does not exist, operator: %s",
			title, serviceID, remoteIP)
		return proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."), nil
	}

	serviceUtil.RemandServiceQuota(ctx)

	log.Infof("%s micro-service[%s] successfully, operator: %s", title, serviceID, remoteIP)
	return proto.CreateResponse(proto.Response_SUCCESS, "Unregister service successfully."), nil
}

func (ds *DataSource) GetDeleteServiceFunc(ctx context.Context, serviceID string, force bool,
	serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		resp, err := ds.DeleteServicePri(ctx, serviceID, force)
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		} else if resp.GetCode() != proto.Response_SUCCESS {
			serviceRst.ErrMessage = resp.GetMessage()
		}

		serviceRespChan <- serviceRst
	}
}

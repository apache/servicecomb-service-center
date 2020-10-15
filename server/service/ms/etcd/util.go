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
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"strconv"
	"strings"
	"time"
)

// service
func capRegisterData(ctx context.Context, request *pb.CreateServiceRequest) (
	[]registry.PluginOp, []registry.CompareOp, []registry.PluginOp, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceBody := request.Service
	serviceFlag := util.StringJoin([]string{
		serviceBody.Environment, serviceBody.AppId, serviceBody.ServiceName, serviceBody.Version}, "/")
	serviceUtil.SetServiceDefaultValue(serviceBody)
	domainProject := util.ParseDomainProject(ctx)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: serviceBody.Environment,
		AppId:       serviceBody.AppId,
		ServiceName: serviceBody.ServiceName,
		Alias:       serviceBody.Alias,
		Version:     serviceBody.Version,
	}
	index := apt.GenerateServiceIndexKey(serviceKey)

	// generate global service id
	requestServiceID := serviceBody.ServiceId
	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, index)
		serviceBody.ServiceId = plugin.Plugins().UUID().GetServiceID(ctx)
	}
	serviceBody.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	serviceBody.ModTimestamp = serviceBody.Timestamp

	data, err := json.Marshal(serviceBody)
	if err != nil {
		log.Errorf(err, "create micro-serviceBody[%s] failed, json marshal serviceBody failed, operator: %s",
			serviceFlag, remoteIP)
		return nil, nil, nil, err
	}

	key := apt.GenerateServiceKey(domainProject, serviceBody.ServiceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(serviceKey))

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithKey(keyBytes), registry.WithValue(data)),
		registry.OpPut(registry.WithKey(indexBytes), registry.WithStrValue(serviceBody.ServiceId)),
	}
	uniqueCmpOpts := []registry.CompareOp{
		registry.OpCmp(registry.CmpVer(indexBytes), registry.CmpEqual, 0),
		registry.OpCmp(registry.CmpVer(keyBytes), registry.CmpEqual, 0),
	}
	failOpts := []registry.PluginOp{
		registry.OpGet(registry.WithKey(indexBytes)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, registry.OpPut(registry.WithKey(aliasBytes), registry.WithStrValue(serviceBody.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			registry.OpCmp(registry.CmpVer(aliasBytes), registry.CmpEqual, 0))
		failOpts = append(failOpts, registry.OpGet(registry.WithKey(aliasBytes)))
	}

	return opts, uniqueCmpOpts, failOpts, nil
}

func newRegisterServiceResp(ctx context.Context, reqService *pb.MicroService, resp *registry.PluginResponse,
	err error) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceFlag := util.StringJoin([]string{
		reqService.Environment, reqService.AppId, reqService.ServiceName, reqService.Version}, "/")
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	if !resp.Succeeded {
		requestServiceID := reqService.ServiceId
		if len(requestServiceID) != 0 {
			if len(resp.Kvs) == 0 ||
				requestServiceID != util.BytesToStringWithNoCopy(resp.Kvs[0].Value) {
				log.Warnf("create micro-service[%s] failed, service already exists, operator: %s",
					serviceFlag, remoteIP)
				return &pb.CreateServiceResponse{
					Response: proto.CreateResponse(scerr.ErrServiceAlreadyExists,
						"ServiceID conflict or found the same service with different id."),
				}, nil
			}
		}

		if len(resp.Kvs) == 0 {
			// internal error?
			log.Errorf(nil, "create micro-service[%s] failed, unexpected txn response, operator: %s",
				serviceFlag, remoteIP)
			return &pb.CreateServiceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, "Unexpected txn response."),
			}, nil
		}

		serviceIDInner := util.BytesToStringWithNoCopy(resp.Kvs[0].Value)
		log.Warnf("create micro-service[%s][%s] failed, service already exists, operator: %s",
			serviceIDInner, serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response:  proto.CreateResponse(proto.Response_SUCCESS, "register service successfully"),
			ServiceId: serviceIDInner,
		}, nil
	}

	log.Infof("create micro-service[%s][%s] successfully, operator: %s",
		reqService.ServiceId, serviceFlag, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "Register service successfully."),
		ServiceId: reqService.ServiceId,
	}, nil
}

// schema
func getSchemaSummary(ctx context.Context, domainProject string, serviceID string, schemaID string) (string, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schemaID)
	resp, err := backend.Store().SchemaSummary().Search(ctx,
		registry.WithStrKey(key),
	)
	if err != nil {
		log.Errorf(err, "get schema[%s/%s] summary failed", serviceID, schemaID)
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func getSchemasFromDatabase(ctx context.Context, domainProject string, serviceID string) ([]*pb.Schema, error) {
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, "")
	resp, err := backend.Store().Schema().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(key))
	if err != nil {
		log.Errorf(err, "get service[%s]'s schema failed", serviceID)
		return nil, err
	}
	schemas := make([]*pb.Schema, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		tmp := strings.Split(key, "/")
		schemaID := tmp[len(tmp)-1]
		schema := util.BytesToStringWithNoCopy(kv.Value.([]byte))
		schemaStruct := &pb.Schema{
			SchemaId: schemaID,
			Schema:   schema,
		}
		schemas = append(schemas, schemaStruct)
	}
	return schemas, nil
}

func schemasAnalysis(schemas []*pb.Schema, schemasFromDb []*pb.Schema, schemaIDsInService []string) (
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
		for _, schemaID := range schemaIDsInService {
			if schema.SchemaId == schemaID {
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

func checkSchemaInfoExist(ctx context.Context, key string) (bool, error) {
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithCountOnly())
	resp, errDo := backend.Store().Schema().Search(ctx, opts...)
	if errDo != nil {
		return false, errDo
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func isExistSchemaSummary(ctx context.Context, domainProject, serviceID, schemaID string) (bool, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schemaID)
	resp, err := backend.Store().SchemaSummary().Search(ctx, registry.WithStrKey(key), registry.WithCountOnly())
	if err != nil {
		return true, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func schemaWithDatabaseOpera(invoke registry.Operation, domainProject string, serviceID string, schema *pb.Schema) []registry.PluginOp {
	pluginOps := make([]registry.PluginOp, 0)
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	opt := invoke(registry.WithStrKey(key), registry.WithStrValue(schema.Schema))
	pluginOps = append(pluginOps, opt)
	keySummary := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schema.SchemaId)
	opt = invoke(registry.WithStrKey(keySummary), registry.WithStrValue(schema.Summary))
	pluginOps = append(pluginOps, opt)
	return pluginOps
}

func isExistSchemaID(service *pb.MicroService, schemas []*pb.Schema) bool {
	serviceSchemaIds := service.Schemas
	for _, schema := range schemas {
		if !containsValueInSlice(serviceSchemaIds, schema.SchemaId) {
			log.Errorf(nil, "schema[%s/%s] does not exist schemaID", service.ServiceId, schema.SchemaId)
			return false
		}
	}
	return true
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

func commitSchemaInfo(domainProject string, serviceID string, schema *pb.Schema) []registry.PluginOp {
	if len(schema.Summary) != 0 {
		return schemaWithDatabaseOpera(registry.OpPut, domainProject, serviceID, schema)
	}
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	opt := registry.OpPut(registry.WithStrKey(key), registry.WithStrValue(schema.Schema))
	return []registry.PluginOp{opt}
}

// instance util
func preProcessRegisterInstance(ctx context.Context, instance *pb.MicroServiceInstance) *scerr.Error {
	if len(instance.Status) == 0 {
		instance.Status = pb.MSI_UP
	}

	if len(instance.InstanceId) == 0 {
		instance.InstanceId = plugin.Plugins().UUID().GetInstanceID(ctx)
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
				return scerr.NewError(scerr.ErrInvalidParams, "Invalid 'healthCheck' settings in request body.")
			}
		case pb.CHECK_BY_PLATFORM:
			// 默认120s
			instance.HealthCheck.Interval = renewalInterval
			instance.HealthCheck.Times = retryTimes
		}
	}

	domainProject := util.ParseDomainProject(ctx)
	microservice, err := serviceUtil.GetService(ctx, domainProject, instance.ServiceId)
	if microservice == nil || err != nil {
		return scerr.NewError(scerr.ErrServiceNotExists, "Invalid 'serviceID' in request body.")
	}
	instance.Version = microservice.Version
	return nil
}

func getHeartbeatFunc(ctx context.Context, domainProject string, instancesHbRst chan<- *pb.InstanceHbRst, element *pb.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &pb.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}
		_, _, err := serviceUtil.HeartbeatUtil(ctx, domainProject, element.ServiceId, element.InstanceId)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Errorf(err, "heartbeat set failed, %s/%s", element.ServiceId, element.InstanceId)
		}
		instancesHbRst <- hbRst
	}
}

func revokeInstance(ctx context.Context, domainProject string, serviceID string, instanceID string) *scerr.Error {
	leaseID, err := serviceUtil.GetLeaseID(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	if leaseID == -1 {
		return scerr.NewError(scerr.ErrInstanceNotExists, "Instance's leaseId not exist.")
	}

	err = backend.Registry().LeaseRevoke(ctx, leaseID)
	if err != nil {
		if _, ok := err.(errorsEx.InternalError); !ok {
			return scerr.NewError(scerr.ErrInstanceNotExists, err.Error())
		}
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	return nil
}

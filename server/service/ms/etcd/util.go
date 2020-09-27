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
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"strconv"
	"strings"
	"time"
)

// service
func getSingleService(ctx context.Context, domainProject string, serviceID string) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domainProject, serviceID)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key))
	serviceResp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, nil
	}
	return serviceResp.Kvs[0].Value.(*pb.MicroService), nil
}

func getAllServiceUtil(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	services, err := getServicesByDomainProject(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func getServicesByDomainProject(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	kvs, err := getServicesRawData(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	var services []*pb.MicroService
	for _, kv := range kvs {
		services = append(services, kv.Value.(*pb.MicroService))
	}
	return services, nil
}

func getServicesRawData(ctx context.Context, domainProject string) ([]*discovery.KeyValue, error) {
	key := apt.GenerateServiceKey(domainProject, "")
	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, err
}

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

func serviceExist(ctx context.Context, domainProject string, serviceID string) bool {
	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(apt.GenerateServiceKey(domainProject, serviceID)),
		registry.WithCountOnly())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func findServiceIds(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, bool, error) {
	// 版本规则
	match := serviceUtil.ParseVersionRule(versionRule)
	if match == nil {
		copyData := *key
		copyData.Version = versionRule
		serviceID, err := getServiceID(ctx, &copyData)
		if err != nil {
			return nil, false, err
		}
		if len(serviceID) > 0 {
			return []string{serviceID}, true, nil
		}
		return nil, false, nil
	}

	searchAlias := false
	alsoFindAlias := len(key.Alias) > 0

FindRule:
	resp, err := getServiceAllVersions(ctx, key, searchAlias)
	if err != nil {
		return nil, false, err
	}
	if len(resp.Kvs) == 0 {
		if !alsoFindAlias {
			return nil, false, nil
		}
		searchAlias = true
		alsoFindAlias = false
		goto FindRule
	}
	return match(resp.Kvs), true, nil
}

func getServiceID(ctx context.Context, key *pb.MicroServiceKey) (serviceID string, err error) {
	serviceID, err = searchServiceID(ctx, key)
	if err != nil {
		return
	}
	if len(serviceID) == 0 {
		// 别名查询
		log.Debugf("could not search microservice[%s/%s/%s/%s] id by 'serviceName', now try 'alias'",
			key.Environment, key.AppId, key.ServiceName, key.Version)
		return searchServiceIDFromAlias(ctx, key)
	}
	return
}

func searchServiceID(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(apt.GenerateServiceIndexKey(key)))
	resp, err := backend.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func searchServiceIDFromAlias(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(apt.GenerateServiceAliasKey(key)))
	resp, err := backend.Store().ServiceAlias().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func getServiceAllVersions(ctx context.Context, key *pb.MicroServiceKey, alias bool) (*discovery.Response, error) {
	copyData := *key
	copyData.Version = ""
	var (
		prefix  string
		indexer discovery.Indexer
	)
	if alias {
		prefix = apt.GenerateServiceAliasKey(&copyData)
		indexer = backend.Store().ServiceAlias()
	} else {
		prefix = apt.GenerateServiceIndexKey(&copyData)
		indexer = backend.Store().ServiceIndex()
	}
	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(prefix),
		registry.WithPrefix(),
		registry.WithDescendOrder())
	resp, err := indexer.Search(ctx, opts...)
	return resp, err
}

func updateService(domainProject string, serviceID string, service *pb.MicroService) (opt registry.PluginOp,
	err error) {
	opt = registry.PluginOp{}
	key := apt.GenerateServiceKey(domainProject, serviceID)
	data, err := json.Marshal(service)
	if err != nil {
		log.Errorf(err, "marshal service file failed")
		return
	}
	opt = registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))
	return
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

func getService(ctx context.Context, domainProject string, serviceID string) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domainProject, serviceID)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key))
	serviceResp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, nil
	}
	return serviceResp.Kvs[0].Value.(*pb.MicroService), nil
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

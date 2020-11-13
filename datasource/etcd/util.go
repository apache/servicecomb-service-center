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

package etcd

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/pkg/registry"
	"strconv"
	"strings"
	"time"
)

type ServiceDetailOpt struct {
	domainProject string
	service       *pb.MicroService
	countOnly     bool
	options       []string
}

type GetInstanceCountByDomainResponse struct {
	err           error
	countByDomain int64
}

// schema
func getSchemaSummary(ctx context.Context, domainProject string, serviceID string, schemaID string) (string, error) {
	key := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schemaID)
	resp, err := kv.Store().SchemaSummary().Search(ctx,
		client.WithStrKey(key),
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
	resp, err := kv.Store().Schema().Search(ctx,
		client.WithPrefix(),
		client.WithStrKey(key))
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
	opts := append(serviceUtil.FromContext(ctx), client.WithStrKey(key), client.WithCountOnly())
	resp, errDo := kv.Store().Schema().Search(ctx, opts...)
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
	resp, err := kv.Store().SchemaSummary().Search(ctx, client.WithStrKey(key), client.WithCountOnly())
	if err != nil {
		return true, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func schemaWithDatabaseOpera(invoke client.Operation, domainProject string, serviceID string, schema *pb.Schema) []client.PluginOp {
	pluginOps := make([]client.PluginOp, 0)
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	opt := invoke(client.WithStrKey(key), client.WithStrValue(schema.Schema))
	pluginOps = append(pluginOps, opt)
	keySummary := apt.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schema.SchemaId)
	opt = invoke(client.WithStrKey(keySummary), client.WithStrValue(schema.Summary))
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

func commitSchemaInfo(domainProject string, serviceID string, schema *pb.Schema) []client.PluginOp {
	if len(schema.Summary) != 0 {
		return schemaWithDatabaseOpera(client.OpPut, domainProject, serviceID, schema)
	}
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	opt := client.OpPut(client.WithStrKey(key), client.WithStrValue(schema.Schema))
	return []client.PluginOp{opt}
}

// instance util
func preProcessRegisterInstance(ctx context.Context, instance *pb.MicroServiceInstance) *scerr.Error {
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

	err = client.Instance().LeaseRevoke(ctx, leaseID)
	if err != nil {
		if _, ok := err.(errorsEx.InternalError); !ok {
			return scerr.NewError(scerr.ErrInstanceNotExists, err.Error())
		}
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	return nil
}

// governServiceCtrl util
func getServiceAllVersions(ctx context.Context, serviceKey *pb.MicroServiceKey) ([]string, error) {
	var versions []string

	copyKey := *serviceKey
	copyKey.Version = ""
	key := GenerateServiceIndexKey(&copyKey)

	opts := append(serviceUtil.FromContext(ctx),
		client.WithStrKey(key),
		client.WithPrefix())

	resp, err := kv.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		return versions, nil
	}
	for _, kv := range resp.Kvs {
		key := GetInfoFromSvcIndexKV(kv.Key)
		versions = append(versions, key.Version)
	}
	return versions, nil
}

func getServiceDetailUtil(ctx context.Context, serviceDetailOpt ServiceDetailOpt) (*pb.ServiceDetail, error) {
	serviceID := serviceDetailOpt.service.ServiceId
	options := serviceDetailOpt.options
	domainProject := serviceDetailOpt.domainProject
	serviceDetail := new(pb.ServiceDetail)
	if serviceDetailOpt.countOnly {
		serviceDetail.Statics = new(pb.Statistics)
	}

	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, serviceID)
			if err != nil {
				log.Errorf(err, "get service[%s]'s all tags failed", serviceID)
				return nil, err
			}
			serviceDetail.Tags = tags
		case "rules":
			rules, err := serviceUtil.GetRulesUtil(ctx, domainProject, serviceID)
			if err != nil {
				log.Errorf(err, "get service[%s]'s all rules failed", serviceID)
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			if serviceDetailOpt.countOnly {
				instanceCount, err := serviceUtil.GetInstanceCountOfOneService(ctx, domainProject, serviceID)
				if err != nil {
					log.Errorf(err, "get number of service[%s]'s instances failed", serviceID)
					return nil, err
				}
				serviceDetail.Statics.Instances = &pb.StInstance{
					Count: instanceCount}
				continue
			}
			instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, domainProject, serviceID)
			if err != nil {
				log.Errorf(err, "get service[%s]'s all instances failed", serviceID)
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			schemas, err := getSchemaInfoUtil(ctx, domainProject, serviceID)
			if err != nil {
				log.Errorf(err, "get service[%s]'s all schemas failed", serviceID)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			service := serviceDetailOpt.service
			dr := serviceUtil.NewDependencyRelation(ctx, domainProject, service, service)
			consumers, err := dr.GetDependencyConsumers(
				serviceUtil.WithoutSelfDependency(),
				serviceUtil.WithSameDomainProject())
			if err != nil {
				log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s all consumers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version)
				return nil, err
			}
			providers, err := dr.GetDependencyProviders(
				serviceUtil.WithoutSelfDependency(),
				serviceUtil.WithSameDomainProject())
			if err != nil {
				log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s all providers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version)
				return nil, err
			}

			serviceDetail.Consumers = consumers
			serviceDetail.Providers = providers
		case "":
			continue
		default:
			log.Errorf(nil, "request option[%s] is invalid", opt)
		}
	}
	return serviceDetail, nil
}

func getSchemaInfoUtil(ctx context.Context, domainProject string, serviceID string) ([]*pb.Schema, error) {
	key := apt.GenerateServiceSchemaKey(domainProject, serviceID, "")

	resp, err := kv.Store().Schema().Search(ctx,
		client.WithStrKey(key),
		client.WithPrefix())
	if err != nil {
		log.Errorf(err, "get service[%s]'s schemas failed", serviceID)
		return make([]*pb.Schema, 0), err
	}
	schemas := make([]*pb.Schema, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		schemaInfo := &pb.Schema{}
		schemaInfo.Schema = util.BytesToStringWithNoCopy(kv.Value.([]byte))
		schemaInfo.SchemaId = util.BytesToStringWithNoCopy(kv.Key[len(key):])
		schemas = append(schemas, schemaInfo)
	}
	return schemas, nil
}

func statistics(ctx context.Context, withShared bool) (*pb.Statistics, error) {
	result := &pb.Statistics{
		Services:  &pb.StService{},
		Instances: &pb.StInstance{},
		Apps:      &pb.StApp{},
	}
	domainProject := util.ParseDomainProject(ctx)
	opts := serviceUtil.FromContext(ctx)

	// services
	key := apt.GetServiceIndexRootKey(domainProject) + "/"
	svcOpts := append(opts,
		client.WithStrKey(key),
		client.WithPrefix())
	respSvc, err := kv.Store().ServiceIndex().Search(ctx, svcOpts...)
	if err != nil {
		return nil, err
	}

	app := make(map[string]struct{}, respSvc.Count)
	svcWithNonVersion := make(map[string]struct{}, respSvc.Count)
	svcIDToNonVerKey := make(map[string]string, respSvc.Count)
	for _, kv := range respSvc.Kvs {
		key := apt.GetInfoFromSvcIndexKV(kv.Key)
		if !withShared && apt.IsShared(key) {
			continue
		}
		if _, ok := app[key.AppId]; !ok {
			app[key.AppId] = struct{}{}
		}

		key.Version = ""
		svcWithNonVersionKey := apt.GenerateServiceIndexKey(key)
		if _, ok := svcWithNonVersion[svcWithNonVersionKey]; !ok {
			svcWithNonVersion[svcWithNonVersionKey] = struct{}{}
		}
		svcIDToNonVerKey[kv.Value.(string)] = svcWithNonVersionKey
	}

	result.Services.Count = int64(len(svcWithNonVersion))
	result.Apps.Count = int64(len(app))

	respGetInstanceCountByDomain := make(chan GetInstanceCountByDomainResponse, 1)
	gopool.Go(func(_ context.Context) {
		getInstanceCountByDomain(ctx, svcIDToNonVerKey, respGetInstanceCountByDomain)
	})

	// instance
	key = apt.GetInstanceRootKey(domainProject) + "/"
	instOpts := append(opts,
		client.WithStrKey(key),
		client.WithPrefix(),
		client.WithKeyOnly())
	respIns, err := kv.Store().Instance().Search(ctx, instOpts...)
	if err != nil {
		return nil, err
	}

	onlineServices := make(map[string]struct{}, respSvc.Count)
	for _, kv := range respIns.Kvs {
		serviceID, _, _ := apt.GetInfoFromInstKV(kv.Key)
		key, ok := svcIDToNonVerKey[serviceID]
		if !ok {
			continue
		}
		result.Instances.Count++
		if _, ok := onlineServices[key]; !ok {
			onlineServices[key] = struct{}{}
		}
	}
	result.Services.OnlineCount = int64(len(onlineServices))

	data := <-respGetInstanceCountByDomain
	close(respGetInstanceCountByDomain)
	if data.err != nil {
		return nil, data.err
	}
	result.Instances.CountByDomain = data.countByDomain
	return result, nil
}

func getInstanceCountByDomain(ctx context.Context, svcIDToNonVerKey map[string]string, resp chan GetInstanceCountByDomainResponse) {
	domainID := util.ParseDomain(ctx)
	key := apt.GetInstanceRootKey(domainID) + "/"
	instOpts := append([]client.PluginOpOption{},
		client.WithStrKey(key),
		client.WithPrefix(),
		client.WithKeyOnly())
	respIns, err := kv.Store().Instance().Search(ctx, instOpts...)
	ret := GetInstanceCountByDomainResponse{
		err: err,
	}

	if err != nil {
		log.Errorf(err, "get number of instances by domain[%s]", domainID)
	} else {
		for _, kv := range respIns.Kvs {
			serviceID, _, _ := apt.GetInfoFromInstKV(kv.Key)
			_, ok := svcIDToNonVerKey[serviceID]
			if !ok {
				continue
			}
			ret.countByDomain++
		}
	}

	resp <- ret
}

// dep util
func toDependencyFilterOptions(in *pb.GetDependenciesRequest) (opts []serviceUtil.DependencyRelationFilterOption) {
	if in.SameDomain {
		opts = append(opts, serviceUtil.WithSameDomainProject())
	}
	if in.NoSelf {
		opts = append(opts, serviceUtil.WithoutSelfDependency())
	}
	return opts
}

func checkQuota(ctx context.Context, domainProject string) *quota.ApplyQuotaResult {
	if core.IsSCInstance(ctx) {
		log.Debugf("skip quota check")
		return nil
	}
	res := quota.NewApplyQuotaResource(quota.MicroServiceQuotaType, domainProject, "", 1)
	rst := quota.Apply(ctx, res)
	return rst
}

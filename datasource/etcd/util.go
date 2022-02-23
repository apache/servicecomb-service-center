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
	"fmt"
	"strings"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/foundation/gopool"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type ServiceDetailOpt struct {
	domainProject string
	service       *pb.MicroService
	countOnly     bool
	options       []string
}

// schema
func getSchemaSummary(ctx context.Context, domainProject string, serviceID string, schemaID string) (string, error) {
	key := path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schemaID)
	resp, err := sd.SchemaSummary().Search(ctx,
		etcdadpt.WithStrKey(key),
	)
	if err != nil {
		log.Error(fmt.Sprintf("get schema[%s/%s] summary failed", serviceID, schemaID), err)
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func getSchemasFromDatabase(ctx context.Context, domainProject string, serviceID string) ([]*pb.Schema, error) {
	key := path.GenerateServiceSchemaKey(domainProject, serviceID, "")
	resp, err := sd.Schema().Search(ctx,
		etcdadpt.WithPrefix(),
		etcdadpt.WithStrKey(key))
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s schema failed", serviceID), err)
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

func checkSchemaInfoExist(ctx context.Context, key string) (bool, error) {
	opts := append(serviceUtil.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithCountOnly())
	resp, errDo := sd.Schema().Search(ctx, opts...)
	if errDo != nil {
		return false, errDo
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func putSchema(ctx context.Context, domainProject string, serviceID string, schema *pb.Schema) ([]etcdadpt.OpOptions, error) {
	opts := make([]etcdadpt.OpOptions, 0)
	key := path.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	onPutOpt := etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithStrValue(schema.Schema))
	opts = append(opts, onPutOpt)
	syncOpts, err := sync.GenUpdateOpts(ctx, datasource.ResourceKV, schema.Schema, sync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		log.Error("fail to create update opts", err)
		return opts, err
	}
	opts = append(opts, syncOpts...)
	keySummary := path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schema.SchemaId)
	onPutOpt = etcdadpt.OpPut(etcdadpt.WithStrKey(keySummary), etcdadpt.WithStrValue(schema.Summary))
	opts = append(opts, onPutOpt)
	syncOpts, err = sync.GenUpdateOpts(ctx, datasource.ResourceKV, schema.Summary, sync.WithOpts(map[string]string{"key": keySummary}))
	if err != nil {
		log.Error("fail to create update opts", err)
		return opts, err
	}
	opts = append(opts, syncOpts...)
	return opts, nil
}

func deleteSchema(ctx context.Context, domainProject string, serviceID string, schema *pb.Schema) ([]etcdadpt.OpOptions, error) {
	opts := make([]etcdadpt.OpOptions, 0)
	key := path.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	onDelOpt := etcdadpt.OpDel(etcdadpt.WithStrKey(key), etcdadpt.WithStrValue(schema.Schema))
	opts = append(opts, onDelOpt)
	syncOpts, err := sync.GenDeleteOpts(ctx, datasource.ResourceKV, key, schema.Schema, sync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		log.Error("fail to create update opts", err)
		return opts, err
	}
	opts = append(opts, syncOpts...)
	keySummary := path.GenerateServiceSchemaSummaryKey(domainProject, serviceID, schema.SchemaId)
	onDelOpt = etcdadpt.OpDel(etcdadpt.WithStrKey(keySummary), etcdadpt.WithStrValue(schema.Summary))
	opts = append(opts, onDelOpt)
	syncOpts, err = sync.GenDeleteOpts(ctx, datasource.ResourceKV, keySummary, schema.Summary, sync.WithOpts(map[string]string{"key": keySummary}))
	if err != nil {
		log.Error("fail to create update opts", err)
		return opts, err
	}
	opts = append(opts, syncOpts...)
	return opts, nil
}

func isExistSchemaID(service *pb.MicroService, schemas []*pb.Schema) bool {
	serviceSchemaIds := service.Schemas
	for _, schema := range schemas {
		if !util.SliceHave(serviceSchemaIds, schema.SchemaId) {
			log.Error(fmt.Sprintf("schema[%s/%s] does not exist schemaID", service.ServiceId, schema.SchemaId), nil)
			return false
		}
	}
	return true
}

func commitSchemaInfo(ctx context.Context, domainProject string, serviceID string, schema *pb.Schema) ([]etcdadpt.OpOptions, error) {
	if len(schema.Summary) != 0 {
		opts, err := putSchema(ctx, domainProject, serviceID, schema)
		return opts, err
	}
	opts := make([]etcdadpt.OpOptions, 0)
	key := path.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)
	opt := etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithStrValue(schema.Schema))
	opts = append(opts, opt)
	syncOpts, err := sync.GenUpdateOpts(ctx, datasource.ResourceKV, schema.Schema, sync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		log.Error("fail to create update opts", err)
		return opts, err
	}
	opts = append(opts, syncOpts...)
	return opts, nil
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
			log.Error(fmt.Sprintf("heartbeat set failed, %s/%s", element.ServiceId, element.InstanceId), err)
		}
		instancesHbRst <- hbRst
	}
}

func revokeInstance(ctx context.Context, domainProject string, serviceID string, instanceID string) *errsvc.Error {
	leaseID, err := serviceUtil.GetLeaseID(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if leaseID == -1 {
		return pb.NewError(pb.ErrInstanceNotExists, "Instance's leaseId not exist.")
	}

	err = etcdadpt.Instance().LeaseRevoke(ctx, leaseID)
	if err != nil {
		if err == etcdadpt.ErrLeaseNotFound {
			return pb.NewError(pb.ErrInstanceNotExists, err.Error())
		}
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
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
				log.Error(fmt.Sprintf("get service[%s]'s all tags failed", serviceID), err)
				return nil, err
			}
			serviceDetail.Tags = tags
		case "instances":
			if serviceDetailOpt.countOnly {
				instanceCount, err := serviceUtil.GetInstanceCountOfOneService(ctx, domainProject, serviceID)
				if err != nil {
					log.Error(fmt.Sprintf("get number of service[%s]'s instances failed", serviceID), err)
					return nil, err
				}
				serviceDetail.Statics.Instances = &pb.StInstance{
					Count: instanceCount}
				continue
			}
			instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, domainProject, serviceID)
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all instances failed", serviceID), err)
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			schemas, err := getSchemaInfoUtil(ctx, domainProject, serviceID)
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all schemas failed", serviceID), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			service := serviceDetailOpt.service
			consumers, err := serviceUtil.GetConsumers(ctx, domainProject, service,
				serviceUtil.WithoutSelfDependency(),
				serviceUtil.WithSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all consumers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
				return nil, err
			}
			providers, err := serviceUtil.GetProviders(ctx, domainProject, service,
				serviceUtil.WithoutSelfDependency(),
				serviceUtil.WithSameDomainProject())
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s all providers failed",
					service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version), err)
				return nil, err
			}

			serviceDetail.Consumers = consumers
			serviceDetail.Providers = providers
		case "":
			continue
		default:
			log.Error(fmt.Sprintf("request option[%s] is invalid", opt), nil)
		}
	}
	return serviceDetail, nil
}

func getSchemaInfoUtil(ctx context.Context, domainProject string, serviceID string) ([]*pb.Schema, error) {
	key := path.GenerateServiceSchemaKey(domainProject, serviceID, "")

	resp, err := sd.Schema().Search(ctx,
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s schemas failed", serviceID), err)
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
	key := path.GetServiceIndexRootKey(domainProject) + "/"
	svcOpts := append(opts,
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	respSvc, err := sd.ServiceIndex().Search(ctx, svcOpts...)
	if err != nil {
		return nil, err
	}

	var svcIDs []string
	var svcKeys []*pb.MicroServiceKey
	for _, keyValue := range respSvc.Kvs {
		key := path.GetInfoFromSvcIndexKV(keyValue.Key)
		svcKeys = append(svcKeys, key)
		svcIDs = append(svcIDs, keyValue.Value.(string))
	}

	svcIDToNonVerKey := datasource.SetStaticServices(result, svcKeys, svcIDs, withShared)

	respGetInstanceCountByDomain := make(chan datasource.GetInstanceCountByDomainResponse, 1)
	gopool.Go(func(_ context.Context) {
		getInstanceCountByDomain(ctx, svcIDToNonVerKey, respGetInstanceCountByDomain)
	})

	// instance
	key = path.GetInstanceRootKey(domainProject) + "/"
	instOpts := append(opts,
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix(),
		etcdadpt.WithKeyOnly())
	respIns, err := sd.Instance().Search(ctx, instOpts...)
	if err != nil {
		return nil, err
	}

	var instIDs []string
	for _, keyValue := range respIns.Kvs {
		serviceID, _, _ := path.GetInfoFromInstKV(keyValue.Key)
		instIDs = append(instIDs, serviceID)
	}
	datasource.SetStaticInstances(result, svcIDToNonVerKey, instIDs)

	data := <-respGetInstanceCountByDomain
	close(respGetInstanceCountByDomain)
	if data.Err != nil {
		return nil, data.Err
	}
	result.Instances.CountByDomain = data.CountByDomain
	return result, nil
}

func getInstanceCountByDomain(ctx context.Context, svcIDToNonVerKey map[string]string, resp chan datasource.GetInstanceCountByDomainResponse) {
	domainID := util.ParseDomain(ctx)
	key := path.GetInstanceRootKey(domainID) + "/"
	instOpts := append(serviceUtil.FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix(),
		etcdadpt.WithKeyOnly())
	respIns, err := sd.Instance().Search(ctx, instOpts...)
	ret := datasource.GetInstanceCountByDomainResponse{
		Err: err,
	}

	if err != nil {
		log.Error(fmt.Sprintf("get number of instances by domain[%s]", domainID), err)
	} else {
		for _, keyValue := range respIns.Kvs {
			serviceID, _, _ := path.GetInfoFromInstKV(keyValue.Key)
			_, ok := svcIDToNonVerKey[serviceID]
			if !ok {
				continue
			}
			ret.CountByDomain++
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

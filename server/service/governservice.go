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
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/service/dependency"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strings"
)

type GovernServiceController struct {
}

func (governServiceController *GovernServiceController) GetServicesInfo(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	opts := in.Options
	//获取所有服务
	services, err := ms.GetAllServiceUtil(ctx)
	if err != nil {
		util.LOGGER.Errorf(err, "Get all services for govern service faild.")
		return &pb.GetServicesInfoResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get all service failed."),
		}, err
	}

	for _, opt := range opts {
		if opt == "all" {
			opts = []string{"tags", "rules", "instances", "schemas", "dependencies"}
			break
		}
	}
	allServiceDetails := []*pb.ServiceDetail{}
	tenant := util.ParseTenantProject(ctx)
	serviceId := ""
	for _, service := range services {
		serviceId = service.ServiceId
		serviceDetail, err := getServiceDetailUtil(ctx, opts, tenant, serviceId)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get one service detail failed."),
			}, err
		}
		serviceDetail.MicroSerivce = service
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}

	return &pb.GetServicesInfoResponse{
		Response:          pb.CreateResponse(pb.Response_SUCCESS, "register service instance successfully"),
		AllServicesDetail: allServiceDetails,
	}, nil
}

func (governServiceController *GovernServiceController) GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	tenant := util.ParseTenantProject(ctx)
	opts := []string{"tags", "rules", "instances", "schemas", "dependencies"}

	if len(in.ServiceId) == 0 {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid requtest for getting service detail."),
		}, nil
	}

	service, err := ms.GetServiceByServiceId(ctx, tenant, in.ServiceId)
	if service == nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service is not exist."),
		}, nil
	}
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service failed."),
		}, err
	}

	versions, err := getServiceAllVersions(ctx, tenant, service.AppId, service.ServiceName)
	if err != nil {
		util.LOGGER.Errorf(err, "Get service all version fialed.")
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service detail failed."),
		}, err
	}

	serviceInfo, err := getServiceDetailUtil(ctx, opts, tenant, in.ServiceId)
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service detail failed."),
		}, err
	}

	serviceInfo.MicroSerivce = service
	serviceInfo.MicroServiceVersions = versions
	return &pb.GetServiceDetailResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service successful."),
		Service:  serviceInfo,
	}, nil
}

func getServiceAllVersions(ctx context.Context, tenant string, appId string, serviceName string) ([]string, error) {
	versions := []string{}
	key := apt.GenerateServiceIndexKey(&pb.MicroServiceKey{
		Tenant:      tenant,
		AppId:       appId,
		ServiceName: serviceName,
		Version:     "",
	})

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		return versions, nil
	}
	version := ""
	for _, kvs := range resp.Kvs {
		tmpArr := strings.Split(string(kvs.Key), "/")
		version = tmpArr[len(tmpArr)-1]
		versions = append(versions, version)
	}
	return versions, nil
}

func getAllInstancesForOneService(ctx context.Context, tenant string, serviceId string) ([]*pb.MicroServiceInstance, error) {
	key := apt.GenerateInstanceKey(tenant, serviceId, "")

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get one service's instance failed from data source.")
		return nil, err
	}
	instances := []*pb.MicroServiceInstance{}
	for _, kvs := range resp.Kvs {
		util.LOGGER.Debugf("start unmarshal service instance file: %s", string(kvs.Key))
		instance := &pb.MicroServiceInstance{}
		err := json.Unmarshal(kvs.Value, instance)
		if err != nil {
			util.LOGGER.Errorf(err, "Unmarshal service instance failed.")
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func getSchemaInfoUtil(ctx context.Context, tenant string, serviceId string) ([]*pb.SchemaInfos, error) {
	key := apt.GenerateServiceSchemaKey(tenant, serviceId, "")
	schemas := []*pb.SchemaInfos{}
	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "Get schema failded,%s")
		return schemas, err
	}
	schemaId := ""
	schema := ""
	for _, kv := range resp.Kvs {
		schemaInfo := &pb.SchemaInfos{}
		schemaId = string(kv.Key[len(key):])
		schema = string(kv.Value)
		schemaInfo.Schema = schema
		schemaInfo.SchemaId = schemaId
		schemas = append(schemas, schemaInfo)
	}
	return schemas, nil
}

func getServiceDetailUtil(ctx context.Context, opts []string, tenant string, serviceId string) (*pb.ServiceDetail, error) {
	serviceDetail := &pb.ServiceDetail{}
	for _, opt := range opts {
		expr := opt
		switch expr {
		case "tags":
			util.LOGGER.Debugf("is tags")
			tags, err := serviceUtil.GetTagsUtils(ctx, tenant, serviceId)
			if err != nil {
				util.LOGGER.Errorf(err, "Get all tags for govern service faild.")
				return nil, err
			}
			serviceDetail.Tags = tags
		case "rules":
			util.LOGGER.Debugf("is rules")
			rules, err := serviceUtil.GetRulesUtil(ctx, tenant, serviceId)
			if err != nil {
				util.LOGGER.Errorf(err, "Get all rules for govern service faild.")
				return nil, err
			}
			serviceDetail.Rules = rules
		case "instances":
			util.LOGGER.Debugf("is instances")
			instances, err := getAllInstancesForOneService(ctx, tenant, serviceId)
			if err != nil {
				util.LOGGER.Errorf(err, "Get service's all instances for govern service faild.")
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			util.LOGGER.Debugf("is schemas")
			schemas, err := getSchemaInfoUtil(ctx, tenant, serviceId)
			if err != nil {
				util.LOGGER.Errorf(err, "Get service's all schemas for govern service faild.")
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			util.LOGGER.Debugf("is dependencies")
			keyProDependency := apt.GenerateProviderDependencyKey(tenant, serviceId, "")
			consumers, err := dependency.GetDependencies(ctx, keyProDependency, tenant)
			if err != nil {
				util.LOGGER.Errorf(err, "Get service's all consumers for govern service faild.")
				return nil, err
			}
			consumers = deleteSelfDenpendency(consumers, serviceId)
			keyConDependency := apt.GenerateConsumerDependencyKey(tenant, serviceId, "")
			providers, err := dependency.GetDependencies(ctx, keyConDependency, tenant)
			if err != nil {
				util.LOGGER.Errorf(err, "Get service's all providers for govern service faild.")
				return nil, err
			}
			providers = deleteSelfDenpendency(providers, serviceId)
			serviceDetail.Consumers = consumers
			serviceDetail.Providers = providers
		case "":
			continue
		default:
			util.LOGGER.Errorf(nil, "option %s from request is invalid.", opt)
		}
	}
	return serviceDetail, nil
}

func deleteSelfDenpendency(services []*pb.MicroService, serviceId string) []*pb.MicroService {
	for key, service := range services {
		if service.ServiceId == serviceId {
			services = append(services[:key], services[key+1:]...)
		}
	}
	return services
}

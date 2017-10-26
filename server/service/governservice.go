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
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

type GovernServiceController struct {
}

func (governServiceController *GovernServiceController) GetServicesInfo(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	opts := serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))

	optionMap := make(map[string]struct{}, len(in.Options))
	for _, opt := range in.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = []string{"tags", "rules", "instances", "schemas", "dependencies"}
	} else {
		for opt := range optionMap {
			options = append(options, opt)
		}
	}

	var st *pb.Statistics
	if _, ok := optionMap["statistics"]; ok {
		var err error
		st, err = statistics(ctx, opts...)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Statistics failed."),
			}, err
		}
		if len(optionMap) == 1 {
			return &pb.GetServicesInfoResponse{
				Response:   pb.CreateResponse(pb.Response_SUCCESS, "Statistics successfully."),
				Statistics: st,
			}, nil
		}
	}

	//获取所有服务
	services, err := serviceUtil.GetAllServiceUtil(ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "Get all services for govern service faild.")
		return &pb.GetServicesInfoResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get all service failed."),
		}, err
	}

	allServiceDetails := []*pb.ServiceDetail{}
	tenant := util.ParseTenantProject(ctx)
	for _, service := range services {
		if apt.Service.ServiceId == service.ServiceId {
			continue
		}
		serviceDetail, err := getServiceDetailUtil(ctx, options, tenant, service.ServiceId, service, opts...)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "Get one service detail failed."),
			}, err
		}
		serviceDetail.MicroSerivce = service
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}

	return &pb.GetServicesInfoResponse{
		Response:          pb.CreateResponse(pb.Response_SUCCESS, "Get services info successfully."),
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func (governServiceController *GovernServiceController) GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	tenant := util.ParseTenantProject(ctx)
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}

	if len(in.ServiceId) == 0 {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid requtest for getting service detail."),
		}, nil
	}

	opts := serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))

	service, err := serviceUtil.GetService(ctx, tenant, in.ServiceId, opts...)
	if service == nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service failed."),
		}, err
	}

	versions, err := getServiceAllVersions(ctx, tenant, service.AppId, service.ServiceName, opts...)
	if err != nil {
		util.Logger().Errorf(err, "Get service all version fialed.")
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get service detail failed."),
		}, err
	}

	serviceInfo, err := getServiceDetailUtil(ctx, options, tenant, in.ServiceId, service, opts...)
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

func getServiceAllVersions(ctx context.Context, tenant string, appId string, serviceName string, opts ...registry.PluginOpOption) ([]string, error) {
	versions := []string{}
	key := apt.GenerateServiceIndexKey(&pb.MicroServiceKey{
		Tenant:      tenant,
		AppId:       appId,
		ServiceName: serviceName,
		Version:     "",
	})

	opts = append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix())

	resp, err := store.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		return versions, nil
	}
	version := ""
	for _, kvs := range resp.Kvs {
		tmpArr := strings.Split(util.BytesToStringWithNoCopy(kvs.Key), "/")
		version = tmpArr[len(tmpArr)-1]
		versions = append(versions, version)
	}
	return versions, nil
}

func getSchemaInfoUtil(ctx context.Context, tenant string, serviceId string) ([]*pb.Schema, error) {
	key := apt.GenerateServiceSchemaKey(tenant, serviceId, "")
	schemas := []*pb.Schema{}
	resp, err := store.Store().Schema().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		util.Logger().Errorf(err, "Get schema failded,%s")
		return schemas, err
	}
	for _, kv := range resp.Kvs {
		schemaInfo := &pb.Schema{}
		schemaInfo.Schema = util.BytesToStringWithNoCopy(kv.Value)
		schemaInfo.SchemaId = util.BytesToStringWithNoCopy(kv.Key[len(key):])
		schemas = append(schemas, schemaInfo)
	}
	return schemas, nil
}

func getServiceDetailUtil(ctx context.Context, options []string, tenant string, serviceId string, service *pb.MicroService, opts ...registry.PluginOpOption) (*pb.ServiceDetail, error) {
	serviceDetail := &pb.ServiceDetail{}
	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			util.Logger().Debugf("is tags")
			tags, err := serviceUtil.GetTagsUtils(ctx, tenant, serviceId, opts...)
			if err != nil {
				util.Logger().Errorf(err, "Get all tags for govern service faild.")
				return nil, err
			}
			serviceDetail.Tags = tags
		case "rules":
			util.Logger().Debugf("is rules")
			rules, err := serviceUtil.GetRulesUtil(ctx, tenant, serviceId, opts...)
			if err != nil {
				util.Logger().Errorf(err, "Get all rules for govern service faild.")
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			util.Logger().Debugf("is instances")
			instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, tenant, serviceId, "", opts...)
			if err != nil {
				util.Logger().Errorf(err, "Get service's all instances for govern service faild.")
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			util.Logger().Debugf("is schemas")
			schemas, err := getSchemaInfoUtil(ctx, tenant, serviceId)
			if err != nil {
				util.Logger().Errorf(err, "Get service's all schemas for govern service faild.")
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			util.Logger().Debugf("is dependencies")
			dr := serviceUtil.NewDependencyRelation(ctx, tenant, serviceId, service, serviceId, service, opts...)
			consumers, err := dr.GetDependencyConsumers()
			if err != nil {
				util.Logger().Errorf(err, "Get service's all consumers for govern service faild.")
				return nil, err
			}
			consumers = skipSelfDependency(consumers, serviceId)

			providers, err := dr.GetDependencyProviders()
			if err != nil {
				util.Logger().Errorf(err, "Get service's all providers for govern service faild.")
				return nil, err
			}
			providers = skipSelfDependency(providers, serviceId)
			serviceDetail.Consumers = consumers
			serviceDetail.Providers = providers
		case "":
			continue
		default:
			util.Logger().Errorf(nil, "option %s from request is invalid.", opt)
		}
	}
	return serviceDetail, nil
}

func skipSelfDependency(services []*pb.MicroService, serviceId string) []*pb.MicroService {
	for key, service := range services {
		if service.ServiceId == serviceId {
			services = append(services[:key], services[key+1:]...)
		}
	}
	return services
}

func statistics(ctx context.Context, opts ...registry.PluginOpOption) (*pb.Statistics, error) {
	result := &pb.Statistics{
		Services:  &pb.StService{},
		Instances: &pb.StInstance{},
		Apps:      &pb.StApp{},
	}
	tenantProject := util.ParseTenantProject(ctx)

	// services
	key := apt.GetServiceIndexRootKey(tenantProject)
	svcOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := store.Store().Service().Search(ctx, svcOpts...)
	if err != nil {
		return nil, err
	}
	if resp.Count > 0 {
		result.Services.Count = resp.Count
		if apt.IsDefaultDomain(tenantProject) {
			result.Services.Count = result.Services.Count - 1
		}
	}

	app := map[string]interface{}{}
	for _, kv := range resp.Kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		keySpilted := strings.Split(key, "/")
		if len(keySpilted) > 6 {
			appId := keySpilted[6]
			app[appId] = nil
		}
	}
	result.Apps.Count = int32(len(app))

	// instance
	key = apt.GetInstanceRootKey(tenantProject)
	instOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix())
	respIns, err := store.Store().Instance().Search(ctx, instOpts...)
	if err != nil {
		return nil, err
	}
	result.Instances.Count = respIns.Count
	key = apt.GenerateInstanceKey(tenantProject, apt.Service.ServiceId, "")
	scOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	resp, err = store.Store().Instance().Search(ctx, scOpts...)
	if err != nil {
		return nil, err
	}
	result.Instances.Count = result.Instances.Count - resp.Count

	onlineServices := map[string] interface{}{}
	for _, kv := range respIns.Kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		keySpilted := strings.Split(key, "/")
		if len(keySpilted) > 6 {
			servieId := keySpilted[6]
			onlineServices[servieId] = nil
		}
	}
	result.Services.OnlineCount = int32(len(onlineServices))
	return result, err
}

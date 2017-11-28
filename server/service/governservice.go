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
	"github.com/ServiceComb/service-center/server/core/backend/store"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/infra/registry"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
)

type GovernServiceController struct {
}

func (governServiceController *GovernServiceController) GetServicesInfo(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
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
		st, err = statistics(ctx)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Statistics failed."),
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
	services, err := serviceUtil.GetAllServiceUtil(ctx)
	if err != nil {
		util.Logger().Errorf(err, "Get all services for govern service faild.")
		return &pb.GetServicesInfoResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get all service failed."),
		}, err
	}

	allServiceDetails := []*pb.ServiceDetail{}
	domainProject := util.ParseDomainProject(ctx)
	for _, service := range services {
		if apt.Service.ServiceId == service.ServiceId {
			continue
		}
		if len(in.AppId) > 0 {
			if in.AppId != service.AppId {
				continue
			}
			if len(in.ServiceName) > 0 && in.ServiceName != service.ServiceName {
				continue
			}
		}

		serviceDetail, err := getServiceDetailUtil(ctx, options, domainProject, service.ServiceId, service)
		if err != nil {
			return &pb.GetServicesInfoResponse{
				Response: pb.CreateResponse(scerr.ErrInternal, "Get one service detail failed."),
			}, err
		}
		serviceDetail.MicroService = service
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}

	return &pb.GetServicesInfoResponse{
		Response:          pb.CreateResponse(pb.Response_SUCCESS, "Get services info successfully."),
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func (governServiceController *GovernServiceController) GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	options := []string{"tags", "rules", "instances", "schemas", "dependencies"}

	if len(in.ServiceId) == 0 {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid requtest for getting service detail."),
		}, nil
	}

	service, err := serviceUtil.GetService(ctx, domainProject, in.ServiceId)
	if service == nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get service failed."),
		}, err
	}

	key := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     "",
	}
	versions, err := getServiceAllVersions(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "Get service all version fialed.")
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get all versions of the service failed."),
		}, err
	}

	serviceInfo, err := getServiceDetailUtil(ctx, options, domainProject, in.ServiceId, service)
	if err != nil {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Get service detail failed."),
		}, err
	}

	serviceInfo.MicroService = service
	serviceInfo.MicroServiceVersions = versions
	return &pb.GetServiceDetailResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service successfully."),
		Service:  serviceInfo,
	}, nil
}

func (governServiceController *GovernServiceController) GetApplications(ctx context.Context, in *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	err := apt.Validate(in)
	if err != nil {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	key := util.StringJoin([]string{
		apt.GetServiceIndexRootKey(domainProject),
		in.Environment,
	}, "/")
	if key[len(key)-1:] != "/" {
		key += "/"
	}

	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithKeyOnly())

	resp, err := store.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	l := len(resp.Kvs)
	if l == 0 {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(pb.Response_SUCCESS, "Get all applications successfully."),
		}, nil
	}

	apps := make([]string, 0, l)
	appMap := make(map[string]struct{}, l)
	for _, kv := range resp.Kvs {
		key, _ := pb.GetInfoFromSvcIndexKV(kv)
		if _, ok := appMap[key.AppId]; ok {
			continue
		}
		if apt.IsSCKey(key) {
			continue
		}
		appMap[key.AppId] = struct{}{}
		apps = append(apps, key.AppId)
	}

	return &pb.GetAppsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get all applications successfully."),
		AppIds:   apps,
	}, nil
}

func getServiceAllVersions(ctx context.Context, serviceKey *pb.MicroServiceKey) ([]string, error) {
	versions := []string{}
	key := apt.GenerateServiceIndexKey(serviceKey)

	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())

	resp, err := store.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Kvs) == 0 {
		return versions, nil
	}
	for _, kv := range resp.Kvs {
		key, _ := pb.GetInfoFromSvcIndexKV(kv)
		versions = append(versions, key.Version)
	}
	return versions, nil
}

func getSchemaInfoUtil(ctx context.Context, domainProject string, serviceId string) ([]*pb.Schema, error) {
	key := apt.GenerateServiceSchemaKey(domainProject, serviceId, "")
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

func getServiceDetailUtil(ctx context.Context, options []string, domainProject string, serviceId string, service *pb.MicroService) (*pb.ServiceDetail, error) {
	serviceDetail := &pb.ServiceDetail{}
	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, serviceId)
			if err != nil {
				util.Logger().Errorf(err, "Get all tags for govern service faild.")
				return nil, err
			}
			serviceDetail.Tags = tags
		case "rules":
			rules, err := serviceUtil.GetRulesUtil(ctx, domainProject, serviceId)
			if err != nil {
				util.Logger().Errorf(err, "Get all rules for govern service faild.")
				return nil, err
			}
			for _, rule := range rules {
				rule.Timestamp = rule.ModTimestamp
			}
			serviceDetail.Rules = rules
		case "instances":
			instances, err := serviceUtil.GetAllInstancesOfOneService(ctx, domainProject, serviceId)
			if err != nil {
				util.Logger().Errorf(err, "Get service's all instances for govern service faild.")
				return nil, err
			}
			serviceDetail.Instances = instances
		case "schemas":
			schemas, err := getSchemaInfoUtil(ctx, domainProject, serviceId)
			if err != nil {
				util.Logger().Errorf(err, "Get service's all schemas for govern service faild.")
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			dr := serviceUtil.NewDependencyRelation(ctx, domainProject, serviceId, service, serviceId, service)
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

func statistics(ctx context.Context) (*pb.Statistics, error) {
	result := &pb.Statistics{
		Services:  &pb.StService{},
		Instances: &pb.StInstance{},
		Apps:      &pb.StApp{},
	}
	domainProject := util.ParseDomainProject(ctx)
	opts := serviceUtil.FromContext(ctx)

	// services
	key := apt.GetServiceIndexRootKey(domainProject)
	svcOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithKeyOnly())
	respSvc, err := store.Store().ServiceIndex().Search(ctx, svcOpts...)
	if err != nil {
		return nil, err
	}

	app := make(map[string]interface{}, respSvc.Count)
	scSvc := make(map[string]interface{}, respSvc.Count)
	for _, kv := range respSvc.Kvs {
		key, _ := pb.GetInfoFromSvcIndexKV(kv)
		if _, ok := app[key.AppId]; !ok {
			if !apt.IsSCKey(key) {
				app[key.AppId] = nil
			}
		}

		if apt.IsSCKey(key) {
			k := util.BytesToStringWithNoCopy(kv.Key)
			if _, ok := scSvc[k]; !ok {
				scSvc[k] = nil
			}
		}
	}
	result.Services.Count = respSvc.Count - int64(len(scSvc))
	result.Apps.Count = int64(len(app))

	// instance
	key = apt.GetInstanceRootKey(domainProject)
	instOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithKeyOnly())
	respIns, err := store.Store().Instance().Search(ctx, instOpts...)
	if err != nil {
		return nil, err
	}

	onlineServices := make(map[string]interface{}, respSvc.Count)
	for _, kv := range respIns.Kvs {
		serviceId, _, _, _ := pb.GetInfoFromInstKV(kv)
		if _, ok := onlineServices[serviceId]; !ok {
			onlineServices[serviceId] = nil
		}
	}

	key = apt.GenerateInstanceKey(domainProject, apt.Service.ServiceId, "")
	scOpts := append(opts,
		registry.WithStrKey(key),
		registry.WithPrefix(),
		registry.WithCountOnly())
	respScIns, err := store.Store().Instance().Search(ctx, scOpts...)
	if err != nil {
		return nil, err
	}

	result.Instances.Count = respIns.Count - respScIns.Count
	result.Services.OnlineCount = removeSCSelf(domainProject, int64(len(onlineServices)), 1)
	return result, err
}

func removeSCSelf(domainProject string, count int64, removeNum int64) int64 {
	if count > 0 {
		if apt.IsDefaultDomainProject(domainProject) {
			count = count - removeNum
		}
	}
	return count
}

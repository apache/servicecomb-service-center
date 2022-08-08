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

package govern

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/jinzhu/copier"
)

var defaultOptions = []string{"tags", "instances", "schemas", "dependencies"}

type ServiceDetailOpt struct {
	domainProject string
	service       *pb.MicroService
	countOnly     bool
	options       []string
}

func ListServiceDetail(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	ctx = util.WithCacheOnly(ctx)

	optionMap := make(map[string]struct{}, len(in.Options))
	for _, opt := range in.Options {
		optionMap[opt] = struct{}{}
	}

	options := make([]string, 0, len(optionMap))
	if _, ok := optionMap["all"]; ok {
		optionMap["statistics"] = struct{}{}
		options = defaultOptions
	} else {
		for opt := range optionMap {
			options = append(options, opt)
		}
	}

	var st *pb.Statistics
	if _, ok := optionMap["statistics"]; ok {
		var err error
		st, err = datasource.GetMetadataManager().Statistics(ctx, in.WithShared)
		if err != nil {
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
		if len(optionMap) == 1 {
			return &pb.GetServicesInfoResponse{
				Statistics: st,
			}, nil
		}
	}

	//获取所有服务
	resp, err := discosvc.ListService(ctx, &pb.GetServicesRequest{})
	if err != nil {
		log.Error("get all services by domain failed", err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	services := resp.Services

	allServiceDetails := make([]*pb.ServiceDetail, 0, len(services))
	domainProject := util.ParseDomainProject(ctx)
	instanceProperties := config.GetStringMap("registry.instance.properties")
	for _, service := range services {
		if !filterServices(domainProject, in, service) {
			continue
		}

		serviceDetail, err := getServiceDetailUtil(ctx, ServiceDetailOpt{
			domainProject: domainProject,
			service:       service,
			countOnly:     in.CountOnly,
			options:       options,
		})
		if err != nil {
			return nil, pb.NewError(pb.ErrInternal, err.Error())
		}
		serviceDetail.MicroService = service
		tmpServiceDetail, err := NewServiceOverview(serviceDetail, instanceProperties)
		if err != nil {
			return nil, err
		}
		allServiceDetails = append(allServiceDetails, tmpServiceDetail)
	}

	return &pb.GetServicesInfoResponse{
		AllServicesDetail: allServiceDetails,
		Statistics:        st,
	}, nil
}

func getServiceDetailUtil(ctx context.Context, opts ServiceDetailOpt) (*pb.ServiceDetail, error) {
	service := opts.service
	serviceID := service.ServiceId
	serviceLogName := fmt.Sprintf("%s][%s/%s/%s/%s", service.ServiceId, service.Environment, service.AppId, service.ServiceName, service.Version)
	options := opts.options
	serviceDetail := new(pb.ServiceDetail)
	if opts.countOnly {
		serviceDetail.Statics = new(pb.Statistics)
	}

	for _, opt := range options {
		expr := opt
		switch expr {
		case "tags":
			resp, err := discosvc.ListTag(ctx, &pb.GetServiceTagsRequest{
				ServiceId: serviceID,
			})
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all tags failed", serviceLogName), err)
				return nil, err
			}
			serviceDetail.Tags = resp.Tags
		case "instances":
			resp, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{
				ProviderServiceId: serviceID,
			})
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all instances failed", serviceLogName), err)
				return nil, err
			}
			if opts.countOnly {
				if err != nil {
					log.Error(fmt.Sprintf("get number of service[%s]'s instances failed", serviceLogName), err)
					return nil, err
				}
				serviceDetail.Statics.Instances = &pb.StInstance{
					Count: int64(len(resp.Instances)),
				}
				continue
			}
			serviceDetail.Instances = resp.Instances
		case "schemas":
			schemas, err := discosvc.ListSchema(ctx, &pb.GetAllSchemaRequest{
				ServiceId:  serviceID,
				WithSchema: true,
			})
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all schemas failed", serviceLogName), err)
				return nil, err
			}
			serviceDetail.SchemaInfos = schemas
		case "dependencies":
			consumerResp, err := discosvc.ListConsumers(ctx, &pb.GetDependenciesRequest{
				ServiceId:  serviceID,
				NoSelf:     true,
				SameDomain: true,
			})
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all consumers failed", serviceLogName), err)
				return nil, err
			}
			providerResp, err := discosvc.ListProviders(ctx, &pb.GetDependenciesRequest{
				ServiceId:  serviceID,
				NoSelf:     true,
				SameDomain: true,
			})
			if err != nil {
				log.Error(fmt.Sprintf("get service[%s]'s all providers failed", serviceLogName), err)
				return nil, err
			}
			serviceDetail.Consumers = consumerResp.Consumers
			serviceDetail.Providers = providerResp.Providers
		case "":
			continue
		default:
			log.Error(fmt.Sprintf("request option[%s] is invalid", opt), nil)
		}
	}
	return serviceDetail, nil
}

func filterServices(domainProject string, request *pb.GetServicesInfoRequest, service *pb.MicroService) bool {
	if !request.WithShared && datasource.IsGlobal(pb.MicroServiceToKey(domainProject, service)) {
		return false
	}
	if len(request.Environment) > 0 && request.Environment != service.Environment {
		return false
	}
	if len(request.AppId) > 0 && request.AppId != service.AppId {
		return false
	}
	if len(request.ServiceName) > 0 && request.ServiceName != service.ServiceName {
		return false
	}
	if len(request.Properties) > 0 && !matchAllProperties(request.Properties, service) {
		return false
	}
	return true
}

func matchAllProperties(properties map[string]string, service *pb.MicroService) bool {
	for k, v := range properties {
		val, ok := service.Properties[k]
		if !ok || v != val {
			return false
		}
	}
	return true
}

func NewServiceOverview(serviceDetail *pb.ServiceDetail, innerProperties map[string]string) (*pb.ServiceDetail, error) {
	tmpServiceDetail := &pb.ServiceDetail{}
	err := copier.CopyWithOption(tmpServiceDetail, serviceDetail, copier.Option{DeepCopy: true})
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	tmpServiceDetail.MicroService.Schemas = nil
	instances := tmpServiceDetail.Instances
	for _, instance := range instances {
		instance.Properties = removeCustomProperties(instance.Properties, innerProperties)
	}
	return tmpServiceDetail, nil
}

func removeCustomProperties(properties, innerProperties map[string]string) map[string]string {
	if len(innerProperties) == 0 {
		return nil
	}
	props := make(map[string]string)
	for k, v := range properties {
		if _, ok := innerProperties[k]; ok {
			props[k] = v
		}
	}
	return props
}

func GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.ServiceDetail, error) {
	ctx = util.WithCacheOnly(ctx)

	serviceID := in.ServiceId
	if len(serviceID) == 0 {
		return nil, pb.NewError(pb.ErrInvalidParams, "Invalid request for getting service detail.")
	}

	service, err := discosvc.GetService(ctx, in)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed", serviceID), err)
		return nil, err
	}

	serviceInfo := new(pb.ServiceDetail)
	serviceInfo.MicroService = service

	key := &pb.MicroServiceKey{
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
	}
	versions, err := getServiceAllVersions(ctx, key)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s/%s/%s] all versions failed",
			service.Environment, service.AppId, service.ServiceName), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	serviceInfo.MicroServiceVersions = versions

	tagsResp, err := discosvc.ListTag(ctx, &pb.GetServiceTagsRequest{ServiceId: serviceID})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] tags failed", serviceID), err)
		return nil, err
	}
	serviceInfo.Tags = tagsResp.Tags

	schemas, err := discosvc.ListSchema(ctx, &pb.GetAllSchemaRequest{ServiceId: serviceID, WithSchema: true})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] schemas failed", serviceID), err)
		return nil, err
	}
	serviceInfo.SchemaInfos = schemas

	providerResp, err := discosvc.ListProviders(ctx, &pb.GetDependenciesRequest{ServiceId: serviceID})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] providers failed", serviceID), err)
		return nil, err
	}
	serviceInfo.Providers = providerResp.Providers

	consumerResp, err := discosvc.ListConsumers(ctx, &pb.GetDependenciesRequest{ServiceId: serviceID})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] consumers failed", serviceID), err)
		return nil, err
	}
	serviceInfo.Consumers = consumerResp.Consumers

	instResp, err := discosvc.ListInstance(ctx, &pb.GetInstancesRequest{ProviderServiceId: serviceID})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] instances failed", serviceID), err)
		return nil, err
	}
	serviceInfo.Instances = instResp.Instances

	return serviceInfo, nil
}

func getServiceAllVersions(ctx context.Context, key *pb.MicroServiceKey) ([]string, error) {
	resp, err := discosvc.FindService(ctx, key)
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(resp.Services))
	for _, svc := range resp.Services {
		versions = append(versions, svc.Version)
	}
	return versions, nil
}

func ListApp(ctx context.Context, in *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	if err := validator.ValidateGetAppsRequest(in); err != nil {
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().ListApp(ctx, in)
}

func GetOverview(ctx context.Context, in *pb.GetServicesRequest) (*pb.Statistics, error) {
	ctx = util.WithCacheOnly(ctx)
	return datasource.GetMetadataManager().GetOverview(ctx, in)
}

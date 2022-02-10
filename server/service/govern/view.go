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
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

func ListServiceDetail(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	ctx = util.WithCacheOnly(ctx)
	return datasource.GetMetadataManager().ListServiceDetail(ctx, in)
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

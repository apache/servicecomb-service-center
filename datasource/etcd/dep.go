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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/event"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/little-cui/etcdadpt"
)

type DepManager struct {
}

func (dm *DepManager) SearchProviderDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	provider, err := serviceUtil.GetService(ctx, domainProject, providerServiceID)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider[%s] does not exist in db", providerServiceID))
			return &pb.GetProDependenciesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, "Provider does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("query provider service from db failed, provider is %s", providerServiceID), err)
		return nil, err
	}

	services, err := serviceUtil.GetConsumers(ctx, domainProject, provider, toDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("query provider failed, provider is %s/%s/%s/%s",
			provider.Environment, provider.AppId, provider.ServiceName, provider.Version), err)
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetProDependenciesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Get all consumers successful."),
		Consumers: services,
	}, nil
}

func (dm *DepManager) SearchConsumerDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	consumerID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)
	consumer, err := serviceUtil.GetService(ctx, domainProject, consumerID)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("consumer[%s] does not exist", consumerID))
			return &pb.GetConDependenciesResponse{
				Response: pb.CreateResponse(pb.ErrServiceNotExists, "Consumer does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("query consumer failed, consumer is %s", consumerID), err)
		return nil, err
	}

	services, err := serviceUtil.GetProviders(ctx, domainProject, consumer, toDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("query consumer failed, consumer is %s/%s/%s/%s",
			consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version), err)
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(pb.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetConDependenciesResponse{
		Response:  pb.CreateResponse(pb.ResponseSuccess, "Get all providers successfully."),
		Providers: services,
	}, nil
}

func (dm *DepManager) DependencyHandle(ctx context.Context) error {
	var dep *event.DependencyEventHandler
	err := dep.Handle()
	if err != nil {
		return err
	}
	for {
		key := path.GetServiceDependencyQueueRootKey("")
		resp, err := sd.DependencyQueue().Search(ctx,
			etcdadpt.WithStrKey(key), etcdadpt.WithPrefix(), etcdadpt.WithCountOnly())
		if err != nil {
			return err
		}
		// maintain dependency rules.
		if resp.Count == 0 {
			break
		}
	}
	return nil
}

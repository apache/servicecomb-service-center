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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/event"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
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

	dr := serviceUtil.NewProviderDependencyRelation(ctx, domainProject, provider)
	services, err := dr.GetDependencyConsumers(toDependencyFilterOptions(request)...)

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

	dr := serviceUtil.NewConsumerDependencyRelation(ctx, domainProject, consumer)
	services, err := dr.GetDependencyProviders(toDependencyFilterOptions(request)...)
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

func (dm *DepManager) DeleteDependency() {
	panic("implement me")
}

func (dm *DepManager) DependencyHandle(ctx context.Context) error {
	var dep *event.DependencyEventHandler
	err := dep.Handle()
	if err != nil {
		return err
	}
	for {
		key := path.GetServiceDependencyQueueRootKey("")
		resp, err := kv.Store().DependencyQueue().Search(ctx,
			client.WithStrKey(key), client.WithPrefix(), client.WithCountOnly())
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

func (dm *DepManager) AddOrUpdateDependencies(ctx context.Context, dependencyInfos []*pb.ConsumerDependency, override bool) (*pb.Response, error) {
	opts := make([]client.PluginOp, 0, len(dependencyInfos))
	domainProject := util.ParseDomainProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerFlag := util.StringJoin([]string{dependencyInfo.Consumer.Environment, dependencyInfo.Consumer.AppId, dependencyInfo.Consumer.ServiceName, dependencyInfo.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		rsp := serviceUtil.ParamsChecker(consumerInfo, providersInfo)
		if rsp != nil {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer is %s, %s",
				override, consumerFlag, rsp.Response.GetMessage())
			return rsp.Response, nil
		}

		consumerID, err := serviceUtil.GetServiceID(ctx, consumerInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, get consumer[%s] id failed",
				override, consumerFlag)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}
		if len(consumerID) == 0 {
			log.Errorf(nil, "put request into dependency queue failed, override: %t, consumer[%s] does not exist",
				override, consumerFlag)
			return pb.CreateResponse(pb.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag)), nil
		}

		dependencyInfo.Override = override
		data, err := json.Marshal(dependencyInfo)
		if err != nil {
			log.Errorf(err, "put request into dependency queue failed, override: %t, marshal consumer[%s] dependency failed",
				override, consumerFlag)
			return pb.CreateResponse(pb.ErrInternal, err.Error()), err
		}

		id := path.DepsQueueUUID
		if !override {
			id = util.GenerateUUID()
		}
		key := path.GenerateConsumerDependencyQueueKey(domainProject, consumerID, id)
		opts = append(opts, client.OpPut(client.WithStrKey(key), client.WithValue(data)))
	}

	err := client.BatchCommit(ctx, opts)
	if err != nil {
		log.Errorf(err, "put request into dependency queue failed, override: %t, %v", override, dependencyInfos)
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}

	log.Infof("put request into dependency queue successfully, override: %t, %v, from remote %s",
		override, dependencyInfos, util.GetIPFromContext(ctx))
	return pb.CreateResponse(pb.ResponseSuccess, "Create dependency successfully."), nil
}

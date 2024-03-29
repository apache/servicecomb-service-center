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

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/event"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	eutil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type DepManager struct {
}

func (dm *DepManager) ListConsumers(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	domainProject := util.ParseDomainProject(ctx)
	providerServiceID := request.ServiceId
	provider, err := eutil.GetService(ctx, domainProject, providerServiceID)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("provider[%s] does not exist in db", providerServiceID))
			return nil, pb.NewError(pb.ErrServiceNotExists, "Provider does not exist")
		}
		log.Error(fmt.Sprintf("query provider service from db failed, provider is %s", providerServiceID), err)
		return nil, err
	}

	services, err := eutil.GetConsumers(ctx, domainProject, provider, toDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("query provider failed, provider is %s/%s/%s/%s",
			provider.Environment, provider.AppId, provider.ServiceName, provider.Version), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.GetProDependenciesResponse{
		Consumers: services,
	}, nil
}

func (dm *DepManager) ListProviders(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	consumerID := request.ServiceId
	domainProject := util.ParseDomainProject(ctx)
	consumer, err := eutil.GetService(ctx, domainProject, consumerID)

	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("consumer[%s] does not exist", consumerID))
			return nil, pb.NewError(pb.ErrServiceNotExists, "Consumer does not exist")
		}
		log.Error(fmt.Sprintf("query consumer failed, consumer is %s", consumerID), err)
		return nil, err
	}

	services, err := eutil.GetProviders(ctx, domainProject, consumer, toDependencyFilterOptions(request)...)
	if err != nil {
		log.Error(fmt.Sprintf("query consumer failed, consumer is %s/%s/%s/%s",
			consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version), err)
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.GetConDependenciesResponse{
		Providers: services,
	}, nil
}

func (dm *DepManager) DependencyHandle(ctx context.Context) error {
	var dep *event.DependencyEventHandler
	err := dep.Handle()
	if err != nil {
		return err
	}

	key := path.GetServiceDependencyQueueRootKey("")
	resp, err := sd.DependencyQueue().Search(ctx,
		etcdadpt.WithStrKey(key), etcdadpt.WithPrefix(), etcdadpt.WithCountOnly())
	if err != nil {
		return err
	}
	// maintain dependency rules.
	if resp.Count != 0 {
		return fmt.Errorf("residual records[%d]", resp.Count)
	}
	return nil
}

func (dm *DepManager) PutDependencies(ctx context.Context, dependencyInfos []*pb.ConsumerDependency, override bool) error {
	opts := make([]etcdadpt.OpOptions, 0, len(dependencyInfos))
	domainProject := util.ParseDomainProject(ctx)
	for _, dependencyInfo := range dependencyInfos {
		consumerFlag := util.StringJoin([]string{dependencyInfo.Consumer.Environment, dependencyInfo.Consumer.AppId, dependencyInfo.Consumer.ServiceName, dependencyInfo.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{dependencyInfo.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(dependencyInfo.Providers, domainProject)

		if err := datasource.ParamsChecker(consumerInfo, providersInfo); err != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, consumer is %s",
				override, consumerFlag), err)
			return err
		}

		consumerID, err := eutil.GetServiceID(ctx, consumerInfo)
		if err != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, get consumer[%s] id failed",
				override, consumerFlag), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
		if len(consumerID) == 0 {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, consumer[%s] does not exist",
				override, consumerFlag), nil)
			return pb.NewError(pb.ErrServiceNotExists, fmt.Sprintf("Consumer %s does not exist.", consumerFlag))
		}

		dependencyInfo.Override = override
		data, err := json.Marshal(dependencyInfo)
		if err != nil {
			log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, marshal consumer[%s] dependency failed",
				override, consumerFlag), err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}

		id := path.DepsQueueUUID
		if !override {
			id = util.GenerateUUID()
		}
		key := path.GenerateConsumerDependencyQueueKey(domainProject, consumerID, id)
		opts = append(opts, etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)))
		syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceKV, data, esync.WithOpts(map[string]string{"key": key}))
		if err != nil {
			log.Error("fail to create sync opts", err)
			return pb.NewError(pb.ErrInternal, err.Error())
		}
		opts = append(opts, syncOpts...)
	}

	err := etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error(fmt.Sprintf("put request into dependency queue failed, override: %t, %v",
			override, dependencyInfos), err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	log.Info(fmt.Sprintf("put request into dependency queue successfully, override: %t, %v, from remote %s",
		override, dependencyInfos, util.GetIPFromContext(ctx)))
	return nil
}

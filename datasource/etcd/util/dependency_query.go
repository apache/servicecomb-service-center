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

package util

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// DependencyRelationFilterOpt contains SameDomainProject and NonSelf flag
type DependencyRelationFilterOpt struct {
	SameDomainProject bool
	NonSelf           bool
}

type DependencyRelationFilterOption func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt

func WithSameDomainProject() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.SameDomainProject = true
		return opt
	}
}
func WithoutSelfDependency() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.NonSelf = true
		return opt
	}
}

func ToDependencyRelationFilterOpt(opts ...DependencyRelationFilterOption) (op DependencyRelationFilterOpt) {
	for _, opt := range opts {
		op = opt(op)
	}
	return
}

type DependencyRelation struct {
	ctx           context.Context
	domainProject string
	consumer      *pb.MicroService
	provider      *pb.MicroService
}

func (dr *DependencyRelation) GetDependencyProviders(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	services := make([]*pb.MicroService, 0, len(keys))
	op := ToDependencyRelationFilterOpt(opts...)
	for _, key := range keys {
		if op.SameDomainProject && key.Tenant != dr.domainProject {
			continue
		}

		providerIDs, err := dr.parseDependencyRule(key)
		if err != nil {
			return nil, err
		}

		for _, providerID := range providerIDs {
			provider, err := GetService(dr.ctx, key.Tenant, providerID)
			if err != nil {
				if errors.Is(err, datasource.ErrNoData) {
					log.Warn(fmt.Sprintf("provider[%s/%s/%s/%s] does not exist",
						key.Environment, key.AppId, key.ServiceName, key.Version))
				} else {
					log.Warn(fmt.Sprintf("get provider[%s/%s/%s/%s] failed",
						key.Environment, key.AppId, key.ServiceName, key.Version))
				}
				continue
			}
			if op.NonSelf && providerID == dr.consumer.ServiceId {
				continue
			}
			services = append(services, provider)
		}
	}
	return services, nil
}

func (dr *DependencyRelation) getDependencyProviderIds() ([]string, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	return dr.GetProviderIdsByRules(keys)
}

func (dr *DependencyRelation) getProviderKeys() ([]*pb.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, errors.New("invalid consumer")
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)

	conKey := path.GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)
	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, conKey)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) GetProviderIdsByRules(providerRules []*pb.MicroServiceKey) ([]string, error) {
	provideServiceIds := make([]string, 0, len(providerRules))
	for _, provider := range providerRules {
		serviceIDs, err := dr.parseDependencyRule(provider)
		if err != nil {
			log.Errorf(err, "get service[%s/%s/%s/%s]'s providerIDs failed",
				provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
			return provideServiceIds, err
		}
		if len(serviceIDs) == 0 {
			log.Warnf("get service[%s/%s/%s/%s]'s providerIDs is empty",
				provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
			continue
		}
		provideServiceIds = append(provideServiceIds, serviceIDs...)
	}
	return provideServiceIds, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *pb.MicroServiceKey) (serviceIDs []string, err error) {
	serviceIDs, _, err = FindServiceIds(dr.ctx, dependencyRule, false)
	return
}

func (dr *DependencyRelation) GetDependencyConsumers(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	consumerDependAllList, err := dr.GetDependencyConsumersOfProvider()
	if err != nil {
		log.Errorf(err, "get service[%s]'s consumers failed", dr.provider.ServiceId)
		return nil, err
	}
	consumers := make([]*pb.MicroService, 0)
	op := ToDependencyRelationFilterOpt(opts...)
	for _, consumer := range consumerDependAllList {
		if op.SameDomainProject && consumer.Tenant != dr.domainProject {
			continue
		}

		service, err := dr.GetServiceByMicroServiceKey(consumer)
		if err != nil {
			if errors.Is(err, datasource.ErrNoData) {
				log.Warn(fmt.Sprintf("consumer[%s/%s/%s/%s] does not exist",
					consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version))
				continue
			} else {
				return nil, err
			}
		}

		if op.NonSelf && service.ServiceId == dr.provider.ServiceId {
			continue
		}

		consumers = append(consumers, service)
	}
	return consumers, nil
}

func (dr *DependencyRelation) GetServiceByMicroServiceKey(service *pb.MicroServiceKey) (*pb.MicroService, error) {
	serviceID, err := GetServiceID(dr.ctx, service)
	if err != nil {
		return nil, err
	}
	if len(serviceID) == 0 {
		log.Warn(fmt.Sprintf("service[%s/%s/%s/%s] not exist",
			service.Environment, service.AppId, service.ServiceName, service.Version))
		return nil, datasource.ErrNoData
	}
	return GetService(dr.ctx, service.Tenant, serviceID)
}

func (dr *DependencyRelation) getDependencyConsumerIds() ([]string, error) {
	consumerDependAllList, err := dr.GetDependencyConsumersOfProvider()
	if err != nil {
		return nil, err
	}
	consumerIDs := make([]string, 0, len(consumerDependAllList))
	for _, consumer := range consumerDependAllList {
		consumerID, err := GetServiceID(dr.ctx, consumer)
		if err != nil {
			log.Errorf(err, "get consumer[%s/%s/%s/%s] failed",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
			return nil, err
		}
		if len(consumerID) == 0 {
			log.Warnf("get consumer[%s/%s/%s/%s] not exist",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
			continue
		}
		consumerIDs = append(consumerIDs, consumerID)
	}
	return consumerIDs, nil

}

func (dr *DependencyRelation) GetDependencyConsumersOfProvider() ([]*pb.MicroServiceKey, error) {
	if dr.provider == nil {
		return nil, errors.New("invalid provider")
	}
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependList, err := dr.GetConsumerOfSameServiceNameAndAppID(providerService)
	if err != nil {
		log.Errorf(err, "get consumers that depend on rule[%s/%s/%s/%s] failed",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return nil, err
	}
	return consumerDependList, nil
}

func (dr *DependencyRelation) GetConsumerOfSameServiceNameAndAppID(provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	copy := *provider
	copy.Version = ""
	prefix := path.GenerateProviderDependencyRuleKey(dr.domainProject, &copy)

	opts := append(FromContext(dr.ctx),
		client.WithStrKey(prefix),
		client.WithPrefix())
	rsp, err := kv.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service[%s/%s/%s]'s dependency rules failed",
			provider.Environment, provider.AppId, provider.ServiceName)
		return nil, err
	}

	allConsumers := make([]*pb.MicroServiceKey, 0, len(rsp.Kvs))
	for _, kv := range rsp.Kvs {
		allConsumers = append(allConsumers, kv.Value.(*pb.MicroServiceDependency).Dependency...)
	}
	return allConsumers, nil
}

func NewProviderDependencyRelation(ctx context.Context, domainProject string, provider *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, nil, provider)
}

func NewConsumerDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, consumer, nil)
}

func NewDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroService) *DependencyRelation {
	return &DependencyRelation{
		ctx:           ctx,
		domainProject: domainProject,
		consumer:      consumer,
		provider:      provider,
	}
}

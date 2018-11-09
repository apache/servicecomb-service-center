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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"golang.org/x/net/context"
	"strings"
)

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

func toDependencyRelationFilterOpt(opts ...DependencyRelationFilterOption) (op DependencyRelationFilterOpt) {
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
	op := toDependencyRelationFilterOpt(opts...)
	for _, key := range keys {
		if op.SameDomainProject && key.Tenant != dr.domainProject {
			continue
		}

		providerIds, err := dr.parseDependencyRule(key)
		if err != nil {
			return nil, err
		}

		if key.ServiceName == "*" {
			services = services[:0]
		}

		for _, providerId := range providerIds {
			provider, err := GetService(dr.ctx, key.Tenant, providerId)
			if err != nil {
				log.Warnf("get provider[%s/%s/%s/%s] failed",
					key.Environment, key.AppId, key.ServiceName, key.Version)
				continue
			}
			if provider == nil {
				log.Warnf("provider[%s/%s/%s/%s] does not exist",
					key.Environment, key.AppId, key.ServiceName, key.Version)
				continue
			}
			if op.NonSelf && providerId == dr.consumer.ServiceId {
				continue
			}
			services = append(services, provider)
		}

		if key.ServiceName == "*" {
			break
		}
	}
	return services, nil
}

func (dr *DependencyRelation) GetDependencyProviderIds() ([]string, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	return dr.getDependencyProviderIds(keys)
}

func (dr *DependencyRelation) getProviderKeys() ([]*pb.MicroServiceKey, error) {
	if dr.consumer == nil {
		return nil, fmt.Errorf("Invalid consumer")
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)

	conKey := apt.GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)
	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, conKey)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) getDependencyProviderIds(providerRules []*pb.MicroServiceKey) ([]string, error) {
	provideServiceIds := make([]string, 0, len(providerRules))
	for _, provider := range providerRules {
		serviceIds, err := dr.parseDependencyRule(provider)
		switch {
		case provider.ServiceName == "*":
			if err != nil {
				log.Errorf(err, "get all serviceIds failed")
				return provideServiceIds, err
			}
			return serviceIds, nil
		default:
			if err != nil {
				log.Errorf(err, "get service[%s/%s/%s/%s]'s providerIds failed",
					provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
				return provideServiceIds, err
			}
			if len(serviceIds) == 0 {
				log.Warnf("get service[%s/%s/%s/%s]'s providerIds is empty",
					provider.Environment, provider.AppId, provider.ServiceName, provider.Version)
				continue
			}
			provideServiceIds = append(provideServiceIds, serviceIds...)
		}
	}
	return provideServiceIds, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *pb.MicroServiceKey) (serviceIds []string, err error) {
	opts := FromContext(dr.ctx)
	switch {
	case dependencyRule.ServiceName == "*":
		log.Infof("service[%s/%s/%s/%s] rely all service",
			dr.consumer.Environment, dr.consumer.AppId, dr.consumer.ServiceName, dr.consumer.Version)
		splited := strings.Split(apt.GenerateServiceIndexKey(dependencyRule), "/")
		allServiceKey := util.StringJoin(splited[:len(splited)-3], "/") + "/"
		sopts := append(opts,
			registry.WithStrKey(allServiceKey),
			registry.WithPrefix())
		resp, err := backend.Store().ServiceIndex().Search(dr.ctx, sopts...)
		if err != nil {
			return nil, err
		}

		for _, kv := range resp.Kvs {
			serviceIds = append(serviceIds, kv.Value.(string))
		}
	default:
		serviceIds, _, err = FindServiceIds(dr.ctx, dependencyRule.Version, dependencyRule)
	}
	return
}

func (dr *DependencyRelation) GetDependencyConsumers(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		log.Errorf(err, "get service[%s]'s consumers failed", dr.provider.ServiceId)
		return nil, err
	}
	consumers := make([]*pb.MicroService, 0)
	op := toDependencyRelationFilterOpt(opts...)
	for _, consumer := range consumerDependAllList {
		if op.SameDomainProject && consumer.Tenant != dr.domainProject {
			continue
		}

		service, err := dr.getServiceByMicroServiceKey(consumer)
		if err != nil {
			return nil, err
		}
		if service == nil {
			log.Warnf("consumer[%s/%s/%s/%s] does not exist",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
			continue
		}

		if op.NonSelf && service.ServiceId == dr.provider.ServiceId {
			continue
		}

		consumers = append(consumers, service)
	}
	return consumers, nil
}

func (dr *DependencyRelation) getServiceByMicroServiceKey(service *pb.MicroServiceKey) (*pb.MicroService, error) {
	serviceId, err := GetServiceId(dr.ctx, service)
	if err != nil {
		return nil, err
	}
	if len(serviceId) == 0 {
		log.Warnf("service[%s/%s/%s/%s] not exist",
			service.Environment, service.AppId, service.ServiceName, service.Version)
		return nil, nil
	}
	return GetService(dr.ctx, service.Tenant, serviceId)
}

func (dr *DependencyRelation) GetDependencyConsumerIds() ([]string, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		return nil, err
	}
	consumerIds := make([]string, 0, len(consumerDependAllList))
	for _, consumer := range consumerDependAllList {
		consumerId, err := GetServiceId(dr.ctx, consumer)
		if err != nil {
			log.Errorf(err, "get consumer[%s/%s/%s/%s] failed",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
			return nil, err
		}
		if len(consumerId) == 0 {
			log.Warnf("get consumer[%s/%s/%s/%s] not exist",
				consumer.Environment, consumer.AppId, consumer.ServiceName, consumer.Version)
			continue
		}
		consumerIds = append(consumerIds, consumerId)
	}
	return consumerIds, nil

}

func (dr *DependencyRelation) getDependencyConsumersOfProvider() ([]*pb.MicroServiceKey, error) {
	if dr.provider == nil {
		return nil, fmt.Errorf("Invalid provider")
	}
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		log.Errorf(err, "get consumers that depend on all services failed, %s", dr.provider.ServiceId)
		return nil, err
	}

	consumerDependList, err := dr.getConsumerOfSameServiceNameAndAppId(providerService)
	if err != nil {
		log.Errorf(err, "get consumers that depend on rule[%s/%s/%s/%s] failed",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return nil, err
	}
	consumerDependAllList = append(consumerDependAllList, consumerDependList...)
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*pb.MicroServiceKey, error) {
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	relyAllKey := apt.GenerateProviderDependencyRuleKey(dr.domainProject, providerService)
	opts := append(FromContext(dr.ctx), registry.WithStrKey(relyAllKey))
	rsp, err := backend.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		log.Errorf(err, "get consumers that rely all service failed, %s/%s/%s/%s",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return nil, err
	}
	if len(rsp.Kvs) != 0 {
		log.Infof("exist consumers that rely all service, %s/%s/%s/%s",
			dr.provider.Environment, dr.provider.AppId, dr.provider.ServiceName, dr.provider.Version)
		return rsp.Kvs[0].Value.(*pb.MicroServiceDependency).Dependency, nil
	}
	return nil, nil
}

func (dr *DependencyRelation) getConsumerOfSameServiceNameAndAppId(provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	prefix := apt.GenerateProviderDependencyRuleKey(dr.domainProject, provider)
	provider.Version = providerVersion

	opts := append(FromContext(dr.ctx),
		registry.WithStrKey(prefix),
		registry.WithPrefix())
	rsp, err := backend.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service[%s/%s/%s]'s dependency rules failed",
			provider.Environment, provider.AppId, provider.ServiceName)
		return nil, err
	}

	allConsumers := make([]*pb.MicroServiceKey, 0, len(rsp.Kvs))
	var latestServiceId []string

	for _, kv := range rsp.Kvs {
		providerVersionRuleArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		providerVersionRule := providerVersionRuleArr[len(providerVersionRuleArr)-1]
		if providerVersionRule == "latest" {
			if latestServiceId == nil {
				latestServiceId, _, err = FindServiceIds(dr.ctx, providerVersionRule, provider)
				if err != nil {
					log.Errorf(err, "get service[%s/%s/%s/%s]'s serviceId failed",
						provider.Environment, provider.AppId, provider.ServiceName, providerVersionRule)
					return nil, err
				}
			}
			if len(latestServiceId) == 0 {
				log.Infof("service[%s/%s/%s/%s] does not exist",
					provider.Environment, provider.AppId, provider.ServiceName, providerVersionRule)
				continue
			}
			if dr.provider.ServiceId != latestServiceId[0] {
				continue
			}

		} else {
			if !VersionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}

		log.Debugf("providerETCD is %s", providerVersionRuleArr)
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

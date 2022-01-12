/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cache

import (
	"context"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

func GetProviderServiceOfDeps(provider *discovery.MicroService) (*discovery.MicroServiceDependency, bool) {
	res := sd.Store().Dep().Cache().GetValue(genDepServiceKey(datasource.Provider, provider))
	deps, ok := transCacheToDep(res)
	if !ok {
		return nil, false
	}
	return deps[0], true
}

func transCacheToDep(cache []interface{}) ([]*discovery.MicroServiceDependency, bool) {
	res := make([]*discovery.MicroServiceDependency, 0, len(cache))
	for _, v := range cache {
		t, ok := v.(model.DependencyRule)
		if !ok {
			return nil, false
		}
		res = append(res, t.Dep)
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func genDepServiceKey(ruleType string, service *discovery.MicroService) string {
	return strings.Join([]string{ruleType, service.AppId, service.ServiceName, service.Version}, "/")
}

func GetMicroServiceInstancesByID(ctx context.Context, serviceID string) ([]*discovery.MicroServiceInstance, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	index := genServiceIDIndex(ctx, serviceID)
	cacheInstances := sd.Store().Instance().Cache().GetValue(index)
	insts, ok := transCacheToMicroInsts(cacheInstances)
	if !ok {
		return nil, false
	}
	return insts, true
}

func GetServiceByID(ctx context.Context, serviceID string) (*model.Service, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	index := genServiceIDIndex(ctx, serviceID)
	cacheRes := sd.Store().Service().Cache().GetValue(index)
	if len(cacheRes) == 0 {
		return nil, false
	}
	res, ok := transCacheToService(cacheRes)
	if !ok {
		return nil, false
	}
	return res[0], true
}

func GetServiceByName(ctx context.Context, key *discovery.MicroServiceKey) ([]*model.Service, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	cacheIndex := genServiceNameIndex(ctx, key)
	res := sd.Store().Service().Cache().GetValue(cacheIndex)
	cacheService, ok := transCacheToService(res)
	if !ok {
		return nil, false
	}
	return cacheService, true
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (serviceID string, exist bool) {
	if util.NoCache(ctx) {
		return
	}
	cacheIndex := genServiceKeyIndex(ctx, key)
	res := sd.Store().Service().Cache().GetValue(cacheIndex)
	cacheService, ok := transCacheToService(res)
	if !ok {
		return
	}
	return cacheService[0].Service.ServiceId, true
}

func GetServiceByIDAcrossDomain(ctx context.Context, serviceID string) (*model.Service, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	index := genServiceIDIndexAcrossDomain(ctx, serviceID)
	cacheRes := sd.Store().Service().Cache().GetValue(index)

	if len(cacheRes) == 0 {
		return nil, false
	}

	res, ok := transCacheToService(cacheRes)
	if !ok {
		return nil, false
	}

	return res[0], true
}

func GetServicesByDomainProject(domainProject string) (service []*model.Service, exist bool) {
	services := make([]*model.Service, 0)
	sd.Store().Service().Cache().GetValue(domainProject)
	if len(services) == 0 {
		return services, false
	}
	return services, true
}

func GetMicroServicesByDomainProject(domainProject string) (service []*discovery.MicroService, exist bool) {
	services, exist := GetServicesByDomainProject(domainProject)
	if !exist || len(services) == 0 {
		return nil, false
	}
	ms := make([]*discovery.MicroService, len(services))
	for i, s := range services {
		ms[i] = s.Service
	}
	return ms, true
}

func transCacheToService(services []interface{}) ([]*model.Service, bool) {
	res := make([]*model.Service, 0, len(services))
	for _, v := range services {
		t, ok := v.(model.Service)
		if !ok {
			return nil, false
		}
		res = append(res, &model.Service{
			Domain:  t.Domain,
			Project: t.Project,
			Tags:    t.Tags,
			Service: t.Service,
		})
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func genServiceIDIndexAcrossDomain(ctx context.Context, serviceID string) string {
	return strings.Join([]string{util.ParseTargetDomainProject(ctx), serviceID}, datasource.SPLIT)
}

func genServiceIDIndex(ctx context.Context, serviceID string) string {
	return strings.Join([]string{util.ParseDomainProject(ctx), serviceID}, datasource.SPLIT)
}

func genServiceKeyIndex(ctx context.Context, key *discovery.MicroServiceKey) string {
	return strings.Join([]string{util.ParseDomain(ctx), util.ParseProject(ctx), key.AppId, key.ServiceName, key.Version}, datasource.SPLIT)
}

func genServiceNameIndex(ctx context.Context, key *discovery.MicroServiceKey) string {
	return strings.Join([]string{util.ParseDomain(ctx), util.ParseProject(ctx), key.AppId, key.ServiceName}, datasource.SPLIT)
}

func CountInstances(ctx context.Context, serviceID string) (int, bool) {
	if util.NoCache(ctx) {
		return 0, false
	}
	index := genServiceIDIndex(ctx, serviceID)
	cacheInstances := sd.Store().Instance().Cache().GetValue(index)
	if len(cacheInstances) == 0 {
		return 0, false
	}
	return len(cacheInstances), true
}

func GetInstance(ctx context.Context, serviceID string, instanceID string) (*model.Instance, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	index := generateInstanceIDIndex(util.ParseDomainProject(ctx), serviceID, instanceID)
	cacheInstance := sd.Store().Instance().Cache().GetValue(index)
	insts, ok := transCacheToInsts(cacheInstance)
	if !ok {
		return nil, false
	}
	return insts[0], true
}

func GetInstances(ctx context.Context) ([]*model.Instance, bool) {
	if util.NoCache(ctx) {
		return nil, false
	}
	index := util.ParseDomainProject(ctx)
	cacheInstance := sd.Store().Instance().Cache().GetValue(index)
	insts, ok := transCacheToInsts(cacheInstance)
	if !ok {
		return nil, false
	}
	return insts, true
}

func transCacheToMicroInsts(cache []interface{}) ([]*discovery.MicroServiceInstance, bool) {
	res := make([]*discovery.MicroServiceInstance, 0, len(cache))
	for _, iter := range cache {
		inst, ok := iter.(model.Instance)
		if !ok {
			return nil, false
		}
		res = append(res, inst.Instance)
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func transCacheToInsts(cache []interface{}) ([]*model.Instance, bool) {
	res := make([]*model.Instance, 0, len(cache))
	for _, iter := range cache {
		inst, ok := iter.(model.Instance)
		if !ok {
			return nil, false
		}
		res = append(res, &inst)
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func generateInstanceIDIndex(domainProject string, serviceID string, instanceID string) string {
	return util.StringJoin([]string{
		domainProject,
		serviceID,
		instanceID,
	}, datasource.SPLIT)
}

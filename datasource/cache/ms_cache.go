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
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"strings"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	Provider = "p"
)

func GetProviderServiceOfDeps(provider *discovery.MicroService) (*discovery.MicroServiceDependency, bool) {
	res := sd.Store().Dep().Cache().GetValue(genDepserivceKey(Provider, provider))
	deps, ok := transCacheToDep(res)
	if !ok {
		return nil, false
	}
	return deps[0], true
}

func transCacheToDep(cache []interface{}) ([]*discovery.MicroServiceDependency, bool) {
	res := make([]*discovery.MicroServiceDependency, 0, len(cache))
	for _, v := range cache {
		t, ok := v.(dao.DependencyRule)
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

func genDepserivceKey(ruleType string, service *discovery.MicroService) string {
	return strings.Join([]string{ruleType, service.AppId, service.ServiceName, service.Version}, "/")
}

func GetMicroServiceInstancesByID(serviceID string) ([]*discovery.MicroServiceInstance, bool) {
	cacheInstances := sd.Store().Instance().Cache().GetValue(serviceID)
	insts, ok := transCacheToInsts(cacheInstances)
	if !ok {
		return nil, false
	}
	return insts, true
}

func transCacheToInsts(cache []interface{}) ([]*discovery.MicroServiceInstance, bool) {
	res := make([]*discovery.MicroServiceInstance, 0, len(cache))
	for _, iter := range cache {
		inst, ok := iter.(dao.Instance)
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

func GetRulesByServiceID(serviceID string) ([]*dao.Rule, bool) {
	cacheRes := sd.Store().Rule().Cache().GetValue(serviceID)
	return transCacheToRules(cacheRes)
}

func transCacheToRules(cacheRules []interface{}) ([]*dao.Rule, bool) {
	res := make([]*dao.Rule, 0, len(cacheRules))
	for _, v := range cacheRules {
		t, ok := v.(dao.Rule)
		if !ok {
			return nil, false
		}
		res = append(res, &dao.Rule{
			Domain:    t.Domain,
			Project:   t.Project,
			ServiceID: t.ServiceID,
			Rule:      t.Rule,
		})
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func GetServiceByID(serviceID string) (*dao.Service, bool) {
	cacheRes := sd.Store().Service().Cache().GetValue(serviceID)
	res, ok := transCacheToService(cacheRes)
	if !ok {
		return nil, false
	}
	return res[0], true
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (serviceID string, exist bool) {
	cacheIndex := strings.Join([]string{util.ParseDomain(ctx), util.ParseProject(ctx), key.AppId, key.ServiceName, key.Version}, "/")
	res := sd.Store().Service().Cache().GetValue(cacheIndex)
	cacheService, ok := transCacheToService(res)
	if !ok {
		return
	}
	return cacheService[0].Service.ServiceId, true
}

func transCacheToService(services []interface{}) ([]*dao.Service, bool) {
	res := make([]*dao.Service, 0, len(services))
	for _, v := range services {
		t, ok := v.(dao.Service)
		if !ok {
			return nil, false
		}
		res = append(res, &dao.Service{
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

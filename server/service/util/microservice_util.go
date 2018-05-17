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
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
)

/*
	get Service by service id
*/
func GetServiceWithRev(ctx context.Context, domain string, id string, rev int64) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domain, id)
	serviceResp, err := backend.Store().Service().Search(ctx,
		registry.WithStrKey(key),
		registry.WithRev(rev))
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, nil
	}
	service := &pb.MicroService{}
	err = json.Unmarshal(serviceResp.Kvs[0].Value, &service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func GetServiceInCache(ctx context.Context, domain string, id string) (*pb.MicroService, error) {
	return GetService(ctx, domain, id)
}

func GetService(ctx context.Context, domainProject string, serviceId string) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domainProject, serviceId)
	opts := append(FromContext(ctx), registry.WithStrKey(key))
	serviceResp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, nil
	}
	service := &pb.MicroService{}
	err = json.Unmarshal(serviceResp.Kvs[0].Value, &service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func GetServicesRawData(ctx context.Context, domainProject string) ([]*mvccpb.KeyValue, error) {
	key := apt.GenerateServiceKey(domainProject, "")
	opts := append(FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, err
}

func GetServicesByDomain(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	kvs, err := GetServicesRawData(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	services := []*pb.MicroService{}
	for _, kvs := range kvs {
		service := &pb.MicroService{}
		err := json.Unmarshal(kvs.Value, service)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

func GetServiceId(ctx context.Context, key *pb.MicroServiceKey) (serviceId string, err error) {
	serviceId, err = searchServiceId(ctx, key)
	if err != nil {
		return
	}
	if len(serviceId) == 0 {
		// 别名查询
		util.Logger().Debugf("could not search microservice %s/%s/%s id by field 'serviceName', now try field 'alias'.",
			key.AppId, key.ServiceName, key.Version)
		return searchServiceIdFromAlias(ctx, key)
	}
	return
}

func searchServiceId(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(FromContext(ctx), registry.WithStrKey(apt.GenerateServiceIndexKey(key)))
	resp, err := backend.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return util.BytesToStringWithNoCopy(resp.Kvs[0].Value), nil
}

func searchServiceIdFromAlias(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(FromContext(ctx), registry.WithStrKey(apt.GenerateServiceAliasKey(key)))
	resp, err := backend.Store().ServiceAlias().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return util.BytesToStringWithNoCopy(resp.Kvs[0].Value), nil
}

func GetServiceAllVersions(ctx context.Context, key *pb.MicroServiceKey, alias bool) (*registry.PluginResponse, error) {
	key.Version = ""
	var prefix string
	if alias {
		prefix = apt.GenerateServiceAliasKey(key)
	} else {
		prefix = apt.GenerateServiceIndexKey(key)
	}
	opts := append(FromContext(ctx),
		registry.WithStrKey(prefix),
		registry.WithPrefix(),
		registry.WithDescendOrder())
	resp, err := backend.Store().ServiceIndex().Search(ctx, opts...)
	return resp, err
}

func FindServiceIds(ctx context.Context, versionRule string, key *pb.MicroServiceKey) ([]string, error) {
	// 版本规则
	ids := []string{}
	match := ParseVersionRule(versionRule)
	if match == nil {
		key.Version = versionRule
		serviceId, err := GetServiceId(ctx, key)
		if err != nil {
			return nil, err
		}
		if len(serviceId) > 0 {
			ids = append(ids, serviceId)
		}
		return ids, nil
	}

	searchAlias := false
	alsoFindAlias := len(key.Alias) > 0

FIND_RULE:
	resp, err := GetServiceAllVersions(ctx, key, searchAlias)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) > 0 {
		ids = match(resp.Kvs)
	}
	if len(ids) == 0 && alsoFindAlias {
		searchAlias = true
		alsoFindAlias = false
		goto FIND_RULE
	}
	return ids, nil
}

func ServiceExist(ctx context.Context, domainProject string, serviceId string) bool {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateServiceKey(domainProject, serviceId)),
		registry.WithCountOnly())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetAllServiceUtil(ctx context.Context) ([]*pb.MicroService, error) {
	domainProject := util.ParseDomainProject(ctx)
	services, err := GetServicesByDomain(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func RemandServiceQuota(ctx context.Context) {
	plugin.Plugins().Quota().RemandQuotas(ctx, quota.MicroServiceQuotaType)
}

func RemandInstanceQuota(ctx context.Context) {
	plugin.Plugins().Quota().RemandQuotas(ctx, quota.MicroServiceInstanceQuotaType)
}

func UpdateService(domainProject string, serviceId string, service *pb.MicroService) (opt registry.PluginOp, err error) {
	opt = registry.PluginOp{}
	key := apt.GenerateServiceKey(domainProject, serviceId)
	data, err := json.Marshal(service)
	if err != nil {
		util.Logger().Errorf(err, "marshal service failed.")
		return
	}
	opt = registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))
	return
}

func GetOneDomainProjectServiceCount(ctx context.Context, domainProject string) (int64, error) {
	key := apt.GenerateServiceKey(domainProject, "")
	opts := append(FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithCountOnly(),
		registry.WithPrefix())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func GetOneDomainProjectInstanceCount(ctx context.Context, domainProject string) (int64, error) {
	key := apt.GetInstanceRootKey(domainProject) + "/"
	opts := append(FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithCountOnly(),
		registry.WithPrefix())
	resp, err := backend.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func SetServiceDefaultValue(service *pb.MicroService) {
	if len(service.Level) == 0 {
		service.Level = "BACK"
	}
	if len(service.Status) == 0 {
		service.Status = pb.MS_UP
	}
}

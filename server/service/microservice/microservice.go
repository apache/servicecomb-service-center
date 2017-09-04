//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package microservice

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/pkg/common/cache"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"time"
)

var msCache *cache.Cache

func MsCache() *cache.Cache {
	return msCache
}

func init() {
	d, _ := time.ParseDuration("1m")
	msCache = cache.New(d, d)
}

/*
	get Service by service id
*/
func GetServiceWithRev(ctx context.Context, domain string, id string, rev int64) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domain, id)
	serviceResp, err := store.Store().Service().Search(ctx, &registry.PluginOp{
		Action:  registry.GET,
		Key:     util.StringToBytesWithNoCopy(key),
		WithRev: rev,
	})
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
	ms, ok := msCache.Get(id)
	if !ok {
		ms, err := SearchService(ctx, domain, id, registry.MODE_BOTH)
		if ms == nil {
			return nil, err
		}
		msCache.Set(id, ms, 0)
		return ms, nil
	}

	return ms.(*pb.MicroService), nil
}

func GetService(ctx context.Context, tenant string, serviceId string) (*pb.MicroService, error) {
	return SearchService(ctx, tenant, serviceId, registry.MODE_BOTH)
}

func SearchService(ctx context.Context, tenant, serviceId string, mode registry.CacheMode) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(tenant, serviceId)
	serviceResp, err := store.Store().Service().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(key),
		Mode:   mode,
	})
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

func GetServicesRawData(ctx context.Context, tenant string) ([]*mvccpb.KeyValue, error) {
	key := apt.GenerateServiceKey(tenant, "")
	resp, err := store.Store().Service().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(key),
		WithPrefix: true,
	})
	return resp.Kvs, err
}

func GetServicesByTenant(ctx context.Context, tenant string) ([]*pb.MicroService, error) {
	kvs, err := GetServicesRawData(ctx, tenant)
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
	serviceId, err = SearchServiceId(ctx, key, registry.MODE_BOTH)
	if err != nil {
		return
	}
	if len(serviceId) == 0 {
		// 别名查询
		util.Logger().Debugf("could not search microservice %s/%s/%s id by field 'serviceName', now try field 'alias'.",
			key.AppId, key.ServiceName, key.Version)
		return SearchServiceIdFromAlias(ctx, key, registry.MODE_BOTH)
	}
	return
}

func SearchServiceId(ctx context.Context, key *pb.MicroServiceKey, mode registry.CacheMode) (string, error) {
	resp, err := store.Store().ServiceIndex().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceIndexKey(key)),
		Mode:   mode,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return util.BytesToStringWithNoCopy(resp.Kvs[0].Value), nil
}

func SearchServiceIdFromAlias(ctx context.Context, key *pb.MicroServiceKey, mode registry.CacheMode) (string, error) {
	resp, err := store.Store().ServiceAlias().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(key)),
		Mode:   mode,
	})
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
	resp, err := store.Store().ServiceIndex().Search(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        util.StringToBytesWithNoCopy(prefix),
		WithPrefix: true,
		SortOrder:  registry.SORT_DESCEND,
	})
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

func ServiceExist(ctx context.Context, tenant string, serviceId string) bool {
	resp, err := store.Store().Service().Search(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       util.StringToBytesWithNoCopy(apt.GenerateServiceKey(tenant, serviceId)),
		CountOnly: true,
	})
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetAllServiceUtil(ctx context.Context) ([]*pb.MicroService, error) {
	tenant := util.ParseTenantProject(ctx)
	services, err := GetServicesByTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return services, nil
}

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
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ServiceComb/service-center/pkg/common/cache"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"strings"
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
func GetById(domain string, id string, rev int64) (*proto.MicroService, error) {
	key := core.GenerateServiceKey(domain, id)
	serviceResp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		util.LOGGER.Errorf(err, "query service %s file with revision %d failed", id, rev)
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		util.LOGGER.Errorf(nil, "service %s with revision %d does not exist.", id, rev)
		return nil, nil
	}
	service := &proto.MicroService{}
	err = json.Unmarshal(serviceResp.Kvs[0].Value, &service)
	if err != nil {
		util.LOGGER.Errorf(err, "unmarshal provider service %s file with revision %d failed", id, rev)
		return nil, err
	}
	return service, nil
}

func GetByIdInCache(domain string, id string) (*proto.MicroService, error) {
	uid := domain + ":::" + id
	ms, ok := msCache.Get(uid)
	if !ok {
		ms, err := GetById(domain, id, 0)
		if ms == nil {
			return nil, err
		}
		msCache.Set(uid, ms, 0)
		return ms, nil
	}

	return ms.(*proto.MicroService), nil
}

func GetServicesRawData(ctx context.Context, tenant string) ([]*mvccpb.KeyValue, error) {
	key := strings.Join([]string{
		core.GetServiceRootKey(tenant),
		"",
	}, "/")

	resp, err := registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
	})
	return resp.Kvs, err
}

func GetServicesByTenant(ctx context.Context, tenant string) ([]*proto.MicroService, error) {
	kvs, err := GetServicesRawData(ctx, tenant)
	if err != nil {
		return nil, err
	}
	services := []*proto.MicroService{}
	for _, kvs := range kvs {
		util.LOGGER.Debugf("start unmarshal service file: %s", string(kvs.Value))
		service := &proto.MicroService{}
		err := json.Unmarshal(kvs.Value, service)
		if err != nil {
			util.LOGGER.Error(fmt.Sprintf("Can not unmarshal %s", err), err)
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

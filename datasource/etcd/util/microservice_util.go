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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/little-cui/etcdadpt"
)

/*
	get Service by service id
*/
func GetServiceWithRev(ctx context.Context, domain string, id string, rev int64) (*pb.MicroService, error) {
	key := path.GenerateServiceKey(domain, id)
	serviceResp, err := kv.Store().Service().Search(ctx,
		etcdadpt.WithStrKey(key),
		etcdadpt.WithRev(rev))
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, datasource.ErrNoData
	}
	return serviceResp.Kvs[0].Value.(*pb.MicroService), nil
}

func GetService(ctx context.Context, domainProject string, serviceID string) (*pb.MicroService, error) {
	key := path.GenerateServiceKey(domainProject, serviceID)
	opts := append(FromContext(ctx), etcdadpt.WithStrKey(key))
	serviceResp, err := kv.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, datasource.ErrNoData
	}
	return serviceResp.Kvs[0].Value.(*pb.MicroService), nil
}

func getServicesRawData(ctx context.Context, domainProject string) ([]*sd.KeyValue, error) {
	key := path.GenerateServiceKey(domainProject, "")
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	resp, err := kv.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, err
}

//GetAllServicesAcrossDomainProject get services of all domains, projects
//the map's key is domainProject
func GetAllServicesAcrossDomainProject(ctx context.Context) (map[string][]*pb.MicroService, error) {
	key := path.GetServiceRootKey("")
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	serviceResp, err := kv.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}

	services := make(map[string][]*pb.MicroService)
	if len(serviceResp.Kvs) == 0 {
		return services, nil
	}

	for _, value := range serviceResp.Kvs {
		prefix := util.BytesToStringWithNoCopy(value.Key)
		parts := strings.Split(prefix, path.SPLIT)
		if len(parts) != 7 {
			continue
		}
		domainProject := parts[4] + path.SPLIT + parts[5]
		microService, ok := value.Value.(*pb.MicroService)
		if !ok {
			log.Error(fmt.Sprintf("backend key[%s]'s value is not type *pb.MicroService", prefix), nil)
			continue
		}
		services[domainProject] = append(services[domainProject], microService)
	}
	return services, nil
}

func GetServicesByDomainProject(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	kvs, err := getServicesRawData(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	services := []*pb.MicroService{}
	for _, kv := range kvs {
		services = append(services, kv.Value.(*pb.MicroService))
	}
	return services, nil
}

func GetServiceID(ctx context.Context, key *pb.MicroServiceKey) (serviceID string, err error) {
	serviceID, err = searchServiceID(ctx, key)
	if err != nil {
		return
	}
	if len(serviceID) == 0 {
		// 别名查询
		log.Debug(fmt.Sprintf("could not search microservice[%s/%s/%s/%s] id by 'serviceName', now try 'alias'",
			key.Environment, key.AppId, key.ServiceName, key.Version))
		return searchServiceIDFromAlias(ctx, key)
	}
	return
}

func searchServiceID(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(FromContext(ctx), etcdadpt.WithStrKey(path.GenerateServiceIndexKey(key)))
	resp, err := kv.Store().ServiceIndex().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func searchServiceIDFromAlias(ctx context.Context, key *pb.MicroServiceKey) (string, error) {
	opts := append(FromContext(ctx), etcdadpt.WithStrKey(path.GenerateServiceAliasKey(key)))
	resp, err := kv.Store().ServiceAlias().Search(ctx, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Value.(string), nil
}

func GetServiceAllVersions(ctx context.Context, key *pb.MicroServiceKey, alias bool) (*sd.Response, error) {
	copy := *key
	copy.Version = ""
	var (
		prefix  string
		indexer sd.Indexer
	)
	if alias {
		prefix = path.GenerateServiceAliasKey(&copy)
		indexer = kv.Store().ServiceAlias()
	} else {
		prefix = path.GenerateServiceIndexKey(&copy)
		indexer = kv.Store().ServiceIndex()
	}
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(prefix),
		etcdadpt.WithPrefix(),
		etcdadpt.WithDescendOrder())
	resp, err := indexer.Search(ctx, opts...)
	return resp, err
}

// FindServiceIds return serviceIDs match the key, the existence of the micro-service without consider of version
func FindServiceIds(ctx context.Context, key *pb.MicroServiceKey, matchVersion bool) ([]string, bool, error) {
	resp, err := GetServiceAllVersions(ctx, key, false)
	if err != nil {
		return nil, false, err
	}
	fromEtcdKey := path.GetInfoFromSvcIndexKV
	kvs := resp.Kvs
	if len(kvs) == 0 {
		resp, err := GetServiceAllVersions(ctx, key, true)
		if err != nil {
			return nil, false, err
		}
		fromEtcdKey = path.GetInfoFromSvcAliasKV
		kvs = resp.Kvs
	}
	var (
		etcdKeys   [][]byte
		serviceIDs []string
	)
	for _, kv := range kvs {
		etcdKeys = append(etcdKeys, kv.Key)
		serviceIDs = append(serviceIDs, kv.Value.(string))
	}
	if !matchVersion {
		return serviceIDs, len(serviceIDs) > 0, nil
	}
	for i, etcdKey := range etcdKeys {
		serviceKey := fromEtcdKey(etcdKey)
		if serviceKey.Version == key.Version {
			return []string{serviceIDs[i]}, len(serviceIDs) > 0, nil
		}
	}
	return nil, len(serviceIDs) > 0, nil
}

func ServiceExist(ctx context.Context, domainProject string, serviceID string) bool {
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(path.GenerateServiceKey(domainProject, serviceID)),
		etcdadpt.WithCountOnly())
	resp, err := kv.Store().Service().Search(ctx, opts...)
	if err != nil || resp.Count == 0 {
		return false
	}
	return true
}

func GetAllServiceUtil(ctx context.Context) ([]*pb.MicroService, error) {
	domainProject := util.ParseDomainProject(ctx)
	services, err := GetServicesByDomainProject(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func RemandServiceQuota(ctx context.Context) {
	quota.Remand(ctx, quota.TypeService)
}

func RemandInstanceQuota(ctx context.Context) {
	quota.Remand(ctx, quota.TypeInstance)
}

func UpdateService(domainProject string, serviceID string, service *pb.MicroService) (opt etcdadpt.OpOptions, err error) {
	opt = etcdadpt.OpOptions{}
	key := path.GenerateServiceKey(domainProject, serviceID)
	data, err := json.Marshal(service)
	if err != nil {
		log.Error("marshal service file failed", err)
		return
	}
	opt = etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data))
	return
}

func GetOneDomainProjectServiceCount(ctx context.Context, domainProject string) (int64, error) {
	key := path.GetServiceRootKey(domainProject) + path.SPLIT
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithCountOnly(),
		etcdadpt.WithPrefix())
	resp, err := kv.Store().Service().Search(ctx, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func GetOneDomainProjectInstanceCount(ctx context.Context, domainProject string) (int64, error) {
	key := path.GetInstanceRootKey(domainProject) + path.SPLIT
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithCountOnly(),
		etcdadpt.WithPrefix())
	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func GetGlobalInstanceCount(ctx context.Context) (int64, error) {
	serviceIDs, err := GetGlobalServiceIDs(ctx)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, serviceID := range serviceIDs {
		n, err := GetInstanceCountOfOneService(ctx, datasource.RegistryDomainProject, serviceID)
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}

func GetGlobalServiceIDs(ctx context.Context) ([]string, error) {
	var serviceIDs []string
	for name := range datasource.GlobalServiceNames {
		key := path.GenerateServiceIndexKey(&pb.MicroServiceKey{
			Tenant:      datasource.RegistryDomainProject,
			Environment: getGlobalEnvironment(),
			AppId:       datasource.RegistryAppID,
			ServiceName: name,
		})
		opts := append(FromContext(ctx),
			etcdadpt.WithStrKey(key),
			etcdadpt.WithPrefix())
		resp, err := kv.Store().ServiceIndex().Search(ctx, opts...)
		if err != nil {
			return nil, err
		}
		for _, id := range resp.Kvs {
			serviceIDs = append(serviceIDs, id.Value.(string))
		}
	}
	return serviceIDs, nil
}

func GetGlobalServiceCount(ctx context.Context) (int64, error) {
	var count int64
	for name := range datasource.GlobalServiceNames {
		key := path.GenerateServiceIndexKey(&pb.MicroServiceKey{
			Tenant:      datasource.RegistryDomainProject,
			Environment: getGlobalEnvironment(),
			AppId:       datasource.RegistryAppID,
			ServiceName: name,
		})
		opts := append(FromContext(ctx),
			etcdadpt.WithStrKey(key),
			etcdadpt.WithCountOnly(),
			etcdadpt.WithPrefix())
		resp, err := kv.Store().ServiceIndex().Search(ctx, opts...)
		if err != nil {
			return 0, err
		}
		count += resp.Count
	}
	return count, nil
}

func getGlobalEnvironment() string {
	env := pb.ENV_PROD
	if config.GetProfile().IsDev() {
		env = pb.ENV_DEV
	}
	return env
}

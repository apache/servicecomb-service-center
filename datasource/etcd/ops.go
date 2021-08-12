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
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
	"github.com/little-cui/etcdadpt"
)

const poolSizeOfCleanup = 5

type CleanupServiceIDKey struct {
	DomainProject string
	ServiceID     string
}

func (ds *MetadataManager) GetServiceCount(ctx context.Context, request *pb.GetServiceCountRequest) (*pb.GetServiceCountResponse, error) {
	domainProject := request.Domain
	if request.Project != "" {
		domainProject += path.SPLIT + request.Project
	}
	all, err := serviceUtil.GetOneDomainProjectServiceCount(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	global, err := ds.getGlobalServiceCount(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return &pb.GetServiceCountResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get service count by domain project successfully"),
		Count:    all - global,
	}, nil
}

func (ds *MetadataManager) getGlobalServiceCount(ctx context.Context, domainProject string) (int64, error) {
	if strings.Index(datasource.RegistryDomainProject+datasource.SPLIT, domainProject+datasource.SPLIT) != 0 {
		return 0, nil
	}
	global, err := serviceUtil.GetGlobalServiceCount(ctx)
	if err != nil {
		return 0, err
	}
	return global, nil
}

func (ds *MetadataManager) GetInstanceCount(ctx context.Context, request *pb.GetServiceCountRequest) (*pb.GetServiceCountResponse, error) {
	domainProject := request.Domain
	if request.Project != "" {
		domainProject += path.SPLIT + request.Project
	}
	all, err := serviceUtil.GetOneDomainProjectInstanceCount(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	global, err := ds.getGlobalInstanceCount(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return &pb.GetServiceCountResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Get instance count by domain/project successfully"),
		Count:    all - global,
	}, nil
}

func (ds *MetadataManager) getGlobalInstanceCount(ctx context.Context, domainProject string) (int64, error) {
	if strings.Index(datasource.RegistryDomainProject+path.SPLIT, domainProject+path.SPLIT) != 0 {
		return 0, nil
	}
	global, err := serviceUtil.GetGlobalInstanceCount(ctx)
	if err != nil {
		return 0, err
	}
	return global, nil
}

func (ds *MetadataManager) CleanupUnusedMicroservice(ctx context.Context, reserveVersionCount int) error {
	key := path.GetServiceIndexRootKey("")
	indexesResp, err := sd.ServiceIndex().Search(ctx, etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	if err != nil {
		log.Error("query all microservices indexes failed", err)
		return err
	}
	serviceIDKeys := GetOldServiceIDs(indexesResp, reserveVersionCount)
	if len(serviceIDKeys) == 0 {
		return nil
	}
	serviceIDKeys = FilterUnused(ctx, serviceIDKeys)
	if len(serviceIDKeys) == 0 {
		return nil
	}

	log.Warn(fmt.Sprintf("start clean up %d microservices", len(serviceIDKeys)))
	n := UnregisterManyService(ctx, serviceIDKeys)
	if n > 0 {
		log.Warn(fmt.Sprintf("auto clean up %d microservices", n))
	}
	return nil
}

func FilterUnused(ctx context.Context, serviceIDKeys []*CleanupServiceIDKey) []*CleanupServiceIDKey {
	matched := make([]*CleanupServiceIDKey, 0, len(serviceIDKeys))
	for _, serviceIDKey := range serviceIDKeys {
		serviceID := serviceIDKey.ServiceID
		instanceKey := path.GenerateInstanceKey(serviceIDKey.DomainProject, serviceID, "")
		resp, err := sd.Instance().Search(ctx, etcdadpt.WithStrKey(instanceKey),
			etcdadpt.WithPrefix(), etcdadpt.WithCountOnly())
		if err != nil {
			log.Error(fmt.Sprintf("count microservice %s instance failed", serviceID), err)
			continue
		}
		if resp.Count > 0 {
			continue
		}
		matched = append(matched, serviceIDKey)
	}
	return matched
}

func UnregisterManyService(ctx context.Context, serviceIDKeys []*CleanupServiceIDKey) (deleted int64) {
	pool := goutil.New(gopool.Configure().WithContext(ctx).Workers(poolSizeOfCleanup))
	defer pool.Done()

	for _, key := range serviceIDKeys {
		domainProject := key.DomainProject
		serviceID := key.ServiceID
		pool.Do(func(ctx context.Context) {
			resp, err := datasource.GetMetadataManager().UnregisterService(util.SetDomainProjectString(ctx, domainProject),
				&pb.DeleteServiceRequest{ServiceId: serviceID})
			if err == nil && resp.Response.IsSucceed() {
				atomic.AddInt64(&deleted, 1)
			}
		})
	}
	return
}

func GetOldServiceIDs(indexesResp *kvstore.Response, reserveVersionCount int) []*CleanupServiceIDKey {
	total := indexesResp.Count
	if total == 0 {
		return nil
	}
	var (
		counter           = make(map[string]int, total)
		matched           = make(map[string]*pb.MicroServiceKey)
		serviceVersionMap = make(map[string][]*kvstore.KeyValue, total)
	)
	for _, kv := range indexesResp.Kvs {
		serviceVersionKey := path.GetInfoFromSvcIndexKV(kv.Key)
		serviceVersionKey.Version = ""
		serviceKey := path.GenerateServiceIndexKey(serviceVersionKey)
		count := counter[serviceKey]
		count++
		counter[serviceKey] = count
		serviceVersionMap[serviceKey] = append(serviceVersionMap[serviceKey], kv)
		if count > reserveVersionCount {
			matched[serviceKey] = serviceVersionKey
		}
	}

	if len(matched) == 0 {
		return nil
	}

	var serviceIDs []*CleanupServiceIDKey
	for key, serviceKey := range matched {
		kvs := serviceVersionMap[key]
		serviceUtil.Sort(kvs, serviceUtil.Larger)
		for _, kv := range kvs[reserveVersionCount:] {
			serviceIDs = append(serviceIDs, &CleanupServiceIDKey{
				DomainProject: serviceKey.Tenant,
				ServiceID:     kv.Value.(string),
			})
		}
	}
	return serviceIDs
}

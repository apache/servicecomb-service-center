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

const poolSizeOfRotation = 5

type RotateServiceIDKey struct {
	DomainProject string
	ServiceID     string
}

func (ds *MetadataManager) RetireService(ctx context.Context, plan *datasource.RetirePlan) error {
	key := path.GetServiceIndexRootKey("")
	indexesResp, err := sd.ServiceIndex().Search(ctx, etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	if err != nil {
		log.Error("query all microservices indexes failed", err)
		return err
	}
	serviceIDKeys := GetRetireServiceIDs(indexesResp, plan.Reserve)
	if len(serviceIDKeys) == 0 {
		return nil
	}
	serviceIDKeys = FilterNoInstance(ctx, serviceIDKeys)
	if len(serviceIDKeys) == 0 {
		return nil
	}

	log.Warn(fmt.Sprintf("start retire %d microservices", len(serviceIDKeys)))
	n := UnregisterManyService(ctx, serviceIDKeys)
	if n > 0 {
		log.Warn(fmt.Sprintf("%d microservices retired", n))
	}
	return nil
}

func FilterNoInstance(ctx context.Context, serviceIDKeys []*RotateServiceIDKey) []*RotateServiceIDKey {
	matched := make([]*RotateServiceIDKey, 0, len(serviceIDKeys))
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

func UnregisterManyService(ctx context.Context, serviceIDKeys []*RotateServiceIDKey) (deleted int64) {
	pool := goutil.New(gopool.Configure().WithContext(ctx).Workers(poolSizeOfRotation))
	defer pool.Done()

	for _, key := range serviceIDKeys {
		domainProject := key.DomainProject
		serviceID := key.ServiceID
		pool.Do(func(ctx context.Context) {
			resp, err := datasource.GetMetadataManager().UnregisterService(util.SetDomainProjectString(ctx, domainProject),
				&pb.DeleteServiceRequest{ServiceId: serviceID, Force: true})
			if err == nil && resp.Response.IsSucceed() {
				atomic.AddInt64(&deleted, 1)
			}
		})
	}
	return
}

func GetRetireServiceIDs(indexesResp *kvstore.Response, reserveVersionCount int) []*RotateServiceIDKey {
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

	var serviceIDs []*RotateServiceIDKey
	for key, serviceKey := range matched {
		kvs := serviceVersionMap[key]
		serviceUtil.Sort(kvs, serviceUtil.Larger)
		for _, kv := range kvs[reserveVersionCount:] {
			serviceIDs = append(serviceIDs, &RotateServiceIDKey{
				DomainProject: serviceKey.Tenant,
				ServiceID:     kv.Value.(string),
			})
		}
	}
	return serviceIDs
}

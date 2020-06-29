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
package cache

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"sort"
)

var clustersIndex = make(map[string]int)

func init() {
	var clusters []string
	for name := range registry.Configuration().Clusters {
		clusters = append(clusters, name)
	}
	sort.Strings(clusters)
	for i, name := range clusters {
		clustersIndex[name] = i
	}
}

type InstancesFilter struct {
}

func (f *InstancesFilter) Name(ctx context.Context, _ *cache.Node) string {
	instanceKey, ok := ctx.Value(CTX_FIND_PROVIDER_INSTANCE).(*pb.HeartbeatSetElement)
	if ok {
		return instanceKey.ServiceId + apt.SPLIT + instanceKey.InstanceId
	}
	return ""
}

func (f *InstancesFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	pCopy := *parent.Cache.Get(CACHE_FIND).(*VersionRuleCacheItem)

	pCopy.Instances, pCopy.Rev, err = f.Find(ctx, parent)
	if err != nil {
		return
	}

	pCopy.InitBrokenQueue()
	node = cache.NewNode()
	node.Cache.Set(CACHE_FIND, &pCopy)
	return
}

func (f *InstancesFilter) Find(ctx context.Context, parent *cache.Node) (
	instances []*pb.MicroServiceInstance, rev string, err error) {
	pCache := parent.Cache.Get(CACHE_FIND).(*VersionRuleCacheItem)
	provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)

	instanceKey, ok := ctx.Value(CTX_FIND_PROVIDER_INSTANCE).(*pb.HeartbeatSetElement)
	if ok {
		if len(pCache.ServiceIds) == 0 {
			// can not find by instanceKey.ServiceId after pre-filters init
			return
		}
		instances, rev, err = f.FindInstances(ctx, provider.Tenant, instanceKey)
	} else {
		instances, rev, err = f.BatchFindInstances(ctx, provider.Tenant, pCache.ServiceIds)
	}
	if err != nil {
		consumer := ctx.Value(CTX_FIND_CONSUMER).(*pb.MicroService)
		findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumer.ServiceId,
			provider.AppId, provider.ServiceName, provider.Version)
		log.Errorf(err, "Find failed, %s", findFlag)
	}
	return
}

func (f *InstancesFilter) findInstances(ctx context.Context, domainProject, serviceId, instanceId string, maxRevs []int64, counts []int64) (instances []*pb.MicroServiceInstance, err error) {
	key := apt.GenerateInstanceKey(domainProject, serviceId, instanceId)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
	resp, err := backend.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return
	}

	for _, kv := range resp.Kvs {
		if i, ok := clustersIndex[kv.ClusterName]; ok {
			if kv.ModRevision > maxRevs[i] {
				maxRevs[i] = kv.ModRevision
			}
			counts[i]++
		}
		instances = append(instances, kv.Value.(*pb.MicroServiceInstance))
	}
	return
}

func (f *InstancesFilter) FindInstances(ctx context.Context, domainProject string, instanceKey *pb.HeartbeatSetElement) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs = make([]int64, len(clustersIndex))
		counts  = make([]int64, len(clustersIndex))
	)
	instances, err = f.findInstances(ctx, domainProject, instanceKey.ServiceId, instanceKey.InstanceId, maxRevs, counts)
	if err != nil {
		return
	}
	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

func (f *InstancesFilter) BatchFindInstances(ctx context.Context, domainProject string, serviceIds []string) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		maxRevs = make([]int64, len(clustersIndex))
		counts  = make([]int64, len(clustersIndex))
	)
	for _, providerServiceId := range serviceIds {
		insts, err := f.findInstances(ctx, domainProject, providerServiceId, "", maxRevs, counts)
		if err != nil {
			return nil, "", err
		}
		instances = append(instances, insts...)
	}

	return instances, serviceUtil.FormatRevision(maxRevs, counts), nil
}

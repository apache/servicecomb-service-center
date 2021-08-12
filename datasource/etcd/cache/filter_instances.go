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

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/little-cui/etcdadpt"
)

type InstancesFilter struct {
}

func (f *InstancesFilter) Name(ctx context.Context, _ *cache.Node) string {
	instanceKey, ok := ctx.Value(CtxProviderInstanceKey).(*pb.HeartbeatSetElement)
	if ok {
		return instanceKey.ServiceId + path.SPLIT + instanceKey.InstanceId
	}
	return ""
}

func (f *InstancesFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	pCopy := *parent.Cache.Get(FindResult).(*VersionRuleCacheItem)

	pCopy.Instances, pCopy.Rev, err = f.Find(ctx, parent)
	if err != nil {
		return
	}

	pCopy.InitBrokenQueue()
	node = cache.NewNode()
	node.Cache.Set(FindResult, &pCopy)
	return
}

func (f *InstancesFilter) Find(ctx context.Context, parent *cache.Node) (
	instances []*pb.MicroServiceInstance, rev string, err error) {
	pCache := parent.Cache.Get(FindResult).(*VersionRuleCacheItem)
	provider := ctx.Value(CtxProviderKey).(*pb.MicroServiceKey)

	instanceKey, ok := ctx.Value(CtxProviderInstanceKey).(*pb.HeartbeatSetElement)
	if ok {
		if len(pCache.ServiceIds) == 0 {
			// can not find by instanceKey.ServiceID after pre-filters init
			return
		}
		instances, rev, err = f.FindInstances(ctx, provider.Tenant, instanceKey)
	} else {
		instances, rev, err = f.BatchFindInstances(ctx, provider.Tenant, pCache.ServiceIds)
	}
	if err != nil {
		consumer := ctx.Value(CtxConsumerID).(*pb.MicroService)
		findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s", consumer.ServiceId,
			provider.AppId, provider.ServiceName)
		log.Error(fmt.Sprintf("Find failed, %s", findFlag), err)
	}
	return
}

func (f *InstancesFilter) findInstances(ctx context.Context, domainProject, serviceID, instanceID string, maxRevs []int64, counts []int64) (instances []*pb.MicroServiceInstance, err error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, instanceID)
	opts := append(util.FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	resp, err := sd.Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return
	}

	for _, kv := range resp.Kvs {
		if i, ok := getOrCreateClustersIndex()[kv.ClusterName]; ok {
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
		index   = getOrCreateClustersIndex()
		maxRevs = make([]int64, len(index))
		counts  = make([]int64, len(index))
	)
	instances, err = f.findInstances(ctx, domainProject, instanceKey.ServiceId, instanceKey.InstanceId, maxRevs, counts)
	if err != nil {
		return
	}
	return instances, util.FormatRevision(maxRevs, counts), nil
}

func (f *InstancesFilter) BatchFindInstances(ctx context.Context, domainProject string, serviceIDs []string) (instances []*pb.MicroServiceInstance, rev string, err error) {
	var (
		index   = getOrCreateClustersIndex()
		maxRevs = make([]int64, len(index))
		counts  = make([]int64, len(index))
	)
	for _, providerServiceID := range serviceIDs {
		insts, err := f.findInstances(ctx, domainProject, providerServiceID, "", maxRevs, counts)
		if err != nil {
			return nil, "", err
		}
		instances = append(instances, insts...)
	}

	return instances, util.FormatRevision(maxRevs, counts), nil
}

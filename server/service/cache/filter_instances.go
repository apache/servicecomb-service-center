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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/cache"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

type InstancesFilter struct {
}

func (f *InstancesFilter) Name(ctx context.Context) string {
	return ""
}

func (f *InstancesFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)
	pCopy := *parent.Cache.Get(CACHE_FIND).(*VersionRuleCacheItem)

	var (
		instances []*pb.MicroServiceInstance
		rev       int64
	)
	for _, providerServiceId := range pCopy.ServiceIds {
		key := apt.GenerateInstanceKey(provider.Tenant, providerServiceId, "")
		opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key), registry.WithPrefix())
		resp, err := backend.Store().Instance().Search(ctx, opts...)
		if err != nil {
			consumer := ctx.Value(CTX_FIND_CONSUMER).(*pb.MicroService)
			findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumer.ServiceId,
				provider.AppId, provider.ServiceName, provider.Version)
			log.Errorf(err, "Instance().Search failed, %s", findFlag)
			return nil, err
		}

		for _, kv := range resp.Kvs {
			if kv.ModRevision > rev {
				rev = kv.ModRevision
			}
			instances = append(instances, kv.Value.(*pb.MicroServiceInstance))
		}
	}

	pCopy.Instances = instances
	pCopy.Rev = serviceUtil.FormatRevision(rev, int64(len(instances)))

	node = cache.NewNode()
	node.Cache.Set(CACHE_FIND, &pCopy)
	return
}

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

	pb "github.com/go-chassis/cari/discovery"

	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type VersionFilter struct {
}

func (f *VersionFilter) Name(ctx context.Context, _ *cache.Node) string {
	instanceKey, ok := ctx.Value(CtxProviderInstanceKey).(*pb.HeartbeatSetElement)
	if ok {
		return instanceKey.ServiceId
	}
	return ""
}

func (f *VersionFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	instance, ok := ctx.Value(CtxProviderInstanceKey).(*pb.HeartbeatSetElement)
	if ok {
		node = cache.NewNode()
		node.Cache.Set(FindResult, &VersionRuleCacheItem{
			ServiceIds: []string{instance.ServiceId},
		})
		return
	}

	provider := ctx.Value(CtxProviderKey).(*pb.MicroServiceKey)
	ids, exist, err := serviceUtil.FindServiceIds(ctx, provider, false)
	if err != nil {
		consumer := ctx.Value(CtxConsumerID).(*pb.MicroService)
		findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s", consumer.ServiceId,
			provider.AppId, provider.ServiceName)
		log.Error(fmt.Sprintf("FindServiceIds failed, %s", findFlag), err)
		return
	}
	if !exist {
		return
	}

	node = cache.NewNode()
	node.Cache.Set(FindResult, &VersionRuleCacheItem{
		ServiceIds: ids,
	})
	return
}

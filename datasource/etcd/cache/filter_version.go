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

	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
)

type VersionRuleFilter struct {
}

func (f *VersionRuleFilter) Name(ctx context.Context, _ *cache.Node) string {
	provider := ctx.Value(CtxFindProvider).(*pb.MicroServiceKey)
	return provider.Version
}

func (f *VersionRuleFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	instance, ok := ctx.Value(CtxFindProviderInstance).(*pb.HeartbeatSetElement)
	if ok {
		node = cache.NewNode()
		node.Cache.Set(Find, &VersionRuleCacheItem{
			ServiceIds: []string{instance.ServiceId},
		})
		return
	}

	provider := ctx.Value(CtxFindProvider).(*pb.MicroServiceKey)
	// 版本规则
	ids, exist, err := serviceUtil.FindServiceIds(ctx, provider.Version, provider)
	if err != nil {
		consumer := ctx.Value(CtxFindConsumer).(*pb.MicroService)
		findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumer.ServiceId,
			provider.AppId, provider.ServiceName, provider.Version)
		log.Errorf(err, "FindServiceIds failed, %s", findFlag)
		return
	}
	if !exist {
		return
	}

	node = cache.NewNode()
	node.Cache.Set(Find, &VersionRuleCacheItem{
		ServiceIds: ids,
	})
	return
}

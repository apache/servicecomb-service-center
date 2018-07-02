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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

type VersionRuleFilter struct {
}

func (f *VersionRuleFilter) Name(ctx context.Context) string {
	provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)
	return provider.Version
}

func (f *VersionRuleFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)
	// 版本规则
	ids, err := serviceUtil.FindServiceIds(ctx, provider.Version, provider)
	if err != nil {
		consumer := ctx.Value(CTX_FIND_CONSUMER).(*pb.MicroService)
		findFlag := fmt.Sprintf("consumer %s find provider %s/%s/%s", consumer.ServiceId,
			provider.AppId, provider.ServiceName, provider.Version)
		util.Logger().Errorf(err, "VersionRuleFilter failed, %s", findFlag)
		return
	}
	if len(ids) == 0 {
		return
	}

	node = cache.NewNode()
	node.Cache.Set(cacheFindProviderIds, ids)
	return
}

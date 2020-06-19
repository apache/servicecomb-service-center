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
	"github.com/apache/servicecomb-service-center/pkg/cache"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"math"
)

var DependencyRule = &DependencyRuleCache{
	Tree: cache.NewTree(cache.Configure().
		WithMaxSize(math.MaxInt64))}

func init() {
	DependencyRule.AddFilter(
		&ServiceFilter{},
		&ConsumerFilter{})
}

type DependencyRuleItem struct {
	VersionRule string
}

type DependencyRuleCache struct {
	*cache.Tree
}

func (f *DependencyRuleCache) ExistVersionRule(ctx context.Context, consumerId string, provider *pb.MicroServiceKey) bool {
	cloneCtx := context.WithValue(context.WithValue(ctx,
		CTX_FIND_CONSUMER, consumerId),
		CTX_FIND_PROVIDER, provider)

	node, _ := f.Tree.Get(cloneCtx, cache.Options().Temporary(ctx.Value(serviceUtil.CTX_NOCACHE) == "1"))
	if node == nil {
		return false
	}
	v := node.Cache.Get(CACHE_DEP).(*DependencyRuleItem)
	if v.VersionRule != provider.Version {
		v.VersionRule = provider.Version
		return false
	}
	return true
}

func (f *DependencyRuleCache) Remove(provider *pb.MicroServiceKey) {
	f.Tree.Remove(context.WithValue(context.Background(), CTX_FIND_PROVIDER, provider))
}

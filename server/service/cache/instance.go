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
	"github.com/apache/servicecomb-service-center/pkg/cache"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"math"
	"time"
)

var FindInstances = &FindInstancesCache{
	Tree: cache.NewTree(cache.Configure().
		WithTTL(2 * time.Minute).
		WithMaxSize(math.MaxInt64))}

func init() {
	FindInstances.AddFilter(
		&ServiceFilter{},
		&VersionRuleFilter{},
		&TagsFilter{},
		&AccessibleFilter{},
		&InstancesFilter{},
		&ConsistencyFilter{},
	)
}

type VersionRuleCacheItem struct {
	ServiceIds []string
	Instances  []*pb.MicroServiceInstance
	Rev        string

	broken bool
	queue  chan struct{}
}

func (vi *VersionRuleCacheItem) InitBrokenQueue() {
	if vi.queue == nil {
		vi.queue = make(chan struct{}, 1)
	}
	vi.broken = false
	vi.queue <- struct{}{}
}

func (vi *VersionRuleCacheItem) BrokenWait() bool {
	<-vi.queue
	return vi.broken
}

func (vi *VersionRuleCacheItem) Broken() {
	vi.broken = true
	close(vi.queue)
}

type FindInstancesCache struct {
	*cache.Tree
}

func (f *FindInstancesCache) Get(ctx context.Context, consumer *pb.MicroService, provider *pb.MicroServiceKey,
	tags []string, rev string) (*VersionRuleCacheItem, error) {
	cloneCtx := context.WithValue(context.WithValue(context.WithValue(context.WithValue(ctx,
		CTX_FIND_CONSUMER, consumer),
		CTX_FIND_PROVIDER, provider),
		CTX_FIND_TAGS, tags),
		CTX_FIND_REQUEST_REV, rev)

	node, err := f.Tree.Get(cloneCtx, cache.Options().Temporary(ctx.Value(serviceUtil.CTX_NOCACHE) == "1"))
	if node == nil {
		return nil, err
	}
	return node.Cache.Get(CACHE_FIND).(*VersionRuleCacheItem), nil
}

func (f *FindInstancesCache) GetWithProviderId(ctx context.Context, consumer *pb.MicroService, provider *pb.MicroServiceKey,
	instanceKey *pb.HeartbeatSetElement, tags []string, rev string) (*VersionRuleCacheItem, error) {
	cloneCtx := context.WithValue(ctx, CTX_FIND_PROVIDER_INSTANCE, instanceKey)
	return f.Get(cloneCtx, consumer, provider, tags, rev)
}

func (f *FindInstancesCache) Remove(provider *pb.MicroServiceKey) {
	f.Tree.Remove(context.WithValue(context.Background(), CTX_FIND_PROVIDER, provider))
	if len(provider.Alias) > 0 {
		copy := *provider
		copy.ServiceName = copy.Alias
		f.Tree.Remove(context.WithValue(context.Background(), CTX_FIND_PROVIDER, &copy))
	}
}

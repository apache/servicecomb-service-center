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

	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// ConsistencyFilter improves consistency.
// Scenario: cache maybe different between several service-centers.
type ConsistencyFilter struct {
	InstancesFilter
}

func (f *ConsistencyFilter) Name(ctx context.Context, parent *cache.Node) string {
	item := parent.Cache.Get(FindResult).(*VersionRuleCacheItem)
	requestRev := ctx.Value(CtxRequestRev).(string)
	if len(requestRev) == 0 || requestRev == item.Rev {
		return ""
	}
	return requestRev
}

// Init generates cache.
// We think cache inconsistency happens and correction is needed only when the
// revision in the request is not empty and different from the revision of
// parent cache. To correct inconsistency, RevisionFilter skips cache and get
// data from the backend directly to response.
// It's impossible to guarantee consistency if the backend is not creditable,
// thus in this condition RevisionFilter uses cache only.
func (f *ConsistencyFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	pCache := parent.Cache.Get(FindResult).(*VersionRuleCacheItem)
	requestRev := ctx.Value(CtxRequestRev).(string)
	// do not need to check consistency between sc instances:
	// 1. request without rev param
	// 2. request rev is the same as cache current sc instance
	// 3. datasource has no cache indexer
	if len(requestRev) == 0 || requestRev == pCache.Rev ||
		!(kv.Store().Instance().Creditable()) {
		node = cache.NewNode()
		node.Cache.Set(FindResult, pCache)
		return
	}

	if pCache.BrokenWait() {
		node = cache.NewNode()
		node.Cache.Set(FindResult, pCache)
		return
	}

	cloneCtx := util.WithNoCache(util.CloneContext(ctx))
	insts, rev, err := f.Find(cloneCtx, parent)
	if err != nil {
		pCache.InitBrokenQueue()
		return nil, err
	}

	log.Warn(fmt.Sprintf("inconsistent rev! %s, req[%s], cache[%s], datasource[%s]",
		parent.Name, requestRev, pCache.Rev, rev))
	pCache.Instances, pCache.Rev = insts, rev
	pCache.Broken()

	node = cache.NewNode()
	node.Cache.Set(FindResult, pCache)
	return
}

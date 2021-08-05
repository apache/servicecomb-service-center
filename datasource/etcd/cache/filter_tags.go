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
	"sort"
	"strings"

	pb "github.com/go-chassis/cari/discovery"

	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type TagsFilter struct {
}

func (f *TagsFilter) Name(ctx context.Context, _ *cache.Node) string {
	tags, _ := ctx.Value(CtxTags).([]string)
	sort.Strings(tags)
	return strings.Join(tags, ",")
}

func (f *TagsFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	tags, _ := ctx.Value(CtxTags).([]string)
	if len(tags) == 0 {
		node = cache.NewNode()
		node.Cache = parent.Cache
		return
	}

	var ids []string

	targetDomainProject := util.ParseTargetDomainProject(ctx)
	pCopy := *parent.Cache.Get(FindResult).(*VersionRuleCacheItem)

loopProviderIds:
	for _, providerServiceID := range pCopy.ServiceIds {
		tagsFromETCD, err := serviceUtil.GetTagsUtils(ctx, targetDomainProject, providerServiceID)
		if err != nil {
			consumer := ctx.Value(CtxConsumerID).(*pb.MicroService)
			provider := ctx.Value(CtxProviderKey).(*pb.MicroServiceKey)
			findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumer.ServiceId,
				provider.AppId, provider.ServiceName, provider.Version)
			log.Error(fmt.Sprintf("TagsFilter failed, %s", findFlag), err)
			return nil, err
		}
		if len(tagsFromETCD) == 0 {
			continue
		}
		for _, tag := range tags {
			if _, ok := tagsFromETCD[tag]; !ok {
				continue loopProviderIds
			}
		}
		ids = append(ids, providerServiceID)
	}

	pCopy.ServiceIds = ids

	node = cache.NewNode()
	node.Cache.Set(FindResult, &pCopy)
	return
}

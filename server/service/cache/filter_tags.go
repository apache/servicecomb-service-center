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
	"sort"
	"strings"
)

type TagsFilter struct {
}

func (f *TagsFilter) Name(ctx context.Context) string {
	tags, _ := ctx.Value(CTX_FIND_TAGS).([]string)
	sort.Strings(tags)
	return strings.Join(tags, ",")
}

func (f *TagsFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	tags, _ := ctx.Value(CTX_FIND_TAGS).([]string)
	if len(tags) == 0 {
		node = cache.NewNode()
		node.Cache = parent.Cache
		return
	}

	var ids []string

	targetDomainProject := util.ParseTargetDomainProject(ctx)
	providerIds := parent.Cache.Get(cacheFindProviderIds).([]string)

loopProviderIds:
	for _, providerServiceId := range providerIds {
		tagsFromETCD, err := serviceUtil.GetTagsUtils(ctx, targetDomainProject, providerServiceId)
		if err != nil {
			consumer := ctx.Value(CTX_FIND_CONSUMER).(*pb.MicroService)
			provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)
			findFlag := fmt.Sprintf("consumer %s find provider %s/%s/%s", consumer.ServiceId,
				provider.AppId, provider.ServiceName, provider.Version)
			util.Logger().Errorf(err, "TagsFilter failed, %s", findFlag)
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
		ids = append(ids, providerServiceId)
	}

	if len(ids) == 0 {
		return
	}

	node = cache.NewNode()
	node.Cache.Set(cacheFindProviderIds, ids)
	return
}

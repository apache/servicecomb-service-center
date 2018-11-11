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
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

type AccessibleFilter struct {
}

func (f *AccessibleFilter) Name(ctx context.Context) string {
	consumer := ctx.Value(CTX_FIND_CONSUMER).(*pb.MicroService)
	return consumer.ServiceId
}

func (f *AccessibleFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	var ids []string
	consumerId := f.Name(ctx)
	pCopy := *parent.Cache.Get(CACHE_FIND).(*VersionRuleCacheItem)
	for _, providerServiceId := range pCopy.ServiceIds {
		if err := serviceUtil.Accessible(ctx, consumerId, providerServiceId); err != nil {
			provider := ctx.Value(CTX_FIND_PROVIDER).(*pb.MicroServiceKey)
			findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s/%s", consumerId,
				provider.AppId, provider.ServiceName, provider.Version)
			log.Errorf(err, "AccessibleFilter failed, %s", findFlag)
			continue
		}
		ids = append(ids, providerServiceId)
	}

	pCopy.ServiceIds = ids

	node = cache.NewNode()
	node.Cache.Set(CACHE_FIND, &pCopy)
	return
}

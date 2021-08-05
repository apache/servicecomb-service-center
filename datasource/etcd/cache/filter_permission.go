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

	"github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type AccessibleFilter struct {
}

func (f *AccessibleFilter) Name(ctx context.Context, _ *cache.Node) string {
	consumer := ctx.Value(CtxConsumerID).(*pb.MicroService)
	return consumer.ServiceId
}

func (f *AccessibleFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	var ids []string
	consumerID := ctx.Value(CtxConsumerID).(*pb.MicroService).ServiceId
	pCopy := *parent.Cache.Get(FindResult).(*VersionRuleCacheItem)
	for _, providerServiceID := range pCopy.ServiceIds {
		if err := util.Accessible(ctx, consumerID, providerServiceID); err != nil {
			provider := ctx.Value(CtxProviderKey).(*pb.MicroServiceKey)
			findFlag := fmt.Sprintf("consumer '%s' find provider %s/%s", consumerID,
				provider.AppId, provider.ServiceName)
			log.Error(fmt.Sprintf("AccessibleFilter failed, %s", findFlag), err)
			continue
		}
		ids = append(ids, providerServiceID)
	}

	pCopy.ServiceIds = ids

	node = cache.NewNode()
	node.Cache.Set(FindResult, &pCopy)
	return
}

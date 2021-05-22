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

package mongo

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
)

func GetAllConsumerIds(ctx context.Context, provider *pb.MicroService) (allow []string, deny []string, err error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, ok := cache.GetRulesByServiceID(provider.ServiceId)
	if !ok {
		providerRules, err = dao.GetRulesByServiceID(ctx, provider.ServiceId)
		if err != nil {
			return nil, nil, err
		}
	}

	allow, deny, err = GetConsumerIDsWithFilter(ctx, provider, providerRules)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return nil, nil, err
	}
	return allow, deny, nil
}

func GetConsumerIDsWithFilter(ctx context.Context, provider *pb.MicroService, rules []*dao.Rule) (allow []string, deny []string, err error) {
	providerServiceKey, err := dao.GetProviderDeps(ctx, provider)
	if err != nil {
		return nil, nil, err
	}
	consumerIDs := make([]string, len(providerServiceKey))
	for _, serviceKeys := range providerServiceKey {
		id, ok := cache.GetServiceID(ctx, serviceKeys)
		if !ok {
			id, err = dao.GetServiceID(ctx, serviceKeys)
			if err != nil {
				return nil, nil, err
			}
		}
		consumerIDs = append(consumerIDs, id)
	}
	return FilterAll(ctx, consumerIDs, rules)
}

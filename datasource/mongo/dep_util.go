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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/dao"
	pb "github.com/go-chassis/cari/discovery"
)

func GetConsumerIDs(ctx context.Context, provider *pb.MicroService) ([]string, error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, fmt.Errorf("invalid provider")
	}

	var err error
	serviceDeps, ok := cache.GetProviderServiceOfDeps(provider)
	if !ok {
		serviceDeps, err = dao.GetProviderDeps(ctx, provider)
		if err != nil {
			return nil, err
		}
	}
	consumerIDs := make([]string, 0, len(serviceDeps.Dependency))
	for _, serviceKeys := range serviceDeps.Dependency {
		id, err := GetServiceID(ctx, serviceKeys)
		if err != nil {
			return nil, err
		}
		consumerIDs = append(consumerIDs, id)
	}
	return consumerIDs, nil
}

func GetConsumers(ctx context.Context, domainProject string, provider *pb.MicroService,
	opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	return dr.GetDependencyConsumers(opts...)
}

func GetProviders(ctx context.Context, domainProject string, consumer *pb.MicroService,
	opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer)
	return dr.GetDependencyProviders(opts...)
}

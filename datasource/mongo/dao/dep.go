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

package dao

import (
	"context"

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetProviderDeps(ctx context.Context, provider *discovery.MicroService) (*discovery.MicroServiceDependency, error) {
	return getServiceOfDeps(ctx, datasource.Provider, provider)
}

func getServiceOfDeps(ctx context.Context, ruleType string, provider *discovery.MicroService) (*discovery.MicroServiceDependency, error) {
	filter := mutil.NewFilter(
		mutil.ServiceType(ruleType),
		mutil.ServiceKeyTenant(util.ParseDomainProject(ctx)),
		mutil.ServiceKeyAppID(provider.AppId),
		mutil.ServiceKeyServiceName(provider.ServiceName),
		mutil.ServiceKeyServiceVersion(provider.Version),
	)
	depRule, err := getDeps(ctx, filter)
	if err != nil {
		return nil, err
	}
	return depRule.Dep, nil
}

func getDeps(ctx context.Context, filter interface{}) (*model.DependencyRule, error) {
	findRes := mongo.GetClient().GetDB().Collection(model.CollectionDep).FindOne(ctx, filter)

	var depRule *model.DependencyRule
	if findRes.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err := findRes.Decode(&depRule)
	if err != nil {
		return nil, err
	}
	return depRule, nil
}

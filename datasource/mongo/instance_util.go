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

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func ExistInstance(ctx context.Context, serviceID string, instanceID string) (bool, error) {
	inst, ok := cache.GetInstance(ctx, serviceID, instanceID)
	if ok && inst != nil {
		return true, nil
	}

	return dao.ExistInstance(ctx, serviceID, instanceID)
}

func GetInstances(ctx context.Context) ([]*model.Instance, error) {
	insts, ok := cache.GetInstances(ctx)
	if ok {
		return insts, nil
	}
	filter := mutil.NewBasicFilter(ctx)
	return dao.GetInstances(ctx, filter)

}

func CountInstance(ctx context.Context, serviceID string) (int64, error) {
	count, ok := cache.CountInstances(ctx, serviceID)
	if ok {
		return int64(count), nil
	}
	filter := mutil.NewDomainProjectFilter(util.ParseDomain(ctx), util.ParseProject(ctx), mutil.InstanceServiceID(serviceID))
	return dao.CountInstance(ctx, filter)
}

func GetAllInstancesOfOneService(ctx context.Context, serviceID string) ([]*discovery.MicroServiceInstance, error) {
	inst, ok := cache.GetMicroServiceInstancesByID(ctx, serviceID)
	if ok && inst != nil {
		return inst, nil
	}

	return dao.GetMicroServiceInstancesByID(ctx, serviceID)
}

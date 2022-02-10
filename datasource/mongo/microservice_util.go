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

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetServiceByID(ctx context.Context, serviceID string) (*model.Service, error) {
	svc, exist := cache.GetServiceByID(ctx, serviceID)
	if exist && svc != nil {
		return svc, nil
	}

	return dao.GetServiceByID(ctx, serviceID)
}

func GetServiceByIDAcrossDomain(ctx context.Context, serviceID string) (*model.Service, error) {
	svc, exist := cache.GetServiceByIDAcrossDomain(ctx, serviceID)
	if exist && svc != nil {
		return svc, nil
	}
	providerDomain, providerProject := util.ParseTargetDomain(ctx), util.ParseTargetProject(ctx)
	filter := mutil.NewDomainProjectFilter(providerDomain, providerProject, mutil.ServiceServiceID(serviceID))
	return dao.GetService(ctx, filter)
}

func ServiceExistID(ctx context.Context, serviceID string) (bool, error) {
	svc, exist := cache.GetServiceByID(ctx, serviceID)
	if exist && svc != nil {
		return true, nil
	}
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(serviceID))
	num, err := mongo.GetClient().GetDB().Collection(model.CollectionService).CountDocuments(ctx, filter)
	return num != 0, err
}

func GetAllMicroServicesByDomainProject(ctx context.Context) ([]*discovery.MicroService, error) {
	services, exist := cache.GetMicroServicesByDomainProject(util.ParseDomainProject(ctx))
	if exist {
		return services, nil
	}

	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{model.ColumnDomain: domain, model.ColumnProject: project}
	return dao.GetMicroServices(ctx, filter)
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
	serviceID, exist := cache.GetServiceID(ctx, key)
	if exist && len(serviceID) != 0 {
		return serviceID, nil
	}
	return dao.GetServiceID(ctx, key)
}

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
	"errors"

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

func GetServiceByID(ctx context.Context, serviceID string) (*model.Service, error) {
	return GetService(ctx, util.NewBasicFilter(ctx, util.ServiceServiceID(serviceID)))
}

func GetService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.Service, error) {
	result := mongo.GetClient().GetDB().Collection(model.CollectionService).FindOne(ctx, filter, opts...)
	var svc *model.Service
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err := result.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
	filter := util.NewBasicFilter(
		ctx,
		util.ServiceEnv(key.Environment),
		util.ServiceAppID(key.AppId),
		util.ServiceServiceName(key.ServiceName),
		util.ServiceVersion(key.Version),
	)
	id, err := getServiceID(ctx, filter)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return "", err
	}
	if len(id) == 0 && len(key.Alias) != 0 {
		filter = util.NewBasicFilter(
			ctx,
			util.ServiceEnv(key.Environment),
			util.ServiceAppID(key.AppId),
			util.ServiceAlias(key.Alias),
			util.ServiceVersion(key.Version),
		)
		return getServiceID(ctx, filter)
	}
	return id, nil
}

func getServiceID(ctx context.Context, filter bson.M) (serviceID string, err error) {
	svc, err := GetService(ctx, filter)
	if err != nil {
		return
	}
	if svc != nil {
		serviceID = svc.Service.ServiceId
		return
	}
	return
}

func GetServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.Service, error) {
	res, err := mongo.GetClient().GetDB().Collection(model.CollectionService).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*model.Service
	for res.Next(ctx) {
		var tmp *model.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp)
	}
	return services, nil
}

func GetMicroServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroService, error) {
	res, err := mongo.GetClient().GetDB().Collection(model.CollectionService).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*discovery.MicroService
	for res.Next(ctx) {
		var tmp model.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.Service)
	}
	return services, nil
}

func UpdateService(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) error {
	res := mongo.GetClient().GetDB().Collection(model.CollectionService).FindOneAndUpdate(ctx, filter, update, opts...)
	if res.Err() != nil {
		// means no doc find, if the operation is update,should return err
		return ErrNoDocuments
	}
	return nil
}

func CountService(ctx context.Context, filter interface{}) (int64, error) {
	count, err := mongo.GetClient().GetDB().Collection(model.CollectionService).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

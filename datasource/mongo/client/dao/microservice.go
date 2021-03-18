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
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
)

func GetService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.Service, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var svc *model.Service
	if result.Err() != nil {
		//not get any service,not db err
		return nil, datasource.ErrNoData
	}
	err = result.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func GetServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.Service, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
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
	res, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
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
	res, err := client.GetMongoClient().FindOneAndUpdate(ctx, model.CollectionService, filter, update, opts...)
	if err != nil {
		return err
	}
	if res.Err() != nil {
		// means no doc find, if the operation is update,should return err
		return client.ErrNoDocuments
	}
	return nil
}

func GetServicesVersions(ctx context.Context, filter interface{}) ([]string, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter)
	if err != nil {
		return nil, nil
	}
	var versions []string
	for res.Next(ctx) {
		var tmp model.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		versions = append(versions, tmp.Service.Version)
	}
	return versions, nil
}

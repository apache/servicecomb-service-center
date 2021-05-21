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

	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
)

func GetInstance(ctx context.Context, filter interface{}) (*model.Instance, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, model.CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instance *model.Instance
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	err = findRes.Decode(&instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func GetInstances(ctx context.Context, filter interface{}) ([]*model.Instance, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instances []*model.Instance
	for res.Next(ctx) {
		var tmp *model.Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp)
	}
	return instances, nil
}

func GetMicroServiceInstancesByID(ctx context.Context, serviceID string) ([]*discovery.MicroServiceInstance, error) {
	filter := mutil.NewFilter(mutil.InstanceServiceID(serviceID))
	option := &options.FindOptions{Sort: bson.M{mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}): -1}}
	return GetMicroServiceInstances(ctx, filter, option)
}

func GetMicroServiceInstances(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroServiceInstance, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionInstance, filter, opts...)
	if err != nil {
		return nil, err
	}
	var instances []*discovery.MicroServiceInstance
	for res.Next(ctx) {
		var tmp model.Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp.Instance)
	}
	return instances, nil
}

func CountInstance(ctx context.Context, filter interface{}) (int64, error) {
	count, err := client.GetMongoClient().Count(ctx, model.CollectionInstance, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateInstance(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) *discovery.Error {
	_, err := client.GetMongoClient().Update(ctx, model.CollectionInstance, filter, update, opts...)
	if err != nil {
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func ExistInstance(ctx context.Context, serviceID string, instanceID string) (bool, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID), mutil.InstanceInstanceID(instanceID))
	return client.GetMongoClient().DocExist(ctx, model.CollectionInstance, filter)
}

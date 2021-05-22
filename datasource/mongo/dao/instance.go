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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/dao/util"
)

func UpdateInstance(ctx context.Context, filter interface{}, updateFilter interface{}) error {
	result, err := client.GetMongoClient().FindOneAndUpdate(ctx, CollectionInstance, filter, updateFilter)
	if err != nil {
		return err
	}
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func GetInstance(ctx context.Context, filter interface{}) (*Instance, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instance *Instance
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	err = findRes.Decode(&instance)
	if err != nil {
		log.Error("decode instance failed", err)
		return nil, err
	}
	return instance, nil
}

func GetInstances(ctx context.Context, filter interface{}) ([]*Instance, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instances []*Instance
	for res.Next(ctx) {
		var tmp *Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp)
	}
	return instances, nil
}

func DeleteInstance(ctx context.Context, filter interface{}) error {
	res, err := client.GetMongoClient().DeleteOne(ctx, CollectionInstance, filter)
	if err != nil {
		log.Error("failed to delete instance", err)
		return err
	}
	log.Info(fmt.Sprintf("delete from mongodb:%+v", res))
	return nil
}

func GetMicroServiceInstancesByID(ctx context.Context, serviceID string) ([]*discovery.MicroServiceInstance, error) {
	filter := bson.D{
		{mutil.ConnectWithDot([]string{ColumnInstance, ColumnServiceID}), serviceID},
	}
	option := &options.FindOptions{Sort: bson.D{{mutil.ConnectWithDot([]string{ColumnInstance, ColumnVersion}), -1}}}
	return GetMicroServiceInstances(ctx, filter, option)
}

func GetMicroServiceInstances(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroServiceInstance, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionInstance, filter, opts...)
	if err != nil {
		return nil, err
	}
	var instances []*discovery.MicroServiceInstance
	for res.Next(ctx) {
		var tmp Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp.Instance)
	}
	return instances, nil
}

func CountInstance(ctx context.Context, filter interface{}) (int64, error) {
	count, err := client.GetMongoClient().Count(ctx, CollectionInstance, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func ExistInstance(ctx context.Context, serviceID string, instanceID string) (bool, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.InstanceServiceID(serviceID), mutil.InstanceInstanceID(instanceID))
	return client.GetMongoClient().DocExist(ctx, CollectionInstance, filter)
}

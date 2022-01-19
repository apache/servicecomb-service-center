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
	"github.com/go-chassis/cari/pkg/errsvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

func GetInstances(ctx context.Context, filter interface{}) ([]*model.Instance, error) {
	res, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).Find(ctx, filter)
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
	filter := util.NewFilter(util.InstanceServiceID(serviceID))
	option := &options.FindOptions{Sort: bson.M{util.ConnectWithDot([]string{model.ColumnInstance, model.ColumnVersion}): -1}}
	return GetMicroServiceInstances(ctx, filter, option)
}

func GetMicroServiceInstances(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroServiceInstance, error) {
	res, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).Find(ctx, filter, opts...)
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
	count, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateInstance(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) *errsvc.Error {
	_, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func ExistInstance(ctx context.Context, serviceID string, instanceID string) (bool, error) {
	filter := util.NewBasicFilter(ctx, util.InstanceServiceID(serviceID), util.InstanceInstanceID(instanceID))
	result := mongo.GetClient().GetDB().Collection(model.CollectionInstance).FindOne(ctx, filter)
	if result.Err() != nil {
		return false, nil
	}
	return true, nil
}

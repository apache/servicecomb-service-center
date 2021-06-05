/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package heartbeatcache

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func getInstance(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.Instance, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionInstance, filter, opts...)
	if err != nil {
		log.Error("failed to query instance", err)
		return nil, err
	}
	var ins model.Instance
	err = result.Decode(&ins)
	if err != nil {
		log.Error("decode instance failed", err)
		return nil, err
	}
	return &ins, nil
}

func deleteInstance(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) error {
	res, err := client.GetMongoClient().DeleteOne(ctx, model.CollectionInstance, filter, opts...)
	if err != nil {
		log.Error("failed to clean instance", err)
		return err
	}
	log.Info(fmt.Sprintf("delete from mongodb:%+v", res))
	return nil
}

func findAndUpdateInstance(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) error {
	result, err := client.GetMongoClient().FindOneAndUpdate(ctx, model.CollectionInstance, filter, update, opts...)
	if err != nil {
		log.Error("failed to update refresh time of instance", err)
		return err
	}
	return result.Err()
}

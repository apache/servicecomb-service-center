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

package checker

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func updateInstanceRefreshTime(ctx context.Context, serviceID string, instanceID string) error {
	filter := bson.M{
		mongo.StringBuilder([]string{model.ColumnInstance, model.ColumnServiceID}):  serviceID,
		mongo.StringBuilder([]string{model.ColumnInstance, model.ColumnInstanceID}): instanceID,
	}
	update := bson.M{
		"$set": bson.M{model.ColumnRefreshTime: time.Now()},
	}
	result, err := client.GetMongoClient().FindOneAndUpdate(ctx, model.CollectionInstance, filter, update)
	if err != nil {
		log.Error("failed to update refresh time of instance: ", err)
		return err
	}
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

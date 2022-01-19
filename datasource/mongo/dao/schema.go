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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

func GetSchema(ctx context.Context, filter interface{}) (*model.Schema, error) {
	result := mongo.GetClient().GetDB().Collection(model.CollectionSchema).FindOne(ctx, filter)
	if result.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	var schema *model.Schema
	err := result.Decode(&schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func GetSchemas(ctx context.Context, filter interface{}) ([]*discovery.Schema, error) {
	getRes, err := mongo.GetClient().GetDB().Collection(model.CollectionSchema).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var schemas []*discovery.Schema
	for getRes.Next(ctx) {
		var tmp *model.Schema
		err = getRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, &discovery.Schema{
			SchemaId: tmp.SchemaID,
			Summary:  tmp.SchemaSummary,
			Schema:   tmp.Schema,
		})
	}
	return schemas, nil
}

func SchemaSummaryExist(ctx context.Context, serviceID, schemaID string) (bool, error) {
	filter := util.NewBasicFilter(ctx, util.ServiceID(serviceID), util.SchemaID(schemaID))
	res := mongo.GetClient().GetDB().Collection(model.CollectionSchema).FindOne(ctx, filter)
	if res.Err() != nil {
		return false, nil
	}
	var s model.Schema
	err := res.Decode(&s)
	if err != nil {
		return false, err
	}
	return len(s.SchemaSummary) != 0, nil
}

func CountSchema(ctx context.Context, serviceID string) (int64, error) {
	filter := util.NewBasicFilter(ctx, util.ServiceID(serviceID))
	return mongo.GetClient().GetDB().Collection(model.CollectionSchema).CountDocuments(ctx, filter)
}

func UpdateSchema(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	_, err := mongo.GetClient().GetDB().Collection(model.CollectionSchema).UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return err
	}
	return nil
}

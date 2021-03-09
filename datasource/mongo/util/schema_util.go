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

package util

import (
	"context"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetSchema(ctx context.Context, filter bson.M) (*db.Schema, error) {
	findRes, err := client.GetMongoClient().FindOne(ctx, db.CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	var schema *db.Schema
	err = findRes.Decode(&schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func GetSchemas(ctx context.Context, filter bson.M) ([]*discovery.Schema, error) {
	getRes, err := client.GetMongoClient().Find(ctx, db.CollectionSchema, filter)
	if err != nil {
		return nil, err
	}
	var schemas []*discovery.Schema
	for getRes.Next(ctx) {
		var tmp *db.Schema
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

func GeneratorSchemaFilter(ctx context.Context, serviceID, schemaID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	return bson.M{db.ColumnDomain: domain, db.ColumnProject: project, db.ColumnServiceID: serviceID, db.ColumnSchemaID: schemaID}
}

func SchemaSummaryExist(ctx context.Context, serviceID, schemaID string) (bool, error) {
	res, err := client.GetMongoClient().FindOne(ctx, db.CollectionSchema, GeneratorSchemaFilter(ctx, serviceID, schemaID))
	if err != nil {
		return false, err
	}
	if res.Err() != nil {
		return false, nil
	}
	var s db.Schema
	err = res.Decode(&s)
	if err != nil {
		return false, err
	}
	return len(s.SchemaSummary) != 0, nil
}

func DeleteSchema(ctx context.Context, filter interface{}) error {
	res, err := client.GetMongoClient().DocDelete(ctx, db.CollectionSchema, filter)
	if err != nil {
		return err
	}
	if !res {
		return ErrDeleteSchemaFailed
	}
	return nil
}

func UpdateSchema(ctx context.Context, filter interface{}, m bson.M, opts ...*options.FindOneAndUpdateOptions) error {
	_, err := client.GetMongoClient().FindOneAndUpdate(ctx, db.CollectionSchema, filter, m, opts...)
	if err != nil {
		return err
	}
	return nil
}

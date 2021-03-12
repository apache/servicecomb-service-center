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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
)

type Option func(filter bson.M)

func Domain(domain string) Option {
	return func(filter bson.M) {
		filter[model.ColumnDomain] = domain
	}
}

func Project(project string) Option {
	return func(filter bson.M) {
		filter[model.ColumnProject] = project
	}
}

func NewFilter(options ...func(filter bson.M)) bson.M {
	filter := bson.M{}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func InstanceServiceID(serviceID string) Option {
	return func(filter bson.M) {
		filter[StringBuilder([]string{model.ColumnInstance, model.ColumnServiceID})] = serviceID
	}
}

func InstanceInstanceID(instanceID string) Option {
	return func(filter bson.M) {
		filter[StringBuilder([]string{model.ColumnInstance, model.ColumnInstanceID})] = instanceID
	}
}

func BuildIndexDoc(keys ...string) mongo.IndexModel {
	keysDoc := bsonx.Doc{}
	for _, key := range keys {
		keysDoc = keysDoc.Append(key, bsonx.Int32(1))
	}
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	return index
}

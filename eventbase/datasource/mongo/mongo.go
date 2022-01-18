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

package mongo

import (
	dmongo "github.com/go-chassis/cari/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/task"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/tombstone"
)

func init() {
	datasource.RegisterPlugin("mongo", NewDatasource)
}

type Datasource struct {
	taskDao      datasource.TaskDao
	tombstoneDao datasource.TombstoneDao
}

func (d *Datasource) TaskDao() datasource.TaskDao {
	return d.taskDao
}

func (d *Datasource) TombstoneDao() datasource.TombstoneDao {
	return d.tombstoneDao
}

func NewDatasource() datasource.DataSource {
	ensureDB()
	return &Datasource{taskDao: &task.Dao{}, tombstoneDao: &tombstone.Dao{}}
}

func ensureDB() {
	ensureTask()
	ensureTombstone()
}

func ensureTask() {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{model.ColumnID, model.ColumnDomain, model.ColumnProject, model.ColumnTimestamp},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	dmongo.EnsureCollection(model.CollectionTask, validator, []mongo.IndexModel{buildIndexDoc(
		model.ColumnDomain, model.ColumnProject, model.ColumnID, model.ColumnTimestamp)})
}

func ensureTombstone() {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{model.ColumnResourceID, model.ColumnDomain, model.ColumnProject, model.ColumnResourceType},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	dmongo.EnsureCollection(model.CollectionTombstone, validator, []mongo.IndexModel{buildIndexDoc(
		model.ColumnDomain, model.ColumnProject, model.ColumnResourceID, model.ColumnResourceType)})
}

func buildIndexDoc(keys ...string) mongo.IndexModel {
	keysDoc := bsonx.Doc{}
	for _, key := range keys {
		keysDoc = keysDoc.Append(key, bsonx.Int32(1))
	}
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	return index
}

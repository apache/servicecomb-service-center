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

package tombstone

import (
	"context"

	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/mongo/model"
	emodel "github.com/apache/servicecomb-service-center/eventbase/model"
)

type Dao struct {
}

func (d *Dao) Get(ctx context.Context, req *emodel.GetTombstoneRequest) (*sync.Tombstone, error) {
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTombstone)
	filter := bson.M{model.ColumnDomain: req.Domain, model.ColumnProject: req.Project,
		model.ColumnResourceType: req.ResourceType, model.ColumnResourceID: req.ResourceID}
	result := collection.FindOne(ctx, filter)
	if result != nil && result.Err() != nil {
		openlog.Error("fail to get tombstone" + result.Err().Error())
		return nil, result.Err()
	}
	if result == nil {
		openlog.Error(datasource.ErrTombstoneNotExists.Error())
		return nil, datasource.ErrTombstoneNotExists
	}
	var tombstone sync.Tombstone

	err := result.Decode(&tombstone)
	if err != nil {
		openlog.Error("fail to decode tombstone" + err.Error())
		return nil, err
	}
	return &tombstone, nil
}

func (d *Dao) Create(ctx context.Context, tombstone *sync.Tombstone) (*sync.Tombstone, error) {
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTombstone)
	_, err := collection.InsertOne(ctx, tombstone)
	if err != nil {
		openlog.Error("fail to create tombstone" + err.Error())
		return nil, err
	}
	return tombstone, nil
}

func (d *Dao) Delete(ctx context.Context, tombstones ...*sync.Tombstone) error {
	tombstonesIDs := make([]string, len(tombstones))
	filter := bson.A{}
	for i, tombstone := range tombstones {
		tombstonesIDs[i] = tombstone.ResourceID
		dFilter := bson.D{
			{model.ColumnResourceID, tombstone.ResourceID},
			{model.ColumnResourceType, tombstone.ResourceType},
			{model.ColumnDomain, tombstone.Domain},
			{model.ColumnProject, tombstone.Project},
		}
		filter = append(filter, dFilter)
	}
	var deleteFunc = func(sessionContext mongo.SessionContext) error {
		collection := client.GetMongoClient().GetDB().Collection(model.CollectionTombstone)
		_, err := collection.DeleteMany(sessionContext, bson.M{"$or": filter})
		return err
	}
	err := client.GetMongoClient().ExecTxn(ctx, deleteFunc)
	if err != nil {
		openlog.Error(err.Error())
	}
	return err
}

func (d *Dao) List(ctx context.Context, options ...datasource.TombstoneFindOption) ([]*sync.Tombstone, error) {
	opts := datasource.NewTombstoneFindOptions()
	for _, o := range options {
		o(&opts)
	}
	collection := client.GetMongoClient().GetDB().Collection(model.CollectionTombstone)
	filter := bson.M{}
	if opts.Domain != "" {
		filter[model.ColumnDomain] = opts.Domain
	}
	if opts.Project != "" {
		filter[model.ColumnProject] = opts.Project
	}
	if opts.ResourceType != "" {
		filter[model.ColumnResourceType] = opts.ResourceType
	}
	if opts.BeforeTimestamp != 0 {
		filter[model.ColumnTimestamp] = bson.M{"$lte": opts.BeforeTimestamp}
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	tombstones := make([]*sync.Tombstone, 0)
	for cur.Next(ctx) {
		tombstone := &sync.Tombstone{}
		if err := cur.Decode(tombstone); err != nil {
			openlog.Error("decode to tombstone error: " + err.Error())
			return nil, err
		}
		tombstones = append(tombstones, tombstone)
	}
	return tombstones, nil

}

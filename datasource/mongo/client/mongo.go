// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-chassis/v2/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	MongoDB             = "servicecenter"
	MongoCheckDelay     = 2 * time.Second
	HeathChekRetryTimes = 3
)

var (
	mc *MongoClient
)

type MongoClient struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
	db          *mongo.Database
	dbconfig    storage.Options

	err       chan error
	ready     chan struct{}
	goroutine *gopool.Pool
}

func GetMongoClient() *MongoClient {
	return mc
}

func NewMongoClient(config storage.Options, cols []string) {
	inst := &MongoClient{}
	if err := inst.Initialize(config, cols); err != nil {
		log.Errorf(err, "failed to init mongodb")
		inst.err <- err
	}
	mc = inst
}

func (mc *MongoClient) Initialize(config storage.Options, cols []string) (err error) {
	mc.err = make(chan error, 1)
	mc.ready = make(chan struct{})
	mc.goroutine = gopool.New(context.Background())
	mc.dbconfig = config
	err = mc.newClient(context.Background(), cols)
	if err != nil {
		return
	}
	mc.StartHealthCheck()
	close(mc.ready)
	return nil
}

func (mc *MongoClient) Err() <-chan error {
	return mc.err
}

func (mc *MongoClient) Ready() <-chan struct{} {
	return mc.ready
}

func (mc *MongoClient) Close() {
	if mc.client != nil {
		if err := mc.client.Disconnect(context.TODO()); err != nil {
			log.Errorf(err, "[close mongo client] failed disconnect the mongo client")
		}
	}
}

func (mc *MongoClient) StartHealthCheck() {
	mc.goroutine.Do(mc.HealthCheck)
}

func (mc *MongoClient) HealthCheck(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			mc.Close()
			return
		case <-time.After(MongoCheckDelay):
			for i := 0; i < HeathChekRetryTimes; i++ {
				err := mc.client.Ping(context.Background(), nil)
				if err == nil {
					break
				}
				log.Errorf(err, "retry to connect to mongodb %s after %s", mc.dbconfig.URI, MongoCheckDelay)
				select {
				case <-ctx.Done():
					mc.Close()
					return
				case <-time.After(MongoCheckDelay):
				}
			}
		}
	}
}

func (mc *MongoClient) newClient(ctx context.Context, cols []string) (err error) {
	clientOptions := options.Client().ApplyURI(mc.dbconfig.URI)
	mc.client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		if derr := mc.client.Disconnect(ctx); derr != nil {
			log.Errorf(derr, "[init mongo client] failed to disconnect mongo client ")
		}
		return
	}
	mc.db = mc.client.Database(MongoDB)
	if mc.db == nil {
		return errors.New("open db failed")
	}
	mc.collections = make(map[string]*mongo.Collection)
	for _, name := range cols {
		mc.collections[name] = mc.db.Collection(name)
		if mc.collections[name] == nil {
			err = errors.New("failed to open table")
			return
		}
	}
	return nil
}

func (mc *MongoClient) Insert(ctx context.Context, Table string, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.InsertOne(ctx, document, opts...)
}

func (mc *MongoClient) BatchInsert(ctx context.Context, Table string, document []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.InsertMany(ctx, document, opts...)
}

func (mc *MongoClient) Delete(ctx context.Context, Table string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.DeleteMany(ctx, filter, opts...)
}

func (mc *MongoClient) BatchDelete(ctx context.Context, Table string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.BulkWrite(ctx, models, opts...)
}

func (mc *MongoClient) Update(ctx context.Context, Table string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.UpdateMany(ctx, filter, update, opts...)
}

func (mc *MongoClient) FindOneAndUpdate(ctx context.Context, Table string, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) (*mongo.SingleResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.FindOneAndUpdate(ctx, filter, update, opts...), nil
}

func (mc *MongoClient) BatchUpdate(ctx context.Context, Table string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.BulkWrite(ctx, models, opts...)
}

func (mc *MongoClient) Find(ctx context.Context, Table string, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.Find(ctx, filter, opts...)
}

func (mc *MongoClient) FindOne(ctx context.Context, Table string, filter interface{}, opts ...*options.FindOneOptions) (*mongo.SingleResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.FindOne(ctx, filter, opts...), nil
}

func (mc *MongoClient) Count(ctx context.Context, Table string, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	col := mc.collections[Table]
	if col == nil {
		return 0, ErrCollectionsNil
	}
	return col.CountDocuments(ctx, filter, opts...)
}

func (mc *MongoClient) Aggregate(ctx context.Context, Table string, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.Aggregate(ctx, pipeline, opts...)
}

func (mc *MongoClient) Watch(ctx context.Context, Table string, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.Watch(ctx, pipeline, opts...)
}

func (mc *MongoClient) StartSession(ctx context.Context) (mongo.Session, error) {
	return mc.client.StartSession()
}

func (mc *MongoClient) DocExist(ctx context.Context, table string, filter bson.M) (bool, error) {
	num, err := mc.Count(ctx, table, filter)
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

func (mc *MongoClient) DocUpdate(ctx context.Context, table string, filter interface{}, m bson.M, opts ...*options.FindOneAndUpdateOptions) error {
	res, err := mc.FindOneAndUpdate(ctx, table, filter, m, opts...)
	if err != nil {
		return err
	}
	if res.Err() != nil {
		// means no doc find, if the operation is update,should return err
		return res.Err()
	}
	return nil
}

func (mc *MongoClient) DocDelete(ctx context.Context, table string, filter interface{}) (bool, error) {
	res, err := mc.DeleteOne(ctx, table, filter)
	if err != nil {
		return false, err
	}
	return res.DeletedCount != 0, nil
}

func (mc *MongoClient) DocDeleteMany(ctx context.Context, table string, filter interface{}) error {
	_, err := mc.Delete(ctx, table, filter)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MongoClient) DeleteOne(ctx context.Context, Table string, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	col := mc.collections[Table]
	if col == nil {
		return nil, ErrCollectionsNil
	}
	return col.DeleteOne(ctx, filter, opts...)
}

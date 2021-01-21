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
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/go-chassis/v2/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func init() {
	datasource.Install("mongo", NewDataSource)
}

type DataSource struct {
	// SchemaEditable determines whether schema modification is allowed for
	SchemaEditable bool
	// TTL options
	ttlFromEnv int64
}

func NewDataSource(opts datasource.Options) (datasource.DataSource, error) {
	// TODO: construct a reasonable DataSource instance

	inst := &DataSource{
		SchemaEditable: opts.SchemaEditable,
		ttlFromEnv:     opts.InstanceTTL,
	}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return nil, err
	}
	return inst, nil
}

func (ds *DataSource) initialize() error {
	var err error
	// init heartbeat plugins
	err = ds.initPlugins()
	if err != nil {
		return err
	}
	// init mongo client
	err = ds.initClient()
	if err != nil {
		return err
	}
	err = ds.createIndexes()
	if err != nil {
		return err
	}
	// init mongo cache
	ds.initStore()
	return nil
}

func (ds *DataSource) initPlugins() error {
	kind := config.GetString("registry.heartbeat.kind", "heartbeatchecker", config.WithStandby("heartbeat_plugin"))
	err := heartbeat.Init(heartbeat.Options{PluginImplName: heartbeat.ImplName(kind)})
	if err != nil {
		log.Fatal("heartbeat init failed", err)
		return err
	}
	return nil
}

func (ds *DataSource) initClient() error {
	uri := config.GetString("registry.mongo.cluster.uri", "mongodb://localhost:27017", config.WithStandby("manager_cluster"))
	cfg := storage.NewConfig(uri)
	client.NewMongoClient(cfg)
	select {
	case err := <-client.GetMongoClient().Err():
		return err
	case <-client.GetMongoClient().Ready():
		return nil
	}
}

//{Key: StringBuilder([]string{ColumnServiceInfo, ColumnAlias}), Value: bsonx.Int32(1)}
func (ds *DataSource) createIndexes() (err error) {
	err = client.GetMongoClient().CreateIndexes(context.TODO(), CollectionService, []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: StringBuilder([]string{ColumnServiceInfo, ColumnServiceID}), Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	}, {
		Keys: bsonx.Doc{{Key: StringBuilder([]string{ColumnServiceInfo, ColumnAppID}), Value: bsonx.Int32(1)},
			{Key: StringBuilder([]string{ColumnServiceInfo, ColumnServiceName}), Value: bsonx.Int32(1)},
			{Key: StringBuilder([]string{ColumnServiceInfo, ColumnEnv}), Value: bsonx.Int32(1)},
			{Key: StringBuilder([]string{ColumnServiceInfo, ColumnVersion}), Value: bsonx.Int32(1)},
			{Key: ColumnDomain, Value: bsonx.Int32(1)},
			{Key: ColumnProject, Value: bsonx.Int32(1)},
		},
		Options: options.Index().SetUnique(true).SetName(serviceAppIDNameEnvVersionIndex),
	}})
	if err != nil {
		return
	}
	err = client.GetMongoClient().CreateIndexes(context.TODO(), CollectionInstance, []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: StringBuilder([]string{ColumnInstanceInfo, ColumnInstanceID}), Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	}, {
		Keys: bsonx.Doc{{Key: StringBuilder([]string{ColumnInstanceID, ColumnServiceID}), Value: bsonx.Int32(1)}},
	}, {
		Keys:    bsonx.Doc{{Key: ColumnRefreshTime, Value: bsonx.Int32(1)}},
		Options: options.Index().SetExpireAfterSeconds(60),
	}})
	if err != nil {
		return
	}
	err = client.GetMongoClient().CreateIndexes(context.TODO(), CollectionSchema, []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{Key: ColumnDomain, Value: bsonx.Int32(1)},
			{Key: ColumnProject, Value: bsonx.Int32(1)},
			{Key: ColumnServiceID, Value: bsonx.Int32(1)}},
	}})
	if err != nil {
		return
	}
	err = client.GetMongoClient().CreateIndexes(context.TODO(), CollectionRule, []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{Key: ColumnDomain, Value: bsonx.Int32(1)},
			{Key: ColumnProject, Value: bsonx.Int32(1)},
			{Key: ColumnServiceID, Value: bsonx.Int32(1)}},
	}})
	if err != nil {
		return
	}
	err = client.GetMongoClient().CreateIndexes(context.TODO(), CollectionDep, []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{Key: ColumnDomain, Value: bsonx.Int32(1)},
			{Key: ColumnProject, Value: bsonx.Int32(1)},
			{Key: ColumnServiceKey, Value: bsonx.Int32(1)}},
	}})
	if err != nil {
		return
	}
	return
}

func (ds *DataSource) initStore() {
	if !config.GetRegistry().EnableCache {
		log.Debug("cache is disabled")
		return
	}
	sd.Store().Run()
	<-sd.Store().Ready()
}

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
	// create db index and validator
	EnsureDB()
	// init cache
	ds.initStore()
	return nil
}

func (ds *DataSource) initPlugins() error {
	kind := config.GetString("registry.mongo.heartbeat.kind", "cache")
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

func EnsureDB() {
	EnsureService()
	EnsureInstance()
	EnsureRule()
	EnsureSchema()
	EnsureDep()
}

func EnsureService() {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), CollectionService, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	serviceIDIndex := BuildIndexDoc(
		StringBuilder([]string{ColumnService, ColumnServiceID}))
	serviceIDIndex.Options = options.Index().SetUnique(true)

	serviceIndex := BuildIndexDoc(
		StringBuilder([]string{ColumnService, ColumnAppID}),
		StringBuilder([]string{ColumnService, ColumnServiceName}),
		StringBuilder([]string{ColumnService, ColumnEnv}),
		StringBuilder([]string{ColumnService, ColumnVersion}),
		ColumnDomain,
		ColumnProject)
	serviceIndex.Options = options.Index().SetUnique(true)

	var serviceIndexs []mongo.IndexModel
	serviceIndexs = append(serviceIndexs, serviceIDIndex, serviceIndex)

	err = client.GetMongoClient().CreateIndexes(context.Background(), CollectionService, serviceIndexs)
	if err != nil {
		log.Fatal("failed to create service collection indexs", err)
		return
	}
}

func EnsureInstance() {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), CollectionInstance, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	instanceIndex := BuildIndexDoc(ColumnRefreshTime)
	instanceIndex.Options = options.Index().SetExpireAfterSeconds(60)

	instanceServiceIndex := BuildIndexDoc(StringBuilder([]string{ColumnInstanceID, ColumnServiceID}))

	var instanceIndexs []mongo.IndexModel
	instanceIndexs = append(instanceIndexs, instanceIndex, instanceServiceIndex)

	err = client.GetMongoClient().CreateIndexes(context.Background(), CollectionInstance, instanceIndexs)
	if err != nil {
		log.Fatal("failed to create instance collection indexs", err)
		return
	}
}

func EnsureSchema() {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), CollectionSchema, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	schemaServiceIndex := BuildIndexDoc(
		ColumnDomain,
		ColumnProject,
		ColumnServiceID)

	var schemaIndexs []mongo.IndexModel
	schemaIndexs = append(schemaIndexs, schemaServiceIndex)

	err = client.GetMongoClient().CreateIndexes(context.Background(), CollectionSchema, schemaIndexs)
	if err != nil {
		log.Fatal("failed to create schema collection indexs", err)
		return
	}
}

func EnsureRule() {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), CollectionRule, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	ruleServiceIndex := BuildIndexDoc(
		ColumnDomain,
		ColumnProject,
		ColumnServiceID)

	var ruleIndexs []mongo.IndexModel
	ruleIndexs = append(ruleIndexs, ruleServiceIndex)

	err = client.GetMongoClient().CreateIndexes(context.Background(), CollectionRule, ruleIndexs)
	if err != nil {
		log.Fatal("failed to create rule collection indexs", err)
		return
	}
}

func EnsureDep() {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), CollectionDep, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	depServiceIndex := BuildIndexDoc(
		ColumnDomain,
		ColumnProject,
		ColumnServiceKey)

	var depIndexs []mongo.IndexModel
	depIndexs = append(depIndexs, depServiceIndex)

	err = client.GetMongoClient().CreateIndexes(context.Background(), CollectionDep, depIndexs)
	if err != nil {
		log.Fatal("failed to create dep collection indexs", err)
		return
	}
}

func wrapCreateCollectionError(err error) {
	if err != nil {
		// commandError can be returned by any operation
		cmdErr, ok := err.(mongo.CommandError)
		if ok && cmdErr.Code == client.CollectionsExists {
			return
		}
		log.Fatal("failed to create collection with validation", err)
	}
}

func (ds *DataSource) initStore() {
	if !config.GetRegistry().EnableCache {
		log.Debug("cache is disabled")
		return
	}
	sd.Store().Run()
	<-sd.Store().Ready()
}

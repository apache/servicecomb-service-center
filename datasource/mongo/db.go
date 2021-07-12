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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureDB() {
	EnsureService()
	EnsureInstance()
	EnsureRule()
	EnsureSchema()
	EnsureDep()
	EnsureAccountLock()
}

func EnsureService() {
	serviceIDIndex := mutil.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceID}))
	serviceIDIndex.Options = options.Index().SetUnique(true)

	serviceIndex := mutil.BuildIndexDoc(
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnAppID}),
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceName}),
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnEnv}),
		mutil.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}),
		model.ColumnDomain,
		model.ColumnProject)
	serviceIndex.Options = options.Index().SetUnique(true)

	var serviceIndexes []mongo.IndexModel
	serviceIndexes = append(serviceIndexes, serviceIDIndex, serviceIndex)

	EnsureCollection(model.CollectionService, serviceIndexes)
}

func EnsureInstance() {
	instanceIndex := mutil.BuildIndexDoc(model.ColumnRefreshTime)
	instanceIndex.Options = options.Index().SetExpireAfterSeconds(defaultExpireTime)

	instanceServiceIndex := mutil.BuildIndexDoc(mutil.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}))

	var instanceIndexes []mongo.IndexModel
	instanceIndexes = append(instanceIndexes, instanceIndex, instanceServiceIndex)

	EnsureCollection(model.CollectionInstance, instanceIndexes)
}

func EnsureSchema() {
	EnsureCollection(model.CollectionSchema, []mongo.IndexModel{mutil.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		model.ColumnServiceID)})
}

func EnsureRule() {
	EnsureCollection(model.CollectionRule, []mongo.IndexModel{mutil.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		model.ColumnServiceID)})
}

func EnsureDep() {
	EnsureCollection(model.CollectionDep, []mongo.IndexModel{mutil.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		model.ColumnServiceKey)})
}

func EnsureAccountLock() {
	EnsureCollection(model.CollectionAccountLock, []mongo.IndexModel{
		mutil.BuildIndexDoc(model.ColumnAccountLockKey)})
}

func EnsureCollection(col string, indexes []mongo.IndexModel) {
	err := client.GetMongoClient().GetDB().CreateCollection(context.Background(), col, options.CreateCollection().SetValidator(nil))
	wrapCreateCollectionError(err)

	err = client.GetMongoClient().CreateIndexes(context.Background(), col, indexes)
	wrapCreateIndexesError(err)
}

func wrapCreateCollectionError(err error) {
	if err != nil {
		if client.IsCollectionsExist(err) {
			log.Warn(fmt.Sprintf("collection already exist, err type: %s", util.Reflect(err).FullName))
			return
		}
		log.Fatal(fmt.Sprintf("failed to create collection with validation, err type: %s", util.Reflect(err).FullName), err)
	}
}

func wrapCreateIndexesError(err error) {
	if err != nil {
		if client.IsDuplicateKey(err) {
			log.Warn(fmt.Sprintf("indexes already exist, err type: %s", util.Reflect(err).FullName))
			return
		}
		log.Fatal(fmt.Sprintf("failed to create indexes, err type: %s", util.Reflect(err).FullName), err)
	}
}

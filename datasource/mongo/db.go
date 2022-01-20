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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

func ensureDB() {
	ensureService()
	ensureInstance()
	ensureSchema()
	ensureDep()
	ensureAccountLock()
}

func ensureService() {
	serviceIDIndex := util.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		util.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceID}))
	serviceIDIndex.Options = options.Index().SetUnique(true)

	serviceIndex := util.BuildIndexDoc(
		util.ConnectWithDot([]string{model.ColumnService, model.ColumnAppID}),
		util.ConnectWithDot([]string{model.ColumnService, model.ColumnServiceName}),
		util.ConnectWithDot([]string{model.ColumnService, model.ColumnEnv}),
		util.ConnectWithDot([]string{model.ColumnService, model.ColumnVersion}),
		model.ColumnDomain,
		model.ColumnProject)
	serviceIndex.Options = options.Index().SetUnique(true)

	var serviceIndexes []mongo.IndexModel
	serviceIndexes = append(serviceIndexes, serviceIDIndex, serviceIndex)

	dmongo.EnsureCollection(model.CollectionService, nil, serviceIndexes)
}

func ensureInstance() {
	instanceIndex := util.BuildIndexDoc(model.ColumnRefreshTime)
	instanceIndex.Options = options.Index().SetExpireAfterSeconds(defaultExpireTime)

	instanceServiceIndex := util.BuildIndexDoc(util.ConnectWithDot([]string{model.ColumnInstance, model.ColumnServiceID}))

	var instanceIndexes []mongo.IndexModel
	instanceIndexes = append(instanceIndexes, instanceIndex, instanceServiceIndex)

	dmongo.EnsureCollection(model.CollectionInstance, nil, instanceIndexes)
}

func ensureSchema() {
	dmongo.EnsureCollection(model.CollectionSchema, nil, []mongo.IndexModel{util.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		model.ColumnServiceID)})
}

func ensureDep() {
	dmongo.EnsureCollection(model.CollectionDep, nil, []mongo.IndexModel{util.BuildIndexDoc(
		model.ColumnDomain,
		model.ColumnProject,
		model.ColumnServiceKey)})
}

func ensureAccountLock() {
	dmongo.EnsureCollection(model.CollectionAccountLock, nil, []mongo.IndexModel{
		util.BuildIndexDoc(model.ColumnAccountLockKey)})
}

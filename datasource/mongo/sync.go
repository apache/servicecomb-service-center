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

	dmongo "github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	putil "github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	SyncAllKey = "sync-all"
)

type SyncManager struct {
}

// SyncAll will list all services,accounts,roles,schemas,tags,deps and use tasks to store
func (s *SyncManager) SyncAll(ctx context.Context) error {
	enable := config.GetBool("sync.enableOnStart", false)
	if !enable {
		return nil
	}
	exist, err := syncAllKeyExist(ctx)
	if err != nil {
		return err
	}
	if exist {
		log.Info(fmt.Sprintf("%s key already exists, do not need to do tasks", SyncAllKey))
		return datasource.ErrSyncAllKeyExists
	}
	// TODO use mongo distributed lock
	err = syncAllAccounts(ctx)
	if err != nil {
		return err
	}
	err = syncAllRoles(ctx)
	if err != nil {
		return err
	}
	err = syncAllServices(ctx)
	if err != nil {
		return err
	}
	return insertSyncAllKey(ctx)
}

func syncAllKeyExist(ctx context.Context) (bool, error) {
	count, err := dmongo.GetClient().GetDB().Collection(model.CollectionSync).CountDocuments(ctx, bson.M{"key": SyncAllKey})
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func syncAllAccounts(ctx context.Context) error {
	cursor, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccount).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error("fail to close mongo cursor", err)
		}
	}(cursor, ctx)
	for cursor.Next(ctx) {
		var account rbacmodel.Account
		err = cursor.Decode(&account)
		if err != nil {
			log.Error("failed to decode account", err)
			return err
		}
		err = sync.DoCreateOpts(ctx, datasource.ResourceAccount, &account)
		if err != nil {
			log.Error("failed to create account task", err)
			return err
		}
	}
	return nil
}

func syncAllRoles(ctx context.Context) error {
	cursor, err := dmongo.GetClient().GetDB().Collection(model.CollectionRole).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error("fail to close mongo cursor", err)
		}
	}(cursor, ctx)
	for cursor.Next(ctx) {
		var role rbacmodel.Role
		err = cursor.Decode(&role)
		if err != nil {
			log.Error("failed to decode role", err)
			return err
		}
		err = sync.DoCreateOpts(ctx, datasource.ResourceRole, &role)
		if err != nil {
			log.Error("failed to create role task", err)
			return err
		}
	}
	return nil
}

func syncAllServices(ctx context.Context) error {
	cursor, err := dmongo.GetClient().GetDB().Collection(model.CollectionService).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error("fail to close mongo cursor", err)
		}
	}(cursor, ctx)
	for cursor.Next(ctx) {
		var tmp model.Service
		err := cursor.Decode(&tmp)
		if err != nil {
			return err
		}
		request := &discovery.CreateServiceRequest{
			Service: tmp.Service,
			Tags:    tmp.Tags,
		}
		putil.SetDomain(ctx, tmp.Domain)
		putil.SetProject(ctx, tmp.Project)
		err = sync.DoCreateOpts(ctx, datasource.ResourceService, request)
		if err != nil {
			log.Error("failed to create service task", err)
			return err
		}
	}
	return nil
}

func insertSyncAllKey(ctx context.Context) error {
	_, err := dmongo.GetClient().GetDB().Collection(model.CollectionSync).InsertOne(ctx, bson.M{"key": SyncAllKey})
	return err
}

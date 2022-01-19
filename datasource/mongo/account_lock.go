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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func (al *RbacDAO) UpsertLock(ctx context.Context, lock *rbac.Lock) error {
	key := lock.Key
	releaseAt := lock.ReleaseAt
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	updateFilter := mutil.NewFilter(mutil.Set(mutil.NewFilter(
		mutil.AccountLockKey(key),
		mutil.AccountLockStatus(lock.Status),
		mutil.AccountLockReleaseAt(releaseAt),
	)))
	result := dmongo.GetClient().GetDB().Collection(model.CollectionAccountLock).FindOneAndUpdate(ctx, filter, updateFilter,
		options.FindOneAndUpdate().SetUpsert(true))
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		log.Error(fmt.Sprintf("can not save account lock %s", key), result.Err())
		return result.Err()
	}
	log.Info(fmt.Sprintf("%s is locked, release at %d", key, releaseAt))
	return nil
}

func (al *RbacDAO) GetLock(ctx context.Context, key string) (*rbac.Lock, error) {
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	result := dmongo.GetClient().GetDB().Collection(model.CollectionAccountLock).FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, rbac.ErrAccountLockNotExist
		}
		msg := fmt.Sprintf("failed to query account lock, key %s", key)
		log.Error(msg, result.Err())
		return nil, rbac.ErrQueryAccountLockFailed
	}
	var lock rbac.Lock
	err := result.Decode(&lock)
	if err != nil {
		log.Error(fmt.Sprintf("failed to decode account lock %s", key), err)
		return nil, err
	}
	return &lock, nil
}

func (al *RbacDAO) ListLock(ctx context.Context) ([]*rbac.Lock, int64, error) {
	filter := mutil.NewFilter()
	cursor, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccountLock).Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	var locks []*rbac.Lock
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var lock rbac.Lock
		err = cursor.Decode(&lock)
		if err != nil {
			log.Error("failed to decode account lock", err)
			continue
		}
		locks = append(locks, &lock)
	}
	return locks, int64(len(locks)), nil
}

func (al *RbacDAO) DeleteLock(ctx context.Context, key string) error {
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	_, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccountLock).DeleteMany(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("remove lock %s failed", key), err)
		return rbac.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%s is released", key))
	return nil
}

func (al *RbacDAO) DeleteLockList(ctx context.Context, keys []string) error {
	var delKeys []mongo.WriteModel
	for _, key := range keys {
		delKeys = append(delKeys, mongo.NewDeleteOneModel().SetFilter(mutil.NewFilter(mutil.AccountLockKey(key))))
	}
	if len(delKeys) == 0 {
		return nil
	}
	_, err := dmongo.GetClient().GetDB().Collection(model.CollectionAccountLock).BulkWrite(ctx, delKeys)
	if err != nil {
		log.Error(fmt.Sprintf("remove locks %v failed", keys), err)
		return rbac.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%v are released", keys))
	return nil
}

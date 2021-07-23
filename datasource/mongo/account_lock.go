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

package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountLockManager struct {
	releaseAfter time.Duration
}

func (al *AccountLockManager) GetLock(ctx context.Context, key string) (*datasource.AccountLock, error) {
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionAccountLock, filter)
	if err != nil {
		return nil, err
	}
	if err = result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, datasource.ErrAccountLockNotExist
		}
		msg := fmt.Sprintf("failed to query account lock, key %s", key)
		log.Error(msg, result.Err())
		return nil, datasource.ErrQueryAccountLockFailed
	}
	var lock datasource.AccountLock
	err = result.Decode(&lock)
	if err != nil {
		log.Error(fmt.Sprintf("failed to decode account lock %s", key), err)
		return nil, err
	}
	return &lock, nil
}

func (al *AccountLockManager) ListLock(ctx context.Context) ([]*datasource.AccountLock, int64, error) {
	filter := mutil.NewFilter()
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionAccountLock, filter)
	if err != nil {
		return nil, 0, err
	}
	var locks []*datasource.AccountLock
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var lock datasource.AccountLock
		err = cursor.Decode(&lock)
		if err != nil {
			log.Error("failed to decode account lock", err)
			continue
		}
		locks = append(locks, &lock)
	}
	return locks, int64(len(locks)), nil
}

func (al *AccountLockManager) DeleteLock(ctx context.Context, key string) error {
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	_, err := client.GetMongoClient().Delete(ctx, model.CollectionAccountLock, filter)
	if err != nil {
		log.Error(fmt.Sprintf("remove lock %s failed", key), err)
		return datasource.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%s is released", key))
	return nil
}

func (al *AccountLockManager) Ban(ctx context.Context, key string) error {
	releaseAt := time.Now().Add(al.releaseAfter).Unix()
	filter := mutil.NewFilter(mutil.AccountLockKey(key))
	updateFilter := mutil.NewFilter(mutil.Set(mutil.NewFilter(
		mutil.AccountLockKey(key),
		mutil.AccountLockStatus(datasource.StatusBanned),
		mutil.AccountLockReleaseAt(releaseAt),
	)))
	result, err := client.GetMongoClient().FindOneAndUpdate(ctx, model.CollectionAccountLock, filter, updateFilter,
		options.FindOneAndUpdate().SetUpsert(true))
	if err != nil {
		log.Error(fmt.Sprintf("can not save account lock %s", key), err)
		return err
	}
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		log.Error(fmt.Sprintf("can not save account lock %s", key), result.Err())
		return result.Err()
	}
	log.Info(fmt.Sprintf("%s is locked, release at %d", key, releaseAt))
	return nil
}

func NewAccountLockManager(ReleaseAfter time.Duration) datasource.AccountLockManager {
	return &AccountLockManager{releaseAfter: ReleaseAfter}
}

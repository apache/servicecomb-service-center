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

package service

import (
	"context"

	"github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func insertAccount(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) error {
	_, err := client.GetMongoClient().Insert(ctx, model.CollectionAccount, document, opts...)
	return err
}

func countAccount(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return client.GetMongoClient().Count(ctx, model.CollectionAccount, filter, opts...)
}

func findAccount(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*rbac.Account, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionAccount, filter, opts...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	var account rbac.Account
	err = result.Decode(&account)
	if err != nil {
		log.Error("failed to decode account", err)
		return nil, err
	}
	return &account, nil
}

func findAccounts(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*rbac.Account, int64, error) {
	cursor, err := client.GetMongoClient().Find(ctx, model.CollectionAccount, filter, opts...)
	if err != nil {
		log.Error("failed to find accounts", err)
		return nil, 0, err
	}
	var accounts []*rbac.Account
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var account rbac.Account
		err = cursor.Decode(&account)
		if err != nil {
			log.Error("failed to decode account", err)
			continue
		}
		account.Password = ""
		accounts = append(accounts, &account)
	}
	return accounts, int64(len(accounts)), nil
}

func deleteAccount(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (bool, error) {
	result, err := client.GetMongoClient().Delete(ctx, model.CollectionAccount, filter, opts...)
	if err != nil {
		log.Error("failed to delete account", err)
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func updateAccount(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	res, err := client.GetMongoClient().Update(ctx, model.CollectionAccount, filter, update, opts...)
	if err != nil {
		log.Error("failed to update account", err)
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoDataToUpdate
	}
	return nil
}

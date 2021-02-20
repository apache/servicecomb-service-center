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
	"errors"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbac.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	a.Password, err = privacy.HashPassword(a.Password)
	if err != nil {
		log.Error("pwd hash failed", err)
		return err
	}
	a.ID = util.GenerateUUID()
	_, err = client.GetMongoClient().Insert(ctx, CollectionAccount, a)
	if err != nil {
		if client.IsDuplicateKey(err) {
			return datasource.ErrAccountDuplicated
		}
		return err
	}
	log.Info("create new account: " + a.ID)
	return nil
}

func (ds *DataSource) AccountExist(ctx context.Context, name string) (bool, error) {
	filter := bson.M{
		ColumnAccountName: name,
	}
	count, err := client.GetMongoClient().Count(ctx, CollectionAccount, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetAccount(ctx context.Context, name string) (*rbac.Account, error) {
	filter := bson.M{
		ColumnAccountName: name,
	}
	result, err := client.GetMongoClient().FindOne(ctx, CollectionAccount, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		log.Error("failed to get account: ", result.Err())
		return nil, errors.New("failed to get account")
	}
	var account rbac.Account
	err = result.Decode(&account)
	if err != nil {
		log.Error("decode account failed: ", err)
		return nil, err
	}
	return &account, nil
}

func (ds *DataSource) ListAccount(ctx context.Context) ([]*rbac.Account, int64, error) {
	cursor, err := client.GetMongoClient().Find(ctx, CollectionAccount, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	var accounts []*rbac.Account
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var account rbac.Account
		err = cursor.Decode(&account)
		if err != nil {
			log.Error("decode account failed: ", err)
			continue
		}
		account.Password = ""
		accounts = append(accounts, &account)
	}
	return accounts, int64(len(accounts)), nil
}

func (ds *DataSource) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	filter := bson.M{
		ColumnAccountName: names[0],
	}
	result, err := client.GetMongoClient().Delete(ctx, CollectionAccount, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) UpdateAccount(ctx context.Context, name string, account *rbac.Account) error {
	filter := bson.M{
		ColumnAccountName: name,
	}
	update := bson.M{
		"$set": bson.M{
			ColumnID:                  account.ID,
			ColumnPassword:            account.Password,
			ColumnRoles:               account.Roles,
			ColumnTokenExpirationTime: account.TokenExpirationTime,
			ColumnCurrentPassword:     account.CurrentPassword,
			ColumnStatus:              account.Status,
		},
	}
	res, err := client.GetMongoClient().Update(ctx, CollectionAccount, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrUpdateNodata
	}
	return nil
}

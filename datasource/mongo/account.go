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
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/util"
	stringutil "github.com/go-chassis/foundation/string"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbacframe.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not save account info")
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), 14)
	if err != nil {
		log.Errorf(err, "pwd hash failed")
		return err
	}
	a.Password = stringutil.Bytes2str(hash)
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

func (ds *DataSource) AccountExist(ctx context.Context, key string) (bool, error) {
	filter := bson.M{
		AccountName: key,
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

func (ds *DataSource) GetAccount(ctx context.Context, key string) (*rbacframe.Account, error) {
	filter := bson.M{
		AccountName: key,
	}
	result, err := client.GetMongoClient().FindOne(ctx, CollectionAccount, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var account rbacframe.Account
	err = result.Decode(&account)
	if err != nil {
		log.Errorf(err, "Decode account failed: ")
		return nil, err
	}
	return &account, nil
}

func (ds *DataSource) ListAccount(ctx context.Context, key string) ([]*rbacframe.Account, int64, error) {
	filter := bson.M{
		AccountName: bson.M{"$regex": key},
	}
	cursor, err := client.GetMongoClient().Find(ctx, CollectionAccount, filter)
	if err != nil {
		return nil, 0, err
	}
	var accounts []*rbacframe.Account
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var account rbacframe.Account
		err = cursor.Decode(&account)
		if err != nil {
			log.Errorf(err, "Decode account failed: ")
			break
		}
		accounts = append(accounts, &account)
	}
	return accounts, int64(len(accounts)), nil
}

func (ds *DataSource) DeleteAccount(ctx context.Context, key string) (bool, error) {
	filter := bson.M{
		AccountName: key,
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

func (ds *DataSource) UpdateAccount(ctx context.Context, key string, account *rbacframe.Account) error {
	filter := bson.M{
		AccountName: key,
	}
	update := bson.M{
		"$set": bson.M{AccountID: account.ID, AccountPassword: account.Name, AccountRole: account.Role, AccountTokenExpirationTime: account.TokenExpirationTime,
			AccountCurrentPassword: account.CurrentPassword, AccountStatus: account.Status,
		},
	}
	result, err := client.GetMongoClient().Update(ctx, CollectionAccount, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("UpdateAccount: no data to update")
	}
	return nil
}

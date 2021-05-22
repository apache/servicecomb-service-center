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

	"github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/dao/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func (ds *DataSource) CreateAccount(ctx context.Context, a *rbac.Account) error {
	exist, err := ds.AccountExist(ctx, a.Name)
	if err != nil {
		msg := fmt.Sprintf("failed to query account, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	if exist {
		return datasource.ErrAccountDuplicated
	}
	a.Password, err = privacy.ScryptPassword(a.Password)
	if err != nil {
		msg := fmt.Sprintf("failed to hash account pwd, account name %s", a.Name)
		log.Error(msg, err)
		return err
	}
	a.ID = util.GenerateUUID()
	_, err = client.GetMongoClient().Insert(ctx, dao.CollectionAccount, a)
	if err != nil {
		if mutil.IsDuplicateKey(err) {
			return datasource.ErrAccountDuplicated
		}
		return err
	}
	log.Info("succeed to create new account: " + a.ID)
	return nil
}

func (ds *DataSource) AccountExist(ctx context.Context, name string) (bool, error) {
	filter := bson.D{
		{dao.ColumnAccountName, name},
	}
	count, err := client.GetMongoClient().Count(ctx, dao.CollectionAccount, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) GetAccount(ctx context.Context, name string) (*rbac.Account, error) {
	filter := bson.D{
		{dao.ColumnAccountName, name},
	}
	result, err := client.GetMongoClient().FindOne(ctx, dao.CollectionAccount, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query account, account name %s", name)
		log.Error(msg, err)
		return nil, err
	}
	if result.Err() != nil {
		msg := fmt.Sprintf("failed to query account, account name %s", name)
		log.Error(msg, result.Err())
		return nil, datasource.ErrQueryAccountFailed
	}
	var account rbac.Account
	err = result.Decode(&account)
	if err != nil {
		log.Error("failed to decode account", err)
		return nil, err
	}
	return &account, nil
}

func (ds *DataSource) ListAccount(ctx context.Context) ([]*rbac.Account, int64, error) {
	filter := bson.D{}
	cursor, err := client.GetMongoClient().Find(ctx, dao.CollectionAccount, filter)
	if err != nil {
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

func (ds *DataSource) DeleteAccount(ctx context.Context, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	inFilter := bson.D{
		{"$in", names},
	}
	filter := bson.D{
		{dao.ColumnAccountName, inFilter},
	}
	result, err := client.GetMongoClient().Delete(ctx, dao.CollectionAccount, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (ds *DataSource) UpdateAccount(ctx context.Context, name string, account *rbac.Account) error {
	filter := bson.D{
		{dao.ColumnAccountName, name},
	}
	setFilter := bson.D{
		{dao.ColumnID, account.ID},
		{dao.ColumnPassword, account.Password},
		{dao.ColumnRoles, account.Roles},
		{dao.ColumnTokenExpirationTime, account.TokenExpirationTime},
		{dao.ColumnCurrentPassword, account.CurrentPassword},
		{dao.ColumnStatus, account.Status},
	}
	updateFilter := bson.D{
		{"$set", setFilter},
	}
	res, err := client.GetMongoClient().Update(ctx, dao.CollectionAccount, filter, updateFilter)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return mutil.ErrNoDataToUpdate
	}
	return nil
}

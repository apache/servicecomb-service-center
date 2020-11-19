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

package mongo_test

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var instance datasource.DataSource

func init() {
	config := storage.DB{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config, []string{mongo.Account})
	instance, _ = mongo.NewDataSource(datasource.Options{})
}

func TestCreateAccount(t *testing.T) {
	account := rbacframe.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
	t.Run("create account: should be able to create account", func(t *testing.T) {
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
		err := instance.CreateAccount(context.Background(), &account)
		assert.Nil(t, err)
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
	})

	t.Run("create account: should not be able to create two same account", func(t *testing.T) {
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
		err := instance.CreateAccount(context.Background(), &account)
		assert.Nil(t, err)
		err = instance.CreateAccount(context.Background(), &account)
		assert.NotNil(t, err)
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
	})
}

func TestGetAccount(t *testing.T) {
	account := rbacframe.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
	t.Run("get account: if the account exists, it should be able to get the account", func(t *testing.T) {
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
		err := instance.CreateAccount(context.Background(), &account)
		assert.Nil(t, err)
		result, err := instance.GetAccount(context.Background(), account.Name)
		assert.Nil(t, err)
		assert.Equal(t, &account, result)
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
	})

	t.Run("get account: if the account not exists, it should not be able to get the account", func(t *testing.T) {
		_, err := instance.GetAccount(context.Background(), account.Name)
		assert.NotNil(t, err)
	})
}

func TestListAccount(t *testing.T) {
	account1 := rbacframe.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
	account2 := rbacframe.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account2",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset2",
	}
	t.Run("list account: if there are multiple accounts exist, it should be able to query multiple accounts", func(t *testing.T) {
		_ = instance.CreateAccount(context.Background(), &account1)
		_ = instance.CreateAccount(context.Background(), &account2)
		_, count, err := instance.ListAccount(context.Background(), "test-account")
		assert.Equal(t, int64(2), count)
		assert.Nil(t, err)
		_, _ = instance.DeleteAccount(context.Background(), account1.Name)
		_, _ = instance.DeleteAccount(context.Background(), account2.Name)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Run("delete account, if the account does not exist,it should not be deleted", func(t *testing.T) {
		flag, _ := instance.DeleteAccount(context.Background(), "not_exist")
		assert.Equal(t, false, flag)
	})

	t.Run("delete account, if the account exists,it should be deleted", func(t *testing.T) {
		account := rbacframe.Account{
			ID:                  "11111-22222-33333",
			Name:                "test-account1",
			Password:            "tnuocca-tset",
			Role:                "admin",
			TokenExpirationTime: "2020-12-30",
			CurrentPassword:     "tnuocca-tset1",
		}
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
		err := instance.CreateAccount(context.Background(), &account)
		assert.Nil(t, err)
		flag, err := instance.DeleteAccount(context.Background(), account.Name)
		assert.Equal(t, true, flag)
		assert.Nil(t, err)
	})
}

func TestUpdateAccount(t *testing.T) {
	t.Run("update account: if the account does not exist,the update should fail", func(t *testing.T) {
		account := rbacframe.Account{
			ID:                  "11111-22222-33333",
			Name:                "test-account1",
			Password:            "tnuocca-tset",
			Role:                "admin",
			TokenExpirationTime: "2020-12-30",
			CurrentPassword:     "tnuocca-tset1",
		}
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
		err := instance.UpdateAccount(context.Background(), account.Name, &account)
		assert.NotNil(t, err)
	})

	t.Run("update account: if the account exists,it should be updated successfully", func(t *testing.T) {
		account := rbacframe.Account{
			ID:                  "11111-22222-33333",
			Name:                "test-account1",
			Password:            "tnuocca-tset",
			Role:                "admin",
			TokenExpirationTime: "2020-12-30",
			CurrentPassword:     "tnuocca-tset1",
		}
		_ = instance.CreateAccount(context.Background(), &account)
		account.ID = "11111-22222-33333-44444"
		err := instance.UpdateAccount(context.Background(), account.Name, &account)
		assert.Nil(t, err)
		result, err := instance.GetAccount(context.Background(), account.Name)
		assert.Nil(t, err)
		assert.Equal(t, account.ID, result.ID)
		_, _ = instance.DeleteAccount(context.Background(), account.Name)
	})
}

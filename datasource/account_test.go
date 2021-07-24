/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package datasource_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/stretchr/testify/assert"
)

var (
	a1 = rbac.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Roles:               []string{"admin"},
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
	a2 = rbac.Account{
		ID:                  "11111-22222-33333-44444",
		Name:                "test-account2",
		Password:            "tnuocca-tset",
		Roles:               []string{"admin"},
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset2",
	}
)

func TestAccount(t *testing.T) {
	t.Run("add and get account", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		err = datasource.GetAccountManager().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		r, err := datasource.GetAccountManager().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		assert.Equal(t, a1, *r)
		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a1.Name, a2.Name})
		assert.NoError(t, err)
	})
	t.Run("account should exist", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		exist, err := datasource.GetAccountManager().AccountExist(context.Background(), a1.Name)
		assert.NoError(t, err)
		assert.True(t, exist)
		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
	})
	t.Run("delete account", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a2.Name})
		assert.NoError(t, err)
	})
	t.Run("add and update accounts then list", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		err = datasource.GetAccountManager().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		a2.Password = "new-password"
		err = datasource.GetAccountManager().UpdateAccount(context.Background(), a2.Name, &a2)
		assert.NoError(t, err)
		accounts, _, err := datasource.GetAccountManager().ListAccount(context.Background())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(accounts), 2)
		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a2.Name})
		assert.NoError(t, err)
	})
	t.Run("add and update accounts, should have create/update time", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)

		r, err := datasource.GetAccountManager().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		dt, _ := strconv.Atoi(r.CreateTime)
		assert.Less(t, 0, dt)
		assert.Equal(t, r.CreateTime, r.UpdateTime)

		time.Sleep(time.Second)
		a1.Password = "new-password"
		err = datasource.GetAccountManager().UpdateAccount(context.Background(), a1.Name, &a1)
		assert.NoError(t, err)

		old, _ := strconv.Atoi(r.UpdateTime)
		r, err = datasource.GetAccountManager().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		last, _ := strconv.Atoi(r.UpdateTime)
		assert.Less(t, old, last)
		assert.NotEqual(t, r.CreateTime, r.UpdateTime)

		_, err = datasource.GetAccountManager().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
	})
}

func TestAccountLock(t *testing.T) {
	var banTime int64

	t.Run("ban account TestAccountLock, should return no error", func(t *testing.T) {
		err := datasource.GetAccountLockManager().Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := datasource.GetAccountLockManager().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, datasource.StatusBanned, lock.Status)
		assert.Less(t, time.Now().Unix(), lock.ReleaseAt)

		banTime = lock.ReleaseAt
	})

	t.Run("ban account TestAccountLock again, should return a new release time", func(t *testing.T) {
		err := datasource.GetAccountLockManager().Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := datasource.GetAccountLockManager().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, datasource.StatusBanned, lock.Status)
		assert.Less(t, banTime, lock.ReleaseAt)
	})

	t.Run("ban account TestAccountLock again, should refresh releaseAt", func(t *testing.T) {
		lock1, err := datasource.GetAccountLockManager().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, datasource.StatusBanned, lock1.Status)

		time.Sleep(time.Second)
		err = datasource.GetAccountLockManager().Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock2, err := datasource.GetAccountLockManager().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Less(t, lock1.ReleaseAt, lock2.ReleaseAt)
	})

	t.Run("delete account lock, should return no error", func(t *testing.T) {
		err := datasource.GetAccountLockManager().DeleteLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := datasource.GetAccountLockManager().GetLock(context.Background(), "TestAccountLock")
		assert.Equal(t, datasource.ErrAccountLockNotExist, err)
		assert.Nil(t, lock)
	})
}

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

package rbac_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
)

var (
	a1 = rbacmodel.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Roles:               []string{"admin"},
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
		CreateTime:          strconv.FormatInt(time.Now().Unix(), 10),
	}
	a2 = rbacmodel.Account{
		ID:                  "11111-22222-33333-44444",
		Name:                "test-account2",
		Password:            "tnuocca-tset",
		Roles:               []string{"admin"},
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset2",
		CreateTime:          strconv.FormatInt(time.Now().Unix(), 10),
	}
)

func TestAccount(t *testing.T) {
	a1.UpdateTime = a1.CreateTime

	t.Run("add and get account", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		err = rbac.Instance().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		r, err := rbac.Instance().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		assert.Equal(t, a1, *r)
		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a1.Name, a2.Name})
		assert.NoError(t, err)
	})
	t.Run("account should exist", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		exist, err := rbac.Instance().AccountExist(context.Background(), a1.Name)
		assert.NoError(t, err)
		assert.True(t, exist)
		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
	})
	t.Run("delete account", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a2.Name})
		assert.NoError(t, err)
	})
	t.Run("add and update accounts then list", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)
		err = rbac.Instance().CreateAccount(context.Background(), &a2)
		assert.NoError(t, err)
		a2.Password = "new-password"
		err = rbac.Instance().UpdateAccount(context.Background(), a2.Name, &a2)
		assert.NoError(t, err)
		accounts, _, err := rbac.Instance().ListAccount(context.Background())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(accounts), 2)
		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a2.Name})
		assert.NoError(t, err)
	})
	t.Run("add and update accounts, should have create/update time", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a1)
		assert.NoError(t, err)

		r, err := rbac.Instance().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		dt, _ := strconv.Atoi(r.CreateTime)
		assert.Less(t, 0, dt)
		assert.Equal(t, r.CreateTime, r.UpdateTime)

		time.Sleep(time.Second)
		a1.Password = "new-password"
		err = rbac.Instance().UpdateAccount(context.Background(), a1.Name, &a1)
		assert.NoError(t, err)

		old, _ := strconv.Atoi(r.UpdateTime)
		r, err = rbac.Instance().GetAccount(context.Background(), a1.Name)
		assert.NoError(t, err)
		last, _ := strconv.Atoi(r.UpdateTime)
		assert.Less(t, old, last)
		assert.NotEqual(t, r.CreateTime, r.UpdateTime)

		_, err = rbac.Instance().DeleteAccount(context.Background(), []string{a1.Name})
		assert.NoError(t, err)
	})
}

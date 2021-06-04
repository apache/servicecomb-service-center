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
	"testing"

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
}

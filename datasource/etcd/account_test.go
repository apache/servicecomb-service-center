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

package etcd_test

import (
	"context"
	"github.com/go-chassis/cari/rbac"
	"testing"

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
		CurrentPassword:     "tnuocca-tset1",
	}
)

func TestAccount(t *testing.T) {
	t.Run("add and get account", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "test-account-key", &a1)
		assert.NoError(t, err)
		r, err := datasource.Instance().GetAccount(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.Equal(t, a1, *r)
	})
	t.Run("account should exist", func(t *testing.T) {
		exist, err := datasource.Instance().AccountExist(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.True(t, exist)
	})
	t.Run("delete account", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "test-account-key222", &a1)
		assert.NoError(t, err)
		_, err = datasource.Instance().DeleteAccount(context.Background(), "test-account-key222")
		assert.NoError(t, err)
	})
	t.Run("add two accounts and list", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "key1", &a1)
		assert.NoError(t, err)
		err = datasource.Instance().UpdateAccount(context.Background(), "key2", &a2)
		assert.NoError(t, err)
		accs, n, err := datasource.Instance().ListAccount(context.Background(), "key")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)
		t.Log(accs)
	})
}

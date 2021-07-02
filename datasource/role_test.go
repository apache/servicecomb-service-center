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

package datasource_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-chassis/cari/rbac"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
)

var (
	r1 = rbac.Role{
		ID:    "11111-22222-33333",
		Name:  "test-role1",
		Perms: nil,
	}

	r2 = rbac.Role{
		ID:    "11111-22222-33333-44444",
		Name:  "test-role2",
		Perms: nil,
	}

	a = rbac.Account{
		Name:     "account-role-test",
		Password: "abc",
		Roles:    []string{"test-role1"},
	}
)

func TestRole(t *testing.T) {
	t.Run("create role should success", func(t *testing.T) {
		err := datasource.GetRoleManager().CreateRole(context.Background(), &r1)
		assert.NoError(t, err)
		r, err := datasource.GetRoleManager().GetRole(context.Background(), "test-role1")
		assert.NoError(t, err)
		assert.Equal(t, r1, *r)
		dt, _ := strconv.Atoi(r.CreateTime)
		assert.Less(t, 0, dt)
		assert.Equal(t, r.CreateTime, r.UpdateTime)
	})
	t.Run("role should exist", func(t *testing.T) {
		exist, err := datasource.GetRoleManager().RoleExist(context.Background(), "test-role1")
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("repeated create role should failed", func(t *testing.T) {
		err := datasource.GetRoleManager().CreateRole(context.Background(), &r1)
		assert.Error(t, err)
	})

	t.Run("update role should success", func(t *testing.T) {
		r, err := datasource.GetRoleManager().GetRole(context.Background(), "test-role1")
		assert.NoError(t, err)
		old, _ := strconv.Atoi(r.UpdateTime)

		time.Sleep(time.Second)
		r1.ID = "11111-22222-33333-4"
		err = datasource.GetRoleManager().UpdateRole(context.Background(), "test-role1", &r1)
		assert.NoError(t, err)

		r, err = datasource.GetRoleManager().GetRole(context.Background(), "test-role1")
		assert.NoError(t, err)
		last, _ := strconv.Atoi(r.UpdateTime)
		assert.Less(t, old, last)
		assert.NotEqual(t, r.CreateTime, r.UpdateTime)
	})
	t.Run("add new role should success", func(t *testing.T) {
		err := datasource.GetRoleManager().CreateRole(context.Background(), &r2)
		assert.NoError(t, err)
		_, n, err := datasource.GetRoleManager().ListRole(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)
	})

	t.Run("delete role bind to user should failed", func(t *testing.T) {
		err := datasource.GetAccountManager().CreateAccount(context.Background(), &a)
		assert.NoError(t, err)

		_, err = datasource.GetRoleManager().DeleteRole(context.Background(), "test-role1")
		assert.Error(t, err)
	})

	t.Run("update account role and delete old role should success", func(t *testing.T) {
		a.Roles = []string{"test-role2"}
		err := datasource.GetAccountManager().UpdateAccount(context.Background(), "account-role-test", &a)
		assert.NoError(t, err)

		_, err = datasource.GetRoleManager().DeleteRole(context.Background(), "test-role1")
		assert.NoError(t, err)
	})

	t.Run("delete role should success", func(t *testing.T) {
		b, err := datasource.GetAccountManager().DeleteAccount(context.Background(), []string{"account-role-test"})
		assert.True(t, b)
		assert.NoError(t, err)

		_, err = datasource.GetRoleManager().DeleteRole(context.Background(), "test-role2")
		assert.NoError(t, err)
		_, n, err := datasource.GetRoleManager().ListRole(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})
}

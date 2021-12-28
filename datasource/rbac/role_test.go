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

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
)

var (
	r1 = rbacmodel.Role{
		ID:    "11111-22222-33333",
		Name:  "test-role",
		Perms: nil,
	}

	r2 = rbacmodel.Role{
		ID:    "11111-22222-33333-44444",
		Name:  "test-role-ex",
		Perms: nil,
	}

	a = rbacmodel.Account{
		Name:     "account-role-test",
		Password: "abc",
		Roles:    []string{"test-role"},
	}
)

func TestRole(t *testing.T) {
	t.Run("create role should success", func(t *testing.T) {
		err := rbac.Instance().CreateRole(context.Background(), &r1)
		assert.NoError(t, err)
		r, err := rbac.Instance().GetRole(context.Background(), "test-role")
		assert.NoError(t, err)
		assert.Equal(t, r1, *r)
		dt, _ := strconv.Atoi(r.CreateTime)
		assert.Less(t, 0, dt)
		assert.Equal(t, r.CreateTime, r.UpdateTime)
	})
	t.Run("role should exist", func(t *testing.T) {
		exist, err := rbac.Instance().RoleExist(context.Background(), "test-role")
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("repeated create role should failed", func(t *testing.T) {
		err := rbac.Instance().CreateRole(context.Background(), &r1)
		assert.Error(t, err)
	})

	t.Run("update role should success", func(t *testing.T) {
		r, err := rbac.Instance().GetRole(context.Background(), "test-role")
		assert.NoError(t, err)
		old, _ := strconv.Atoi(r.UpdateTime)

		time.Sleep(time.Second)
		r1.ID = "11111-22222-33333-4"
		err = rbac.Instance().UpdateRole(context.Background(), "test-role", &r1)
		assert.NoError(t, err)

		r, err = rbac.Instance().GetRole(context.Background(), "test-role")
		assert.NoError(t, err)
		last, _ := strconv.Atoi(r.UpdateTime)
		assert.Less(t, old, last)
		assert.NotEqual(t, r.CreateTime, r.UpdateTime)
	})

	t.Run("add new role should success", func(t *testing.T) {
		err := rbac.Instance().CreateRole(context.Background(), &r2)
		assert.NoError(t, err)
		_, n, err := rbac.Instance().ListRole(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)
	})

	t.Run("delete role bind to user should failed", func(t *testing.T) {
		err := rbac.Instance().CreateAccount(context.Background(), &a)
		assert.NoError(t, err)

		_, err = rbac.Instance().DeleteRole(context.Background(), "test-role")
		assert.Error(t, err)
	})

	t.Run("update account role and delete old role should success", func(t *testing.T) {
		a.Roles = []string{"test-role-ex"}
		err := rbac.Instance().UpdateAccount(context.Background(), "account-role-test", &a)
		assert.NoError(t, err)

		_, err = rbac.Instance().DeleteRole(context.Background(), "test-role")
		assert.NoError(t, err)
	})

	t.Run("delete role should success", func(t *testing.T) {
		b, err := rbac.Instance().DeleteAccount(context.Background(), []string{"account-role-test"})
		assert.True(t, b)
		assert.NoError(t, err)

		_, err = rbac.Instance().DeleteRole(context.Background(), "test-role-ex")
		assert.NoError(t, err)

		_, n, err := rbac.Instance().ListRole(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), n)
	})
}

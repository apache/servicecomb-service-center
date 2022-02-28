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
	"testing"

	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
)

func TestListSelfPerms(t *testing.T) {
	const selfAccount = "testSelfPerms"
	const role1 = "rolePerms1"
	const role2 = "rolePerms2"

	t.Run("init test account & role, should pass", func(t *testing.T) {
		err := rbacsvc.CreateRole(context.TODO(), &rbac.Role{
			Name: role1,
			Perms: []*rbac.Permission{
				{
					Resources: []*rbac.Resource{{Type: rbacsvc.ResourceService}},
					Verbs:     []string{"create"},
				},
			},
		})
		assert.NoError(t, err)

		err = rbacsvc.CreateRole(context.TODO(), &rbac.Role{
			Name: role2,
			Perms: []*rbac.Permission{
				{
					Resources: []*rbac.Resource{{Type: rbacsvc.ResourceAccount}},
					Verbs:     []string{"get"},
				},
			},
		})
		assert.NoError(t, err)

		err = rbacsvc.CreateAccount(context.Background(), &rbac.Account{
			Name:     selfAccount,
			Password: "Ab@00000",
			Roles:    []string{role1},
		})
		assert.NoError(t, err)
	})
	t.Run("list account perms which has rolePerms1, should return 1 perm", func(t *testing.T) {
		ctx := getSelfCtx()
		perms, err := rbacsvc.ListSelfPerms(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, perms)
		assert.Equal(t, 1, len(perms))
		assert.Equal(t, []string{"create"}, perms[0].Verbs)
	})
	t.Run("list account perms which has rolePerms1 & 2, should return 2 perms", func(t *testing.T) {
		err := rbacsvc.UpdateAccount(context.Background(), selfAccount, &rbac.Account{
			Roles: []string{role1, role2},
		})
		assert.NoError(t, err)

		ctx := getSelfCtx()
		perms, err := rbacsvc.ListSelfPerms(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(perms))
	})
	t.Run("cleanup, should pass", func(t *testing.T) {
		err := rbacsvc.DeleteAccount(context.Background(), selfAccount)
		assert.NoError(t, err)
		err = rbacsvc.DeleteRole(context.Background(), role1)
		assert.NoError(t, err)
		err = rbacsvc.DeleteRole(context.Background(), role2)
		assert.NoError(t, err)
	})
}

func getSelfCtx() context.Context {
	claims := map[string]interface{}{
		rbac.ClaimsUser: "testSelfPerms",
	}
	return context.WithValue(context.Background(), rbacsvc.CtxRequestClaims, claims)
}

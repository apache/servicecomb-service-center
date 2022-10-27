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

// initialize
import (
	"context"
	"testing"

	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"

	_ "github.com/apache/servicecomb-service-center/test"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
}

const (
	testPwd0 = "Ab@00000"
	testPwd1 = "Ab@11111"
)

func newAccount(name string) *rbac.Account {
	return &rbac.Account{
		Name:     name,
		Password: testPwd0,
		Roles:    []string{rbac.RoleAdmin},
		Status:   "active",
	}
}

func TestCreateAccount(t *testing.T) {
	ctx := context.TODO()
	t.Run("create account, should succeed", func(t *testing.T) {
		a := newAccount("TestCreateAccount_create_account")
		err := rbacsvc.CreateAccount(ctx, a)
		assert.Nil(t, err)
	})
	t.Run("create account twice, should return: "+rbac.NewError(rbac.ErrAccountConflict, "").Error(), func(t *testing.T) {
		name := "TestCreateAccount_create_account_twice"
		a := newAccount(name)
		err := rbacsvc.CreateAccount(ctx, a)
		assert.Nil(t, err)

		a = newAccount(name)
		err = rbacsvc.CreateAccount(ctx, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountConflict, svcErr.Code)
	})
	t.Run("account has invalid role, should return: "+rbac.NewError(rbac.ErrAccountHasInvalidRole, "").Error(), func(t *testing.T) {
		a := newAccount("TestCreateAccount_account_has_invalid_role")
		a.Roles = append(a.Roles, "invalid_role")
		err := rbacsvc.CreateAccount(ctx, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountHasInvalidRole, svcErr.Code)
	})
	t.Run("account has id, should succeed", func(t *testing.T) {
		accountName := "TestCreateAccount_account_has_id"
		a := newAccount(accountName)
		a.ID = "specifyID"
		err := rbacsvc.CreateAccount(ctx, a)
		assert.NoError(t, err)

		defer rbacsvc.DeleteAccount(ctx, accountName)

		account, err := rbacsvc.GetAccount(ctx, accountName)
		assert.NoError(t, err)
		assert.Equal(t, "specifyID", account.ID)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Run("delete account, should succeed", func(t *testing.T) {
		a := newAccount("TestDeleteAccount_delete_account")
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)

		exist, err := rbacsvc.AccountExist(context.TODO(), a.Name)
		assert.Nil(t, err)
		assert.True(t, exist)
		err = rbacsvc.DeleteAccount(context.TODO(), a.Name)
		assert.Nil(t, err)
		exist, err = rbacsvc.AccountExist(context.TODO(), a.Name)
		assert.Nil(t, err)
		assert.False(t, exist)
	})
	t.Run("delete no exist account, should return: "+rbac.NewError(rbac.ErrAccountNotExist, "").Error(), func(t *testing.T) {
		err := rbacsvc.DeleteAccount(context.TODO(), "TestDeleteAccount_delete_no_exist_account")
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountNotExist, svcErr.Code)
	})
	t.Run("delete root, should return: "+rbac.NewError(rbac.ErrForbidOperateBuildInAccount, "").Error(), func(t *testing.T) {
		err := rbacsvc.DeleteAccount(context.TODO(), "root")
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrForbidOperateBuildInAccount, svcErr.Code)
	})
	t.Run("delete self, should return: "+rbac.NewError(rbac.ErrForbidOperateSelfAccount, "").Error(), func(t *testing.T) {
		a := newAccount("TestDeleteAccount_delete_self")
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)
		claims := map[string]interface{}{
			rbac.ClaimsUser: a.Name,
		}
		ctx := context.WithValue(context.TODO(), rbacsvc.CtxRequestClaims, claims)
		err = rbacsvc.DeleteAccount(ctx, a.Name)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrForbidOperateSelfAccount, svcErr.Code)
	})
}

func TestUpdateAccount(t *testing.T) {
	t.Run("update account, should succeed", func(t *testing.T) {
		name := "TestUpdateAccount_update_account"
		a := newAccount(name)
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)

		a = newAccount(name)
		a.Roles = []string{rbac.RoleAdmin, rbac.RoleDeveloper}
		err = rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.Nil(t, err)
		resp, err := rbacsvc.GetAccount(context.TODO(), a.Name)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(resp.Roles))
	})
	t.Run("update no exist account, should return: "+rbac.NewError(rbac.ErrAccountNotExist, "").Error(), func(t *testing.T) {
		name := "TestUpdateAccount_update_no_exist_account"
		a := newAccount(name)
		err := rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountNotExist, svcErr.Code)
	})
	t.Run("update root, should return: "+discovery.NewError(rbac.ErrForbidOperateBuildInAccount, "").Error(), func(t *testing.T) {
		a := newAccount("root")
		err := rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrForbidOperateBuildInAccount, svcErr.Code)
	})
	t.Run("account has invalid role, should return: "+rbac.NewError(rbac.ErrAccountHasInvalidRole, "").Error(), func(t *testing.T) {
		name := "TestUpdateAccount_account_has_invalid_role"
		a := newAccount(name)
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)

		a.Roles = append(a.Roles, "invalid_role")
		err = rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountHasInvalidRole, svcErr.Code)
	})
	t.Run("roles status empty both, should return: "+discovery.NewError(discovery.ErrInvalidParams, "").Error(), func(t *testing.T) {
		name := "TestUpdateAccount_roles_status_empty_both"
		a := newAccount(name)
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)

		a = newAccount(name)
		a.Roles = nil
		a.Status = ""
		err = rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, discovery.ErrInvalidParams, svcErr.Code)
	})
	t.Run("update self, should return: "+rbac.NewError(rbac.ErrForbidOperateSelfAccount, "").Error(), func(t *testing.T) {
		name := "TestDeleteAccount_update_self"
		a := newAccount(name)
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)

		a = newAccount(name)
		claims := map[string]interface{}{
			rbac.ClaimsUser: a.Name,
		}
		ctx := context.WithValue(context.TODO(), rbacsvc.CtxRequestClaims, claims)
		err = rbacsvc.UpdateAccount(ctx, a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrForbidOperateSelfAccount, svcErr.Code)
	})
}

func TestEditAccount(t *testing.T) {
	t.Run("edit no exist account, should return: "+rbac.NewError(rbac.ErrAccountNotExist, "").Error(), func(t *testing.T) {
		a := newAccount("TestEditAccount_edit_no_exist_account")
		err := rbacsvc.UpdateAccount(context.TODO(), a.Name, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountNotExist, svcErr.Code)
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("get account, should succeed", func(t *testing.T) {
		a, err := rbacsvc.GetAccount(context.TODO(), "root")
		assert.Nil(t, err)
		assert.Equal(t, "root", a.Name)
	})
	t.Run("get no exist account, should return: "+rbac.NewError(rbac.ErrAccountNotExist, "").Error(), func(t *testing.T) {
		a, err := rbacsvc.GetAccount(context.TODO(), "TestGetAccount_no_exist_account")
		assert.Nil(t, a)
		assert.NotNil(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, rbac.ErrAccountNotExist, svcErr.Code)
	})
}

func TestListAccount(t *testing.T) {
	t.Run("list account, should succeed", func(t *testing.T) {
		accounts, n, err := rbacsvc.ListAccount(context.TODO())
		assert.Nil(t, err)
		assert.True(t, n > 0)
		assert.Equal(t, n, int64(len(accounts)))
	})
}

func TestBatchCreateAccounts(t *testing.T) {
	ctx := context.TODO()

	t.Run("batch create invalid accounts, should failed", func(t *testing.T) {
		resp, err := rbacsvc.BatchCreateAccounts(ctx, &rbac.BatchCreateAccountsRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
		svcErr := err.(*errsvc.Error)
		assert.Equal(t, discovery.ErrInvalidParams, svcErr.Code)
	})
	t.Run("batch create accounts, should succeed", func(t *testing.T) {
		a1 := newAccount("TestBatchCreateAccounts_account_1")
		a2 := newAccount("TestBatchCreateAccounts_account_no_pwd")
		a2.Password = ""
		a3 := newAccount("TestBatchCreateAccounts_account_invalid_pwd")
		a3.Password = "1"

		defer func() {
			rbacsvc.DeleteAccount(ctx, "TestBatchCreateAccounts_account_1")
			rbacsvc.DeleteAccount(ctx, "TestBatchCreateAccounts_account_no_pwd")
			rbacsvc.DeleteAccount(ctx, "TestBatchCreateAccounts_account_invalid_pwd")
		}()

		resp, err := rbacsvc.BatchCreateAccounts(ctx, &rbac.BatchCreateAccountsRequest{
			Accounts: []*rbac.Account{a1, a2, a3},
		})
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(resp.Accounts))

		item := resp.Accounts[0]
		assert.Equal(t, "TestBatchCreateAccounts_account_1", item.Name)
		assert.Nil(t, item.Error)

		item = resp.Accounts[1]
		assert.Equal(t, "TestBatchCreateAccounts_account_no_pwd", item.Name)
		assert.Nil(t, item.Error)

		item = resp.Accounts[2]
		assert.Equal(t, "TestBatchCreateAccounts_account_invalid_pwd", item.Name)
		assert.NotEmpty(t, item.Code)
	})
}

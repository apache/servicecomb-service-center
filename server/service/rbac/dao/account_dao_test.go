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

package dao_test

// initialize
import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/astaxie/beego"
	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
}

func newAccount(name string) *rbac.Account {
	return &rbac.Account{
		Name:     name,
		Password: "Ab@11111",
		Roles:    []string{rbacframe.RoleAdmin},
	}
}

func TestAccountDao_CreateAccount(t *testing.T) {
	account := newAccount("createAccountTest")
	dao.DeleteAccount(context.TODO(), account.Name)
	_ = dao.CreateAccount(context.Background(), account)
	t.Run("get account", func(t *testing.T) {
		r, err := dao.GetAccount(context.Background(), account.Name)
		assert.NoError(t, err)
		assert.Equal(t, account.Name, r.Name)
		hash, err := bcrypt.GenerateFromPassword([]byte(account.Password), 14)
		err = bcrypt.CompareHashAndPassword(hash, []byte(account.Password))
		assert.NoError(t, err)
	})
}
func TestAccountDao_UpdateAccount(t *testing.T) {
	account := newAccount("updateAccountTest")
	t.Run("update an none exist account", func(t *testing.T) {
		newAccount := &rbac.Account{Roles: []string{"admin"}}
		err := dao.UpdateAccount(context.Background(), "noExist", newAccount)
		assert.Error(t, err)
	})

	dao.DeleteAccount(context.TODO(), account.Name)
	err := dao.CreateAccount(context.Background(), account)
	assert.NoError(t, err)

	t.Run("update account", func(t *testing.T) {
		newAccount := &rbac.Account{
			Roles: []string{rbacframe.RoleDeveloper},
		}
		err = dao.UpdateAccount(context.Background(), account.Name, newAccount)
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), account.Name)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(a.Roles))
		assert.Equal(t, rbacframe.RoleDeveloper, a.Roles[0])
	})
}

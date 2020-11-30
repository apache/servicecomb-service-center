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

package v4_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	v4 "github.com/apache/servicecomb-service-center/server/resource/v4"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
import (
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/astaxie/beego"
)

func init() {
	beego.AppConfig.Set("rbac_enabled", "true")
	beego.AppConfig.Set("rbac_rsa_public_key_file", "./rbac.pub")
	beego.AppConfig.Set("rbac_rsa_private_key_file", "./private.key")
	config.Init()
}
func TestAuthResource_Login(t *testing.T) {
	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	assert.NoError(t, err)

	pri, pub, err := secret.GenRSAKeyPair(4096)
	assert.NoError(t, err)

	b, err := secret.RSAPrivate2Bytes(pri)
	assert.NoError(t, err)
	ioutil.WriteFile("./private.key", b, 0600)
	b, err = secret.RSAPublicKey2Bytes(pub)
	err = ioutil.WriteFile("./rbac.pub", b, 0600)
	assert.NoError(t, err)

	archaius.Set(rbac.InitPassword, "Complicated_password1")

	ctx := context.TODO()
	dao.DeleteAccount(ctx, "root")
	archaius.Init(archaius.WithMemorySource())

	rbac.Init()
	rest.RegisterServant(&v4.AuthResource{})
	rest.RegisterServant(&v4.RoleResource{})

	dao.DeleteAccount(ctx, "dev_account")

	t.Run("invalid user login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	// root account token
	var to = &rbacframe.Token{}
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1", Roles: []string{"admin"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("invalid password", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create dev_account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/account", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("dev_account login and change pwd, then login again", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b2, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", CurrentPassword: "Complicated_password1", Password: "Complicated_password2"})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/account/dev_account/password", bytes.NewBuffer(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		b3, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password2", Roles: []string{"developer"}})
		r3, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b3))
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})

}
func TestAuthResource_DeleteAccount(t *testing.T) {
	t.Run("dev_account can not even delete him self", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password2", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devTo := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), devTo)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/account/dev_account", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+devTo.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})
	t.Run("root can delete account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		b, _ = json.Marshal(&rbacframe.Account{Name: "delete_account", Password: "Complicated_password1"})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/account", bytes.NewBuffer(b))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/account/delete_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusNoContent, w3.Code)
	})
}
func TestAuthResource_GetAccount(t *testing.T) {
	t.Run("get account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account/dev_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacframe.Account{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Equal(t, "dev_account", a.Name)
		assert.Equal(t, []string{"developer"}, a.Roles)
		assert.Empty(t, a.Password)
	})
	t.Run("list account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacframe.AccountResponse{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Greater(t, a.Total, int64(1))
		assert.Empty(t, a.Accounts[0].Password)
	})

	t.Run("get a short expiration token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1", TokenExpirationTime: "10s"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		time.Sleep(11 * time.Second)
		r3, _ := http.NewRequest(http.MethodGet, "/v4/account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusUnauthorized, w3.Code)
	})
}

func TestRoleResource_CreateOrUpdateRole(t *testing.T) {
	var to = &rbacframe.Token{}
	ctx := context.TODO()
	dao.DeleteAccount(ctx, "dev_test")
	dao.DeleteRole(ctx, "tester")
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("create account dev_test and add a role", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_test", Password: "Complicated_password3", Roles: []string{"tester"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/account", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create a role name tester ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_test", Password: "Complicated_password3", Roles: []string{"tester"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devToken := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), devToken)

		b2, _ := json.Marshal(&rbacframe.Role{
			Name: "tester",
			Perms: []*rbacframe.Permission{
				{
					Resources: []string{"service", "instance"},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})

		r2, _ := http.NewRequest(http.MethodPost, "/v4/role", bytes.NewReader(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/role", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		b4, _ := json.Marshal(&rbacframe.Role{
			Name: "tester",
			Perms: []*rbacframe.Permission{
				{
					Resources: []string{"service"},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})
		r4, _ := http.NewRequest(http.MethodPut, "/v4/role/tester", bytes.NewReader(b4))
		r4.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusOK, w4.Code)
	})

	t.Run("Inquire role", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacframe.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/role/admin", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/role/admin", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}

func TestRoleResource_MoreRoles(t *testing.T) {
	var to = &rbacframe.Token{}
	ctx := context.TODO()
	dao.DeleteAccount(ctx, "dev_test")
	dao.DeleteAccount(ctx, "dev_test2")
	dao.DeleteRole(ctx, "tester")
	dao.DeleteRole(ctx, "tester2")
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("create role name tester", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Role{
			Name: "tester",
			Perms: []*rbacframe.Permission{
				{
					Resources: []string{"service"},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/role", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create new role name tester2", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Role{
			Name: "tester2",
			Perms: []*rbacframe.Permission{
				{
					Resources: []string{"rule"},
					Verbs:     []string{"*"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/role", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/role", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("account dev_test2 support more than 1 role ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_test2", Password: "Complicated_password3", Roles: []string{"tester", "tester2"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/account", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
		devToken := &rbacframe.Token{}
		json.Unmarshal(w2.Body.Bytes(), devToken)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		r4, _ := http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices", nil)
		r4.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusUnauthorized, w4.Code)
	})
}

func TestAuthResource_Login2(t *testing.T) {
	t.Run("bock user dev_account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacframe.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		w = httptest.NewRecorder()
		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

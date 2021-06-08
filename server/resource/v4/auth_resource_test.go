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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rbacmodel "github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	v4 "github.com/apache/servicecomb-service-center/server/resource/v4"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
)

var pwd = "Complicated_password1"

func init() {
	beego.AppConfig.Set("rbac_enabled", "true")
	beego.AppConfig.Set("rbac_rsa_public_key_file", "./rbac.pub")
	beego.AppConfig.Set("rbac_rsa_private_key_file", "./private.key")
	config.Init()

	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	if err != nil {
		panic(err)
	}

	pri, pub, err := secret.GenRSAKeyPair(4096)
	if err != nil {
		panic(err)
	}

	b, err := secret.RSAPrivate2Bytes(pri)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("./private.key", b, 0600)
	b, err = secret.RSAPublicKey2Bytes(pub)
	err = ioutil.WriteFile("./rbac.pub", b, 0600)
	if err != nil {
		panic(err)
	}

	archaius.Set(rbacsvc.InitPassword, pwd)
	ctx := context.TODO()
	rbacsvc.DeleteAccount(ctx, "root")

	rbacsvc.Init()
	rest.RegisterServant(&v4.AuthResource{})
	rest.RegisterServant(&v4.RoleResource{})
}
func TestAuthResource_Login(t *testing.T) {
	ctx := context.TODO()

	rbacsvc.DeleteAccount(ctx, "dev_account")

	t.Run("invalid user login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: pwd})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	// root account token
	var rootToken = &rbacmodel.Token{}
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: pwd, Roles: []string{"admin"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), rootToken)
	})

	t.Run("invalid password", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create dev_account,should success", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: "Complicated_password1", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("given a valid token and deleted account, auth should fail", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "admin_account", Password: "Complicated_password1", Roles: []string{"admin"}})
		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b, _ = json.Marshal(&rbacmodel.Account{Name: "admin_account", Password: "Complicated_password1"})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
		deletedToken := &rbacmodel.Token{}
		json.Unmarshal(w2.Body.Bytes(), deletedToken)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/admin_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		r4, _ := http.NewRequest(http.MethodGet, "/v4/accounts/admin_account", nil)
		r4.Header.Set(restful.HeaderAuth, "Bearer "+deletedToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusUnauthorized, w4.Code)
	})
	t.Run("dev_account login and change pwd, then login again", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: "Complicated_password1", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b2, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", CurrentPassword: "Complicated_password1", Password: "Complicated_password2"})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/accounts/dev_account/password", bytes.NewBuffer(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		b3, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: "Complicated_password2", Roles: []string{"developer"}})
		r3, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b3))
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})

}
func TestAuthResource_DeleteAccount(t *testing.T) {
	rootToken := &rbacmodel.Token{}

	t.Run("given root, try to delete it, should fail ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), rootToken)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/root", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusForbidden, w2.Code)
	})
	t.Run("dev_account can not even delete him self", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: "Complicated_password2", Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devTo := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), devTo)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/dev_account", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+devTo.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})

	t.Run("create your_account with admin role should be pass", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "your_account", Password: "Complicated_password3", Roles: []string{"admin"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("account can not even delete him self", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "your_account", Password: "Complicated_password3", Roles: []string{"admin"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		yourToken := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), yourToken)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/your_account", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+yourToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusForbidden, w2.Code)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/your_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})
	t.Run("root can delete account", func(t *testing.T) {
		var err error
		b, err := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})
		assert.NoError(t, err)
		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		b, _ = json.Marshal(&rbacmodel.Account{Name: "delete_account", Password: "Complicated_password1", Roles: []string{rbacmodel.RoleDeveloper}})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/delete_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}
func TestAuthResource_GetAccount(t *testing.T) {
	t.Run("get account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/accounts/dev_account", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacmodel.Account{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Equal(t, "dev_account", a.Name)
		assert.Equal(t, []string{"developer"}, a.Roles)
		assert.Empty(t, a.Password)
	})
	t.Run("list account", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/accounts?noCache=1", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacmodel.AccountResponse{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Greater(t, a.Total, int64(1))
		assert.Empty(t, a.Accounts[0].Password)
	})

	t.Run("get a short expiration token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1", TokenExpirationTime: "10s"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get a expiration token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1", TokenExpirationTime: "15m"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		time.Sleep(11 * time.Second)
		r3, _ := http.NewRequest(http.MethodGet, "/v4/accounts", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})
}

func TestAuthResource_Login2(t *testing.T) {
	t.Run("given wrong pwd, should bock user", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_account", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		w = httptest.NewRecorder()
		r, _ = http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func BenchmarkAuthResource_LoginP(b *testing.B) {
	body, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: pwd})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			rest.GetRouter().ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				panic(w.Code)
			}
		}
	})
	b.ReportAllocs()
}

//
func BenchmarkAuthResource_Login(b *testing.B) {
	body, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: pwd})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			panic(w.Code)
		}

	}
	b.ReportAllocs()
}

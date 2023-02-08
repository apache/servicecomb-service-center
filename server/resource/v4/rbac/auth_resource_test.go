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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	v4 "github.com/apache/servicecomb-service-center/server/resource/v4/rbac"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	beego "github.com/beego/beego/v2/server/web"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
)

const (
	devAccount           = "dev_account"
	testSelfPermsAccount = "test_self_perms"
	devPwd1              = "Complicated_password1"
	devPwd2              = "Complicated_password2"
)

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
	os.WriteFile("./private.key", b, 0600)
	b, err = secret.RSAPublicKey2Bytes(pub)
	err = os.WriteFile("./rbac.pub", b, 0600)
	if err != nil {
		panic(err)
	}

	archaius.Set(rbacsvc.InitPassword, devPwd1)
	ctx := context.TODO()
	rbacsvc.DeleteAccount(ctx, "root")

	rbacsvc.Init()
	rest.RegisterServant(&v4.AuthResource{})
	rest.RegisterServant(&v4.RoleResource{})
}

func TestAuthResource_Login(t *testing.T) {
	ctx := context.TODO()

	rbacsvc.DeleteAccount(ctx, devAccount)

	t.Run("invalid user login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd1})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	// root account token
	var rootToken = &rbacmodel.Token{}
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1, Roles: []string{"admin"}})

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
		b, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd1, Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("given a valid token and deleted account, auth should fail", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "admin_account", Password: devPwd1, Roles: []string{"admin"}})
		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b, _ = json.Marshal(&rbacmodel.Account{Name: "admin_account", Password: devPwd1})
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
		b, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd1, Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		b2, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, CurrentPassword: devPwd1, Password: devPwd2})
		r2, _ := http.NewRequest(http.MethodPost, "/v4/accounts/"+devAccount+"/password", bytes.NewBuffer(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		b3, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd2, Roles: []string{"developer"}})
		r3, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b3))
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})

}

func TestAuthResource_DeleteAccount(t *testing.T) {
	rootToken := getToken(t, "root", devPwd1)

	t.Run("given root, try to delete it, should fail ", func(t *testing.T) {

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/root", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+rootToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusForbidden, w2.Code)
	})
	t.Run("dev_account can not even delete him self", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd2, Roles: []string{"developer"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devTo := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), devTo)

		r2, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/"+devAccount, nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+devTo.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusForbidden, w2.Code)
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
		to := getToken(t, "root", devPwd1)

		b, _ := json.Marshal(&rbacmodel.Account{Name: "delete_account", Password: devPwd1, Roles: []string{rbacmodel.RoleDeveloper}})
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
		to := getToken(t, "root", devPwd1)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/accounts/"+devAccount, nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		a := &rbacmodel.Account{}
		json.Unmarshal(w3.Body.Bytes(), a)
		assert.Equal(t, devAccount, a.Name)
		assert.Equal(t, []string{"developer"}, a.Roles)
		assert.Empty(t, a.Password)
	})

	t.Run("list account", func(t *testing.T) {
		to := getToken(t, "root", devPwd1)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/accounts", nil)
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
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1, TokenExpirationTime: "10s"})
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get a expiration token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1, TokenExpirationTime: "15m"})
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
		b, _ := json.Marshal(&rbacmodel.Account{Name: devAccount, Password: devPwd1})

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

func TestAuthResource_SelfPerms(t *testing.T) {
	t.Run("get self perms with invalid token, should return 401", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/self-perms", nil)
		r.RemoteAddr = "1.1.1.1"
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("get self perms with root token, should return 200", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r, _ = http.NewRequest(http.MethodGet, "/v4/self-perms", nil)
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		r.RemoteAddr = "1.1.1.1"
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		resp := &rbacmodel.SelfPermissionResponse{}
		json.Unmarshal(w.Body.Bytes(), resp)
		assert.Equal(t, "*", resp.Perms[0].Verbs[0])
	})
	t.Run("get self perms with dev token, should return 200", func(t *testing.T) {
		rbacsvc.DeleteAccount(context.Background(), testSelfPermsAccount)
		rbacsvc.CreateAccount(context.Background(), &rbacmodel.Account{Name: testSelfPermsAccount, Password: devPwd1, Roles: []string{"developer"}})
		b, _ := json.Marshal(&rbacmodel.Account{Name: testSelfPermsAccount, Password: devPwd1})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		r.RemoteAddr = "1.1.1.1"
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r, _ = http.NewRequest(http.MethodGet, "/v4/self-perms", nil)
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		r.RemoteAddr = "1.1.1.1"
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		resp := &rbacmodel.SelfPermissionResponse{}
		json.Unmarshal(w.Body.Bytes(), resp)
		assert.Equal(t, "*", resp.Perms[0].Verbs[0])
	})
}

func TestAuthResource_ListLock(t *testing.T) {
	t.Run("admin list account, should pass", func(t *testing.T) {
		to := getToken(t, "root", devPwd1)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account-locks", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)
		resp := &rbac.LockResponse{}
		err := json.Unmarshal(w3.Body.Bytes(), resp)
		assert.NoError(t, err)
	})
	t.Run("not admin list account, should return 403", func(t *testing.T) {
		const testListLock = "list_account_lock"
		err := rbacsvc.CreateAccount(context.TODO(), &rbacmodel.Account{Name: testListLock, Password: devPwd1, Roles: []string{"developer"}})
		assert.NoError(t, err)

		defer rbacsvc.DeleteAccount(context.TODO(), testListLock)

		to := getToken(t, testListLock, devPwd1)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/account-locks", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusForbidden, w3.Code)
	})
}

func TestAuthResource_BatchCreateAccount(t *testing.T) {
	to := getToken(t, "root", devPwd1)

	b, _ := json.Marshal(&rbacmodel.BatchCreateAccountsRequest{Accounts: []*rbacmodel.Account{
		{Name: "TestAuthResource_BatchCreateAccount_1", Password: devPwd1, Roles: []string{"developer"}},
	}})
	r, _ := http.NewRequest(http.MethodPost, "/v4/accounts/batch-create", bytes.NewBuffer(b))
	r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
	w := httptest.NewRecorder()
	rest.GetRouter().ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	defer rbacsvc.DeleteAccount(context.TODO(), "TestAuthResource_BatchCreateAccount_1")

	a := &rbacmodel.BatchCreateAccountsResponse{}
	json.Unmarshal(w.Body.Bytes(), a)

	item := a.Accounts[0]
	assert.Empty(t, item.Error)
	assert.Equal(t, "TestAuthResource_BatchCreateAccount_1", item.Name)
}

func getToken(t *testing.T, name, pwd string) *rbacmodel.Token {
	b, _ := json.Marshal(&rbacmodel.Account{Name: name, Password: pwd})
	r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
	w := httptest.NewRecorder()
	rest.GetRouter().ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	to := &rbacmodel.Token{}
	json.Unmarshal(w.Body.Bytes(), to)
	return to
}

func BenchmarkAuthResource_LoginP(b *testing.B) {
	body, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			rest.GetRouter().ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				b.Fatal(w.Code)
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkAuthResource_Login(b *testing.B) {
	body, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: devPwd1})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			b.Fatal(w.Code)
		}

	}
	b.ReportAllocs()
}

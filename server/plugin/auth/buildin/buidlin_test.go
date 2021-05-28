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

package buildin_test

// initialize
import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/astaxie/beego"
	carirbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	beego.AppConfig.Set("rbac_enabled", "true")
	beego.AppConfig.Set("rbac_rsa_public_key_file", "./rbac.pub")
	beego.AppConfig.Set("rbac_rsa_private_key_file", "./private.key")
	config.Init()
}

func TestTokenAuthenticator_Identify(t *testing.T) {
	dao.DeleteAccount(context.TODO(), "root")
	dao.DeleteAccount(context.TODO(), "non-admin")
	t.Run("init rbac", func(t *testing.T) {
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

		rbac.Init()
	})
	a := buildin.New()
	ta := a.(*buildin.TokenAuthenticator)

	t.Run("without auth header should failed", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/any", nil)
		err := ta.Identify(r)
		t.Log(err)
		assert.Equal(t, carirbac.ErrNoHeader, err)
	})

	t.Run("with wrong auth header should failed", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/any", nil)
		r.Header.Set(restful.HeaderAuth, "Bear")
		err := ta.Identify(r)
		t.Log(err)
		assert.Equal(t, carirbac.ErrInvalidHeader, err)
	})

	t.Run("with valid header and invalid token, should failed", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/any", nil)
		r.Header.Set(restful.HeaderAuth, "Bear fake_token")
		err := ta.Identify(r)
		t.Log(err)
		assert.Error(t, err)
	})
	t.Run("valid admin token, should be able to get account", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v4/accounts", nil)
		to, err := authr.Login(context.TODO(), "root", "Complicated_password1")
		assert.NoError(t, err)
		r.Header.Set(restful.HeaderAuth, "Bear "+to)
		err = ta.Identify(r)
		assert.NoError(t, err)
	})
	t.Run("valid normal token, should no be able to get account", func(t *testing.T) {
		err := dao.CreateAccount(context.TODO(), &rbacmodel.Account{Name: "non-admin", Password: "Complicated_password1"})
		assert.NoError(t, err)
		r := httptest.NewRequest(http.MethodGet, "/v4/accounts", nil)
		to, err := authr.Login(context.TODO(), "non-admin", "Complicated_password1")
		assert.NoError(t, err)
		r.Header.Set(restful.HeaderAuth, "Bear "+to)
		err = ta.Identify(r)
		t.Log(err)
		assert.Error(t, err)
	})
	t.Run("valid normal token, should no be able to delete account", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v4/accounts", nil)
		v := r.URL.Query()
		v.Set(":name", "non-admin")
		to, err := authr.Login(context.TODO(), "non-admin", "Complicated_password1")
		assert.NoError(t, err)
		r.Header.Set(restful.HeaderAuth, "Bear "+to)
		err = ta.Identify(r)
		t.Log(err)
		assert.Error(t, err)
	})
	t.Run("valid admin token, should be able to delete account", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v4/accounts/:name", nil)
		v := r.URL.Query()
		v.Set(":name", "admin")
		to, err := authr.Login(context.TODO(), "root", "Complicated_password1")
		assert.NoError(t, err)
		r.Header.Set(restful.HeaderAuth, "Bear "+to)
		err = ta.Identify(r)
		assert.NoError(t, err)
	})

	t.Run("TestTokenAuthenticator_ResourceScopes", func(t *testing.T) {
		url := "/v4/accounts/:name"
		a := buildin.New()
		ta := a.(*buildin.TokenAuthenticator)

		tests := []struct {
			name     string
			request  *http.Request
			wantType string
			wantVerb string
		}{
			{"method GET should return verb=get",
				httptest.NewRequest(http.MethodPut, url, nil), "account", "update"},
			{"method POST should return verb=create",
				httptest.NewRequest(http.MethodPost, url, nil), "account", "create"},
			{"method PUT should return verb=update",
				httptest.NewRequest(http.MethodPut, url, nil), "account", "update"},
			{"method DELETE should return verb=delete",
				httptest.NewRequest(http.MethodDelete, url, nil), "account", "delete"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				util.SetRequestContext(tt.request, rest.CtxMatchPattern, url)
				scopes := ta.ResourceScopes(tt.request)
				assert.NotNil(t, scopes)
				assert.Equal(t, tt.wantType, scopes.Type)
				assert.Equal(t, tt.wantVerb, scopes.Verb)
			})
		}
	})
}

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
	"os"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/server/config"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/stretchr/testify/assert"
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

	archaius.Set(rbacsvc.InitPassword, "Complicated_password1")
	rbacsvc.Init()
}

func TestInitRBAC(t *testing.T) {
	ctx := context.Background()

	t.Run("login and authenticate", func(t *testing.T) {
		token, err := authr.Login(ctx, "root", "Complicated_password1")
		assert.NoError(t, err)
		claims, err := authr.Authenticate(ctx, token)
		assert.NoError(t, err)
		assert.Equal(t, "root", claims.(map[string]interface{})[rbac.ClaimsUser])
	})

	t.Run("second time init", func(t *testing.T) {
		rbacsvc.Init()
	})

	t.Run("get private key after init should be not nil", func(t *testing.T) {
		key, err := rbacsvc.GetPrivateKey()
		assert.NoError(t, err)
		assert.NotNil(t, key)
	})

	t.Run("change pwd,admin can change any one password", func(t *testing.T) {
		accountName := "admin_change_other_pwd"
		persisted := newAccount(accountName)
		err := rbacsvc.CreateAccount(ctx, persisted)
		assert.NoError(t, err)
		defer rbacsvc.DeleteAccount(ctx, accountName)

		claims := map[string]interface{}{
			rbac.ClaimsUser:  "test",
			rbac.ClaimsRoles: []interface{}{rbac.RoleAdmin},
		}
		ctx := context.WithValue(ctx, rbacsvc.CtxRequestClaims, claims)
		err = rbacsvc.ChangePassword(ctx, &rbac.Account{Name: persisted.Name, Password: "Complicated_password2"})
		assert.NoError(t, err)
		a, err := rbacsvc.GetAccount(ctx, persisted.Name)
		assert.NoError(t, err)
		assert.True(t, privacy.SamePassword(a.Password, "Complicated_password2"))
	})
	t.Run("admin change self, must provide current pwd", func(t *testing.T) {
		accountName := "admin_change_self"
		a := newAccount(accountName)
		a.Roles = []string{rbac.RoleAdmin}
		err := rbacsvc.CreateAccount(context.TODO(), a)
		assert.Nil(t, err)
		defer rbacsvc.DeleteAccount(ctx, accountName)

		claims := map[string]interface{}{
			rbac.ClaimsUser:  accountName,
			rbac.ClaimsRoles: []interface{}{rbac.RoleAdmin},
		}
		ctx := context.WithValue(ctx, rbacsvc.CtxRequestClaims, claims)
		err = rbacsvc.ChangePassword(ctx, &rbac.Account{Name: a.Name, CurrentPassword: "", Password: testPwd1})
		assert.True(t, errsvc.IsErrEqualCode(err, discovery.ErrInvalidParams))

		err = rbacsvc.ChangePassword(ctx, &rbac.Account{Name: a.Name, CurrentPassword: testPwd0, Password: testPwd1})
		assert.Nil(t, err)
	})
	t.Run("change self password", func(t *testing.T) {
		accountName := "change_self_pwd"
		a := newAccount(accountName)
		err := rbacsvc.CreateAccount(ctx, a)
		assert.NoError(t, err)
		defer rbacsvc.DeleteAccount(ctx, accountName)

		claims := map[string]interface{}{
			rbac.ClaimsUser:  accountName,
			rbac.ClaimsRoles: []interface{}{rbac.RoleDeveloper},
		}
		ctx := context.WithValue(ctx, rbacsvc.CtxRequestClaims, claims)
		err = rbacsvc.ChangePassword(ctx, &rbac.Account{Name: a.Name, CurrentPassword: testPwd0, Password: testPwd1})
		assert.NoError(t, err)
		resp, err := rbacsvc.GetAccount(ctx, a.Name)
		assert.NoError(t, err)
		assert.True(t, privacy.SamePassword(resp.Password, testPwd1))
	})
	t.Run("no admin account change other user password, should return: "+discovery.NewError(discovery.ErrForbidden, "").Error(), func(t *testing.T) {
		accountName := "test"
		a := newAccount(accountName)
		defer rbacsvc.DeleteAccount(ctx, accountName)

		claims := map[string]interface{}{
			rbac.ClaimsUser:  "change_other_user_password",
			rbac.ClaimsRoles: []interface{}{rbac.RoleDeveloper},
		}
		ctx := context.WithValue(ctx, rbacsvc.CtxRequestClaims, claims)
		err := rbacsvc.ChangePassword(ctx, &rbac.Account{Name: a.Name, CurrentPassword: testPwd0, Password: testPwd1})
		assert.True(t, errsvc.IsErrEqualCode(err, discovery.ErrForbidden))
	})
}

func BenchmarkAuthResource_LoginP(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := authr.Login(context.TODO(), "root", "Complicated_password1")
			if err != nil {
				panic(err)
			}
		}
	})
	b.ReportAllocs()
}
func BenchmarkAuthResource_Login(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := authr.Login(context.TODO(), "root", "Complicated_password1")
		if err != nil {
			panic(err)
		}

	}
	b.ReportAllocs()
}

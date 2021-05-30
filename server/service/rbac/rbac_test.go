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
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/server/config"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/astaxie/beego"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/secret"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
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
	ioutil.WriteFile("./private.key", b, 0600)
	b, err = secret.RSAPublicKey2Bytes(pub)
	err = ioutil.WriteFile("./rbac.pub", b, 0600)
	if err != nil {
		panic(err)
	}

	archaius.Set(rbacsvc.InitPassword, "Complicated_password1")
	dao.DeleteAccount(context.Background(), "root")
	dao.DeleteAccount(context.Background(), "a")
	dao.DeleteAccount(context.Background(), "b")

	rbacsvc.Init()
}

func TestInitRBAC(t *testing.T) {
	a, err := dao.GetAccount(context.Background(), "root")
	assert.NoError(t, err)
	assert.Equal(t, "root", a.Name)

	t.Run("login and authenticate", func(t *testing.T) {
		token, err := authr.Login(context.Background(), "root", "Complicated_password1")
		assert.NoError(t, err)
		claims, err := authr.Authenticate(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, "root", claims.(map[string]interface{})[rbac.ClaimsUser])
	})

	t.Run("second time init", func(t *testing.T) {
		rbacsvc.Init()
	})

	t.Run("change pwd,admin can change any one password", func(t *testing.T) {
		persisted := &rbac.Account{Name: "a", Password: "Complicated_password1"}
		err := dao.CreateAccount(context.Background(), persisted)
		assert.NoError(t, err)
		err = rbacsvc.ChangePassword(context.Background(), []string{rbac.RoleAdmin}, "admin", &rbac.Account{Name: "a", Password: "Complicated_password2"})
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), "a")
		assert.NoError(t, err)
		assert.True(t, privacy.SamePassword(a.Password, "Complicated_password2"))
	})
	t.Run("change self password", func(t *testing.T) {
		err := dao.CreateAccount(context.Background(), &rbac.Account{Name: "b", Password: "Complicated_password1"})
		assert.NoError(t, err)
		err = rbacsvc.ChangePassword(context.Background(), nil, "b", &rbac.Account{Name: "b", CurrentPassword: "Complicated_password1", Password: "Complicated_password2"})
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), "b")
		assert.NoError(t, err)
		assert.True(t, privacy.SamePassword(a.Password, "Complicated_password2"))

	})
	t.Run("list kv", func(t *testing.T) {
		_, n, err := dao.ListAccount(context.TODO())
		assert.NoError(t, err)
		assert.Greater(t, n, int64(2))
	})
}
func BenchmarkAuthResource_Login(b *testing.B) {
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
func BenchmarkAuthResource_Login2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := authr.Login(context.TODO(), "root", "Complicated_password1")
		if err != nil {
			panic(err)
		}

	}
	b.ReportAllocs()
}

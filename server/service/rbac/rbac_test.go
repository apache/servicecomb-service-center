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
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/astaxie/beego"
	rbacmodel "github.com/go-chassis/cari/rbac"
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

	archaius.Set(rbac.InitPassword, "Complicated_password1")
	dao.DeleteAccount(context.Background(), "root")
	dao.DeleteAccount(context.Background(), "a")
	dao.DeleteAccount(context.Background(), "b")

	rbac.Init()
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
		assert.Equal(t, "root", claims.(map[string]interface{})[rbacmodel.ClaimsUser])
	})

	t.Run("second time init", func(t *testing.T) {
		rbac.Init()
	})

	t.Run("change pwd,admin can change any one password", func(t *testing.T) {
		persisted := &rbacmodel.Account{Name: "a", Password: "Complicated_password1"}
		err := dao.CreateAccount(context.Background(), persisted)
		assert.NoError(t, err)
		err = rbac.ChangePassword(context.Background(), []string{rbacmodel.RoleAdmin}, "admin", &rbacmodel.Account{Name: "a", Password: "Complicated_password2"})
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), "a")
		assert.NoError(t, err)
		assert.True(t, privacy.SamePassword(a.Password, "Complicated_password2"))
	})
	t.Run("change self password", func(t *testing.T) {
		err := dao.CreateAccount(context.Background(), &rbacmodel.Account{Name: "b", Password: "Complicated_password1"})
		assert.NoError(t, err)
		err = rbac.ChangePassword(context.Background(), nil, "b", &rbacmodel.Account{Name: "b", CurrentPassword: "Complicated_password1", Password: "Complicated_password2"})
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
	dao.DeleteRole(context.Background(), "tester")
	t.Run("check is there exist role", func(t *testing.T) {
		exist, err := dao.RoleExist(context.Background(), "admin")
		assert.NoError(t, err)
		assert.Equal(t, true, exist)

		exist, err = dao.RoleExist(context.Background(), "developer")
		assert.NoError(t, err)
		assert.Equal(t, true, exist)

		exist, err = dao.RoleExist(context.Background(), "tester")
		assert.NoError(t, err)
		assert.Equal(t, false, exist)
	})

	t.Run("delete the default role", func(t *testing.T) {
		r, err := dao.DeleteRole(context.Background(), "admin")
		assert.NoError(t, err)
		assert.Equal(t, false, r)

		r, err = dao.DeleteRole(context.Background(), "developer")
		assert.NoError(t, err)
		assert.Equal(t, false, r)
	})

	t.Run("delete the not exist role", func(t *testing.T) {
		_, err := dao.DeleteRole(context.Background(), "tester")
		assert.NoError(t, err)
	})

	t.Run("list exist role", func(t *testing.T) {
		_, _, err := dao.ListRole(context.TODO())
		assert.NoError(t, err)
	})

	tester := &rbacmodel.Role{
		Name: "tester",
		Perms: []*rbacmodel.Permission{
			{
				Resources: []*rbacmodel.Resource{{Type: "service"}, {Type: "instance"}},
				Verbs:     []string{"get", "create", "update"},
			},
			{
				Resources: []*rbacmodel.Resource{{Type: "rule"}},
				Verbs:     []string{"*"},
			},
		},
	}

	t.Run("create new role success", func(t *testing.T) {
		err := dao.CreateRole(context.Background(), tester)
		assert.NoError(t, err)

		exist, err := dao.RoleExist(context.Background(), "tester")
		assert.NoError(t, err)
		assert.Equal(t, true, exist)

		r, err := dao.GetRole(context.Background(), "tester")
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("edit new role success", func(t *testing.T) {
		err := dao.EditRole(context.Background(), tester.Name, tester)
		assert.NoError(t, err)

		r, err := dao.GetRole(context.Background(), "tester")
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("delete the new role", func(t *testing.T) {
		r, err := dao.DeleteRole(context.Background(), "tester")
		assert.NoError(t, err)
		assert.Equal(t, true, r)
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

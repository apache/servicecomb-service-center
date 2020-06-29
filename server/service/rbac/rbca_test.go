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
	"github.com/apache/servicecomb-service-center/pkg/model"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	"github.com/apache/servicecomb-service-center/server/plugin/security/buildin"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing/buildin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/security/authr"
	"github.com/go-chassis/go-chassis/security/secret"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	beego.AppConfig.Set("rbac_enabled", "true")
	beego.AppConfig.Set(rbac.PubFilePath, "./rbac.pub")
	beego.AppConfig.Set("rbac_rsa_private_key_file", "./private.key")
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "etcd", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.CIPHER, "buildin", plain.New})
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", buildin.New})
}

func TestInitRBAC(t *testing.T) {
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

	archaius.Set(rbac.InitPassword, "root")

	rbac.Init()
	a, err := dao.GetAccount(context.Background(), "root")
	assert.NoError(t, err)
	assert.Equal(t, "root", a.Name)

	t.Run("login and authenticate", func(t *testing.T) {
		token, err := authr.Login(context.Background(), "root", "root")
		assert.NoError(t, err)
		t.Log(token)
		claims, err := authr.Authenticate(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, "root", claims.(map[string]interface{})[rbac.ClaimsUser])
	})

	t.Run("second time init", func(t *testing.T) {
		rbac.Init()
	})

	t.Run("change pwd,admin can change any one password", func(t *testing.T) {
		dao.CreateAccount(context.Background(), &model.Account{Name: "a", Password: "123"})
		err := rbac.ChangePassword(context.Background(), model.RoleAdmin, "admin", &model.Account{Name: "a", Password: "1234"})
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), "a")
		assert.NoError(t, err)
		assert.Equal(t, "1234", a.Password)
	})
	t.Run("change own password", func(t *testing.T) {
		dao.CreateAccount(context.Background(), &model.Account{Name: "b", Password: "123"})
		err := rbac.ChangePassword(context.Background(), "", "b", &model.Account{CurrentPassword: "123", Password: "1234"})
		assert.NoError(t, err)
		a, err := dao.GetAccount(context.Background(), "b")
		assert.NoError(t, err)
		assert.Equal(t, "1234", a.Password)
	})
}

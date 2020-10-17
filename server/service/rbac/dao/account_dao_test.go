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

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "etcd", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{tracing.TRACING, "buildin", pzipkin.New})

}
func TestAccountDao_CreateAccount(t *testing.T) {
	dao.DeleteAccount(context.TODO(), "admin")
	_ = dao.CreateAccount(context.Background(), &rbacframe.Account{Name: "admin", Password: "pwd"})
	t.Run("get account", func(t *testing.T) {
		r, err := dao.GetAccount(context.Background(), "admin")
		assert.NoError(t, err)
		assert.Equal(t, "admin", r.Name)
		hash, err := bcrypt.GenerateFromPassword([]byte("pwd"), 14)
		err = bcrypt.CompareHashAndPassword(hash, []byte("pwd"))
		assert.NoError(t, err)
	})
}

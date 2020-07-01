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

import (
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	plain "github.com/apache/servicecomb-service-center/server/plugin/security/buildin"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/security/secret"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", pzipkin.New})

}
func TestTokenAuthenticator_Identify(t *testing.T) {
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

		archaius.Set(rbac.InitPassword, "root")

		rbac.Init()
	})
	a := buildin.New()
	ta := a.(*buildin.TokenAuthenticator)
	r := httptest.NewRequest(http.MethodGet, "/any", nil)
	err := ta.Identify(r)
	t.Log(err)
	assert.Equal(t, auth.ErrNoHeader, err)
}

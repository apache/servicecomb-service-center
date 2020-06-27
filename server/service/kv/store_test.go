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

package kv_test

import (
	"context"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/pkg/registry/etcd"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/tracing/buildin"
	"github.com/apache/servicecomb-service-center/server/service/kv"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "etcd", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", buildin.New})

}
func TestStoreData(t *testing.T) {

	t.Run("put, get string", func(t *testing.T) {
		err := kv.Put(context.Background(), "test", "value")
		assert.NoError(t, err)
		r, err := kv.Get(context.Background(), "test")
		assert.NoError(t, err)
		assert.Equal(t, "value", string(r.Value))
	})
	t.Run("should exist", func(t *testing.T) {
		b, err := kv.Exist(context.Background(), "test")
		assert.NoError(t, err)
		assert.True(t, b)
	})

	t.Run("put, get bytes", func(t *testing.T) {
		err := kv.PutBytes(context.Background(), "test", []byte(`value`))
		assert.NoError(t, err)
		r, err := kv.Get(context.Background(), "test")
		assert.NoError(t, err)
		assert.Equal(t, "value", string(r.Value))
	})
}

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package sd

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

var depCache *MongoCacher

var depRule1 = model.DependencyRule{
	Type:    "p",
	Domain:  "default",
	Project: "default",
	ServiceKey: &discovery.MicroServiceKey{
		Tenant:      "default/default",
		Environment: "prod",
		AppId:       "appid1",
		ServiceName: "svc1",
		Alias:       "alias1",
		Version:     "1.0",
	},
}

var depRule2 = model.DependencyRule{
	Type:    "p",
	Domain:  "default",
	Project: "default",
	ServiceKey: &discovery.MicroServiceKey{
		Tenant:      "default/default",
		Environment: "prod",
		AppId:       "appid1",
		ServiceName: "svc1",
		Alias:       "alias1",
		Version:     "2.0",
	},
}

func init() {
	depCache = newDepStore()
}

func TestDepCacheBasicFunc(t *testing.T) {
	t.Run("init depcache, should pass", func(t *testing.T) {
		depCache := newDepStore()
		assert.NotNil(t, depCache)
		assert.Equal(t, dep, depCache.cache.Name())
	})
	event1 := MongoEvent{
		DocumentID: "id1",
		Value:      depRule1,
	}
	event2 := MongoEvent{
		DocumentID: "id2",
		Value:      depRule2,
	}
	t.Run("update&&delete depcache, should pass", func(t *testing.T) {
		depCache.cache.ProcessUpdate(event1)
		assert.Equal(t, depCache.cache.Size(), 1)
		assert.Nil(t, depCache.cache.Get("id_not_exist"))
		assert.Equal(t, depRule1.ServiceKey.ServiceName, depCache.cache.Get("id1").(model.DependencyRule).ServiceKey.ServiceName)
		assert.Len(t, depCache.cache.GetValue("p/appid1/svc1/1.0"), 1)
		assert.Equal(t, depRule1.ServiceKey.ServiceName, depCache.cache.GetValue("p/appid1/svc1/1.0")[0].(model.DependencyRule).ServiceKey.ServiceName)
		depCache.cache.ProcessUpdate(event2)
		assert.Equal(t, depRule2.ServiceKey.ServiceName, depCache.cache.GetValue("p/appid1/svc1/2.0")[0].(model.DependencyRule).ServiceKey.ServiceName)
		assert.Len(t, depCache.cache.GetValue("p/appid1/svc1/*"), 2)
		depCache.cache.ProcessDelete(event1)
		assert.Nil(t, depCache.cache.Get("id1"))
		assert.Len(t, depCache.cache.GetValue("p/appid1/svc1/1.0"), 0)
		depCache.cache.ProcessDelete(event2)
		assert.Len(t, depCache.cache.GetValue("p/appid1/svc1/*"), 0)
	})
}

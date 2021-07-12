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

var serviceCache *MongoCacher

var svc1 = model.Service{
	Domain:  "default",
	Project: "default",
	Tags:    nil,
	Service: &discovery.MicroService{
		ServiceId:   "123456789",
		AppId:       "appid1",
		ServiceName: "svc1",
		Version:     "1.0",
	},
}

var svc2 = model.Service{
	Domain:  "default",
	Project: "default",
	Tags:    nil,
	Service: &discovery.MicroService{
		ServiceId:   "987654321",
		AppId:       "appid1",
		ServiceName: "svc1",
		Version:     "1.0",
	},
}

func init() {
	serviceCache = newServiceStore()
}

func TestServiceCacheBasicFunc(t *testing.T) {
	t.Run("init serviceCache,should pass", func(t *testing.T) {
		serviceCache := newServiceStore()
		assert.NotNil(t, serviceCache)
		assert.Equal(t, service, serviceCache.cache.Name())
	})
	event1 := MongoEvent{
		DocumentID: "id1",
		Value:      svc1,
	}
	event2 := MongoEvent{
		DocumentID: "id2",
		Value:      svc2,
	}
	t.Run("update&&delete serviceCache, should pass", func(t *testing.T) {
		serviceCache.cache.ProcessUpdate(event1)
		assert.Equal(t, serviceCache.cache.Size(), 1)
		assert.Nil(t, serviceCache.cache.Get("id_not_exist"))
		assert.Equal(t, svc1.Service.ServiceName, serviceCache.cache.Get("id1").(model.Service).Service.ServiceName)
		assert.Len(t, serviceCache.cache.GetValue("default/default/appid1/svc1/1.0"), 1)
		serviceCache.cache.ProcessUpdate(event2)
		assert.Equal(t, serviceCache.cache.Size(), 2)
		assert.Len(t, serviceCache.cache.GetValue("default/default/appid1/svc1/1.0"), 2)
		assert.Len(t, serviceCache.cache.GetValue("default/default/987654321"), 1)
		assert.Len(t, serviceCache.cache.GetValue("default/default/123456789"), 1)
		serviceCache.cache.ProcessDelete(event1)
		assert.Nil(t, serviceCache.cache.Get("id1"))
		assert.Len(t, serviceCache.cache.GetValue("default/default/appid1/svc1/1.0"), 1)
		serviceCache.cache.ProcessDelete(event2)
		assert.Len(t, serviceCache.cache.GetValue("default/default/appid1/svc1/1.0"), 0)
		assert.Nil(t, serviceCache.cache.Get("id2"))
		assert.Len(t, serviceCache.cache.GetValue("default/default/987654321"), 0)
		assert.Len(t, serviceCache.cache.GetValue("default/default/123456789"), 0)
	})
}

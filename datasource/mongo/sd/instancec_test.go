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
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

var instanceCache *MongoCacher

var inst1 = dao.Instance{
	Domain:  "default",
	Project: "default",
	Instance: &discovery.MicroServiceInstance{
		InstanceId: "123456789",
		ServiceId:  "svcid",
	},
}

var inst2 = dao.Instance{
	Domain:  "default",
	Project: "default",
	Instance: &discovery.MicroServiceInstance{
		InstanceId: "987654321",
		ServiceId:  "svcid",
	},
}

func init() {
	instanceCache = newInstanceStore()
}

func TestInstCacheBasicFunc(t *testing.T) {
	t.Run("init instCache,should pass", func(t *testing.T) {
		instanceCache := newInstanceStore()
		assert.NotNil(t, instanceCache)
		assert.Equal(t, instance, instanceCache.cache.Name())
	})
	event1 := MongoEvent{
		DocumentID: "id1",
		Value:      inst1,
	}
	event2 := MongoEvent{
		DocumentID: "id2",
		Value:      inst2,
	}
	t.Run("update&&delete instCache, should pass", func(t *testing.T) {
		instanceCache.cache.ProcessUpdate(event1)
		assert.Equal(t, instanceCache.cache.Size(), 1)
		assert.Nil(t, instanceCache.cache.Get("id_not_exist"))
		assert.Equal(t, inst1.Instance.InstanceId, instanceCache.cache.Get("id1").(dao.Instance).Instance.InstanceId)
		assert.Len(t, instanceCache.cache.GetValue("svcid"), 1)
		instanceCache.cache.ProcessUpdate(event2)
		assert.Equal(t, instanceCache.cache.Size(), 2)
		assert.Len(t, instanceCache.cache.GetValue("svcid"), 2)
		instanceCache.cache.ProcessDelete(event1)
		assert.Nil(t, instanceCache.cache.Get("id1"))
		assert.Len(t, instanceCache.cache.GetValue("svcid"), 1)
		instanceCache.cache.ProcessDelete(event2)
		assert.Len(t, instanceCache.cache.GetValue("svcid"), 0)
		assert.Nil(t, instanceCache.cache.Get("id2"))
		assert.Len(t, instanceCache.cache.GetValue("svcid"), 0)
	})
}

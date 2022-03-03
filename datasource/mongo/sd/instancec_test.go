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

package sd

import (
	"testing"
	"time"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
)

var instanceCache *MongoCacher

var inst1 = model.Instance{
	Domain:  "default",
	Project: "default",
	Instance: &discovery.MicroServiceInstance{
		InstanceId: "123456789",
		ServiceId:  "svcid",
	},
}

var inst2 = model.Instance{
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
	t.Run("add instCache, should pass", func(t *testing.T) {
		instanceCache.cache.ProcessUpdate(event1)
		assert.Equal(t, instanceCache.cache.Size(), 1)
		assert.Nil(t, instanceCache.cache.Get("id_not_exist"))
		assert.Equal(t, inst1.Instance.InstanceId, instanceCache.cache.Get("id1").(model.Instance).Instance.InstanceId)
		assert.Len(t, instanceCache.cache.GetValue("default/default/svcid"), 1)
		instanceCache.cache.ProcessUpdate(event2)
		assert.Equal(t, instanceCache.cache.Size(), 2)
		assert.Len(t, instanceCache.cache.GetValue("default/default/svcid"), 2)

	})

	t.Run("update instCache, should pass", func(t *testing.T) {
		assert.Equal(t, inst1, instanceCache.cache.Get("id1").(model.Instance))
		instUpdate := model.Instance{
			Domain:  "default",
			Project: "default",
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "123456789",
				ServiceId:  "svcid",
				HostName:   "hostUpdate",
			},
		}
		eventUpdate := MongoEvent{
			DocumentID: "id1",
			Value:      instUpdate,
		}
		instanceCache.cache.ProcessUpdate(eventUpdate)
		assert.Equal(t, instUpdate, instanceCache.cache.Get("id1").(model.Instance))
	})

	t.Run("delete instCache, should pass", func(t *testing.T) {
		instanceCache.cache.ProcessDelete(event1)
		assert.Nil(t, instanceCache.cache.Get("id1"))
		assert.Len(t, instanceCache.cache.GetValue("default/default/svcid"), 1)
		instanceCache.cache.ProcessDelete(event2)
		assert.Len(t, instanceCache.cache.GetValue("default/default/svcid"), 0)
		assert.Nil(t, instanceCache.cache.Get("id2"))
		assert.Len(t, instanceCache.cache.GetValue("default/default/svcid"), 0)
	})
}

func TestInstValueUpdate(t *testing.T) {
	inst1 := model.Instance{
		Domain:      "d1",
		Project:     "p1",
		RefreshTime: time.Time{},
		Instance: &discovery.MicroServiceInstance{
			InstanceId: "123",
			Version:    "1.0",
		},
	}
	inst2 := model.Instance{
		Domain:      "d1",
		Project:     "p1",
		RefreshTime: time.Time{},
		Instance: &discovery.MicroServiceInstance{
			InstanceId: "123",
			Version:    "1.0",
		},
	}
	inst3 := model.Instance{
		Domain:      "d2",
		Project:     "p2",
		RefreshTime: time.Time{},
		Instance: &discovery.MicroServiceInstance{
			InstanceId: "123",
			Version:    "1.0",
		},
	}
	inst4 := model.Instance{
		Domain:      "d2",
		Project:     "p2",
		RefreshTime: time.Time{},
		Instance: &discovery.MicroServiceInstance{
			InstanceId: "123",
			Version:    "1.1",
		},
	}
	inst5 := model.Instance{
		Domain:      "d2",
		Project:     "p2",
		RefreshTime: time.Now(),
		Instance: &discovery.MicroServiceInstance{
			InstanceId: "123",
			Version:    "1.1",
		},
	}

	t.Run("given the same instances expect equal result", func(t *testing.T) {
		res := instanceCache.cache.isValueNotUpdated(inst1, inst2)
		assert.True(t, res)
	})
	t.Run("given instances with the different domain expect not equal result", func(t *testing.T) {
		res := instanceCache.cache.isValueNotUpdated(inst2, inst3)
		assert.False(t, res)
	})
	t.Run("given instances with the different version expect not equal result", func(t *testing.T) {
		res := instanceCache.cache.isValueNotUpdated(inst3, inst4)
		assert.False(t, res)
	})
	t.Run("given instances with the different refresh time expect  equal result", func(t *testing.T) {
		res := instanceCache.cache.isValueNotUpdated(inst4, inst5)
		assert.True(t, res)
	})
}

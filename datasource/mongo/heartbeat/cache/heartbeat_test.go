/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package heartbeatcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	_ "github.com/apache/servicecomb-service-center/test"
)

var c = heartbeatcache.Configuration()

func TestAddCacheInstance(t *testing.T) {
	t.Run("add cache instance: set the ttl to 4 seconds", func(t *testing.T) {
		instance1 := model.Instance{
			RefreshTime: time.Now(),
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "instanceID1",
				ServiceId:  "serviceID1",
				HealthCheck: &discovery.HealthCheck{
					Interval: 2,
					Times:    1,
				},
			},
		}
		err := c.AddHeartbeatTask(instance1.Instance.ServiceId, instance1.Instance.InstanceId, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(context.Background(), instance1)
		assert.Equal(t, nil, err)
		time.Sleep(1 * time.Second)
		info, ok := c.InstanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*heartbeatcache.InstanceHeartbeatInfo)
			assert.Equal(t, instance1.Instance.InstanceId, heartBeatInfo.InstanceID)
			assert.Equal(t, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1), heartBeatInfo.TTL)
		}
		time.Sleep(4 * time.Second)
		_, ok = c.InstanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(context.Background(), instance1)
		assert.Equal(t, nil, err)
	})

	t.Run("add cache instance: do not set interval time", func(t *testing.T) {
		instance2 := model.Instance{
			RefreshTime: time.Now(),
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "instanceID2",
				ServiceId:  "serviceID2",
				HealthCheck: &discovery.HealthCheck{
					Interval: 0,
					Times:    0,
				},
			},
		}
		err := c.AddHeartbeatTask(instance2.Instance.ServiceId, instance2.Instance.InstanceId, instance2.Instance.HealthCheck.Interval*(instance2.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(context.Background(), instance2)
		assert.Equal(t, nil, err)
		time.Sleep(1 * time.Second)
		info, ok := c.InstanceHeartbeatStore.Get(instance2.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*heartbeatcache.InstanceHeartbeatInfo)
			assert.Equal(t, instance2.Instance.InstanceId, heartBeatInfo.InstanceID)
			assert.Equal(t, int32(heartbeatcache.DefaultTTL), heartBeatInfo.TTL)
		}
		time.Sleep(heartbeatcache.DefaultTTL * time.Second)
		_, ok = c.InstanceHeartbeatStore.Get(instance2.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(context.Background(), instance2)
		assert.Equal(t, nil, err)
	})
}

func TestRemoveCacheInstance(t *testing.T) {
	t.Run("remove cache instance: the instance has cache and can be deleted successfully", func(t *testing.T) {
		instance3 := model.Instance{
			RefreshTime: time.Now(),
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "instanceID3",
				ServiceId:  "serviceID3",
				HealthCheck: &discovery.HealthCheck{
					Interval: 3,
					Times:    1,
				},
			},
		}
		err := c.AddHeartbeatTask(instance3.Instance.ServiceId, instance3.Instance.InstanceId, instance3.Instance.HealthCheck.Interval*(instance3.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(context.Background(), instance3)
		assert.Equal(t, nil, err)
		time.Sleep(1 * time.Second)
		info, ok := c.InstanceHeartbeatStore.Get(instance3.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*heartbeatcache.InstanceHeartbeatInfo)
			assert.Equal(t, instance3.Instance.InstanceId, heartBeatInfo.InstanceID)
			assert.Equal(t, instance3.Instance.HealthCheck.Interval*(instance3.Instance.HealthCheck.Times+1), heartBeatInfo.TTL)
		}
		time.Sleep(4 * time.Second)
		c.RemoveCacheInstance(instance3.Instance.InstanceId)
		_, ok = c.InstanceHeartbeatStore.Get(instance3.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(context.Background(), instance3)
		assert.Equal(t, nil, err)
	})
}

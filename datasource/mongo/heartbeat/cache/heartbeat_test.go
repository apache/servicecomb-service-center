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

package heartbeatcache

import (
	_ "github.com/apache/servicecomb-service-center/server/init"
)

import (
	"context"
	"testing"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config)
}

func TestAddCacheInstance(t *testing.T) {
	t.Run("add cache instance: set the ttl to 2 seconds", func(t *testing.T) {
		instance1 := db.Instance{
			RefreshTime: time.Now(),
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instanceID1",
				ServiceId:  "serviceID1",
				HealthCheck: &pb.HealthCheck{
					Interval: 1,
					Times:    1,
				},
			},
		}
		err := addHeartbeatTask(instance1.Instance.ServiceId, instance1.Instance.InstanceId, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = client.GetMongoClient().Insert(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
		info, ok := instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*instanceHeartbeatInfo)
			assert.Equal(t, instance1.Instance.InstanceId, heartBeatInfo.instanceID)
			assert.Equal(t, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1), heartBeatInfo.ttl)
		}
		time.Sleep(2 * time.Second)
		_, ok = instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = client.GetMongoClient().Delete(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
	})

	t.Run("add cache instance: do not set interval time", func(t *testing.T) {
		instance1 := db.Instance{
			RefreshTime: time.Now(),
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instanceID1",
				ServiceId:  "serviceID1",
				HealthCheck: &pb.HealthCheck{
					Interval: 0,
					Times:    0,
				},
			},
		}
		err := addHeartbeatTask(instance1.Instance.ServiceId, instance1.Instance.InstanceId, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = client.GetMongoClient().Insert(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
		info, ok := instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*instanceHeartbeatInfo)
			assert.Equal(t, instance1.Instance.InstanceId, heartBeatInfo.instanceID)
			assert.Equal(t, int32(defaultTTL), heartBeatInfo.ttl)
		}
		time.Sleep(defaultTTL * time.Second)
		_, ok = instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = client.GetMongoClient().Delete(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
	})
}

func TestRemoveCacheInstance(t *testing.T) {
	t.Run("remove cache instance: the instance has cache and can be deleted successfully", func(t *testing.T) {
		instance1 := db.Instance{
			RefreshTime: time.Now(),
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instanceID1",
				ServiceId:  "serviceID1",
				HealthCheck: &pb.HealthCheck{
					Interval: 3,
					Times:    1,
				},
			},
		}
		err := addHeartbeatTask(instance1.Instance.ServiceId, instance1.Instance.InstanceId, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		_, err = client.GetMongoClient().Insert(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
		info, ok := instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, true, ok)
		if ok {
			heartBeatInfo := info.(*instanceHeartbeatInfo)
			assert.Equal(t, instance1.Instance.InstanceId, heartBeatInfo.instanceID)
			assert.Equal(t, instance1.Instance.HealthCheck.Interval*(instance1.Instance.HealthCheck.Times+1), heartBeatInfo.ttl)
		}
		time.Sleep(4 * time.Second)
		RemoveCacheInstance(instance1.Instance.InstanceId)
		_, ok = instanceHeartbeatStore.Get(instance1.Instance.InstanceId)
		assert.Equal(t, false, ok)
		_, err = client.GetMongoClient().Delete(context.Background(), db.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
	})
}

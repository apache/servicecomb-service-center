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
	"context"
	"testing"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
)

func TestHeartBeatCheck(t *testing.T) {
	t.Run("heartbeat check: instance does not exist,it should be failed", func(t *testing.T) {
		heartBeatCheck := &HeartBeatCheck{}
		resp, err := heartBeatCheck.Heartbeat(context.Background(), &pb.HeartbeatRequest{
			ServiceId:  "serviceId1",
			InstanceId: "not-exist-ins",
		})
		assert.NotNil(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("heartbeat check: data exists in the cache,but not in db,it should be failed", func(t *testing.T) {
		err := addHeartbeatTask("not-exist-svc", "not-exist-ins", 30)
		assert.Nil(t, err)
		heartBeatCheck := &HeartBeatCheck{}
		resp, err := heartBeatCheck.Heartbeat(context.Background(), &pb.HeartbeatRequest{
			ServiceId:  "serviceId1",
			InstanceId: "not-exist-ins",
		})
		assert.NotNil(t, err)
		assert.NotEqual(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("heartbeat check: data exists in the cache and db,it can be update successfully", func(t *testing.T) {
		heartBeatCheck := &HeartBeatCheck{}
		instanceDB := mongo.Instance{
			RefreshTime: time.Now(),
			Instance: &pb.MicroServiceInstance{
				InstanceId: "instanceIdDB",
				ServiceId:  "serviceIdDB",
				HealthCheck: &pb.HealthCheck{
					Interval: 1,
					Times:    1,
				},
			},
		}
		filter := bson.M{
			mongo.StringBuilder([]string{mongo.ColumnInstance, mongo.ColumnInstanceID}): instanceDB.Instance.InstanceId,
		}
		_, _ = client.GetMongoClient().Delete(context.Background(), mongo.CollectionInstance, filter)
		_, err := client.GetMongoClient().Insert(context.Background(), mongo.CollectionInstance, instanceDB)
		assert.Equal(t, nil, err)
		err = addHeartbeatTask(instanceDB.Instance.ServiceId, instanceDB.Instance.InstanceId, instanceDB.Instance.HealthCheck.Interval*(instanceDB.Instance.HealthCheck.Times+1))
		assert.Equal(t, nil, err)
		resp, err := heartBeatCheck.Heartbeat(context.Background(), &pb.HeartbeatRequest{
			ServiceId:  "serviceIdDB",
			InstanceId: "instanceIdDB",
		})
		assert.Nil(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		_, err = client.GetMongoClient().Delete(context.Background(), mongo.CollectionInstance, filter)
		assert.Nil(t, err)
	})
}

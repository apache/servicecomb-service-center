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

package checker_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat/checker"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
	_ "github.com/apache/servicecomb-service-center/test"
)

func TestHeartbeat(t *testing.T) {
	t.Run("heartbeat: if the instance does not exist,the heartbeat should fail", func(t *testing.T) {
		heartBeatChecker := &checker.HeartBeatChecker{}
		resp, err := heartBeatChecker.Heartbeat(context.Background(), &discovery.HeartbeatRequest{
			ServiceId:  "not-exist-ins",
			InstanceId: "not-exist-ins",
		})
		assert.NotNil(t, err)
		assert.NotEqual(t, discovery.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("heartbeat: if the instance does exist,the heartbeat should succeed", func(t *testing.T) {
		instance1 := model.Instance{
			RefreshTime: time.Now(),
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "instanceId1",
				ServiceId:  "serviceId1",
			},
		}
		_, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(context.Background(), instance1)
		assert.Equal(t, nil, err)
		heartBeatChecker := &checker.HeartBeatChecker{}
		resp, err := heartBeatChecker.Heartbeat(context.Background(), &discovery.HeartbeatRequest{
			ServiceId:  instance1.Instance.ServiceId,
			InstanceId: instance1.Instance.InstanceId,
		})
		assert.Nil(t, err)
		assert.Equal(t, discovery.ResponseSuccess, resp.Response.GetCode())
		filter := util.NewFilter(util.InstanceInstanceID(instance1.Instance.InstanceId))
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(context.Background(), filter)
		assert.Nil(t, err)
	})
}

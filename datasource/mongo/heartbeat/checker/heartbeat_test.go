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
	"github.com/apache/servicecomb-service-center/pkg/log"
	_ "github.com/apache/servicecomb-service-center/test"
)

func TestUpdateInstanceRefreshTime(t *testing.T) {
	t.Run("update instance refresh time: if the instance does not exist,the update should fail", func(t *testing.T) {
		err := checker.UpdateInstanceRefreshTime(context.Background(), "not-exist", "not-exist")
		log.Error("", err)
		assert.NotNil(t, err)
	})

	t.Run("update instance refresh time: if the instance does exist,the update should succeed", func(t *testing.T) {
		instance1 := model.Instance{
			RefreshTime: time.Now(),
			Instance: &discovery.MicroServiceInstance{
				InstanceId: "instanceId1",
				ServiceId:  "serviceId1",
			},
		}
		_, err := mongo.GetClient().GetDB().Collection(model.CollectionInstance).InsertOne(context.Background(), instance1)
		assert.Equal(t, nil, err)
		err = checker.UpdateInstanceRefreshTime(context.Background(), instance1.Instance.ServiceId, instance1.Instance.InstanceId)
		assert.Equal(t, nil, err)
		filter := util.NewFilter(util.InstanceServiceID(instance1.Instance.ServiceId), util.InstanceInstanceID(instance1.Instance.InstanceId))
		result := mongo.GetClient().GetDB().Collection(model.CollectionInstance).FindOne(context.Background(), filter)
		var ins model.Instance
		err = result.Decode(&ins)
		assert.Nil(t, err)
		assert.NotEqual(t, instance1.RefreshTime, ins.RefreshTime)
		filter = util.NewFilter(util.InstanceServiceID(instance1.Instance.ServiceId), util.InstanceInstanceID(instance1.Instance.InstanceId))
		_, err = mongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(context.Background(), filter)
		assert.Nil(t, err)
	})
}

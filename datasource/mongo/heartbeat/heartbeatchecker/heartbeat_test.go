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

package heartbeatchecker

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config)
}

func TestUpdateInstanceRefreshTime(t *testing.T) {
	t.Run("update instance refresh time: if the instance does not exist,the update should fail", func(t *testing.T) {
		err := updateInstanceRefreshTime(context.Background(), "not-exist", "not-exist")
		fmt.Println(err)
		assert.NotNil(t, err)
	})

	t.Run("update instance refresh time: if the instance does exist,the update should succeed", func(t *testing.T) {
		instance1 := mongo.Instance{
			RefreshTime: time.Now(),
			InstanceInfo: &pb.MicroServiceInstance{
				InstanceId: "instanceId1",
				ServiceId:  "serviceId1",
			},
		}
		_, err := client.GetMongoClient().Insert(context.Background(), mongo.CollectionInstance, instance1)
		assert.Equal(t, nil, err)
		err = updateInstanceRefreshTime(context.Background(), instance1.InstanceInfo.ServiceId, instance1.InstanceInfo.InstanceId)
		assert.Equal(t, nil, err)
		filter := bson.M{
			mongo.InstanceID: instance1.InstanceInfo.InstanceId,
			mongo.ServiceID:  instance1.InstanceInfo.ServiceId,
		}
		result, err := client.GetMongoClient().FindOne(context.Background(), mongo.CollectionInstance, filter)
		assert.Nil(t, err)
		var ins mongo.Instance
		err = result.Decode(&ins)
		assert.Nil(t, err)
		assert.NotEqual(t, instance1.RefreshTime, ins.RefreshTime)
		filter = bson.M{
			mongo.InstanceID: instance1.InstanceInfo.InstanceId,
		}
		_, err = client.GetMongoClient().Delete(context.Background(), mongo.CollectionInstance, filter)
		assert.Nil(t, err)
	})
}

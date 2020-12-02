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

package mongo_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config)
	// clean the mongodb
	client.GetMongoClient().Delete(getContext(), mongo.CollectionInstance, bson.M{})
	client.GetMongoClient().Delete(getContext(), mongo.CollectionService, bson.M{})
}

func TestDump(t *testing.T) {
	var serviceID string
	t.Run("Register service&&intance,check dump length, should paas", func(t *testing.T) {
		insertRes, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_dump_service",
				AppId:       "create_dump_instance_appid",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, insertRes.Response.GetCode())
		serviceID = insertRes.ServiceId

		respCreateInst, err := datasource.Instance().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				Endpoints: []string{
					"localhost://127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateInst.Response.GetCode())
		assert.NotEqual(t, "ins_instance", respCreateInst.InstanceId)

		cache := new(dump.Cache)
		datasource.Instance().DumpCache(getContext(), cache)
		assert.Equal(t, 1, len(cache.Microservices))
		assert.Equal(t, 1, len(cache.Instances))
	})
}

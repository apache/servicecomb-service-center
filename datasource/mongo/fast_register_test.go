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

package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestInstance_BatchCreate(t *testing.T) {
	var serviceID string

	mongo.FastRegConfiguration().QueueSize = 100
	fastRegisterService := mongo.NewFastRegisterInstanceService()
	mongo.SetFastRegisterInstanceService(fastRegisterService)

	fastRegisterTimeTask := mongo.NewRegisterTimeTask()
	fastRegisterTimeTask.Start()

	t.Run("given service to register instance expect register success", func(t *testing.T) {
		respCreateService, err := datasource.Instance().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "create_instance_service_ms",
				AppId:       "create_instance_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respCreateService.Response.GetCode())
		serviceID = respCreateService.ServiceId

	})

	t.Run("when instance request is 3 expect batch register success", func(t *testing.T) {
		request := &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: serviceID,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		}

		getInstsReq := &pb.GetInstancesRequest{
			ProviderServiceId: serviceID,
		}

		getInstsResp, err := datasource.Instance().GetInstances(context.TODO(), getInstsReq)
		assert.NoError(t, err)
		beforLen := len(getInstsResp.Instances)

		instanceBatchLen := 3
		for i := 0; i < instanceBatchLen; i++ {
			event := &mongo.InstanceRegisterEvent{Ctx: context.TODO(), Request: request}
			fastRegisterService.AddEvent(event)
		}
		assert.Equal(t, instanceBatchLen, len(mongo.GetFastRegisterInstanceService().InstEventCh))
		time.Sleep(500 * time.Millisecond)

		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().InstEventCh))
		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().FailedInstCh))

		//if mongo is not replSet, batch register will failed, should wait failed instance register
		time.Sleep(5 * time.Second)

		getInstsResp, err = datasource.Instance().GetInstances(context.TODO(), getInstsReq)
		assert.NoError(t, err)
		afterLen := len(getInstsResp.Instances)
		assert.Equal(t, instanceBatchLen, afterLen-beforLen)
	})

	t.Run("when instanceID is custom expect register success", func(t *testing.T) {

		request := &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "customId_ms",
				ServiceId:  serviceID,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		}

		event := &mongo.InstanceRegisterEvent{Ctx: context.TODO(), Request: request}
		fastRegisterService.AddEvent(event)

		assert.Equal(t, 1, len(mongo.GetFastRegisterInstanceService().InstEventCh))
		time.Sleep(500 * time.Millisecond)

		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().InstEventCh))
		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().FailedInstCh))
	})

	t.Run("when has failed instance expect retry register success", func(t *testing.T) {

		request := &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				InstanceId: "failedInstanceId",
				ServiceId:  serviceID,
				Endpoints: []string{
					"createInstance_ms:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		}

		event := &mongo.InstanceRegisterEvent{Ctx: context.TODO(), Request: request}
		fastRegisterService.AddFailedEvent(event)

		time.Sleep(100 * time.Millisecond)

		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().InstEventCh))
		assert.Equal(t, 0, len(mongo.GetFastRegisterInstanceService().FailedInstCh))
	})

	fastRegisterTimeTask.Stop()
}

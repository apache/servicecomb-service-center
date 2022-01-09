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

package etcd_test

import (
	"context"
	"testing"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func microServiceGetContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "sync-micro-service",
		"sync-micro-service"))
}

func TestSyncMicroService(t *testing.T) {
	datasource.EnableSync = true

	var serviceID string

	t.Run("register micro-service", func(t *testing.T) {
		t.Run("register a micro service will create a task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(microServiceGetContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_micro_service_group",
					ServiceName: "sync_micro_service",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			serviceID = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-micro-service",
				Project:      "sync-micro-service",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(microServiceGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("update micro-service", func(t *testing.T) {
		t.Run("update a micro service setting sync-test property will create a task should pass", func(t *testing.T) {
			request := &pb.UpdateServicePropsRequest{
				ServiceId:  serviceID,
				Properties: make(map[string]string),
			}
			request.Properties["sync-test"] = "sync-test"
			resp, err := datasource.GetMetadataManager().UpdateService(microServiceGetContext(), request)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-micro-service",
				Project:      "sync-micro-service",
				ResourceType: datasource.ResourceService,
				Action:       sync.UpdateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(microServiceGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("unregister micro-service", func(t *testing.T) {
		t.Run("unregister a micro service will create a task and a tombstone should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().UnregisterService(microServiceGetContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceID,
				Force:     true,
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-micro-service",
				Project:      "sync-micro-service",
				ResourceType: datasource.ResourceService,
				Action:       sync.DeleteAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(microServiceGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(microServiceGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-micro-service",
				Project:      "sync-micro-service",
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(microServiceGetContext(), tombstones...)
			assert.NoError(t, err)
		})
	})

	datasource.EnableSync = false
}

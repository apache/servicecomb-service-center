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
	_ "github.com/apache/servicecomb-service-center/test"
)

func depGetContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "sync-dep", "sync-dep"))
}

func TestSyncAddOrUpdateDependencies(t *testing.T) {
	datasource.EnableSync = true
	var (
		consumerId  string
		providerId1 string
		providerId2 string
	)

	t.Run("register service", func(t *testing.T) {
		t.Run("create a consumer service will create a service task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_dep_group",
					ServiceName: "sync_dep_consumer",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			consumerId = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(depGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create two provider services will create two service task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_dep_group",
					ServiceName: "sync_dep_provider",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			providerId1 = resp.ServiceId

			resp, err = datasource.GetMetadataManager().RegisterService(depGetContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_dep_group",
					ServiceName: "sync_dep_provider_other",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			providerId2 = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(depGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("create dependencies", func(t *testing.T) {
		t.Run("create dependencies for microServices will create a dependency task should pass", func(t *testing.T) {
			consumer := &pb.MicroServiceKey{
				ServiceName: "sync_dep_consumer",
				AppId:       "sync_dep_group",
				Version:     "1.0.0",
			}
			resp, err := datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "sync_dep_group",
							ServiceName: "sync_dep_provider",
						},
					},
				},
			}, true)
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceDependency,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("add dependencies", func(t *testing.T) {
		t.Run("add dependencies for microServices will create a dependency task should pass", func(t *testing.T) {
			consumer := &pb.MicroServiceKey{
				ServiceName: "sync_dep_consumer",
				AppId:       "sync_dep_group",
				Version:     "1.0.0",
			}
			resp, err := datasource.GetDependencyManager().AddOrUpdateDependencies(depGetContext(), []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "sync_dep_group",
							ServiceName: "sync_dep_provider_other",
						},
					},
				},
			}, false)
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceDependency,
				Action:       sync.UpdateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("unregister consumer and provider", func(t *testing.T) {
		t.Run("unregister consumer and providers will 3 tasks and 3 tombstones should pass", func(t *testing.T) {
			respDelP, err := datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
				ServiceId: consumerId, Force: true,
			})
			assert.NotNil(t, respDelP)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

			respDelP, err = datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
				ServiceId: providerId1, Force: true,
			})
			assert.NotNil(t, respDelP)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

			respDelP, err = datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
				ServiceId: providerId2, Force: true,
			})
			assert.NotNil(t, respDelP)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceService,
				Action:       sync.DeleteAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(depGetContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(depGetContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-dep",
				Project:      "sync-dep",
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.NoError(t, err)
			tombstones, err = tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tombstones))
		})
	})
	datasource.EnableSync = false
}

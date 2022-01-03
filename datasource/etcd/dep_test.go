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
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/test"
)

func depGetContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "sync-dep", "sync-dep"))
}

func TestSyncAddOrUpdateDependencies(t *testing.T) {
	datasource.EnableSync = true
	var (
		consumerId string
		providerId string
	)

	t.Run("register service", func(t *testing.T) {
		t.Run("create a consumer service should pass", func(t *testing.T) {
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
		})
		t.Run("create a provider service should pass", func(t *testing.T) {
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
			providerId = resp.ServiceId
		})
	})

	t.Run("AddOrUpdateDependencies", func(t *testing.T) {
		t.Run("add dependencies should pass", func(t *testing.T) {
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
			}, false)
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceDependency,
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
		t.Run("unregister consumer and provider should pass", func(t *testing.T) {
			respDelP, err := datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
				ServiceId: consumerId, Force: true,
			})
			assert.NotNil(t, respDelP)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

			respDelP, err = datasource.GetMetadataManager().UnregisterService(depGetContext(), &pb.DeleteServiceRequest{
				ServiceId: providerId, Force: true,
			})
			assert.NotNil(t, respDelP)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respDelP.Response.GetCode())

		})
	})
	datasource.EnableSync = false
}

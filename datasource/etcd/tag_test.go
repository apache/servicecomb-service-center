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

func tagContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "sync-tag",
		"sync-tag"))
}

func TestSyncTag(t *testing.T) {
	var serviceID string
	datasource.EnableSync = true
	t.Run("create service", func(t *testing.T) {
		t.Run("register a micro service will create a task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(tagContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_tag_group",
					ServiceName: "sync_micro_service_tag",
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
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(tagContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("add tags", func(t *testing.T) {
		t.Run("add tags for a service will create a task should pass", func(t *testing.T) {
			respAddTages, err := datasource.GetMetadataManager().AddTags(tagContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceID,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, respAddTages.Response.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceKV,
				Action:       sync.UpdateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(tagContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("update a tag", func(t *testing.T) {
		t.Run("update a service tag will create a task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().UpdateTag(tagContext(), &pb.UpdateServiceTagRequest{
				ServiceId: serviceID,
				Key:       "a",
				Value:     "update",
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceKV,
				Action:       sync.UpdateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(tagContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("delete tags", func(t *testing.T) {
		t.Run("delete a service's tags will create a task and a tombstone should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().DeleteTags(tagContext(), &pb.DeleteServiceTagsRequest{
				ServiceId: serviceID,
				Keys:      []string{"a", "b"},
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			respGetTags, err := datasource.GetMetadataManager().GetTags(tagContext(), &pb.GetServiceTagsRequest{
				ServiceId: serviceID,
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			assert.Equal(t, "", respGetTags.Tags["a"])
			assert.Equal(t, "", respGetTags.Tags["b"])
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceKV,
				Action:       sync.DeleteAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(tagContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(tagContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(tagContext(), tombstones...)
			assert.NoError(t, err)
		})
	})

	t.Run("unregister service", func(t *testing.T) {
		t.Run("unregister a service will create a task and a tombstone should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().UnregisterService(tagContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceID,
				Force:     true,
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceService,
				Action:       sync.DeleteAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(tagContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(tagContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-tag",
				Project:      "sync-tag",
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(tagContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(tagContext(), tombstones...)
			assert.NoError(t, err)
		})
	})

	datasource.EnableSync = false
}

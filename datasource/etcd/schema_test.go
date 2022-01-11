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
	csync "github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/test"
)

func schemaContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "sync-schema", "sync-schema"))
}

func TestSyncSchema(t *testing.T) {

	datasource.EnableSync = true
	var serviceID string

	defer schema.Instance().DeleteContent(schemaContext(), &schema.ContentRequest{
		Hash: "hash_1",
	})
	defer schema.Instance().DeleteContent(schemaContext(), &schema.ContentRequest{
		Hash: "hash_2",
	})
	defer schema.Instance().DeleteContent(schemaContext(), &schema.ContentRequest{
		Hash: "hash_2",
	})

	t.Run("register a micro service", func(t *testing.T) {
		t.Run("register a service will create a service task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(schemaContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_schemas_prod",
					ServiceName: "sync_schemas_service",
					Version:     "1.0.1",
					Level:       "FRONT",
					Status:      pb.MS_UP,
					Environment: pb.ENV_PROD,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
			serviceID = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				Action:       csync.CreateAction,
				ResourceType: datasource.ResourceService,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
		})
	})

	t.Run("put schema will execute the PutContent func", func(t *testing.T) {
		t.Run("put content with valid request, will create 3 kv tasks(hash summary content) should pass", func(t *testing.T) {
			err := schema.Instance().PutContent(schemaContext(), &schema.PutContentRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_1",
				Content: &schema.ContentItem{
					Hash:    "hash_1",
					Summary: "summary_1",
					Content: "1111111111",
				},
			})
			assert.NoError(t, err)

			ref, err := schema.Instance().GetRef(schemaContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_1",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "summary_1", ref.Summary)
			assert.Equal(t, "hash_1", ref.Hash)
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
		})
	})

	t.Run("put schemas will execute the PutManyContent func", func(t *testing.T) {
		t.Run("put many content with valid request, will create 7 kv update task (2 ref tasks, 2 content tasks, 2 summary tasks"+
			" 1 service task), two delete kv task, two tombstones(ref and summary) should pass", func(t *testing.T) {
			err := schema.Instance().PutManyContent(schemaContext(), &schema.PutManyContentRequest{
				ServiceID: serviceID,
				SchemaIDs: []string{"schemaID_2", "schemaID_3"},
				Contents: []*schema.ContentItem{
					{
						Hash:    "hash_2",
						Content: "content_2",
						Summary: "summary_2",
					},
					{
						Hash:    "hash_3",
						Content: "content_3",
						Summary: "summary_3",
					},
				},
			})
			assert.NoError(t, err)
			ref, err := schema.Instance().GetRef(schemaContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_2",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "summary_2", ref.Summary)
			assert.Equal(t, "hash_2", ref.Hash)
			ref, err = schema.Instance().GetRef(schemaContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_3",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "summary_3", ref.Summary)
			assert.Equal(t, "hash_3", ref.Hash)
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				Action:       csync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 7, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			listTaskReq = model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err = task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.NoError(t, err)
		})
	})

	t.Run("delete schemas will execute the DeleteRef func and DeleteSchema func ", func(t *testing.T) {
		t.Run("delete schemaID_2 and schemaID_3 will create 4 tasks(2 from DeleteRef, 2 from DeleteSchema ) "+
			"and 4 tombstones (2 from DeleteRef, 2 from DeleteSchema) should pass", func(t *testing.T) {
			err := schema.Instance().DeleteRef(schemaContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_2",
			})
			assert.NoError(t, err)
			err = datasource.GetMetadataManager().DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "schemaID_2",
			})
			assert.Equal(t, schema.ErrSchemaNotFound, err)
			err = schema.Instance().DeleteRef(schemaContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_3",
			})
			assert.NoError(t, err)
			err = datasource.GetMetadataManager().DeleteSchema(schemaContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "schemaID_3",
			})
			assert.Equal(t, schema.ErrSchemaNotFound, err)
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				Action:       csync.DeleteAction,
				ResourceType: datasource.ResourceKV,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				ResourceType: datasource.ResourceKV,
			}
			tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.NoError(t, err)
		})
	})

	t.Run("unregister micro-service", func(t *testing.T) {
		t.Run("unregister a micro service will create a task and a tombstone should pass", func(t *testing.T) {
			err := datasource.GetMetadataManager().UnregisterService(schemaContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceID,
				Force:     true,
			})
			assert.NoError(t, err)

			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				ResourceType: datasource.ResourceService,
				Action:       csync.DeleteAction,
				Status:       csync.PendingStatus,
			}
			tasks, err := task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(schemaContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			tombstoneListReq := model.ListTombstoneRequest{
				Domain:       "sync-schema",
				Project:      "sync-schema",
				ResourceType: datasource.ResourceService,
			}
			tombstones, err := tombstone.List(schemaContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(schemaContext(), tombstones...)
			assert.NoError(t, err)
		})
	})

	datasource.EnableSync = false
}

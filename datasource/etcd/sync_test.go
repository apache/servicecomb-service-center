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
	"strconv"
	"testing"

	pb "github.com/go-chassis/cari/discovery"
	crbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/go-archaius"
	"github.com/little-cui/etcdadpt"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/test"
)

func syncAllContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "sync-all", "sync-all"))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func TestSyncAll(t *testing.T) {
	t.Run("enableOnStart is false will not do sync", func(t *testing.T) {
		_ = archaius.Set("sync.enableOnStart", false)
		err := datasource.GetSyncManager().SyncAll(syncAllContext())
		assert.Nil(t, err)
	})

	t.Run("enableOnStart is true and syncAllKey exists will not do sync", func(t *testing.T) {
		_ = archaius.Set("sync.enableOnStart", true)
		err := etcdadpt.Put(syncAllContext(), etcd.SyncAllKey, "1")
		assert.Nil(t, err)
		err = datasource.GetSyncManager().SyncAll(syncAllContext())
		assert.Equal(t, datasource.ErrSyncAllKeyExists, err)
		isDeleted, err := etcdadpt.Delete(syncAllContext(), etcd.SyncAllKey)
		assert.Equal(t, isDeleted, true)
		assert.Nil(t, err)
	})

	t.Run("enableOnStart is true and syncAllKey not exists will do sync", func(t *testing.T) {
		_ = archaius.Set("sync.enableOnStart", true)
		var serviceID string
		var accountName string
		var roleName string
		var consumerID string
		var providerID string
		t.Run("register a service and delete the task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(syncAllContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_micro_service_group",
					ServiceName: "sync_micro_service_sync_all",
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
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create a account and delete the task should pass", func(t *testing.T) {
			a1 := crbac.Account{
				ID:                  "sync-create-11111-sync-all",
				Name:                "sync-create-account1-sync-all",
				Password:            "tnuocca-tset",
				Roles:               []string{"admin"},
				TokenExpirationTime: "2020-12-30",
				CurrentPassword:     "tnuocca-tset1",
			}
			err := rbac.Instance().CreateAccount(syncAllContext(), &a1)
			assert.NoError(t, err)
			accountName = a1.Name
			r, err := rbac.Instance().GetAccount(syncAllContext(), a1.Name)
			assert.NoError(t, err)
			assert.Equal(t, a1, *r)
			listTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create a role and delete the task should pass", func(t *testing.T) {
			r1 := crbac.Role{
				ID:    "create-11111-sync-all",
				Name:  "create-role-sync-all",
				Perms: nil,
			}
			err := rbac.Instance().CreateRole(syncAllContext(), &r1)
			assert.NoError(t, err)
			r, err := rbac.Instance().GetRole(syncAllContext(), "create-role-sync-all")
			assert.NoError(t, err)
			assert.Equal(t, r1, *r)
			dt, _ := strconv.Atoi(r.CreateTime)
			assert.Less(t, 0, dt)
			assert.Equal(t, r.CreateTime, r.UpdateTime)
			roleName = r1.Name
			listTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceRole,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
		})
		t.Run("put content with valid request and delete three task should pass", func(t *testing.T) {
			err := schema.Instance().PutContent(syncAllContext(), &schema.PutContentRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_sync_all",
				Content: &schema.ContentItem{
					Hash:    "hash_sync_all",
					Summary: "summary_sync_all",
					Content: "1111111111",
				},
			})
			assert.NoError(t, err)
			ref, err := schema.Instance().GetRef(syncAllContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_sync_all",
			})
			assert.NoError(t, err)
			assert.NotNil(t, ref)
			assert.Equal(t, "summary_sync_all", ref.Summary)
			assert.Equal(t, "hash_sync_all", ref.Hash)
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				Action:       sync.UpdateAction,
				ResourceType: datasource.ResourceKV,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
		})
		t.Run("update a service tag and delete the task should pass", func(t *testing.T) {
			err := datasource.GetMetadataManager().PutManyTags(syncAllContext(), &pb.AddServiceTagsRequest{
				ServiceId: serviceID,
				Tags: map[string]string{
					"a": "test",
					"b": "b",
				},
			})
			assert.NoError(t, err)
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceKV,
				Action:       sync.UpdateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create a consumer service will create a service task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(syncAllContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_dep_group_sync_all",
					ServiceName: "sync_dep_consumer_sync_all",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			consumerID = resp.ServiceId
			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create one provider service will create one service task should pass", func(t *testing.T) {
			resp, err := datasource.GetMetadataManager().RegisterService(syncAllContext(), &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					AppId:       "sync_dep_group_sync_all",
					ServiceName: "sync_dep_provider_sync_all",
					Version:     "1.0.0",
					Level:       "FRONT",
					Status:      pb.MS_UP,
				},
			})
			assert.NotNil(t, resp)
			assert.NoError(t, err)
			providerID = resp.ServiceId

			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("create dependencies for microServices will create a dependency task should pass", func(t *testing.T) {
			consumer := &pb.MicroServiceKey{
				ServiceName: "sync_dep_consumer_sync_all",
				AppId:       "sync_dep_group_sync_all",
				Version:     "1.0.0",
			}
			err := datasource.GetDependencyManager().PutDependencies(syncAllContext(), []*pb.ConsumerDependency{
				{
					Consumer: consumer,
					Providers: []*pb.MicroServiceKey{
						{
							AppId:       "sync_dep_group_sync_all",
							ServiceName: "sync_dep_provider_sync_all",
						},
					},
				},
			}, true)
			assert.NoError(t, err)

			listTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceDependency,
				Action:       sync.CreateAction,
				Status:       sync.PendingStatus,
			}
			tasks, err := task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})

		t.Run("do sync will create task should pass", func(t *testing.T) {
			err := datasource.GetSyncManager().SyncAll(syncAllContext())
			assert.Nil(t, err)
			listServiceTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
			}
			tasks, err := task.List(syncAllContext(), &listServiceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			listKVTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceKV,
			}
			tasks, err = task.List(syncAllContext(), &listKVTaskReq)
			assert.NoError(t, err)
			// three schema and one tag
			assert.Equal(t, 4, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			listAccountTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err = task.List(syncAllContext(), &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			listRoleTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceRole,
			}
			tasks, err = task.List(syncAllContext(), &listRoleTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			listDepTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceDependency,
			}
			tasks, err = task.List(syncAllContext(), &listDepTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			exist, err := etcdadpt.Exist(syncAllContext(), etcd.SyncAllKey)
			assert.Equal(t, true, exist)
			assert.Nil(t, err)
		})

		t.Run("delete all resources should pass", func(t *testing.T) {
			err := schema.Instance().DeleteRef(syncAllContext(), &schema.RefRequest{
				ServiceID: serviceID,
				SchemaID:  "schemaID_sync_all",
			})
			assert.NoError(t, err)
			err = datasource.GetMetadataManager().DeleteSchema(syncAllContext(), &pb.DeleteSchemaRequest{
				ServiceId: serviceID,
				SchemaId:  "schemaID_sync_all",
			})
			err = datasource.GetMetadataManager().UnregisterService(syncAllContext(), &pb.DeleteServiceRequest{
				ServiceId: serviceID,
				Force:     true,
			})
			assert.NoError(t, err)
			_, err = rbac.Instance().DeleteAccount(syncAllContext(), []string{accountName})
			assert.NoError(t, err)
			_, err = rbac.Instance().DeleteRole(syncAllContext(), roleName)
			assert.NoError(t, err)
			err = datasource.GetMetadataManager().UnregisterService(syncAllContext(), &pb.DeleteServiceRequest{
				ServiceId: consumerID, Force: true,
			})
			assert.NoError(t, err)

			err = datasource.GetMetadataManager().UnregisterService(syncAllContext(), &pb.DeleteServiceRequest{
				ServiceId: providerID, Force: true,
			})
			assert.NoError(t, err)

			listSeviceTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
			}
			tasks, err := task.List(syncAllContext(), &listSeviceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 3, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listSeviceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
			listAccountTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err = task.List(syncAllContext(), &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))

			listRoleTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceRole,
			}
			tasks, err = task.List(syncAllContext(), &listRoleTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listRoleTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))

			listKVTaskReq := model.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceKV,
			}
			tasks, err = task.List(syncAllContext(), &listKVTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listKVTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))

			tombstoneListReq := model.ListTombstoneRequest{
				Domain:  "sync-all",
				Project: "sync-all",
			}
			tombstones, err := tombstone.List(syncAllContext(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 7, len(tombstones))
			err = tombstone.Delete(syncAllContext(), tombstones...)
			assert.NoError(t, err)
		})
	})
}

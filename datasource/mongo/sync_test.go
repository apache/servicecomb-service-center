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
	"context"
	"strconv"
	"testing"

	dmongo "github.com/go-chassis/cari/db/mongo"
	pb "github.com/go-chassis/cari/discovery"
	crbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	emodel "github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
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
		_, err := dmongo.GetClient().GetDB().Collection(model.CollectionSync).InsertOne(syncAllContext(),
			bson.M{"key": mongo.SyncAllKey})
		assert.Nil(t, err)
		err = datasource.GetSyncManager().SyncAll(syncAllContext())
		assert.Equal(t, datasource.ErrSyncAllKeyExists, err)
		_, err = dmongo.GetClient().GetDB().Collection(model.CollectionSync).DeleteOne(syncAllContext(),
			bson.M{"key": mongo.SyncAllKey})
		assert.Nil(t, err)
	})

	t.Run("enableOnStart is true and syncAllKey not exists will do sync", func(t *testing.T) {
		_ = archaius.Set("sync.enableOnStart", true)
		var accountName string
		var roleName string
		var consumerID string
		var providerID string
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
			listTaskReq := emodel.ListTaskRequest{
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
			listTaskReq := emodel.ListTaskRequest{
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
			listTaskReq := emodel.ListTaskRequest{
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

			listTaskReq := emodel.ListTaskRequest{
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

		t.Run("do sync will create task should pass", func(t *testing.T) {
			err := datasource.GetSyncManager().SyncAll(syncAllContext())
			assert.Nil(t, err)
			listServiceTaskReq := emodel.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
			}
			tasks, err := task.List(syncAllContext(), &listServiceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listServiceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))

			listAccountTaskReq := emodel.ListTaskRequest{
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

			listRoleTaskReq := emodel.ListTaskRequest{
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

			count, err := dmongo.GetClient().GetDB().Collection(model.CollectionSync).CountDocuments(syncAllContext(),
				bson.M{"key": mongo.SyncAllKey})
			assert.Nil(t, err)
			assert.Equal(t, int64(1), count)
			_, err = dmongo.GetClient().GetDB().Collection(model.CollectionSync).DeleteOne(syncAllContext(),
				bson.M{"key": mongo.SyncAllKey})
			assert.Nil(t, err)
		})

		t.Run("delete all resources should pass", func(t *testing.T) {
			_, err := rbac.Instance().DeleteAccount(syncAllContext(), []string{accountName})
			assert.NoError(t, err)
			listAccountTaskReq := emodel.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err := task.List(syncAllContext(), &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))

			_, err = rbac.Instance().DeleteRole(syncAllContext(), roleName)
			assert.NoError(t, err)
			listRoleTaskReq := emodel.ListTaskRequest{
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

			err = datasource.GetMetadataManager().UnregisterService(syncAllContext(), &pb.DeleteServiceRequest{
				ServiceId: consumerID, Force: true,
			})
			assert.NoError(t, err)

			err = datasource.GetMetadataManager().UnregisterService(syncAllContext(), &pb.DeleteServiceRequest{
				ServiceId: providerID, Force: true,
			})
			assert.NoError(t, err)

			listSeviceTaskReq := emodel.ListTaskRequest{
				Domain:       "sync-all",
				Project:      "sync-all",
				ResourceType: datasource.ResourceService,
			}
			tasks, err = task.List(syncAllContext(), &listSeviceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(syncAllContext(), tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(syncAllContext(), &listSeviceTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
	})

	t.Run("enableOnStart is true ,syncAllKey not exists and context is context.Background() will do sync", func(t *testing.T) {
		_ = archaius.Set("sync.enableOnStart", true)
		var accountName string
		ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "sync-all-background", "sync-all-background"))
		ctx = util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
		t.Run("create a account and delete the task should pass", func(t *testing.T) {
			a1 := crbac.Account{
				ID:                  "sync-create-11111-sync-all",
				Name:                "sync-create-account1-sync-all",
				Password:            "tnuocca-tset",
				Roles:               []string{"admin"},
				TokenExpirationTime: "2020-12-30",
				CurrentPassword:     "tnuocca-tset1",
			}
			err := rbac.Instance().CreateAccount(ctx, &a1)
			assert.NoError(t, err)
			accountName = a1.Name
			r, err := rbac.Instance().GetAccount(ctx, a1.Name)
			assert.NoError(t, err)
			assert.Equal(t, a1, *r)
			listTaskReq := emodel.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err := task.List(ctx, &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(ctx, tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(ctx, &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})

		t.Run("do sync will create task should pass", func(t *testing.T) {
			err := datasource.GetSyncManager().SyncAll(context.Background())
			assert.Nil(t, err)
			listAccountTaskReq := emodel.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err := task.List(ctx, &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(ctx, tasks...)
			assert.NoError(t, err)
			tasks, err = task.List(ctx, &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(tasks))
		})
		t.Run("delete account resource should pass", func(t *testing.T) {
			_, err := rbac.Instance().DeleteAccount(ctx, []string{accountName})
			assert.NoError(t, err)
			listAccountTaskReq := emodel.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceAccount,
			}
			tasks, err := task.List(ctx, &listAccountTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tasks))
			err = task.Delete(ctx, tasks...)
			assert.NoError(t, err)

			tombstoneListReq := emodel.ListTombstoneRequest{
				Domain:  "sync-all-background",
				Project: "sync-all-background",
			}
			tombstones, err := tombstone.List(ctx, &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(ctx, tombstones...)
			assert.NoError(t, err)
		})
	})
}

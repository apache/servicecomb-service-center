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
	"strconv"
	"testing"

	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	_ "github.com/apache/servicecomb-service-center/test"
)

func TestSyncRole(t *testing.T) {

	datasource.EnableSync = true

	t.Run("create role", func(t *testing.T) {
		t.Run("creating a role and delete it will create two tasks and a tombstone should pass", func(t *testing.T) {
			r1 := rbacmodel.Role{
				ID:    "create-11111",
				Name:  "create-role",
				Perms: nil,
			}
			err := rbac.Instance().CreateRole(context.Background(), &r1)
			assert.NoError(t, err)
			r, err := rbac.Instance().GetRole(context.Background(), "create-role")
			assert.NoError(t, err)
			assert.Equal(t, r1, *r)
			dt, _ := strconv.Atoi(r.CreateTime)
			assert.Less(t, 0, dt)
			assert.Equal(t, r.CreateTime, r.UpdateTime)
			_, err = rbac.Instance().DeleteRole(context.Background(), r1.Name)
			assert.NoError(t, err)
			listTaskReq := model.ListTaskRequest{
				Domain:       "",
				Project:      "",
				ResourceType: datasource.ResourceRole,
			}
			tasks, err := task.List(context.Background(), &listTaskReq)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tasks))
			err = task.Delete(context.Background(), tasks...)
			assert.NoError(t, err)
			tombstoneListReq := model.ListTombstoneRequest{
				ResourceType: datasource.ResourceRole,
			}
			tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.NoError(t, err)
		})
	})

	t.Run("update role", func(t *testing.T) {
		t.Run("create two roles ,then update them, finally delete them, will create six tasks and two tombstones should pass",
			func(t *testing.T) {
				r2 := rbacmodel.Role{
					ID:    "update-22222",
					Name:  "update-role-22222",
					Perms: nil,
				}
				r3 := rbacmodel.Role{
					ID:    "update-33333",
					Name:  "update-role-33333",
					Perms: nil,
				}
				err := rbac.Instance().CreateRole(context.Background(), &r2)
				assert.NoError(t, err)
				err = rbac.Instance().CreateRole(context.Background(), &r3)
				assert.NoError(t, err)
				r2.ID = "update-22222-33333"
				err = rbac.Instance().UpdateRole(context.Background(), "update-role-22222", &r2)
				assert.NoError(t, err)
				r3.ID = "update-33333-44444"
				err = rbac.Instance().UpdateRole(context.Background(), "update-role-33333", &r3)
				assert.NoError(t, err)
				_, err = rbac.Instance().DeleteRole(context.Background(), r2.Name)
				assert.NoError(t, err)
				_, err = rbac.Instance().DeleteRole(context.Background(), r3.Name)
				assert.NoError(t, err)
				listTaskReq := model.ListTaskRequest{
					Domain:       "",
					Project:      "",
					ResourceType: datasource.ResourceRole,
				}
				tasks, err := task.List(context.Background(), &listTaskReq)
				assert.NoError(t, err)
				assert.Equal(t, 6, len(tasks))
				err = task.Delete(context.Background(), tasks...)
				assert.NoError(t, err)
				tombstoneListReq := model.ListTombstoneRequest{
					ResourceType: datasource.ResourceRole,
				}
				tombstones, err := tombstone.List(context.Background(), &tombstoneListReq)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(tombstones))
				err = tombstone.Delete(context.Background(), tombstones...)
				assert.NoError(t, err)

			})
	})

	datasource.EnableSync = false

}

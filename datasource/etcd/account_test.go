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

	crbac "github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/test"
)

func accountContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "sync-account", "sync-account"))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func TestSyncAccount(t *testing.T) {

	t.Run("create account", func(t *testing.T) {

		t.Run("creating a account then delete it will create two tasks and a tombstone should pass",
			func(t *testing.T) {
				a1 := crbac.Account{
					ID:                  "sync-create-11111",
					Name:                "sync-create-account1",
					Password:            "tnuocca-tset",
					Roles:               []string{"admin"},
					TokenExpirationTime: "2020-12-30",
					CurrentPassword:     "tnuocca-tset1",
				}
				err := rbac.Instance().CreateAccount(accountContext(), &a1)
				assert.NoError(t, err)
				r, err := rbac.Instance().GetAccount(accountContext(), a1.Name)
				assert.NoError(t, err)
				assert.Equal(t, a1, *r)
				_, err = rbac.Instance().DeleteAccount(accountContext(), []string{a1.Name})
				assert.NoError(t, err)
				listTaskReq := model.ListTaskRequest{
					Domain:       "sync-account",
					Project:      "sync-account",
					ResourceType: datasource.ResourceAccount,
				}
				tasks, err := task.List(accountContext(), &listTaskReq)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(tasks))
				err = task.Delete(accountContext(), tasks...)
				assert.NoError(t, err)
				tombstoneListReq := model.ListTombstoneRequest{
					Domain:       "sync-account",
					Project:      "sync-account",
					ResourceType: datasource.ResourceAccount,
				}
				tombstones, err := tombstone.List(accountContext(), &tombstoneListReq)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(tombstones))
				err = tombstone.Delete(accountContext(), tombstones...)
				assert.NoError(t, err)
			})
	})

	t.Run("update account", func(t *testing.T) {
		t.Run("creating two accounts then update them,finally delete them, will create six tasks and two tombstones should pass",
			func(t *testing.T) {
				a2 := crbac.Account{
					ID:                  "sync-update-22222",
					Name:                "sync-update-account2",
					Password:            "tnuocca-tset",
					Roles:               []string{"admin"},
					TokenExpirationTime: "2020-12-30",
					CurrentPassword:     "tnuocca-tset",
				}
				a3 := crbac.Account{
					ID:                  "sync-update-33333",
					Name:                "sync-update-account3",
					Password:            "tnuocca-tset",
					Roles:               []string{"admin"},
					TokenExpirationTime: "2020-12-30",
					CurrentPassword:     "tnuocca-tset",
				}
				err := rbac.Instance().CreateAccount(accountContext(), &a2)
				assert.NoError(t, err)
				err = rbac.Instance().CreateAccount(accountContext(), &a3)
				assert.NoError(t, err)
				a2.Password = "new-password"
				err = rbac.Instance().UpdateAccount(accountContext(), a2.Name, &a2)
				assert.NoError(t, err)
				a3.Password = "new-password"
				err = rbac.Instance().UpdateAccount(accountContext(), a3.Name, &a3)
				assert.NoError(t, err)
				_, err = rbac.Instance().DeleteAccount(accountContext(), []string{a2.Name, a3.Name})
				assert.NoError(t, err)
				listTaskReq := model.ListTaskRequest{
					Domain:       "sync-account",
					Project:      "sync-account",
					ResourceType: datasource.ResourceAccount,
				}
				tasks, err := task.List(accountContext(), &listTaskReq)
				assert.NoError(t, err)
				assert.Equal(t, 6, len(tasks))
				err = task.Delete(accountContext(), tasks...)
				assert.NoError(t, err)
				tombstoneListReq := model.ListTombstoneRequest{
					Domain:       "sync-account",
					Project:      "sync-account",
					ResourceType: datasource.ResourceAccount,
				}
				tombstones, err := tombstone.List(accountContext(), &tombstoneListReq)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(tombstones))
				err = tombstone.Delete(accountContext(), tombstones...)
				assert.NoError(t, err)
			})
	})
}

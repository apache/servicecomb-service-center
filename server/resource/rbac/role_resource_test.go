/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package rbac_test

import (
	"strings"

	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	_ "github.com/apache/servicecomb-service-center/test"
	"github.com/go-chassis/cari/rbac"

	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
)

func newRole(name string) *rbac.Role {
	return &rbac.Role{
		Name: name,
		Perms: []*rbac.Permission{
			{
				Resources: []*rbac.Resource{
					{
						Type: rbacsvc.ResourceService,
					},
				},
				Verbs: []string{"*"},
			},
		},
	}
}

const (
	testPwd0 = "Ab@00000"
	testPwd1 = "Ab@11111"
)

func newAccount(name string) *rbac.Account {
	return &rbac.Account{
		Name:     name,
		Password: testPwd0,
		Roles:    []string{rbac.RoleAdmin},
		Status:   "active",
	}
}

func TestRoleResource_CreateOrUpdateRole(t *testing.T) {
	var superToken = &rbacmodel.Token{}
	ctx := context.TODO()
	rbacsvc.DeleteAccount(ctx, "dev_test")
	rbacsvc.DeleteAccount(ctx, "dev_test2")
	rbacsvc.DeleteRole(ctx, "tester")
	devAccount := newAccount("dev_test")
	testRole := newRole("tester")
	rbacsvc.DeleteAccount(ctx, devAccount.Name)
	rbacsvc.DeleteRole(ctx, testRole.Name)
	t.Run("root login,to get super token", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), superToken)
	})

	t.Run("create a role name tester ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: rbacsvc.RootName, Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		devToken := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), devToken)

		b2, _ := json.Marshal(testRole)

		r2, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b2))
		r2.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/roles", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		newTestRole := newRole(testRole.Name)
		newTestRole.Perms = []*rbac.Permission{
			{
				Resources: []*rbacmodel.Resource{{Type: rbacsvc.ResourceAccount}},
				Verbs:     []string{"*"},
			},
		}
		b4, _ := json.Marshal(newTestRole)
		r4, _ := http.NewRequest(http.MethodPut, "/v4/roles/tester", bytes.NewReader(b4))
		r4.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusOK, w4.Code)
	})
	t.Run("create account dev_test and add a role", func(t *testing.T) {
		devAccount.Roles = []string{testRole.Name}
		b, _ := json.Marshal(devAccount)

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get role", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/roles/admin", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)

	})
	t.Run("delete role, given admin role, should fail", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		to := &rbacmodel.Token{}
		json.Unmarshal(w.Body.Bytes(), to)

		r3, _ := http.NewRequest(http.MethodDelete, "/v4/roles/admin", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusForbidden, w3.Code)
	})

	t.Run("delete tester role, it is bind to account, should fail", func(t *testing.T) {
		r3, _ := http.NewRequest(http.MethodDelete, "/v4/roles/tester", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusBadRequest, w3.Code)
	})
	t.Run("delete account, then delete role should success", func(t *testing.T) {
		r3, _ := http.NewRequest(http.MethodDelete, "/v4/accounts/dev_test", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		r4, _ := http.NewRequest(http.MethodDelete, "/v4/roles/tester", nil)
		r4.Header.Set(restful.HeaderAuth, "Bearer "+superToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusOK, w4.Code)
	})
}
func TestRoleResource_MoreRoles(t *testing.T) {
	var to = &rbacmodel.Token{}
	ctx := context.TODO()
	rbacsvc.DeleteAccount(ctx, "dev_test")
	rbacsvc.DeleteAccount(ctx, "dev_test2")
	rbacsvc.DeleteRole(ctx, "tester")
	rbacsvc.DeleteRole(ctx, "tester2")
	t.Run("root login", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "root", Password: "Complicated_password1"})

		r, err := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
		}
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		json.Unmarshal(w.Body.Bytes(), to)
	})

	t.Run("create role name tester", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "service"}},
					Verbs:     []string{"get", "create", "update"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create new role name tester2", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Role{
			Name: "tester2",
			Perms: []*rbacmodel.Permission{
				{
					Resources: []*rbacmodel.Resource{{Type: "rule"}},
					Verbs:     []string{"*"},
				},
			},
		})

		r, _ := http.NewRequest(http.MethodPost, "/v4/roles", bytes.NewReader(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodGet, "/v4/roles", nil)
		r2.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("account dev_test2 support more than 1 role ", func(t *testing.T) {
		b, _ := json.Marshal(&rbacmodel.Account{Name: "dev_test2", Password: "Complicated_password3", Roles: []string{"tester", "tester2"}})

		r, _ := http.NewRequest(http.MethodPost, "/v4/accounts", bytes.NewBuffer(b))
		r.Header.Set(restful.HeaderAuth, "Bearer "+to.TokenStr)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r2, _ := http.NewRequest(http.MethodPost, "/v4/token", bytes.NewBuffer(b))
		w2 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
		devToken := &rbacmodel.Token{}
		json.Unmarshal(w2.Body.Bytes(), devToken)

		r3, _ := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices", nil)
		r3.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w3 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusOK, w3.Code)

		reader := strings.NewReader("{\n  \"serviceIds\": [\n    \"MOCK\"\n  ]\n}")
		r4, _ := http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices", reader)
		r4.Header.Set(restful.HeaderAuth, "Bearer "+devToken.TokenStr)
		w4 := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w4, r4)
		assert.Equal(t, http.StatusUnauthorized, w4.Code)
	})
}

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
	"context"
	"testing"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"

	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
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

func TestCreateRole(t *testing.T) {
	t.Run("create new role, should succeed", func(t *testing.T) {
		r := newRole("TestCreateRole_createNewRole")
		err := rbacsvc.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
	})
	t.Run("create role twice, should return: "+rbac.NewError(rbac.ErrRoleConflict, "").Error(), func(t *testing.T) {
		r := newRole("TestCreateRole_createRoleTwice")
		err := rbacsvc.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		// twice
		err = rbacsvc.CreateRole(context.TODO(), r)
		assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrRoleConflict))
	})
}

func TestGetRole(t *testing.T) {
	t.Run("get no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		r, err := rbacsvc.GetRole(context.TODO(), "TestGetRole_getNoExistRole")
		assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrRoleNotExist))
		assert.Nil(t, r)
	})
	t.Run("get exist role, should success", func(t *testing.T) {
		r := newRole("TestGetRole_getExistRole")
		err := rbacsvc.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		resp, err := rbacsvc.GetRole(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.Equal(t, r.Name, resp.Name)
	})
}

func TestEditRole(t *testing.T) {
	t.Run("edit no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		r := newRole("TestEditRole_editNoExistRole")
		err := rbacsvc.EditRole(context.TODO(), r.Name, r)
		assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrRoleNotExist))
	})
	t.Run("edit role, should success", func(t *testing.T) {
		r := newRole("TestGetRole_editRole")
		err := rbacsvc.CreateRole(context.TODO(), r)
		assert.Nil(t, err)

		// edit
		assert.Equal(t, 1, len(r.Perms))
		r.Perms = []*rbac.Permission{
			{
				Resources: []*rbac.Resource{
					{
						Type: rbacsvc.ResourceService,
					},
				},
				Verbs: []string{"*"},
			},
			{
				Resources: []*rbac.Resource{
					{
						Type: rbacsvc.ResourceSchema,
					},
				},
				Verbs: []string{"*"},
			},
		}
		err = rbacsvc.EditRole(context.TODO(), r.Name, r)
		assert.Nil(t, err)

		resp, err := rbacsvc.GetRole(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(resp.Perms))
	})
	t.Run("edit build in role, should return: "+rbac.NewError(rbac.ErrForbidOperateBuildInRole, "").Error(), func(t *testing.T) {
		for _, name := range []string{rbac.RoleDeveloper, rbac.RoleDeveloper} {
			err := rbacsvc.EditRole(context.TODO(), name, newRole(""))
			assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrForbidOperateBuildInRole))
		}
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("delete no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		err := rbacsvc.DeleteRole(context.TODO(), "TestDeleteRole_deleteNoExistRole")
		assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrRoleNotExist))
	})
	t.Run("delete role, should success", func(t *testing.T) {
		r := newRole("TestDeleteRole_deleteRole")
		err := rbacsvc.CreateRole(context.TODO(), r)
		assert.Nil(t, err)

		err = rbacsvc.DeleteRole(context.TODO(), r.Name)
		assert.Nil(t, err)

		exist, err := rbacsvc.RoleExist(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.False(t, exist)
	})
	t.Run("delete build in role, should return: "+rbac.NewError(rbac.ErrForbidOperateBuildInRole, "").Error(), func(t *testing.T) {
		for _, name := range []string{rbac.RoleDeveloper, rbac.RoleDeveloper} {
			err := rbacsvc.DeleteRole(context.TODO(), name)
			assert.True(t, errsvc.IsErrEqualCode(err, rbac.ErrForbidOperateBuildInRole))
		}
	})
}

func TestListRole(t *testing.T) {
	t.Run("list role, should success", func(t *testing.T) {
		roles, total, err := rbacsvc.ListRole(context.TODO())
		assert.Nil(t, err)
		assert.True(t, total > 0)
		assert.Equal(t, int64(len(roles)), total)
	})
}

func TestRoleExistt(t *testing.T) {
	t.Run("check no exist role, should success and not exist", func(t *testing.T) {
		exist, err := rbacsvc.RoleExist(context.TODO(), "TestRoleExist_checkNoExistRole")
		assert.Nil(t, err)
		assert.False(t, exist)
	})
}

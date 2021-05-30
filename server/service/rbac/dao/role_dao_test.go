package dao_test

import (
	"context"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"
	"testing"
)

func exampleRole(name string) *rbac.Role {
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
		r := exampleRole("TestCreateRole_createNewRole")
		status, err := dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())
	})
	t.Run("create role twice, should return: "+rbac.NewError(rbac.ErrRoleConflict, "").Error(), func(t *testing.T) {
		r := exampleRole("TestCreateRole_createRoleTwice")
		status, err := dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())
		// twice
		status, err = dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.Equal(t, rbac.ErrRoleConflict, status.GetCode())
	})
}

func TestGetRole(t *testing.T) {
	t.Run("get no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		r, status, err := dao.GetRole(context.TODO(), "TestGetRole_getNoExistRole")
		assert.Nil(t, err)
		assert.Equal(t, rbac.ErrRoleNotExist, status.GetCode())
		assert.Nil(t, r)
	})
	t.Run("get exist role, should success", func(t *testing.T) {
		r := exampleRole("TestGetRole_getExistRole")
		status, err := dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())
		resp, status, err := dao.GetRole(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())
		assert.Equal(t, r.Name, resp.Name)
	})
}

func TestEditRole(t *testing.T) {
	t.Run("edit no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		r := exampleRole("TestEditRole_editNoExistRole")
		status, err := dao.EditRole(context.TODO(), r.Name, r)
		assert.Nil(t, err)
		assert.Equal(t, rbac.ErrRoleNotExist, status.GetCode())
	})
	t.Run("edit role, should success", func(t *testing.T) {
		r := exampleRole("TestGetRole_editRole")
		status, err := dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())

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
		status, err = dao.EditRole(context.TODO(), r.Name, r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())

		resp, status, err := dao.GetRole(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())
		assert.Equal(t, 2, len(resp.Perms))
	})
	t.Run("edit build in role, should return: "+discovery.NewError(discovery.ErrForbidden, "").Error(), func(t *testing.T) {
		for _, name := range []string{rbac.RoleDeveloper, rbac.RoleDeveloper} {
			status, err := dao.EditRole(context.TODO(), name, exampleRole(""))
			assert.Nil(t, err)
			assert.Equal(t, discovery.ErrForbidden, status.GetCode())
		}
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("delete no exist role, should return: "+rbac.NewError(rbac.ErrRoleNotExist, "").Error(), func(t *testing.T) {
		status, err := dao.DeleteRole(context.TODO(), "TestDeleteRole_deleteNoExistRole")
		assert.Nil(t, err)
		assert.Equal(t, rbac.ErrRoleNotExist, status.GetCode())
	})
	t.Run("delete role, should success", func(t *testing.T) {
		r := exampleRole("TestDeleteRole_deleteRole")
		status, err := dao.CreateRole(context.TODO(), r)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())

		status, err = dao.DeleteRole(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.True(t, status.IsSucceed())

		exist, err := dao.RoleExist(context.TODO(), r.Name)
		assert.Nil(t, err)
		assert.False(t, exist)
	})
	t.Run("delete build in role, should return: "+discovery.NewError(discovery.ErrForbidden, "").Error(), func(t *testing.T) {
		for _, name := range []string{rbac.RoleDeveloper, rbac.RoleDeveloper} {
			status, err := dao.DeleteRole(context.TODO(), name)
			assert.Nil(t, err)
			assert.Equal(t, discovery.ErrForbidden, status.GetCode())
		}
	})
}

func TestListRole(t *testing.T) {
	t.Run("list role, should success", func(t *testing.T) {
		roles, total, err := dao.ListRole(context.TODO())
		assert.Nil(t, err)
		assert.True(t, total > 0)
		assert.Equal(t, int64(len(roles)), total)
	})
}

func TestRoleExistt(t *testing.T) {
	t.Run("check no exist role, should success and not exist", func(t *testing.T) {
		exist, err := dao.RoleExist(context.TODO(), "TestRoleExist_checkNoExistRole")
		assert.Nil(t, err)
		assert.False(t, exist)
	})
}

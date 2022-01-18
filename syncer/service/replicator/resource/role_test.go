package resource

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

type forkRole struct {
	roles map[string]*rbac.Role
}

func (f forkRole) GetRole(_ context.Context, name string) (*rbac.Role, error) {
	result, ok := f.roles[name]
	if !ok {
		return nil, rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}
	return result, nil
}

func (f forkRole) EditRole(_ context.Context, name string, r *rbac.Role) error {
	_, ok := f.roles[name]
	if !ok {
		return rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}
	f.roles[name] = r
	return nil
}

func (f forkRole) CreateRole(_ context.Context, r *rbac.Role) error {
	_, ok := f.roles[r.Name]
	if ok {
		return rbacmodel.NewError(rbacmodel.ErrRoleConflict, "role exist")
	}

	f.roles[r.Name] = r
	return nil
}

func (f forkRole) DeleteRole(_ context.Context, name string) error {
	_, ok := f.roles[name]
	if !ok {
		return rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}

	delete(f.roles, name)
	return nil
}

func TestOperateRole(t *testing.T) {
	createTime := strconv.FormatInt(time.Now().Unix(), 10)
	input := &rbac.Role{
		ID:   "ffba57db4a094240aa573641044d16a6",
		Name: "admin",
		Perms: []*rbac.Permission{
			{
				Resources: []*rbac.Resource{
					{
						Type: "service",
					},
				},
			},
		},
		CreateTime: createTime,
		UpdateTime: createTime,
	}
	value, _ := json.Marshal(input)
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.CreateAction,
		Subject:   Role,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a := &role{
		event: e,
	}
	a.manager = &forkRole{
		roles: make(map[string]*rbac.Role),
	}
	ctx := context.Background()
	result := a.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				data, err := a.manager.GetRole(ctx, "admin")
				assert.Nil(t, err)
				assert.NotNil(t, data)
			}
		}
	}

	updateTime := strconv.FormatInt(time.Now().Unix(), 10)
	input = &rbac.Role{
		ID:   "ffba57db4a094240aa573641044d16a6",
		Name: "admin",
		Perms: []*rbac.Permission{
			{
				Resources: []*rbac.Resource{
					{
						Type: "governance",
					},
				},
			},
		},
		CreateTime: createTime,
		UpdateTime: updateTime,
	}
	value, _ = json.Marshal(input)

	id, _ = v1sync.NewEventID()
	e1 := &v1sync.Event{
		Id:        id,
		Action:    sync.UpdateAction,
		Subject:   Role,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a1 := &role{
		event:   e1,
		manager: a.manager,
	}
	result = a1.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a1.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a1.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				data, err := a1.manager.GetRole(ctx, "admin")
				assert.Nil(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, "governance", data.Perms[0].Resources[0].Type)
			}
		}
	}

	id, _ = v1sync.NewEventID()
	e2 := &v1sync.Event{
		Id:        id,
		Action:    sync.DeleteAction,
		Subject:   Role,
		Opts:      nil,
		Value:     []byte("admin"),
		Timestamp: v1sync.Timestamp(),
	}
	a2 := &role{
		event:   e2,
		manager: a.manager,
	}
	result = a2.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a2.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a2.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				_, err := a1.manager.GetRole(ctx, "admin")
				assert.NotNil(t, err)
				assert.True(t, errsvc.IsErrEqualCode(err, rbacmodel.ErrRoleNotExist))
			}
		}
	}
}

func TestNewRole(t *testing.T) {
	r := NewRole(nil)
	assert.NotNil(t, r)
}

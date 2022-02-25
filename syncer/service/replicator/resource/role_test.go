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

package resource

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	datasourcerbac "github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

type mockRole struct {
	roles map[string]*rbac.Role
}

func (f mockRole) GetRole(_ context.Context, name string) (*rbac.Role, error) {
	result, ok := f.roles[name]
	if !ok {
		return nil, datasourcerbac.ErrRoleNotExist
	}
	return result, nil
}

func (f mockRole) EditRole(_ context.Context, name string, r *rbac.Role) error {
	_, ok := f.roles[name]
	if !ok {
		return datasourcerbac.ErrRoleNotExist
	}
	f.roles[name] = r
	return nil
}

func (f mockRole) CreateRole(_ context.Context, r *rbac.Role) error {
	_, ok := f.roles[r.Name]
	if ok {
		return datasourcerbac.ErrRoleDuplicated
	}

	f.roles[r.Name] = r
	return nil
}

func (f mockRole) DeleteRole(_ context.Context, name string) error {
	_, ok := f.roles[name]
	if !ok {
		return datasourcerbac.ErrRoleNotExist
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
	a.manager = &mockRole{
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
				assert.True(t, errors.Is(err, datasourcerbac.ErrRoleNotExist))
			}
		}
	}
}

func TestNewRole(t *testing.T) {
	r := NewRole(nil)
	assert.NotNil(t, r)
}

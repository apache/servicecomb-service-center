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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	crbac "github.com/go-chassis/cari/rbac"
)

const (
	Role = "role"
)

func NewRole(e *v1sync.Event) Resource {
	r := &role{
		event: e,
	}
	r.manager = r
	return r
}

type role struct {
	event *v1sync.Event

	input       *crbac.Role
	deleteInput *string

	roleName string

	cur *crbac.Role

	manager roleManager

	defaultFailHandler
}

type roleManager interface {
	GetRole(ctx context.Context, name string) (*crbac.Role, error)
	EditRole(ctx context.Context, name string, r *crbac.Role) error
	CreateRole(ctx context.Context, r *crbac.Role) error
	DeleteRole(ctx context.Context, name string) error
}

func (r *role) loadInput() error {
	r.input = new(crbac.Role)
	callback := func() {
		r.roleName = r.input.Name
	}

	r.deleteInput = new(string)

	createOrUpdateParam := newInputParam(r.input, callback)
	deleteParam := newInputParam(r.deleteInput, func() {
		r.roleName = *r.deleteInput
	})

	return newInputLoader(
		r.event,
		createOrUpdateParam,
		createOrUpdateParam,
		deleteParam,
	).loadInput()
}

func (r *role) LoadCurrentResource(ctx context.Context) *Result {
	err := r.loadInput()
	if err != nil {
		return FailResult(err)
	}

	cur, err := r.manager.GetRole(ctx, r.roleName)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotExist) {
			return nil
		}
		return FailResult(err)
	}
	r.cur = cur
	return nil
}

func (r *role) NeedOperate(ctx context.Context) *Result {
	checker := &checker{
		curNotNil: r.cur != nil,
		event:     r.event,
		updateTime: func() (int64, error) {
			return formatUpdateTimeSecond(r.cur.UpdateTime)
		},
		resourceID: r.input.Name,
	}
	checker.tombstoneLoader = checker
	return checker.needOperate(ctx)
}

func (r *role) CreateHandle(ctx context.Context) error {
	if r.cur != nil {
		log.Warn(fmt.Sprintf("create action but resource exist, %s, %s",
			r.roleName, r.event.Id))
		return r.UpdateHandle(ctx)
	}
	return r.manager.CreateRole(ctx, r.input)
}

func (r *role) UpdateHandle(ctx context.Context) error {
	if r.cur == nil {
		log.Warn(fmt.Sprintf("update action but resource not exist, %s, %s",
			r.roleName, r.event.Id))
		return r.CreateHandle(ctx)
	}
	return r.manager.EditRole(ctx, r.roleName, r.input)
}

func (r *role) DeleteHandle(ctx context.Context) error {
	return r.manager.DeleteRole(ctx, r.roleName)
}

func (r *role) Operate(ctx context.Context) *Result {
	return newOperator(r).operate(ctx, r.event.Action)
}

func (r *role) GetRole(ctx context.Context, name string) (*crbac.Role, error) {
	return rbac.Instance().GetRole(ctx, name)
}

func (r *role) EditRole(ctx context.Context, name string, role *crbac.Role) error {
	return rbac.Instance().UpdateRole(ctx, name, role)
}

func (r *role) CreateRole(ctx context.Context, role *crbac.Role) error {
	return rbac.Instance().CreateRole(ctx, role)
}

func (r *role) DeleteRole(ctx context.Context, name string) error {
	_, err := rbac.Instance().DeleteRole(ctx, name)
	return err
}

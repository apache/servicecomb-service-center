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

package datasource

import (
	"context"
	"errors"
	"github.com/go-chassis/cari/rbac"
)

var (
	ErrRoleDuplicated = errors.New("role is duplicated")
	ErrRoleCanNotEdit = errors.New("role can not be edited")
)

// RoleManager contains the RBAC CRUD
type RoleManager interface {
	CreateRole(ctx context.Context, r *rbac.Role) error
	RoleExist(ctx context.Context, name string) (bool, error)
	GetRole(ctx context.Context, name string) (*rbac.Role, error)
	ListRole(ctx context.Context) ([]*rbac.Role, int64, error)
	DeleteRole(ctx context.Context, name string) (bool, error)
	UpdateRole(ctx context.Context, name string, role *rbac.Role) error
}

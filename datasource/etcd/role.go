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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/little-cui/etcdadpt"
)

func (rm *RbacDAO) CreateRole(ctx context.Context, r *rbacmodel.Role) error {
	lock, err := etcdsync.Lock("/role-creating/"+r.Name, -1, false)
	if err != nil {
		return fmt.Errorf("role %s is creating", r.Name)
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Error("can not release role lock", err)
		}
	}()
	key := path.GenerateRBACRoleKey(r.Name)
	exist, err := rm.RoleExist(ctx, r.Name)
	if err != nil {
		log.Error("can not save role info", err)
		return err
	}
	if exist {
		return rbac.ErrRoleDuplicated
	}
	r.ID = util.GenerateUUID()
	r.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	r.UpdateTime = r.CreateTime
	value, err := json.Marshal(r)
	if err != nil {
		log.Error("role info is invalid", err)
		return err
	}
	err = etcdadpt.PutBytes(ctx, key, value)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	log.Info("create new role: " + r.ID)
	return nil
}

func (rm *RbacDAO) RoleExist(ctx context.Context, name string) (bool, error) {
	return etcdadpt.Exist(ctx, path.GenerateRBACRoleKey(name))
}

func (rm *RbacDAO) GetRole(ctx context.Context, name string) (*rbacmodel.Role, error) {
	kv, err := etcdadpt.Get(ctx, path.GenerateRBACRoleKey(name))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, rbac.ErrRoleNotExist
	}
	role := &rbacmodel.Role{}
	err = json.Unmarshal(kv.Value, role)
	if err != nil {
		log.Error("role info format invalid", err)
		return nil, err
	}
	return role, nil
}
func (rm *RbacDAO) ListRole(ctx context.Context) ([]*rbacmodel.Role, int64, error) {
	kvs, n, err := etcdadpt.List(ctx, path.GenerateRBACRoleKey(""))
	if err != nil {
		return nil, 0, err
	}
	roles := make([]*rbacmodel.Role, 0, n)
	for _, v := range kvs {
		r := &rbacmodel.Role{}
		err = json.Unmarshal(v.Value, r)
		if err != nil {
			log.Error("role info format invalid:", err)
			continue //do not fail if some role is invalid
		}

		roles = append(roles, r)
	}
	return roles, n, nil
}
func (rm *RbacDAO) DeleteRole(ctx context.Context, name string) (bool, error) {
	exists, err := RoleBindingExists(ctx, name)
	if err != nil {
		log.Error("", err)
		return false, err
	}
	if exists {
		return false, rbac.ErrRoleBindingExist
	}
	del, err := etcdadpt.Delete(ctx, path.GenerateRBACRoleKey(name))
	if err != nil {
		return false, err
	}
	return del, nil
}
func RoleBindingExists(ctx context.Context, role string) (bool, error) {
	_, total, err := etcdadpt.List(ctx, path.GenRoleAccountPrefixIdxKey(role))
	if err != nil {
		log.Error("", err)
		return false, err
	}
	return total > 0, nil
}
func (rm *RbacDAO) UpdateRole(ctx context.Context, name string, role *rbacmodel.Role) error {
	role.UpdateTime = strconv.FormatInt(time.Now().Unix(), 10)
	value, err := json.Marshal(role)
	if err != nil {
		log.Error("role info is invalid", err)
		return err
	}
	return etcdadpt.PutBytes(ctx, path.GenerateRBACRoleKey(name), value)
}

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

	crbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
)

const isMigrated = "/cse-sr/role-migrated"

var (
	resources   = crbac.BuildResourceList(rbacsvc.ResourceConfig)
	configPerms = &crbac.Permission{
		Resources: resources,
		Verbs:     []string{"*"},
	}
)

func (rm *RbacDAO) CreateRole(ctx context.Context, r *crbac.Role) error {
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
	opts := []etcdadpt.OpOptions{
		etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(value)),
	}
	syncOpts, err := esync.GenCreateOpts(ctx, datasource.ResourceRole, r)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return err
	}
	opts = append(opts, syncOpts...)
	err = etcdadpt.Txn(ctx, opts)
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

func (rm *RbacDAO) GetRole(ctx context.Context, name string) (*crbac.Role, error) {
	kv, err := etcdadpt.Get(ctx, path.GenerateRBACRoleKey(name))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, rbac.ErrRoleNotExist
	}
	role := &crbac.Role{}
	err = json.Unmarshal(kv.Value, role)
	if err != nil {
		log.Error("role info format invalid", err)
		return nil, err
	}
	return role, nil
}
func (rm *RbacDAO) ListRole(ctx context.Context) ([]*crbac.Role, int64, error) {
	kvs, n, err := etcdadpt.List(ctx, path.GenerateRBACRoleKey(""))
	if err != nil {
		return nil, 0, err
	}
	roles := make([]*crbac.Role, 0, n)
	for _, v := range kvs {
		r := &crbac.Role{}
		err = json.Unmarshal(v.Value, r)
		if err != nil {
			log.Error("role info format invalid:", err)
			continue // do not fail if some role is invalid
		}

		roles = append(roles, r)
	}
	return roles, n, nil
}
func (rm *RbacDAO) DeleteRole(ctx context.Context, name string) (bool, error) {
	exists, err := RoleBindingExists(ctx, name)
	if err != nil {
		log.Error("check role binding existence failed", err)
		return false, err
	}
	if exists {
		return false, rbac.ErrRoleBindingExist
	}
	opts := []etcdadpt.OpOptions{etcdadpt.OpDel(etcdadpt.WithStrKey(path.GenerateRBACRoleKey(name)))}
	syncOpts, err := esync.GenDeleteOpts(ctx, datasource.ResourceRole, name, name)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return false, err
	}
	opts = append(opts, syncOpts...)
	err = etcdadpt.Txn(ctx, opts)
	if err != nil {
		return false, err
	}
	return true, nil
}
func RoleBindingExists(ctx context.Context, role string) (bool, error) {
	_, total, err := etcdadpt.List(ctx, path.GenRoleAccountPrefixIdxKey(role))
	if err != nil {
		log.Error("list role account failed", err)
		return false, err
	}
	return total > 0, nil
}
func (rm *RbacDAO) UpdateRole(ctx context.Context, name string, role *crbac.Role) error {
	role.UpdateTime = strconv.FormatInt(time.Now().Unix(), 10)
	value, err := json.Marshal(role)
	if err != nil {
		log.Error("role info is invalid", err)
		return err
	}
	opts := []etcdadpt.OpOptions{
		etcdadpt.OpPut(etcdadpt.WithStrKey(path.GenerateRBACRoleKey(name)), etcdadpt.WithValue(value)),
	}
	syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceRole, role)
	if err != nil {
		log.Error("fail to create sync opts", err)
		return err
	}
	opts = append(opts, syncOpts...)
	return etcdadpt.Txn(ctx, opts)
}
func (rm *RbacDAO) MigrateOldRoles(ctx context.Context) error {
	exist, err := etcdadpt.Exist(ctx, isMigrated)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	rs, _, err := rbac.Instance().ListRole(ctx)
	if err != nil {
		return err
	}
	for _, role := range rs {
		role.Perms = append(role.Perms, configPerms)
		err = rbac.Instance().UpdateRole(ctx, role.Name, role)
		if err != nil {
			log.Error(fmt.Sprintf("edit role [%s] info faied", role.Name), err)
			return err
		}
	}
	err = etcdadpt.Put(ctx, isMigrated, "true")
	if err != nil {
		log.Error("can not save migrated flag", err)
		return err
	}
	return nil
}

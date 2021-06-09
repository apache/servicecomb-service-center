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

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type RoleManager struct {
}

func (rm *RoleManager) CreateRole(ctx context.Context, r *rbac.Role) error {
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
		return datasource.ErrRoleDuplicated
	}
	r.ID = util.GenerateUUID()
	r.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	r.UpdateTime = r.CreateTime
	value, err := json.Marshal(r)
	if err != nil {
		log.Error("role info is invalid", err)
		return err
	}
	err = client.PutBytes(ctx, key, value)
	if err != nil {
		log.Error("can not save account info", err)
		return err
	}
	log.Info("create new role: " + r.ID)
	return nil
}

func (rm *RoleManager) RoleExist(ctx context.Context, name string) (bool, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACRoleKey(name)))
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}
func (rm *RoleManager) GetRole(ctx context.Context, name string) (*rbac.Role, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACRoleKey(name)))
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, datasource.ErrRoleNotExist
	}
	if resp.Count != 1 {
		return nil, client.ErrNotUnique
	}
	role := &rbac.Role{}
	err = json.Unmarshal(resp.Kvs[0].Value, role)
	if err != nil {
		log.Errorf(err, "role info format invalid")
		return nil, err
	}
	return role, nil
}
func (rm *RoleManager) ListRole(ctx context.Context) ([]*rbac.Role, int64, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateRBACRoleKey("")), client.WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	roles := make([]*rbac.Role, 0, resp.Count)
	for _, v := range resp.Kvs {
		r := &rbac.Role{}
		err = json.Unmarshal(v.Value, r)
		if err != nil {
			log.Error("role info format invalid:", err)
			continue //do not fail if some role is invalid
		}

		roles = append(roles, r)
	}
	return roles, resp.Count, nil
}
func (rm *RoleManager) DeleteRole(ctx context.Context, name string) (bool, error) {
	exists, err := RoleBindingExists(ctx, name)
	if err != nil {
		log.Error("", err)
		return false, err
	}
	if exists {
		return false, datasource.ErrRoleBindingExist
	}
	resp, err := client.Instance().Do(ctx, client.DEL,
		client.WithStrKey(path.GenerateRBACRoleKey(name)))
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}
func RoleBindingExists(ctx context.Context, role string) (bool, error) {
	_, total, err := client.List(ctx, path.GenRoleAccountPrefixIdxKey(role))
	if err != nil {
		log.Error("", err)
		return false, err
	}
	return total > 0, nil
}
func (rm *RoleManager) UpdateRole(ctx context.Context, name string, role *rbac.Role) error {
	role.UpdateTime = strconv.FormatInt(time.Now().Unix(), 10)
	value, err := json.Marshal(role)
	if err != nil {
		log.Errorf(err, "role info is invalid")
		return err
	}
	_, err = client.Instance().Do(ctx, client.PUT,
		client.WithStrKey(path.GenerateRBACRoleKey(name)),
		client.WithValue(value))
	return err
}

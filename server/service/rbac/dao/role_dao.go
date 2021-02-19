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

package dao

import (
	"context"
	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func CreateRole(ctx context.Context, r *rbac.Role) error {
	return datasource.Instance().CreateRole(ctx, r)
}

func GetRole(ctx context.Context, name string) (*rbac.Role, error) {
	return datasource.Instance().GetRole(ctx, name)
}

func ListRole(ctx context.Context) ([]*rbac.Role, int64, error) {
	return datasource.Instance().ListRole(ctx)
}

func RoleExist(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().RoleExist(ctx, name)
}

func DeleteRole(ctx context.Context, name string) (bool, error) {
	if name == "admin" || name == "developer" {
		log.Warnf("role %s can not be delete", name)
		return false, nil
	}
	return datasource.Instance().DeleteRole(ctx, name)
}

func EditRole(ctx context.Context, a *rbac.Role) error {
	exist, err := datasource.Instance().RoleExist(ctx, a.Name)
	if err != nil {
		log.Errorf(err, "can not edit account info")
		return err
	}
	if !exist {
		return datasource.ErrRoleCanNotEdit
	}

	err = datasource.Instance().UpdateRole(ctx, a.Name, a)
	if err != nil {
		log.Errorf(err, "can not edit role info")
		return err
	}
	log.Info("role is edit")
	return nil
}

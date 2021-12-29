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

package rbac

import (
	"context"
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	"github.com/go-chassis/cari/discovery"
	rbacmodel "github.com/go-chassis/cari/rbac"
)

func CreateRole(ctx context.Context, r *rbacmodel.Role) error {
	err := validator.ValidateCreateRole(r)
	if err != nil {
		log.Error(fmt.Sprintf("create role [%s] failed", r.Name), err)
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}
	quotaErr := quotasvc.ApplyRole(ctx, 1)
	if quotaErr != nil {
		return rbacmodel.NewError(rbacmodel.ErrRoleNoQuota, quotaErr.Error())
	}
	err = rbac.Instance().CreateRole(ctx, r)
	if err == nil {
		log.Info(fmt.Sprintf("create role [%s] success", r.Name))
		return nil
	}

	log.Error(fmt.Sprintf("create role [%s] failed", r.Name), err)
	if err == rbac.ErrRoleDuplicated {
		return rbacmodel.NewError(rbacmodel.ErrRoleConflict, err.Error())
	}

	return err
}

func GetRole(ctx context.Context, name string) (*rbacmodel.Role, error) {
	resp, err := rbac.Instance().GetRole(ctx, name)
	if err == nil {
		return resp, nil
	}
	if err == rbac.ErrRoleNotExist {
		return nil, rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}
	return nil, err
}

func ListRole(ctx context.Context) ([]*rbacmodel.Role, int64, error) {
	return rbac.Instance().ListRole(ctx)
}

func RoleExist(ctx context.Context, name string) (bool, error) {
	return rbac.Instance().RoleExist(ctx, name)
}

func DeleteRole(ctx context.Context, name string) error {
	if err := illegalRoleCheck(name); err != nil {
		return err
	}
	exist, err := RoleExist(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("check role [%s] exist failed", name), err)
		return err
	}
	if !exist {
		log.Error(fmt.Sprintf("role [%s] not exist", name), err)
		return rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}
	succeed, err := rbac.Instance().DeleteRole(ctx, name)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleBindingExist) {
			return rbacmodel.NewError(rbacmodel.ErrRoleIsBound, "")
		}
		return err
	}
	if !succeed {
		return errors.New("delete role failed, please retry")
	}
	return nil
}

func EditRole(ctx context.Context, name string, a *rbacmodel.Role) error {
	if err := illegalRoleCheck(name); err != nil {
		return err
	}
	exist, err := RoleExist(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("check role [%s] exist failed", name), err)
		return err
	}
	if !exist {
		log.Error(fmt.Sprintf("role [%s] not exist", name), err)
		return rbacmodel.NewError(rbacmodel.ErrRoleNotExist, "")
	}
	oldRole, err := GetRole(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("get role [%s] failed", name), err)
		return err
	}

	oldRole.Perms = a.Perms

	err = rbac.Instance().UpdateRole(ctx, name, oldRole)
	if err != nil {
		log.Error("can not edit role info", err)
		return err
	}
	log.Info(fmt.Sprintf("role [%s] is edit", oldRole.ID))
	return nil
}

func illegalRoleCheck(role string) error {
	if role == rbacmodel.RoleAdmin || role == rbacmodel.RoleDeveloper {
		return rbacmodel.NewError(rbacmodel.ErrForbidOperateBuildInRole, errorsEx.MsgCantOperateBuildInRole)
	}
	return nil
}

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

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
)

func CreateRole(ctx context.Context, r *rbac.Role) error {
	err := validator.ValidateCreateRole(r)
	if err != nil {
		log.Error(fmt.Sprintf("create role [%s] failed", r.Name), err)
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}
	quotaErr := quota.Apply(ctx, quota.NewApplyQuotaResource(quota.TypeRole,
		util.ParseDomainProject(ctx), "", 1))
	if quotaErr != nil {
		return rbac.NewError(rbac.ErrRoleNoQuota, quotaErr.Error())
	}
	err = datasource.GetRoleManager().CreateRole(ctx, r)
	if err == nil {
		log.Info(fmt.Sprintf("create role [%s] success", r.Name))
		return nil
	}

	log.Error(fmt.Sprintf("create role [%s] failed", r.Name), err)
	if err == datasource.ErrRoleDuplicated {
		return rbac.NewError(rbac.ErrRoleConflict, err.Error())
	}

	return err
}

func GetRole(ctx context.Context, name string) (*rbac.Role, error) {
	resp, err := datasource.GetRoleManager().GetRole(ctx, name)
	if err == nil {
		return resp, nil
	}
	if err == datasource.ErrRoleNotExist {
		return nil, rbac.NewError(rbac.ErrRoleNotExist, "")
	}
	return nil, err
}

func ListRole(ctx context.Context) ([]*rbac.Role, int64, error) {
	return datasource.GetRoleManager().ListRole(ctx)
}

func RoleExist(ctx context.Context, name string) (bool, error) {
	return datasource.GetRoleManager().RoleExist(ctx, name)
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
		return rbac.NewError(rbac.ErrRoleNotExist, "")
	}
	succeed, err := datasource.GetRoleManager().DeleteRole(ctx, name)
	if err != nil {
		if errors.Is(err, datasource.ErrRoleBindingExist) {
			return rbac.NewError(rbac.ErrRoleIsBound, "")
		}
		return err
	}
	if !succeed {
		return errors.New("delete role failed, please retry")
	}
	return nil
}

func EditRole(ctx context.Context, name string, a *rbac.Role) error {
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
		return rbac.NewError(rbac.ErrRoleNotExist, "")
	}
	oldRole, err := GetRole(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("get role [%s] failed", name), err)
		return err
	}

	oldRole.Perms = a.Perms

	err = datasource.GetRoleManager().UpdateRole(ctx, name, oldRole)
	if err != nil {
		log.Error("can not edit role info", err)
		return err
	}
	log.Info(fmt.Sprintf("role [%s] is edit", oldRole.ID))
	return nil
}

func illegalRoleCheck(role string) error {
	if role == rbac.RoleAdmin || role == rbac.RoleDeveloper {
		return rbac.NewError(rbac.ErrForbidOperateBuildInRole, errorsEx.MsgCantOperateBuildInRole)
	}
	return nil
}

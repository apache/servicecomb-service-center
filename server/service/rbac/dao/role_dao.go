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
	"errors"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/go-chassis/cari/rbac"
)

func CreateRole(ctx context.Context, r *rbac.Role) (*discovery.Response, error) {
	err := validator.ValidateCreateRole(r)
	if err != nil {
		log.Errorf(err, "create role [%s] failed", r.Name)
		return discovery.CreateResponse(discovery.ErrInvalidParams, err.Error()), nil
	}
	quotaErr := quota.Apply(ctx, quota.NewApplyQuotaResource(quota.TypeRole,
		util.ParseDomainProject(ctx), "", 1))
	if quotaErr != nil {
		return discovery.CreateResponse(discovery.ErrNotEnoughQuota, quotaErr.Error()), nil
	}
	err = datasource.Instance().CreateRole(ctx, r)
	if err == nil {
		log.Infof("create role [%s] success", r.Name)
		return nil, nil
	}

	log.Errorf(err, "create role [%s] failed", r.Name)
	if err == datasource.ErrRoleDuplicated {
		return discovery.CreateResponse(rbac.ErrRoleConflict, err.Error()), nil
	}

	return nil, err
}

func GetRole(ctx context.Context, name string) (*rbac.Role, *discovery.Response, error) {
	resp, err := datasource.Instance().GetRole(ctx, name)
	if err == nil {
		return resp, nil, nil
	}
	if err == datasource.ErrRoleNotExist {
		return nil, discovery.CreateResponse(rbac.ErrRoleNotExist, ""), nil
	}
	return nil, nil, err
}

func ListRole(ctx context.Context) ([]*rbac.Role, int64, error) {
	return datasource.Instance().ListRole(ctx)
}

func RoleExist(ctx context.Context, name string) (bool, error) {
	return datasource.Instance().RoleExist(ctx, name)
}

func DeleteRole(ctx context.Context, name string) (*discovery.Response, error) {
	if isBuildInRole(name) {
		return discovery.CreateResponse(discovery.ErrForbidden, errorsEx.MsgCantOperateBuildInRole), nil
	}
	exist, err := datasource.Instance().RoleExist(ctx, name)
	if err != nil {
		log.Errorf(err, "check role [%s] exist failed", name)
		return nil, err
	}
	if !exist {
		log.Errorf(err, "role [%s] not exist", name)
		return discovery.CreateResponse(rbac.ErrRoleNotExist, ""), nil
	}
	succeed, err := datasource.Instance().DeleteRole(ctx, name)
	if err != nil {
		return nil, err
	}
	if !succeed {
		return nil, errors.New("delete role failed, please retry")
	}
	return nil, nil
}

func EditRole(ctx context.Context, name string, a *rbac.Role) (*discovery.Response, error) {
	if isBuildInRole(name) {
		return discovery.CreateResponse(discovery.ErrForbidden, errorsEx.MsgCantOperateBuildInRole), nil
	}
	exist, err := datasource.Instance().RoleExist(ctx, name)
	if err != nil {
		log.Errorf(err, "check role [%s] exist failed", name)
		return nil, err
	}
	if !exist {
		log.Errorf(err, "role [%s] not exist", name)
		return discovery.CreateResponse(rbac.ErrRoleNotExist, ""), nil
	}
	oldRole, status, err := GetRole(ctx, name)
	if err != nil {
		log.Errorf(err, "get role [%s] failed", name)
		return nil, err
	}
	if status != nil && status.GetCode() != discovery.ResponseSuccess {
		return status, nil
	}

	oldRole.Perms = a.Perms

	err = datasource.Instance().UpdateRole(ctx, name, oldRole)
	if err != nil {
		log.Errorf(err, "can not edit role info")
		return nil, err
	}
	log.Infof("role [%s] is edit", oldRole.ID)
	return nil, nil
}

func isBuildInRole(role string) bool {
	if role == rbac.RoleAdmin || role == rbac.RoleDeveloper {
		return true
	}
	return false
}

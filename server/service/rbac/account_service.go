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

//Package rbac is dao layer API to help service center manage account, policy and role info
package rbac

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	"github.com/go-chassis/cari/discovery"
	rbacmodel "github.com/go-chassis/cari/rbac"
)

//CreateAccount save account info
func CreateAccount(ctx context.Context, a *rbacmodel.Account) error {
	quotaErr := quotasvc.ApplyAccount(ctx, 1)
	if quotaErr != nil {
		return rbacmodel.NewError(rbacmodel.ErrAccountNoQuota, quotaErr.Error())
	}
	err := validator.ValidateCreateAccount(a)
	if err != nil {
		log.Error(fmt.Sprintf("create account [%s] failed", a.Name), err)
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}
	if len(a.Status) == 0 {
		a.Status = "active"
	}
	err = a.Check()
	if err != nil {
		log.Error(fmt.Sprintf("create account [%s] failed", a.Name), err)
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}
	if err = checkRoleNames(ctx, a.Roles); err != nil {
		return rbacmodel.NewError(rbacmodel.ErrAccountHasInvalidRole, err.Error())
	}

	err = rbac.Instance().CreateAccount(ctx, a)
	if err == nil {
		log.Info(fmt.Sprintf("create account [%s] success", a.Name))
		return nil
	}
	log.Error(fmt.Sprintf("create account [%s] failed", a.Name), err)
	if err == rbac.ErrAccountDuplicated {
		return rbacmodel.NewError(rbacmodel.ErrAccountConflict, err.Error())
	}
	return err
}

// UpdateAccount updates an account's info, except the password
func UpdateAccount(ctx context.Context, name string, a *rbacmodel.Account) error {
	// todo params validation
	err := validator.ValidateUpdateAccount(a)
	if err != nil {
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}
	if err = illegalAccountCheck(ctx, name); err != nil {
		return err
	}
	if len(a.Status) == 0 && len(a.Roles) == 0 {
		return discovery.NewError(discovery.ErrInvalidParams, "status and roles cannot be empty both")
	}

	oldAccount, err := GetAccount(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("get account [%s] failed", name), err)
		return err
	}
	if len(a.Status) != 0 {
		oldAccount.Status = a.Status
	}
	if len(a.Roles) != 0 {
		oldAccount.Roles = a.Roles
	}
	if err = checkRoleNames(ctx, oldAccount.Roles); err != nil {
		return rbacmodel.NewError(rbacmodel.ErrAccountHasInvalidRole, err.Error())
	}
	err = rbac.Instance().UpdateAccount(ctx, name, oldAccount)
	if err != nil {
		log.Error("can not edit account info", err)
		return err
	}
	log.Info(fmt.Sprintf("account [%s] is edit", oldAccount.ID))
	return nil
}

func GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error) {
	r, err := rbac.Instance().GetAccount(ctx, name)
	if err != nil {
		if err == rbac.ErrAccountNotExist {
			msg := fmt.Sprintf("account [%s] not exist", name)
			return nil, rbacmodel.NewError(rbacmodel.ErrAccountNotExist, msg)
		}
		return nil, err
	}
	return r, nil
}
func ListAccount(ctx context.Context) ([]*rbacmodel.Account, int64, error) {
	return rbac.Instance().ListAccount(ctx)
}
func AccountExist(ctx context.Context, name string) (bool, error) {
	return rbac.Instance().AccountExist(ctx, name)
}
func DeleteAccount(ctx context.Context, name string) error {
	if err := illegalAccountCheck(ctx, name); err != nil {
		return err
	}
	exist, err := rbac.Instance().AccountExist(ctx, name)
	if err != nil {
		log.Error(fmt.Sprintf("check account [%s] exit failed", name), err)
		return err
	}
	if !exist {
		msg := fmt.Sprintf("account [%s] not exist", name)
		return rbacmodel.NewError(rbacmodel.ErrAccountNotExist, msg)
	}
	_, err = rbac.Instance().DeleteAccount(ctx, []string{name})
	return err
}

//EditAccount save account info
func EditAccount(ctx context.Context, a *rbacmodel.Account) error {
	exist, err := rbac.Instance().AccountExist(ctx, a.Name)
	if err != nil {
		log.Error("can not edit account info", err)
		return err
	}
	if !exist {
		return rbacmodel.NewError(rbacmodel.ErrAccountNotExist, "")
	}

	err = rbac.Instance().UpdateAccount(ctx, a.Name, a)
	if err != nil {
		log.Error("can not edit account info", err)
		return err
	}
	log.Info(fmt.Sprintf("account [%s] is edit", a.ID))
	return nil
}

func checkRoleNames(ctx context.Context, roles []string) error {
	for _, name := range roles {
		exist, err := RoleExist(ctx, name)
		if err != nil {
			log.Error(fmt.Sprintf("check role [%s] exist failed", name), err)
			return err
		}
		if !exist {
			return rbac.ErrRoleNotExist
		}
	}
	return nil
}

func illegalAccountCheck(ctx context.Context, target string) error {
	if target == RootName {
		return rbacmodel.NewError(rbacmodel.ErrForbidOperateBuildInAccount, errorsEx.MsgCantOperateRoot)
	}
	changer := UserFromContext(ctx)
	if target == changer {
		return rbacmodel.NewError(rbacmodel.ErrForbidOperateSelfAccount, errorsEx.MsgCantOperateYour)
	}
	return nil
}

func AccountUsage(ctx context.Context) (int64, error) {
	_, used, err := rbac.Instance().ListAccount(ctx)
	if err != nil {
		return 0, err
	}
	return used, nil
}

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

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/foundation/stringutil"
	"golang.org/x/crypto/bcrypt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/validator"
)

func ChangePassword(ctx context.Context, a *rbac.Account) error {
	err := validator.ValidateChangePWD(a)
	if err != nil {
		return discovery.NewError(discovery.ErrInvalidParams, err.Error())
	}

	changer, err := AccountFromContext(ctx)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}

	// change self password, need to check password mismatch
	if changer.Name == a.Name {
		return changePassword(ctx, a.Name, a.CurrentPassword, a.Password)
	}

	// change other user's password, only admin role can do this and no need
	// supply current password
	for _, r := range changer.Roles {
		if r == rbac.RoleAdmin {
			return changePasswordForcibly(ctx, a.Name, a.Password)
		}
	}

	// other cases, change password is forbidden
	return discovery.NewError(discovery.ErrForbidden, ErrNoPermChangeAccount.Error())
}

func changePasswordForcibly(ctx context.Context, name, pwd string) error {
	old, err := GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	return doChangePassword(ctx, old, pwd)
}
func changePassword(ctx context.Context, name, currentPassword, pwd string) error {
	if currentPassword == "" {
		log.Error("current pwd is empty", nil)
		return discovery.NewError(discovery.ErrInvalidParams, ErrEmptyCurrentPassword.Error())
	}
	ip := util.GetIPFromContext(ctx)
	if IsBanned(MakeBanKey(name, ip)) {
		log.Warnf("ip [%s] is banned, account: %s", ip, name)
		return rbac.NewError(rbac.ErrAccountBlocked, "")
	}
	if currentPassword == pwd {
		return rbac.NewError(rbac.ErrNewPwdBad, ErrSamePassword.Error())
	}
	old, err := GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	same := privacy.SamePassword(old.Password, currentPassword)
	if !same {
		log.Error("current password is wrong", nil)
		TryLockAccount(MakeBanKey(name, ip))
		return rbac.NewError(rbac.ErrOldPwdWrong, "")
	}
	return doChangePassword(ctx, old, pwd)
}

func doChangePassword(ctx context.Context, old *rbac.Account, pwd string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
	if err != nil {
		log.Error("pwd hash failed", err)
		return err
	}
	old.Password = stringutil.Bytes2str(hash)
	err = EditAccount(ctx, old)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	return nil
}

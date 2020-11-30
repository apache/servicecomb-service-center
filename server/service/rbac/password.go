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

	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	stringutil "github.com/go-chassis/foundation/string"
	"golang.org/x/crypto/bcrypt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
)

func ChangePassword(ctx context.Context, changerRole []string, changerName string, a *rbacframe.Account) error {
	if changerName == a.Name {
		if a.CurrentPassword == "" {
			log.Error("current pwd is empty", nil)
			return ErrEmptyCurrentPassword
		}
		return changePassword(ctx, changerName, a.CurrentPassword, a.Password)
	}
	for i := 0; i < len(changerRole); i++ {
		//need to check password mismatch. but admin role can change any user password without supply current password
		if changerRole[i] == rbacframe.RoleAdmin {
			return changePasswordForcibly(ctx, a.Name, a.Password)
		}
	}

	log.Error("can not change other account pwd", nil)
	return ErrNoPermChangeAccount
}
func changePasswordForcibly(ctx context.Context, name, pwd string) error {
	old, err := dao.GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	err = doChangePassword(ctx, old, pwd)
	if err != nil {
		return err
	}
	return nil
}
func changePassword(ctx context.Context, name, currentPassword, pwd string) error {
	if currentPassword == pwd {
		return ErrSamePassword
	}
	old, err := dao.GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	same := SamePassword(old.Password, currentPassword)
	if !same {
		log.Error("current password is wrong", nil)
		return ErrWrongPassword
	}
	err = doChangePassword(ctx, old, pwd)
	if err != nil {
		return err
	}
	return nil
}

func doChangePassword(ctx context.Context, old *rbacframe.Account, pwd string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
	if err != nil {
		log.Error("pwd hash failed", err)
		return err
	}
	old.Password = stringutil.Bytes2str(hash)
	err = dao.EditAccount(ctx, old)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	return nil
}

func SamePassword(hashedPwd, pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		log.Warn("incorrect password attempts")
	}
	return err == nil
}

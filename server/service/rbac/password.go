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

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
)

func ChangePassword(ctx context.Context, changerRole, changerName string, a *model.Account) error {
	if a.Name != "" {
		if changerRole != model.RoleAdmin { //need to check password mismatch. but admin role can change any user password without supply current password
			log.Error("can not change other account pwd", nil)
			return ErrInputChangeAccount
		}
		return changePasswordForcibly(ctx, a.Name, a.Password)
	}
	if a.CurrentPassword == "" {
		log.Error("current pwd is empty", nil)
		return ErrInputCurrentPassword
	}
	return changePassword(ctx, changerName, a.CurrentPassword, a.Password)

}
func changePasswordForcibly(ctx context.Context, name, pwd string) error {
	old, err := dao.GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	old.Password = pwd
	err = dao.EditAccount(ctx, old)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	return nil
}
func changePassword(ctx context.Context, name, currentPassword, pwd string) error {
	old, err := dao.GetAccount(ctx, name)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	if old.Password != currentPassword {
		log.Error("current pwd is wrong", nil)
		return errors.New("can not change pwd")
	}
	old.Password = pwd
	err = dao.EditAccount(ctx, old)
	if err != nil {
		log.Error("can not change pwd", err)
		return err
	}
	return nil
}

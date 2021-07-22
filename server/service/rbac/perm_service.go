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

	"github.com/apache/servicecomb-service-center/datasource"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/go-chassis/cari/rbac"
)

// ListSelfPerms list the user permission from ctx
func ListSelfPerms(ctx context.Context) ([]*rbac.Permission, error) {
	user := UserFromContext(ctx)
	if len(user) == 0 {
		return nil, rbac.NewError(rbac.ErrUnauthorized, errorsEx.MsgListSelfPermsFailed)
	}
	account, err := datasource.GetAccountManager().GetAccount(ctx, user)
	if err != nil {
		return nil, err
	}
	var perms []*rbac.Permission
	for _, roleName := range account.Roles {
		role, err := datasource.GetRoleManager().GetRole(ctx, roleName)
		if err != nil {
			return nil, err
		}
		perms = append(perms, role.Perms...)
	}
	return perms, nil
}

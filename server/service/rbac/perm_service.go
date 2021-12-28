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

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	rbacmodel "github.com/go-chassis/cari/rbac"
)

// ListSelfPerms list the user permission from ctx
func ListSelfPerms(ctx context.Context) ([]*rbacmodel.Permission, error) {
	user := UserFromContext(ctx)
	if len(user) == 0 {
		return nil, rbacmodel.NewError(rbacmodel.ErrUnauthorized, errorsEx.MsgListSelfPermsFailed)
	}
	account, err := rbac.Instance().GetAccount(ctx, user)
	if err != nil {
		return nil, err
	}
	var perms []*rbacmodel.Permission
	for _, roleName := range account.Roles {
		role, err := rbac.Instance().GetRole(ctx, roleName)
		if err != nil {
			return nil, err
		}
		perms = append(perms, role.Perms...)
	}
	return perms, nil
}

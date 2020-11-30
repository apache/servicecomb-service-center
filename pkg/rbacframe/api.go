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

// Package rbacframe help other component which want to use servicecomb rbac system
package rbacframe

import (
	"context"
	"crypto/rsa"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-chassis/v2/security/token"
)

const (
	ClaimsUser  = "account"
	ClaimsRoles = "roles"
)

func AccountFromContext(ctx context.Context) (*Account, error) {
	claims := FromContext(ctx)
	m, ok := claims.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidCtx
	}
	accountNameI := m[ClaimsUser]
	a, ok := accountNameI.(string)
	if !ok {
		return nil, ErrConvertErr
	}
	roles := m[ClaimsRoles]
	roleList, err := GetRolesList(roles)
	if err != nil {
		log.Error("role convert failed ", err)
		return nil, ErrConvertErr
	}
	account := &Account{Name: a, Roles: roleList}
	return account, nil
}

//RoleFromContext only return role name
func RoleFromContext(ctx context.Context) (string, error) {
	claims := FromContext(ctx)
	m, ok := claims.(map[string]interface{})
	if !ok {
		return "", ErrInvalidCtx
	}
	roleI := m[ClaimsRoles]
	role, ok := roleI.(string)
	if !ok {
		return "", ErrConvertErr
	}
	return role, nil
}

//Authenticate parse a token to claims
func Authenticate(tokenStr string, pub *rsa.PublicKey) (interface{}, error) {
	claims, err := token.Verify(tokenStr, func(claims interface{}, method token.SigningMethod) (interface{}, error) {
		return pub, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// GetRolesList return role list string
func GetRolesList(v interface{}) ([]string, error) {
	s, ok := v.([]interface{})
	if !ok {
		return nil, ErrConvertErr
	}
	rolesList := make([]string, 0)
	for _, v := range s {
		role, ok := v.(string)
		if !ok {
			return nil, ErrConvertErr
		}
		rolesList = append(rolesList, role)
	}
	return rolesList, nil
}

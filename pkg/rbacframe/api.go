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
	"crypto/rsa"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/go-chassis/v2/security/token"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	ClaimsUser  = "account"
	ClaimsRoles = "roles"

	RoleAdmin = "admin"
)

var whiteAPIList = sets.NewString()

func Add2WhiteAPIList(path ...string) {
	whiteAPIList.Insert(path...)
}

func MustAuth(pattern string) bool {
	if util.IsVersionOrHealthPattern(pattern) {
		return false
	}
	return !whiteAPIList.Has(pattern)
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

//BuildResourceList join the resource to an array
func BuildResourceList(resourceType ...string) []string {
	rt := make([]string, len(resourceType))
	for i := 0; i < len(resourceType); i++ {
		rt[i] = resourceType[i]
	}
	return rt
}

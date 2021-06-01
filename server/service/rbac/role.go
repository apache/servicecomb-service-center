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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
)

var roleMap = map[string]*rbac.Role{}

func init() {
	// Assign resources to admin role, admin role own all permissions
	roleMap[rbac.RoleAdmin] = &rbac.Role{
		Name:  rbac.RoleAdmin,
		Perms: AdminPerms(),
	}
	roleMap[rbac.RoleDeveloper] = &rbac.Role{
		Name:  rbac.RoleDeveloper,
		Perms: DevPerms(),
	}
}

func initBuildInRole() {
	for _, r := range roleMap {
		createBuildInRole(r)
	}
}

func createBuildInRole(r *rbac.Role) {
	err := CreateRole(context.Background(), r)
	if err == nil {
		log.Infof("create role [%s] success", r.Name)
		return
	}
	if errsvc.IsErrEqualCode(err, rbac.ErrRoleConflict) {
		log.Infof("role [%s] already exists", r.Name)
		return
	}
	log.Fatalf(err, "create role [%s] failed", r.Name)
	return
}

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

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
)

var roleMap = map[string]*rbac.Role{}

// Assign resources to admin role, admin role own all permissions
func initAdminRole() {
	roleMap["admin"] = &rbac.Role{
		Name:  "admin",
		Perms: AdminPerms(),
	}
	err := dao.CreateRole(context.Background(), roleMap["admin"])
	if err != nil {
		log.Errorf(err, "create admin role failed")
	}
}

// Assign resources to developer role
func initDevRole() {
	roleMap["developer"] = &rbac.Role{
		Name:  "developer",
		Perms: DevPerms(),
	}
	err := dao.CreateRole(context.Background(), roleMap["developer"])
	if err != nil {
		log.Errorf(err, "create developer role failed")
	}
}

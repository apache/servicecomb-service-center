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

package dao

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
)

//TODO delete dead code and use etcd
var roleMap = map[string]*rbacframe.Role{
	"admin": {Permissions: map[string]*rbacframe.Permission{
		"account": {Verbs: []string{"*"}},
		"service": {Verbs: []string{"*"}},
	}},
	"developer": {Permissions: map[string]*rbacframe.Permission{
		"service": {Verbs: []string{"*"}},
	}},
}

func GetRole(ctx context.Context, role string) *rbacframe.Role {
	return roleMap[role]
}

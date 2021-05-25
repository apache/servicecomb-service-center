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
	"github.com/go-chassis/cari/rbac"
)

// method to verbs
var (
	MethodToVerbs = map[string]string{
		"GET":    "get",
		"POST":   "create",
		"UPDATE": "update",
		"DELETE": "delete",
	}
)

// AdminPerms allocate all resource permissions
func AdminPerms() []*rbac.Permission {
	resources := rbac.BuildResourceList(
		ResourceAccount, ResourceRole,
		ResourceService, ResourceGovern, ResourceOps, ResourceSchema)
	perm := []*rbac.Permission{
		{
			Resources: resources,
			Verbs:     []string{"*"},
		},
	}
	return perm
}

// DevPerms allocate all resource permissions except account and role resources
func DevPerms() []*rbac.Permission {
	resources := rbac.BuildResourceList(
		ResourceService, ResourceGovern, ResourceOps, ResourceSchema)
	perm := []*rbac.Permission{
		{
			Resources: resources,
			Verbs:     []string{"*"},
		},
	}
	return perm
}

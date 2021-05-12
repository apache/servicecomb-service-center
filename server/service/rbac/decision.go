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

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func Allow(ctx context.Context, roleList []string, project, resource, verbs string) (bool, error) {
	//TODO check project
	if ableToAccessResource(roleList, "admin") {
		return true, nil
	}
	// allPerms combines the roleList permission
	var allPerms = make([]*rbac.Permission, 0)
	for i := 0; i < len(roleList); i++ {
		r, err := datasource.Instance().GetRole(ctx, roleList[i])
		if err != nil {
			log.Error("get role list errors", err)
			return false, err
		}
		if r == nil {
			log.Warnf("role [%s] has no any permissions", roleList[i])
			continue
		}
		allPerms = append(allPerms, r.Perms...)
	}

	if len(allPerms) == 0 {
		log.Warn("role list has no any permissions")
		return false, nil
	}
	for i := 0; i < len(allPerms); i++ {
		if ableToAccessResource(allPerms[i].Resources, resource) && ableToOperateResource(allPerms[i].Verbs, verbs) {
			return true, nil
		}
	}

	log.Warn("role is not allowed to operate resource")
	return false, nil
}

func ableToOperateResource(haystack []string, needle string) bool {
	if ableToAccessResource(haystack, "*") || ableToAccessResource(haystack, needle) {
		return true
	}
	return false
}

func ableToAccessResource(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

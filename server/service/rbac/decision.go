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
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
)

func Allow(ctx context.Context, role, project, resource, verbs string) (bool, error) {
	r := dao.GetRole(ctx, role)
	if r == nil {
		log.Warn("empty role info")
		return false, nil
	}
	ps := r.Permissions
	if len(ps) == 0 {
		log.Warn("role has no any permissions")
		return false, nil
	}
	p, ok := ps[resource]
	if !ok || p == nil {
		log.Warn("role is not allowed to access resource")
		return false, nil
	}
	//TODO check verbs and project
	return true, nil
}

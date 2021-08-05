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
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/go-chassis/cari/rbac"
)

// Allow return: matched labels(empty if no label defined), error
func Allow(ctx context.Context, project string, roleList []string,
	targetResource *auth.ResourceScope) ([]map[string]string, error) {
	//TODO check project
	allPerms, err := getPermsByRoles(ctx, roleList)
	if err != nil {
		log.Error("get role list errors", err)
		return nil, err
	}
	if len(allPerms) == 0 {
		log.Warn("role list has no any permissions")
		return nil, rbac.NewError(rbac.ErrNoPermission, "role has no any permissions")
	}
	allow, labelList := GetLabel(allPerms, targetResource.Type, targetResource.Verb)
	if !allow {
		return nil, rbac.NewError(rbac.ErrNoPermission,
			fmt.Sprintf("role has no permissions[%s:%s]", targetResource.Type, targetResource.Verb))
	}
	// allow, but no label found, means we can ignore the labels
	if len(labelList) == 0 {
		return nil, nil
	}
	// target resource needs no label, return without filter
	if len(targetResource.Labels) == 0 {
		return labelList, nil
	}
	// allow, and labels found, filter the labels
	filteredLabelList := FilterLabel(targetResource.Labels, labelList)
	// target resource label matches no label in permission, means not allow
	if len(filteredLabelList) == 0 {
		return nil, rbac.NewError(rbac.ErrNoPermission,
			fmt.Sprintf("role has no permissions[%s:%s] for labels %v",
				targetResource.Type, targetResource.Verb, targetResource.Labels))
	}
	return filteredLabelList, nil
}

func FilterLabel(targetResourceLabel []map[string]string, permLabelList []map[string]string) []map[string]string {
	l := make([]map[string]string, 0)
	for _, resourceLabel := range targetResourceLabel {
		for _, label := range permLabelList {
			if LabelMatched(resourceLabel, label) {
				l = append(l, label)
			}
		}
	}
	return l
}

func LabelMatched(targetResourceLabel map[string]string, permLabel map[string]string) bool {
	for k, v := range permLabel {
		if vv := targetResourceLabel[k]; vv != v {
			return false
		}
	}
	return true
}

func getPermsByRoles(ctx context.Context, roleList []string) ([]*rbac.Permission, error) {
	var allPerms = make([]*rbac.Permission, 0)
	for _, name := range roleList {
		r, err := datasource.GetRoleManager().GetRole(ctx, name)
		if err == nil {
			allPerms = append(allPerms, r.Perms...)
			continue
		}
		if err == datasource.ErrRoleNotExist {
			log.Warn(fmt.Sprintf("role [%s] not exist", name))
			continue
		}
		log.Error(fmt.Sprintf("get role [%s] failed", name), err)
		return nil, err
	}
	return allPerms, nil
}

// GetLabel checks if the perms have permission to operate the resource(ignore label),
// if one perm have the permission, add it's label to the result.
func GetLabel(perms []*rbac.Permission, targetResource, verb string) (allow bool, labelList []map[string]string) {
	for _, perm := range perms {
		a, l := GetLabelFromSinglePerm(perm, targetResource, verb)
		if !a {
			continue
		}
		allow = true
		// allow and has no label, return fast
		if len(l) == 0 {
			return true, nil
		}
		labelList = append(labelList, l...)
	}
	return
}

// GetLabel checks if the perm have permission to operate the resource(ignore label),
// if the perm have the permission, return it's label.
func GetLabelFromSinglePerm(perm *rbac.Permission, targetResource, verb string) (allow bool, labelList []map[string]string) {
	if !allowVerb(perm.Verbs, verb) {
		return false, nil
	}

	return getResourceLabel(perm.Resources, targetResource)
}

func allowVerb(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == "*" || e == needle {
			return true
		}
	}
	return false
}

func getResourceLabel(resources []*rbac.Resource, needle string) (allow bool, labelList []map[string]string) {
	for _, resource := range resources {
		// filter the same resource
		if resource.Type != needle {
			continue
		}
		// has no label, return fast
		if len(resource.Labels) == 0 {
			return true, nil
		}
		labelList = append(labelList, resource.Labels)
		allow = true
	}
	return
}

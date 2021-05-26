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

package rbac_test

import (
	"testing"

	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"

	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
)

func TestGetLabel(t *testing.T) {
	perms := []*rbac.Permission{
		&rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type:   rbacsvc.ResourceAccount,
					Labels: map[string]string{"environment": "production"},
				},
				{
					Type:   rbacsvc.ResourceService,
					Labels: map[string]string{"serviceName": "service-center"},
				},
			},
			Verbs: []string{"get"},
		},
		&rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type:   rbacsvc.ResourceService,
					Labels: map[string]string{"appId": "default"},
				},
			},
			Verbs: []string{"*"},
		},
		&rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type: rbacsvc.ResourceService,
				},
			},
			Verbs: []string{"delete"},
		},
	}
	t.Run("resource and verb matched, should allow", func(t *testing.T) {
		allow, labelList := rbacsvc.GetLabel(perms, rbacsvc.ResourceService, "create")
		assert.True(t, allow)
		assert.Equal(t, 1, len(labelList))
	})

	t.Run("nums of resource matched, should allow and combine their labels", func(t *testing.T) {
		allow, labelList := rbacsvc.GetLabel(perms, rbacsvc.ResourceService, "get")
		assert.True(t, allow)
		assert.Equal(t, 2, len(labelList))
	})

	t.Run("nums of resource matched, one of them has no label, should allow and no label", func(t *testing.T) {
		allow, labelList := rbacsvc.GetLabel(perms, rbacsvc.ResourceService, "delete")
		assert.True(t, allow)
		assert.Equal(t, 0, len(labelList))
	})
	t.Run("resource not matched, should not allow", func(t *testing.T) {
		allow, labelList := rbacsvc.GetLabel(perms, rbacsvc.ResourceRole, "delete")
		assert.False(t, allow)
		assert.Equal(t, 0, len(labelList))
	})
	t.Run("Verb not matched, should not allow", func(t *testing.T) {
		allow, labelList := rbacsvc.GetLabel(perms, rbacsvc.ResourceAccount, "delete")
		assert.False(t, allow)
		assert.Equal(t, 0, len(labelList))
	})
}

func TestGetLabelFromSinglePerm(t *testing.T) {
	t.Run("resource and verb match, should allow", func(t *testing.T) {
		perms := &rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type:   rbacsvc.ResourceAccount,
					Labels: map[string]string{"environment": "production"},
				},
			},
			Verbs: []string{"*"},
		}
		allow, labelList := rbacsvc.GetLabelFromSinglePerm(perms, rbacsvc.ResourceAccount, "create")
		assert.True(t, allow)
		assert.Equal(t, 1, len(labelList))
		assert.Equal(t, "production", labelList[0]["environment"])
	})

	t.Run("resource not match, should no allow", func(t *testing.T) {
		perms := &rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type:   rbacsvc.ResourceAccount,
					Labels: map[string]string{"environment": "production"},
				},
			},
			Verbs: []string{"*"},
		}
		allow, labelList := rbacsvc.GetLabelFromSinglePerm(perms, rbacsvc.ResourceService, "create")
		assert.False(t, allow)
		assert.Equal(t, 0, len(labelList))
	})

	t.Run("verb not match, should no allow", func(t *testing.T) {
		perms := &rbac.Permission{
			Resources: []*rbac.Resource{
				{
					Type:   rbacsvc.ResourceAccount,
					Labels: map[string]string{"environment": "production"},
				},
			},
			Verbs: []string{"get"},
		}
		allow, labelList := rbacsvc.GetLabelFromSinglePerm(perms, rbacsvc.ResourceAccount, "create")
		assert.False(t, allow)
		assert.Equal(t, 0, len(labelList))
	})
}

func TestLabelMatched(t *testing.T) {
	targetResourceLabel := map[string]string{
		"environment": "production",
		"appId":       "default",
	}
	t.Run("value not match, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"environment": "testing",
		}
		assert.False(t, rbacsvc.LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("key not match, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"serviceName": "default",
		}
		assert.False(t, rbacsvc.LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches no permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"version": "1.0.0",
		}
		assert.False(t, rbacsvc.LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches part permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"version":     "1.0.0",
			"environment": "production",
		}
		assert.False(t, rbacsvc.LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches  permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"environment": "production",
		}
		assert.True(t, rbacsvc.LabelMatched(targetResourceLabel, permResourceLabel))
	})
}
func TestFilterLabel(t *testing.T) {
	targetResourceLabel := []map[string]string{
		{"environment": "production", "appId": "default"},
		{"serviceName": "service-center"},
	}
	permResourceLabel := []map[string]string{
		{"environment": "production", "appId": "default"},
		{"appId": "default"},
		{"environment": "production", "serviceName": "service-center"},
		{"serviceName": "service-center", "version": "1.0.0"},
		{"serviceName": "service-center"},
		{"environment": "testing"},
	}
	l := rbacsvc.FilterLabel(targetResourceLabel, permResourceLabel)
	assert.Equal(t, 3, len(l))
}

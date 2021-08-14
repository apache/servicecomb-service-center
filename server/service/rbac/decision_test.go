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
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	t.Run("admin can operate any resource", func(t *testing.T) {
		ok, _ := rbac.Allow(context.TODO(), "admin", "default", "account", "create")
		assert.True(t, ok)
		ok, _ = rbac.Allow(context.TODO(), "admin", "default", "service", "create")
		assert.True(t, ok)
	})
	t.Run("developer can not operate account", func(t *testing.T) {
		ok, _ := rbac.Allow(context.TODO(), "developer", "default", "account", "create")
		assert.False(t, ok)
	})
	t.Run("developer can not operate service", func(t *testing.T) {
		ok, _ := rbac.Allow(context.TODO(), "developer", "default", "service", "create")
		assert.True(t, ok)
	})
}

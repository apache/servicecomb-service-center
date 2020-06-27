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

package dao_test

import (
	"context"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOverrideSecret(t *testing.T) {
	err := dao.OverrideSecret(context.Background(), "sk")
	assert.NoError(t, err)
	s, err := dao.GetSecret(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sk", string(s))
	b, err := dao.SecretExist(context.Background())
	assert.NoError(t, err)
	assert.True(t, b)
}

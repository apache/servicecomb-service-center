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
	v4 "github.com/apache/servicecomb-service-center/server/resource/v4"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/stretchr/testify/assert"
)

func TestCountFailure(t *testing.T) {
	rbac.BanTime = 3 * time.Second

	key1 := v4.MakeBanKey("root", "127.0.0.1")
	key2 := v4.MakeBanKey("root", "10.0.0.1")
	t.Run("ban root@IP, will not affect other root@another_IP", func(t *testing.T) {
		rbac.CountFailure(key1)
		assert.False(t, rbac.IsBanned(key1))

		rbac.CountFailure(key1)
		assert.False(t, rbac.IsBanned(key1))

		rbac.CountFailure(key1)
		assert.True(t, rbac.IsBanned(key1))

		rbac.CountFailure(key2)
		assert.False(t, rbac.IsBanned(key2))

		rbac.CountFailure(key2)
		assert.False(t, rbac.IsBanned(key2))

		rbac.CountFailure(key2)
		assert.True(t, rbac.IsBanned(key2))
	})
	t.Log(rbac.BannedList()[0].ReleaseAt)
	assert.Equal(t, 2, len(rbac.BannedList()))
	time.Sleep(4 * time.Second)
	assert.Equal(t, 0, len(rbac.BannedList()))
	assert.False(t, rbac.IsBanned(key1))
	assert.False(t, rbac.IsBanned(key2))

}

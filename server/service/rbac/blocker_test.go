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
	"time"

	dao "github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/server/resource/v4/rbac"
	accountsvc "github.com/apache/servicecomb-service-center/server/service/account"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/stretchr/testify/assert"
)

func TestCountFailure(t *testing.T) {
	key1 := rbac.MakeBanKey("root", "127.0.0.1")
	key2 := rbac.MakeBanKey("root", "10.0.0.1")
	t.Run("ban root@IP, will not affect other root@another_IP", func(t *testing.T) {
		rbacsvc.TryLockAccount(key1)
		assert.False(t, rbacsvc.IsBanned(key1))

		rbacsvc.TryLockAccount(key1)
		assert.False(t, rbacsvc.IsBanned(key1))

		rbacsvc.TryLockAccount(key1)
		assert.True(t, rbacsvc.IsBanned(key1))

		rbacsvc.TryLockAccount(key1)
		assert.True(t, rbacsvc.IsBanned(key1))

		rbacsvc.TryLockAccount(key2)
		assert.False(t, rbacsvc.IsBanned(key2))

		rbacsvc.TryLockAccount(key2)
		assert.False(t, rbacsvc.IsBanned(key2))

		rbacsvc.TryLockAccount(key2)
		assert.True(t, rbacsvc.IsBanned(key2))
	})

	t.Run("after ban released, should return false", func(t *testing.T) {
		time.Sleep(4 * time.Second)
		assert.False(t, rbacsvc.IsBanned(key1))
		assert.False(t, rbacsvc.IsBanned(key2))

		_, err := accountsvc.GetLock(context.Background(), key1)
		assert.ErrorIs(t, dao.ErrAccountLockNotExist, err)
		_, err = accountsvc.GetLock(context.Background(), key2)
		assert.ErrorIs(t, dao.ErrAccountLockNotExist, err)
	})
}

func TestTryLockAccount(t *testing.T) {
	key1 := rbac.MakeBanKey("attempted", "127.0.0.1")
	var oldReleaseAt int64
	t.Run("try lock account, should save attempted lock", func(t *testing.T) {
		rbacsvc.TryLockAccount(key1)

		lock, err := accountsvc.GetLock(context.Background(), key1)
		assert.NoError(t, err)
		assert.Equal(t, dao.StatusAttempted, lock.Status)

		assert.False(t, rbacsvc.IsBanned(key1))

		oldReleaseAt = lock.ReleaseAt

	})
	t.Run("try lock account again, should save a new attempted lock", func(t *testing.T) {
		time.Sleep(time.Second)

		rbacsvc.TryLockAccount(key1)

		lock, err := accountsvc.GetLock(context.Background(), key1)
		assert.NoError(t, err)
		assert.Equal(t, dao.StatusAttempted, lock.Status)
		assert.Less(t, oldReleaseAt, lock.ReleaseAt)

		assert.False(t, rbacsvc.IsBanned(key1))
	})
}

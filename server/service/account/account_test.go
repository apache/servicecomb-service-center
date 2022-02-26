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

package account_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	dao "github.com/apache/servicecomb-service-center/datasource/rbac"
	accountsvc "github.com/apache/servicecomb-service-center/server/service/account"
	"github.com/stretchr/testify/assert"
)

func TestIsBanned(t *testing.T) {
	t.Run("ban a key and check status, it should be banned, check other key should not be banned",
		func(t *testing.T) {
			err := accountsvc.Ban(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)

			ok, err := accountsvc.IsBanned(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.True(t, ok)

			ok, err = accountsvc.IsBanned(context.TODO(), "test_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.False(t, ok)

			time.Sleep(4 * time.Second)
			ok, err = accountsvc.IsBanned(context.TODO(), "dev_guy::127.0.0.1")
			assert.NoError(t, err)
			assert.False(t, ok)
		})
}

func TestListLock(t *testing.T) {
	t.Run("list 1 account lock, should return 1 item", func(t *testing.T) {
		err := accountsvc.Ban(context.TODO(), "dev_lock::127.0.0.1")
		assert.NoError(t, err)

		locks, n, err := accountsvc.ListLock(context.Background())
		assert.NoError(t, err)
		assert.NotEqual(t, 0, n)
		for _, lock := range locks {
			if lock.Key == "dev_lock::127.0.0.1" {
				return
			}
		}
		assert.Fail(t, "test key not found")
	})
}

func TestBan(t *testing.T) {
	var banTime int64

	t.Run("ban account TestAccountLock, should return no error", func(t *testing.T) {
		err := accountsvc.Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := dao.Instance().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, dao.StatusBanned, lock.Status)
		assert.Less(t, time.Now().Unix(), lock.ReleaseAt)

		banTime = lock.ReleaseAt
	})

	t.Run("ban account TestAccountLock again, should return a new release time", func(t *testing.T) {
		time.Sleep(time.Second)

		err := accountsvc.Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := dao.Instance().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, dao.StatusBanned, lock.Status)
		assert.Less(t, banTime, lock.ReleaseAt)
	})

	t.Run("ban account TestAccountLock again, should refresh releaseAt", func(t *testing.T) {
		lock1, err := dao.Instance().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Equal(t, dao.StatusBanned, lock1.Status)

		time.Sleep(time.Second)
		err = accountsvc.Ban(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock2, err := dao.Instance().GetLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)
		assert.Less(t, lock1.ReleaseAt, lock2.ReleaseAt)
	})

	t.Run("delete account lock, should return no error", func(t *testing.T) {
		err := dao.Instance().DeleteLock(context.Background(), "TestAccountLock")
		assert.NoError(t, err)

		lock, err := dao.Instance().GetLock(context.Background(), "TestAccountLock")
		assert.Equal(t, dao.ErrAccountLockNotExist, err)
		assert.Nil(t, lock)
	})
}

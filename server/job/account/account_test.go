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

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/server/job/account"
	"github.com/stretchr/testify/assert"
)

func TestCleanupReleasedLockHistory(t *testing.T) {
	t.Run("have a released lock, should be cleanup", func(t *testing.T) {
		const key = "TestCleanupReleasedLockHistory1"
		err := datasource.GetAccountLockManager().UpsertLock(context.Background(), &datasource.AccountLock{
			Key:       key,
			Status:    datasource.StatusBanned,
			ReleaseAt: time.Now().Add(-time.Second).Unix(),
		})
		assert.NoError(t, err)

		err = account.CleanupReleasedLockHistory(context.Background())
		assert.NoError(t, err)

		_, err = datasource.GetAccountLockManager().GetLock(context.Background(), key)
		assert.Equal(t, datasource.ErrAccountLockNotExist, err)
	})
	t.Run("have an unreleased lock, should NOT be cleanup", func(t *testing.T) {
		const key = "TestCleanupReleasedLockHistory2"
		err := datasource.GetAccountLockManager().UpsertLock(context.Background(), &datasource.AccountLock{
			Key:       key,
			Status:    datasource.StatusBanned,
			ReleaseAt: time.Now().Add(time.Minute).Unix(),
		})
		assert.NoError(t, err)

		err = account.CleanupReleasedLockHistory(context.Background())
		assert.NoError(t, err)

		lock, err := datasource.GetAccountLockManager().GetLock(context.Background(), key)
		assert.NoError(t, err)
		assert.NotNil(t, lock)

		err = datasource.GetAccountLockManager().DeleteLock(context.Background(), key)
		assert.NoError(t, err)
	})
}

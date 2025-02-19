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

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
)

func TestDefaultHealthChecker_Healthy(t *testing.T) {
	event.Center().Start()

	// normal case
	var hc DefaultHealthChecker
	if err := hc.Healthy(); err != nil {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}
	alarm.Raise(alarm.IDBackendConnectionRefuse, alarm.AdditionalContext("a"))
	time.Sleep(time.Second)
	if err := hc.Healthy(); err == nil || err.Error() != "a" {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}
	alarm.Clear(alarm.IDBackendConnectionRefuse)
	time.Sleep(time.Second)
	if err := hc.Healthy(); err != nil {
		t.Fatal("TestDefaultHealthChecker_Healthy failed", err)
	}

	// set global hc
	if GlobalHealthChecker() == &hc {
		t.Fatal("TestDefaultHealthChecker_Healthy failed")
	}

	SetGlobalHealthChecker(&hc)

	if GlobalHealthChecker() != &hc {
		t.Fatal("TestDefaultHealthChecker_Healthy failed")
	}
}

func TestHealthy(t *testing.T) {
	now := time.Now()
	t.Run("sync_not_start", func(t *testing.T) {
		err := syncReadinessChecker.Healthy()
		assert.ErrorIs(t, err, scNotReadyError)
	})

	t.Run("no_sync", func(t *testing.T) {
		// 未接受到同步请求，并且nowTime.Before(passTime)
		src := &SyncReadinessChecker{startupTime: now}
		err := src.Healthy()
		assert.ErrorIs(t, err, syncerNotReadyError)
	})

	t.Run("no_sync_but_exceeds_60s", func(t *testing.T) {
		// 未接受到同步请求，但是超出最大等待时间
		src := &SyncReadinessChecker{startupTime: now.Add(-60 * time.Second)}
		err := src.Healthy()
		assert.Nil(t, err)
	})

	t.Run("sync_and_before", func(t *testing.T) {
		// 30s内接受到第一次sync请求，并且nowTime.Before(passTime)
		src := &SyncReadinessChecker{startupTime: now}
		err := src.Healthy()
		rpc.RecordFirstReceivedRequestTime()
		assert.ErrorIs(t, err, syncerNotReadyError)
	})

	t.Run("31s_sync_and_before", func(t *testing.T) {
		// 30s后接受到第一次sync请求，并且nowTime.Before(passTime)
		src := &SyncReadinessChecker{startupTime: now.Add(-31 * time.Second)}
		err := src.Healthy()
		assert.ErrorIs(t, err, syncerNotReadyError)
	})
}

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
	"errors"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
)

var (
	healthChecker        Checker = &NullChecker{}
	readinessChecker     Checker = &DefaultHealthChecker{}
	syncReadinessChecker Checker = &SyncReadinessChecker{}
)

var (
	scNotReadyError     = errors.New("sc api server is not ready")
	syncerNotReadyError = errors.New("the syncer module is not ready")
)

type Checker interface {
	Healthy() error
}

type NullChecker struct {
}

func (n NullChecker) Healthy() error {
	return nil
}

type DefaultHealthChecker struct {
}

type SyncReadinessChecker struct {
	startupTime time.Time
}

func SetStartupTime(startupTime time.Time) {
	syncReadinessChecker.(*SyncReadinessChecker).startupTime = startupTime
}

func (src *SyncReadinessChecker) Healthy() error {
	err := defaultHealth()
	if err != nil {
		return err
	}
	if src.startupTime.IsZero() {
		return scNotReadyError
	}
	passTime := src.startupTime.Add(30 * time.Second)
	if !rpc.IsNotReceiveSyncRequest() && rpc.GetFirstReceiveTime().Sub(src.startupTime) < 30*time.Second {
		passTime = passTime.Add(rpc.GetFirstReceiveTime().Sub(src.startupTime))
	} else {
		log.Warn(fmt.Sprintf("first sync request is not received or received 30 seconds after startup,%s,%s", rpc.GetFirstReceiveTime(), src.startupTime))
		passTime = passTime.Add(30 * time.Second)
	}
	nowTime := time.Now()
	if nowTime.After(passTime) {
		return nil
	}
	return syncerNotReadyError
}

func (hc *DefaultHealthChecker) Healthy() error {
	return defaultHealth()
}

func defaultHealth() error {
	for _, a := range alarm.ListAll() {
		if a.Status == alarm.Cleared {
			continue
		}
		if a.ID == alarm.IDBackendConnectionRefuse || a.ID == alarm.IDScSelfHeartbeatFailed {
			return errors.New(a.FieldString(alarm.FieldAdditionalContext))
		}
	}
	return nil
}

func SetGlobalHealthChecker(hc Checker) {
	healthChecker = hc
}

func GlobalHealthChecker() Checker {
	return healthChecker
}

func SetGlobalReadinessChecker(hc Checker) {
	readinessChecker = hc
}

func GlobalReadinessChecker() Checker {
	if config.GetConfig().Sync.EnableOnStart {
		return syncReadinessChecker
	}
	return readinessChecker
}

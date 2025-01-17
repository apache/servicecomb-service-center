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

	"github.com/apache/servicecomb-service-center/server/alarm"
)

var healthChecker Checker = &NullChecker{}
var readinessChecker Checker = &DefaultHealthChecker{}

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

func (hc *DefaultHealthChecker) Healthy() error {
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
	return readinessChecker
}

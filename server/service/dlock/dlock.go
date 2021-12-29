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

// Package dlock provide distributed lock function
package dlock

import (
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/dlock"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func Lock(key string, ttl int64) error {
	return dlock.Instance().Lock(key, ttl)
}

func TryLock(key string, ttl int64) error {
	return dlock.Instance().TryLock(key, ttl)
}

func Renew(key string) error {
	return dlock.Instance().Renew(key)
}

func IsHoldLock(key string) bool {
	return dlock.Instance().IsHoldLock(key)
}

func Unlock(key string) {
	err := dlock.Instance().Unlock(key)
	if err != nil {
		log.Error(fmt.Sprintf("unlock key %s failed", key), err)
	}
}

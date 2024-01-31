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

package service

import (
	"strings"

	"github.com/apache/servicecomb-service-center/syncer/config"
)

const suffix = "*"

func IsExistInWhiteList(serviceName string) bool {
	if config.GetConfig().Sync == nil || !config.GetConfig().Sync.EnableOnStart {
		return false
	}
	// not in the whitelist for compatibility with older versions
	if config.GetConfig().Sync.WhiteList == nil || config.GetConfig().Sync.WhiteList.Service == nil {
		return true
	}
	serviceRules := config.GetConfig().Sync.WhiteList.Service.Rules
	if len(serviceRules) == 0 {
		return true
	}
	for _, rule := range serviceRules {
		if strings.HasSuffix(rule, suffix) {
			// rules by name with *
			rule = rule[:len(rule)-1]
		}
		if strings.HasPrefix(serviceName, rule) {
			return true
		}
	}
	return false
}

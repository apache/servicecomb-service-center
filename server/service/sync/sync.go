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

package sync

import (
	"context"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
)

var (
	enable bool
	once   sync.Once
)

func Enable() bool {
	once.Do(func() {
		if config.GetBool("sync.enableOnStart", false) {
			enable = true
		}
	})
	return enable
}

func SetContext(ctx context.Context) context.Context {
	var val string
	if Enable() {
		val = "1"
	}
	return util.SetContext(ctx, util.CtxEnableSync, val)
}

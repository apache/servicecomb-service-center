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

package util

import (
	"context"

	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

func FromContext(ctx context.Context) []etcdadpt.OpOption {
	opts := make([]etcdadpt.OpOption, 0, 5)
	if util.NoCache(ctx) {
		opts = append(opts, etcdadpt.WithNoCache())
	} else if util.CacheOnly(ctx) {
		opts = append(opts, etcdadpt.WithCacheOnly())
	}
	if util.Global(ctx) {
		opts = append(opts, etcdadpt.WithGlobal())
	}
	return opts
}

func ContextOptions(ctx context.Context, opts ...etcdadpt.OpOption) []etcdadpt.OpOption {
	return append(FromContext(ctx), opts...)
}

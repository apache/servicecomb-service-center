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

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

func FromContext(ctx context.Context) []registry.PluginOpOption {
	opts := make([]registry.PluginOpOption, 0, 5)
	switch {
	case ctx.Value(util.CtxNocache) == "1":
		opts = append(opts, registry.WithNoCache())
	case ctx.Value(util.CtxCacheOnly) == "1":
		opts = append(opts, registry.WithCacheOnly())
	}
	if ctx.Value(util.CtxGlobal) == "1" {
		opts = append(opts, registry.WithGlobal())
	}
	return opts
}

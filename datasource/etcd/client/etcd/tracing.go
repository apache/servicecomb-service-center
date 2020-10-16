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

package etcd

import (
	"context"
	registry "github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"net/http"
)

func TracingBegin(ctx context.Context, operationName string, op registry.PluginOp) tracing.Span {
	r := &tracing.RegistryRequest{
		Ctx:      ctx,
		Options:  op,
		Endpoint: firstEndpoint,
	}
	return plugin.Plugins().Tracing().ClientBegin(operationName, r)
}

func TracingEnd(span tracing.Span, err error) {
	if err != nil {
		plugin.Plugins().Tracing().ClientEnd(span, http.StatusInternalServerError, err.Error())
		return
	}
	plugin.Plugins().Tracing().ClientEnd(span, http.StatusOK, "")
}

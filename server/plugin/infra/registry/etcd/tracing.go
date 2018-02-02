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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/tracing"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"golang.org/x/net/context"
	"net/http"
)

func TracingBegin(ctx context.Context, operationName string, op registry.PluginOp) tracing.Span {
	var action string
	switch op.Action {
	case registry.Get:
		action = http.MethodGet
	case registry.Put:
		action = http.MethodPut
	case registry.Delete:
		action = http.MethodDelete
	}
	r, err := http.NewRequest(action, util.BytesToStringWithNoCopy(op.Key), nil)
	if err != nil {
		util.Logger().Errorf(err, "new backend request failed")
		return nil
	}
	r = r.WithContext(ctx)
	return plugin.Plugins().Tracing().ClientBegin(operationName, r)
}

func TracingEnd(span tracing.Span, err error) {
	if err != nil {
		plugin.Plugins().Tracing().ClientEnd(span, http.StatusInternalServerError, err.Error())
		return
	}
	plugin.Plugins().Tracing().ClientEnd(span, http.StatusOK, "")
}

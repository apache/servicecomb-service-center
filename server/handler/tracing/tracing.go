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
package tracing

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/chain"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"net/http"
	"strconv"
)

type TracingHandler struct {
}

func (h *TracingHandler) Handle(i *chain.Invocation) {
	w, r := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter),
		i.Context().Value(rest.CTX_REQUEST).(*http.Request)

	plugin.Plugins().Tracing().StartServerSpan("api", r)

	i.Next(chain.WithAsyncFunc(func(ret chain.Result) {
		statusCode := w.Header().Get("X-Response-Status")
		code, _ := strconv.ParseInt(statusCode, 10, 64)
		if code == 0 {
			code = 200
		}
		plugin.Plugins().Tracing().FinishServerSpan(r, int(code), statusCode)
	}))
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &TracingHandler{})
}

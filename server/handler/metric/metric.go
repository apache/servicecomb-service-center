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

package metric

import (
	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/metrics"
	"net/http"
	"time"
)

type MetricsHandler struct {
}

func (h *MetricsHandler) Handle(i *chain.Invocation) {
	i.Next(chain.WithAsyncFunc(func(ret chain.Result) {
		start, ok := i.Context().Value(rest.CtxStartTimestamp).(time.Time)
		if !ok {
			return
		}
		w, r := i.Context().Value(rest.CtxResponse).(http.ResponseWriter),
			i.Context().Value(rest.CtxRequest).(*http.Request)
		metrics.ReportRequestCompleted(w, r, start)
		log.NilOrWarnf(start, "%s %s", r.Method, r.RequestURI)
	}))
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &MetricsHandler{})
}

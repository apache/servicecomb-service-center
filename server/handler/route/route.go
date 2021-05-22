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

package route

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/rest"
)

// Handler call the func of matched route
type Handler struct {
}

func (l *Handler) Handle(i *chain.Invocation) {
	w, r := i.Context().Value(rest.CtxResponse).(http.ResponseWriter),
		i.Context().Value(rest.CtxRequest).(*http.Request)
	ph := i.Context().Value(rest.CtxRouteHandler).(http.Handler)

	ph.ServeHTTP(w, r)

	i.Next()
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

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

package auth

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/scerror"
)

type Handler struct {
}

func (h *Handler) Handle(i *chain.Invocation) {
	r := i.Context().Value(rest.CtxRequest).(*http.Request)
	err := plugin.Plugins().Auth().Identify(r)
	if err == nil {
		i.Next()
		return
	}

	log.Errorf(err, "authenticate request failed, %s %s", r.Method, r.RequestURI)

	w := i.Context().Value(rest.CtxResponse).(http.ResponseWriter)
	controller.WriteError(w, scerror.ErrUnauthorized, err.Error())

	i.Fail(nil)
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

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
	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/go-chassis/security/authr"
	"github.com/go-chassis/go-chassis/server/restful"
	"net/http"
	"strings"
)

type Handler struct {
}

func (h *Handler) Handle(i *chain.Invocation) {
	if !rbac.Enabled() {
		i.Next()
		return
	}
	w := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter)
	req, ok := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	if !ok {
		controller.WriteError(w, scerr.ErrUnauthorized, "internal error")
		i.Fail(nil)
		return
	}
	if !mustAuth(req) {
		i.Next()
		return
	}

	v := req.Header.Get(restful.HeaderAuth)
	if v == "" {
		controller.WriteError(w, scerr.ErrUnauthorized, "should provide token in header")
		i.Fail(nil)
		return
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		controller.WriteError(w, scerr.ErrUnauthorized, "invalid auth header")
		i.Fail(nil)
		return
	}
	to := s[1]
	//TODO rbac
	_, err := authr.Authenticate(i.Context(), to)
	if err == nil {
		log.Info("user access")
		i.Next()
		return
	}
	log.Errorf(err, "authenticate request failed, %s %s", req.Method, req.RequestURI)
	controller.WriteError(w, scerr.ErrUnauthorized, err.Error())
	i.Fail(nil)
}
func mustAuth(req *http.Request) bool {
	if strings.Contains(req.URL.Path, "/v4/token") {
		return false
	}
	if strings.Contains(req.URL.Path, "/health") {
		return false
	}
	if strings.Contains(req.URL.Path, "/version") {
		return false
	}
	return true
}
func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

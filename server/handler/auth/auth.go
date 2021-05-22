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
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/go-chassis/cari/discovery"
)

type Handler struct {
}

func (h *Handler) Handle(i *chain.Invocation) {
	r := i.Context().Value(rest.CtxRequest).(*http.Request)

	if err := auth.Identify(r); err != nil {
		log.Errorf(err, "authenticate request failed, %s %s", r.Method, r.RequestURI)
		i.Fail(discovery.NewError(discovery.ErrUnauthorized, err.Error()))
		return
	}

	i.Next(chain.WithFunc(func(ret chain.Result) {
		if !ret.OK {
			return
		}

		obj := i.Context().Value(rest.CtxResponseObject)
		if obj == nil {
			return
		}

		// TODO filter and rewrite here!
		// data, _ := json.Marshal(obj)
		// w := i.Context().Value(rest.CtxResponse).(http.ResponseWriter)
		// w.Write(data)
	}))
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

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
package context

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/chain"
	roa "github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"net/http"
)

type ContextHandler struct {
}

func (c *ContextHandler) Handle(i *chain.Invocation) {
	var (
		err     error
		v3      v3Context
		v4      v4Context
		r       = i.Context().Value(roa.CTX_REQUEST).(*http.Request)
		pattern = i.Context().Value(roa.CTX_MATCH_PATTERN).(string)
	)

	switch {
	case IsSkip(pattern):
	case v3.IsMatch(r):
		err = v3.Do(r)
	case v4.IsMatch(r):
		err = v4.Do(r)
	}

	if err != nil {
		i.Fail(err)
		return
	}

	i.WithContext("x-remote-ip", util.GetRealIP(r))

	i.Next()
}

func IsSkip(url string) bool {
	l, vl, hl := len(url), len("/version"), len("/health")
	if l >= vl && url[l-vl:] == "/version" {
		return true
	}
	if l >= hl && url[l-hl:] == "/health" {
		return true
	}
	return false
}

func RegisterHandlers() {
	chain.RegisterHandler(roa.SERVER_CHAIN_NAME, &ContextHandler{})
}

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
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	queryGlobal    = "global"
	queryNoCache   = "noCache"
	queryCacheOnly = "cacheOnly"
)

type Handler struct {
}

func (c *Handler) Handle(i *chain.Invocation) {
	var (
		v3 v3Context
		v4 v4Context
		r  = i.Context().Value(rest.CtxRequest).(*http.Request)
	)

	switch {
	case v3.IsMatch(r):
		v3.Write(r)
	case v4.IsMatch(r):
		v4.Write(r)
	}

	c.commonQueryToContext(i)

	i.Next()
}

func (c *Handler) commonQueryToContext(i *chain.Invocation) {
	r := i.Context().Value(rest.CtxRequest).(*http.Request)
	query := r.URL.Query()

	i.WithContext(util.CtxRemoteIP, util.GetRealIP(r))

	global := util.StringTRUE(query.Get(queryGlobal))
	if global && r.Method == http.MethodGet {
		i.WithContext(util.CtxGlobal, "1")
	}

	noCache := util.StringTRUE(query.Get(queryNoCache))
	if noCache {
		i.WithContext(util.CtxNocache, "1")
		return
	}

	cacheOnly := util.StringTRUE(query.Get(queryCacheOnly))
	if cacheOnly {
		i.WithContext(util.CtxCacheOnly, "1")
		return
	}

	rev := query.Get("rev")
	if len(rev) > 0 {
		i.WithContext(util.CtxRequestRevision, rev)
		return
	}
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

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
package cache

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/chain"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"net/http"
	"strconv"
)

type CacheResponse struct {
}

func (l *CacheResponse) Handle(i *chain.Invocation) {
	defer i.Next()

	r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)

	noCache := r.URL.Query().Get(serviceUtil.CTX_NOCACHE) == "1"
	cacheOnly := r.URL.Query().Get(serviceUtil.CTX_CACHEONLY) == "1"
	rev, _ := strconv.ParseInt(r.URL.Query().Get("rev"), 10, 64)

	if noCache {
		i.WithContext(serviceUtil.CTX_NOCACHE, "1")
		return
	}

	if cacheOnly {
		i.WithContext(serviceUtil.CTX_CACHEONLY, "1")
		return
	}

	if rev > 0 {
		i.WithContext(serviceUtil.CTX_REQUEST_REVISION, rev)
		return
	}
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &CacheResponse{})
}

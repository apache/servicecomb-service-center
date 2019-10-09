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
package maxbody

import (
	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/core"
	"net/http"
)

const (
	microserviceSize = 5 * 1024 // 5KB
	instanceSize     = 5 * 1024 // 5KB
	propertiesSize   = 3 * 1024 // 3KB
)

var resourcesMap = map[string]int64{
	"/registry/v3/microservices":          microserviceSize,
	"/v4/:project/registry/microservices": microserviceSize,

	"/registry/v3/microservices/:serviceId/instances":          instanceSize,
	"/v4/:project/registry/microservices/:serviceId/instances": instanceSize,

	"/registry/v3/microservices/:serviceId/properties":          propertiesSize,
	"/v4/:project/registry/microservices/:serviceId/properties": propertiesSize,

	"/registry/v3/microservices/:serviceId/instances/:instanceId/properties":          propertiesSize,
	"/v4/:project/registry/microservices/:serviceId/instances/:instanceId/properties": propertiesSize,
}

type MaxBodyHandler struct {
}

func (c *MaxBodyHandler) Handle(i *chain.Invocation) {
	r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	if r.Method == http.MethodGet {
		i.Next()
		return
	}

	w, pattern := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter),
		i.Context().Value(rest.CTX_MATCH_PATTERN).(string)
	v, ok := resourcesMap[pattern]
	if !ok {
		v = core.ServerInfo.Config.MaxBodyBytes
	}

	r.Body = http.MaxBytesReader(w, r.Body, v)

	i.Next()
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &MaxBodyHandler{})
}

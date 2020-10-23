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

package rest

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	HTTPMethodGet    = http.MethodGet
	HTTPMethodPut    = http.MethodPut
	HTTPMethodPost   = http.MethodPost
	HTTPMethodDelete = http.MethodDelete

	CtxResponse     util.CtxKey = "_server_response"
	CtxRequest      util.CtxKey = "_server_request"
	CtxMatchPattern util.CtxKey = "_server_match_pattern"
	CtxMatchFunc    util.CtxKey = "_server_match_func"

	ServerChainName = "_server_chain"

	HeaderResponseStatus = "X-Response-Status"

	HeaderAllow           = "Allow"
	HeaderHost            = "Host"
	HeaderServer          = "Server"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"

	AcceptAny = "*/*"

	ContentTypeJSON = "application/json; charset=UTF-8"
	ContentTypeText = "text/plain; charset=UTF-8"

	DefaultConnPoolPerHostSize = 5
)

func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete:
		return true
	default:
		return false
	}
}

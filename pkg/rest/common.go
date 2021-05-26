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
	CtxResponse       util.CtxKey = "_server_response"
	CtxRequest        util.CtxKey = "_server_request"
	CtxMatchPattern   util.CtxKey = "_server_match_pattern"
	CtxMatchFunc      util.CtxKey = "_server_match_func"
	CtxStartTimestamp util.CtxKey = "x-start-timestamp"
	CtxResponseStatus util.CtxKey = "_server_response_status"
	CtxResponseObject util.CtxKey = "_server_response_object"
	CtxRouteHandler   util.CtxKey = "_server_route_handler"

	ServerChainName = "_server_chain"

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

func match(s string, f func(c byte) bool, exclude byte, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) && s[j] != exclude {
		j++
	}

	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

func matchParticipial(c byte) bool {
	return c != '/'
}

func isAlpha(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

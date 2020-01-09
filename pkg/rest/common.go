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
)

const (
	HTTP_METHOD_GET    = http.MethodGet
	HTTP_METHOD_PUT    = http.MethodPut
	HTTP_METHOD_POST   = http.MethodPost
	HTTP_METHOD_DELETE = http.MethodDelete

	CTX_RESPONSE      = "_server_response"
	CTX_REQUEST       = "_server_request"
	CTX_MATCH_PATTERN = "_server_match_pattern"
	CTX_MATCH_FUNC    = "_server_match_func"

	ServerChainName = "_server_chain"

	HEADER_RESPONSE_STATUS = "X-Response-Status"

	HEADER_ALLOW            = "Allow"
	HEADER_HOST             = "Host"
	HEADER_SERVER           = "Server"
	HEADER_CONTENT_TYPE     = "Content-Type"
	HEADER_CONTENT_ENCODING = "Content-Encoding"
	HEADER_ACCEPT           = "Accept"
	HEADER_ACCEPT_ENCODING  = "Accept-Encoding"

	ACCEPT_ANY  = "*/*"
	ACCEPT_JSON = "application/json"

	CONTENT_TYPE_JSON = "application/json; charset=UTF-8"
	CONTENT_TYPE_TEXT = "text/plain; charset=UTF-8"

	ENCODING_GZIP = "gzip"

	DEFAULT_CONN_POOL_PER_HOST_SIZE = 5
)

func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete:
		return true
	default:
		return false
	}
}

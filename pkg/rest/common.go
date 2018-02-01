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
	SERVER_CHAIN_NAME = "_server_chain"
)

func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete:
		return true
	default:
		return false
	}
}

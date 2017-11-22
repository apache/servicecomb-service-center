//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package rest

import (
	"time"
)

const (
	DEFAULT_TLS_HANDSHAKE_TIMEOUT = 30 * time.Second
	DEFAULT_HTTP_RESPONSE_TIMEOUT = 60 * time.Second

	HTTP_ERROR_STATUS_CODE = 600

	HTTP_METHOD_GET    = "GET"
	HTTP_METHOD_PUT    = "PUT"
	HTTP_METHOD_POST   = "POST"
	HTTP_METHOD_DELETE = "DELETE"

	CTX_RESPONSE      = "_server_response"
	CTX_REQUEST       = "_server_request"
	CTX_MATCH_PATTERN = "_server_match_pattern"
	SERVER_CHAIN_NAME = "_server_chain"
)

func isValidMethod(method string) bool {
	switch method {
	case HTTP_METHOD_GET, HTTP_METHOD_PUT, HTTP_METHOD_POST, HTTP_METHOD_DELETE:
		return true
	default:
		return false
	}
}

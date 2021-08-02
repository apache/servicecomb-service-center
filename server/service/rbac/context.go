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

package rbac

import (
	"context"
	"errors"
	"net/http"

	rbacmodel "github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	CtxRequestClaims util.CtxKey = "_request_claims"
	CtxRequestToken  util.CtxKey = "_request_token"
)

func UserFromContext(ctx context.Context) string {
	m, ok := ctx.Value(CtxRequestClaims).(map[string]interface{})
	if !ok {
		return ""
	}
	user, ok := m[rbacmodel.ClaimsUser].(string)
	if !ok {
		return ""
	}
	return user
}

func AccountFromContext(ctx context.Context) (*rbacmodel.Account, error) {
	m, ok := ctx.Value(CtxRequestClaims).(map[string]interface{})
	if !ok {
		return nil, errors.New("no claims from request context")
	}
	return rbacmodel.GetAccount(m)
}

func WithToken(req *http.Request, token string) *http.Request {
	return util.SetRequestContext(req, CtxRequestToken, token)
}

func TokenFromRequest(req *http.Request) string {
	token, ok := req.Context().Value(CtxRequestToken).(string)
	if !ok {
		return ""
	}
	return token
}

func SignRequest(req *http.Request) error {
	token := TokenFromRequest(req)
	h := req.Header
	if token == "" {
		return errors.New("request unauthorized")
	}
	h.Set("Authorization", token)
	req.Header = h
	return nil
}

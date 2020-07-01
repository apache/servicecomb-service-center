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
package buildin

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/go-chassis/security/authr"
	"github.com/go-chassis/go-chassis/server/restful"
	"net/http"
	"strings"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.AUTH, Name: "buildin", New: New})
}

func New() mgr.Instance {
	return &TokenAuthenticator{}
}

type TokenAuthenticator struct {
}

func (ba *TokenAuthenticator) Identify(req *http.Request) error {
	if !rbac.Enabled() {
		return nil
	}
	if !mustAuth(req) {
		return nil
	}

	v := req.Header.Get(restful.HeaderAuth)
	if v == "" {
		return auth.ErrNoHeader
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return errors.New("invalid auth header")
	}
	to := s[1]
	//TODO rbac
	claims, err := authr.Authenticate(req.Context(), to)
	if err != nil {
		log.Errorf(err, "authenticate request failed, %s %s", req.Method, req.RequestURI)
		return err
	}
	log.Info("user access")
	req2 := req.WithContext(context.WithValue(req.Context(), "accountInfo", claims))
	*req = *req2
	return nil
}
func mustAuth(req *http.Request) bool {
	for v := range rbac.WhiteAPIList() {
		if strings.Contains(req.URL.Path, v) {
			return false
		}
	}
	return true
}

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
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
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
	pattern, ok := req.Context().Value(rest.CtxMatchPattern).(string)
	if ok && !rbacframe.MustAuth(pattern) {
		return nil
	}
	v := req.Header.Get(restful.HeaderAuth)
	if v == "" {
		return rbacframe.ErrNoHeader
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return rbacframe.ErrInvalidHeader
	}
	to := s[1]

	claims, err := authr.Authenticate(req.Context(), to)
	if err != nil {
		log.Errorf(err, "authenticate request failed, %s %s", req.Method, req.RequestURI)
		return err
	}
	//TODO rbac
	m, ok := claims.(map[string]interface{})
	if !ok {
		log.Error("claims convert failed", rbacframe.ErrConvertErr)
		return rbacframe.ErrConvertErr
	}
	role := m[rbacframe.ClaimsRole]
	roleName, ok := role.(string)
	if !ok {
		log.Error("role convert failed", rbacframe.ErrConvertErr)
		return rbacframe.ErrConvertErr
	}
	r := dao.GetResource(context.TODO(), req.URL.Path)
	//TODO add verbs
	allow, err := rbac.Allow(context.TODO(), roleName, req.URL.Query().Get(":project"), r, "")
	if err != nil {
		log.Error("", err)
		return errors.New(errorsEx.ErrMsgRolePerm)
	}
	if !allow {
		return errors.New(errorsEx.ErrMsgNoPerm)
	}
	req2 := req.WithContext(rbacframe.NewContext(req.Context(), claims))
	*req = *req2
	return nil
}

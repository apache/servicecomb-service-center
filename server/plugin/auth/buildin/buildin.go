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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/go-chassis/cari/rbac"

	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

func init() {
	plugin.RegisterPlugin(plugin.Plugin{Kind: auth.AUTH, Name: "buildin", New: New})
}

func New() plugin.Instance {
	return &TokenAuthenticator{}
}

type TokenAuthenticator struct {
}

func (ba *TokenAuthenticator) Identify(req *http.Request) error {
	if !rbacsvc.Enabled() {
		return nil
	}
	pattern, ok := req.Context().Value(rest.CtxMatchPattern).(string)
	if ok && !mustAuth(pattern) {
		return nil
	}
	v := req.Header.Get(restful.HeaderAuth)
	if v == "" {
		return rbac.ErrNoHeader
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return rbac.ErrInvalidHeader
	}
	to := s[1]

	claims, err := authr.Authenticate(req.Context(), to)
	if err != nil {
		log.Errorf(err, "authenticate request failed, %s %s", req.Method, req.RequestURI)
		return err
	}
	m, ok := claims.(map[string]interface{})
	if !ok {
		log.Error("claims convert failed", rbac.ErrConvertErr)
		return rbac.ErrConvertErr
	}

	roleList, err := rbac.GetRolesList(m)
	if err != nil {
		log.Error("get role list failed", err)
		return err
	}

	var apiPattern string
	a := req.Context().Value(rest.CtxMatchPattern)
	if a == nil { //handle exception
		apiPattern = req.URL.Path
		log.Warn("can not find api pattern")
	} else {
		apiPattern = a.(string)
	}

	project := req.URL.Query().Get(":project")
	verbs := rbacsvc.MethodToVerbs[req.Method]
	err = checkPerm(roleList, project, apiPattern, verbs)
	if err != nil {
		return err
	}
	req2 := req.WithContext(rbac.NewContext(req.Context(), m))
	*req = *req2
	return nil
}

//this method decouple business code and perm checks
func checkPerm(roleList []string, project, apiPattern, verbs string) error {
	resource := rbac.GetResource(apiPattern)
	if resource == "" {
		//fast fail, no need to access role storage
		return errors.New(errorsEx.MsgNoPerm)
	}
	//TODO add verbs,project
	allow, err := rbacsvc.Allow(context.TODO(), roleList, project, resource, verbs)
	if err != nil {
		log.Error("", err)
		return errors.New(errorsEx.MsgRolePerm)
	}
	if !allow {
		return errors.New(errorsEx.MsgNoPerm)
	}
	return nil
}

func mustAuth(pattern string) bool {
	if util.IsVersionOrHealthPattern(pattern) {
		return false
	}
	return rbac.MustAuth(pattern)
}

func (ba *TokenAuthenticator) ResourceScopes(r *http.Request) *auth.ResourceScope {
	return FromRequest(r)
}

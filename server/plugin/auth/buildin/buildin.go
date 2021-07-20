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
	"errors"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	authHandler "github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/cari/rbac"
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
	if !ok {
		pattern = req.URL.Path
		log.Warn("can not find api pattern")
	}

	if !mustAuth(pattern) {
		return nil
	}

	claims, err := ba.VerifyToken(req)
	if err != nil {
		log.Errorf(err, "verify request token failed, %s %s", req.Method, req.RequestURI)
		return err
	}

	m, ok := claims.(map[string]interface{})
	if !ok {
		log.Error("claims convert failed", rbac.ErrConvertErr)
		return rbac.ErrConvertErr
	}
	account, err := rbac.GetAccount(m)
	if err != nil {
		log.Error("get account from token failed", err)
		return err
	}
	util.SetRequestContext(req, rbacsvc.CtxRequestClaims, m)
	// user can change self password
	if isChangeSelfPassword(pattern, account, req) {
		return nil
	}

	if len(account.Roles) == 0 {
		log.Error("no role found in token", nil)
		return errors.New("no role found in token")
	}

	project := req.URL.Query().Get(":project")
	allow, matchedLabels, err := checkPerm(account.Roles, project, req, pattern, req.Method)
	if err != nil {
		return err
	}
	if !allow {
		return rbac.NewError(rbac.ErrNoPermission, "")
	}

	util.SetRequestContext(req, authHandler.CtxResourceLabels, matchedLabels)
	return nil
}

func isChangeSelfPassword(pattern string, a *rbac.Account, req *http.Request) bool {
	if pattern != rbacsvc.APIAccountPassword {
		return false
	}
	changerName := a.Name
	targetName := req.URL.Query().Get(":name")
	return changerName == targetName
}

func filterRoles(roleList []string) (hasAdmin bool, normalRoles []string) {
	for _, r := range roleList {
		if r == rbac.RoleAdmin {
			hasAdmin = true
			return
		}
		normalRoles = append(normalRoles, r)
	}
	return
}

func (ba *TokenAuthenticator) VerifyToken(req *http.Request) (interface{}, error) {
	v := req.Header.Get(restful.HeaderAuth)
	if v == "" {
		return nil, rbac.NewError(rbac.ErrNoAuthHeader, "")
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return nil, rbac.ErrInvalidHeader
	}
	to := s[1]

	return authr.Authenticate(req.Context(), to)
}

//this method decouple business code and perm checks
func checkPerm(roleList []string, project string, req *http.Request, apiPattern, method string) (bool, []map[string]string, error) {
	hasAdmin, normalRoles := filterRoles(roleList)
	if hasAdmin {
		return true, nil, nil
	}
	//todo fast check for dev role
	targetResource := FromRequest(req)
	if targetResource == nil {
		return false, nil, errors.New("no valid resouce scope")
	}
	//TODO add project
	return rbacsvc.Allow(req.Context(), project, normalRoles, targetResource)
}

func mustAuth(pattern string) bool {
	if util.IsVersionOrHealthPattern(pattern) {
		return false
	}
	return rbac.MustAuth(pattern)
}

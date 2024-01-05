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
	"fmt"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	authHandler "github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/token"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

var ErrNoRoles = errors.New("no role found in token")

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

	pattern := getRequestPattern(req)

	account, err := ba.mustAuth(req, pattern)
	if account == nil || err != nil {
		return err
	}

	err = accountExist(req.Context(), account.Name)
	if err != nil {
		return err
	}

	if !rbacsvc.MustCheckPerm(pattern) {
		return nil
	}

	matchedLabels, err := checkPerm(account.Roles, req)
	if err != nil {
		return err
	}

	util.SetRequestContext(req, authHandler.CtxResourceLabels, matchedLabels)
	return nil
}

func getRequestPattern(req *http.Request) string {
	pattern, ok := req.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		pattern = req.URL.Path
		log.Warn("can not find api pattern")
	}
	return pattern
}

func (ba *TokenAuthenticator) mustAuth(req *http.Request, pattern string) (*rbacmodel.Account, error) {
	if !rbacsvc.MustAuth(pattern) {
		return nil, nil
	}
	return ba.VerifyRequest(req)
}

func (ba *TokenAuthenticator) VerifyRequest(req *http.Request) (*rbacmodel.Account, error) {
	claims, err := ba.VerifyToken(req)
	if err != nil {
		log.Error(fmt.Sprintf("verify request token failed, %s %s", req.Method, req.RequestURI), err)
		return nil, err
	}
	m, ok := claims.(map[string]interface{})
	if !ok {
		log.Error("claims convert failed", rbacmodel.ErrConvert)
		return nil, rbacmodel.ErrConvert
	}
	util.SetRequestContext(req, rbacsvc.CtxRequestClaims, m)
	account, err := rbacmodel.GetAccount(m)
	if err != nil {
		log.Error("get account from token failed", err)
		return nil, err
	}
	if len(account.Roles) == 0 {
		log.Error("no role found in token", nil)
		return nil, ErrNoRoles
	}
	return account, nil
}

func accountExist(ctx context.Context, user string) error {
	// if root should pass, cause of root initialization
	if user == rbacsvc.RootName {
		return nil
	}
	exist, err := rbacsvc.AccountExist(ctx, user)
	if err != nil {
		return err
	}
	if !exist {
		msg := fmt.Sprintf("account [%s] is deleted", user)
		return rbacmodel.NewError(rbacmodel.ErrTokenOwnedAccountDeleted, msg)
	}
	return nil
}

func filterRoles(roleList []string) (hasAdmin bool, normalRoles []string) {
	for _, r := range roleList {
		if r == rbacmodel.RoleAdmin {
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
		return nil, rbacmodel.NewError(rbacmodel.ErrNoAuthHeader, "")
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return nil, rbacmodel.ErrInvalidHeader
	}
	to := s[1]

	claims, err := authr.Authenticate(req.Context(), to)
	if err != nil {
		return nil, err
	}
	token.WithRequest(req, to)
	return claims, nil
}

// this method decouple business code and perm checks
func checkPerm(roleList []string, req *http.Request) ([]map[string]string, error) {
	hasAdmin, normalRoles := filterRoles(roleList)
	if hasAdmin {
		return nil, nil
	}
	//todo fast check for dev role
	targetResource := FromRequest(req)
	if targetResource == nil {
		return nil, errors.New("no valid resouce scope")
	}
	//TODO add project
	project := req.URL.Query().Get(":project")
	return rbacsvc.Allow(req.Context(), project, normalRoles, targetResource)
}

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
	"encoding/json"
	"io"
	"net/http"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	accountsvc "github.com/apache/servicecomb-service-center/server/service/account"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	"github.com/go-chassis/cari/discovery"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
)

const DefaultTokenExpirationDuration = "12h"

type AuthResource struct {
}

// URLPatterns define htp pattern
func (ar *AuthResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/token", Func: ar.Login},
		{Method: http.MethodGet, Path: "/v4/self-perms", Func: ar.ListSelfPerms},
		{Method: http.MethodPost, Path: "/v4/accounts", Func: ar.CreateAccount},
		{Method: http.MethodPost, Path: "/v4/accounts/batch-create", Func: ar.BatchCreateAccount},
		{Method: http.MethodGet, Path: "/v4/accounts", Func: ar.ListAccount},
		{Method: http.MethodGet, Path: "/v4/accounts/:name", Func: ar.GetAccount},
		{Method: http.MethodDelete, Path: "/v4/accounts/:name", Func: ar.DeleteAccount},
		{Method: http.MethodPut, Path: "/v4/accounts/:name", Func: ar.UpdateAccount},
		{Method: http.MethodPost, Path: "/v4/accounts/:name/password", Func: ar.ChangePassword},
		{Method: http.MethodGet, Path: "/v4/account-locks", Func: ar.ListLock},
	}
}

func (ar *AuthResource) CreateAccount(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacmodel.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	err = rbacsvc.CreateAccount(req.Context(), a)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) DeleteAccount(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get(":name")
	err := rbacsvc.DeleteAccount(req.Context(), name)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) UpdateAccount(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get(":name")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacmodel.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}

	err = rbacsvc.UpdateAccount(req.Context(), name, a)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) ListAccount(w http.ResponseWriter, r *http.Request) {
	as, n, err := rbacsvc.ListAccount(r.Context())
	if err != nil {
		log.Error(errorsEx.MsgGetAccountFailed, err)
		rest.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetAccountFailed)
		return
	}
	resp := &rbacmodel.AccountResponse{
		Total:    n,
		Accounts: as,
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (ar *AuthResource) GetAccount(w http.ResponseWriter, r *http.Request) {
	a, err := rbacsvc.GetAccount(r.Context(), r.URL.Query().Get(":name"))
	if err != nil {
		log.Error(errorsEx.MsgGetAccountFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	a.Password = ""
	rest.WriteResponse(w, r, nil, a)
}

func (ar *AuthResource) ChangePassword(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacmodel.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	a.Name = req.URL.Query().Get(":name")

	err = rbacsvc.ChangePassword(req.Context(), a)
	if err != nil {
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacmodel.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	if a.TokenExpirationTime == "" {
		a.TokenExpirationTime = DefaultTokenExpirationDuration
	}
	err = validator.ValidateAccountLogin(a)
	if err != nil {
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	t, err := authr.Login(r.Context(), a.Name, a.Password,
		authr.ExpireAfter(a.TokenExpirationTime))
	if err != nil {
		log.Error("not authorized", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &rbacmodel.Token{TokenStr: t})
}

func (ar *AuthResource) ListSelfPerms(w http.ResponseWriter, r *http.Request) {
	perms, err := rbacsvc.ListSelfPerms(r.Context())
	if err != nil {
		log.Error(errorsEx.MsgGetAccountFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	resp := &rbacmodel.SelfPermissionResponse{
		Perms: perms,
	}
	rest.WriteResponse(w, r, nil, resp)

}

func (ar *AuthResource) ListLock(w http.ResponseWriter, r *http.Request) {
	al, n, err := accountsvc.ListLock(r.Context())
	if err != nil {
		log.Error("get account lock failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	resp := &rbac.LockResponse{
		Total: n,
		Locks: al,
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (ar *AuthResource) BatchCreateAccount(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacmodel.BatchCreateAccountsRequest{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	resp, err := rbacsvc.BatchCreateAccounts(r.Context(), a)
	if err != nil {
		log.Error(errorsEx.MsgBatchCreateAccountsFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func MakeBanKey(name, ip string) string {
	return name + "::" + ip
}

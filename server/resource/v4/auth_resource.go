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

package v4

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"

	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/validator"
)

type AuthResource struct {
}

//URLPatterns define htp pattern
func (ar *AuthResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/token", Func: ar.Login},
		{Method: http.MethodPost, Path: "/v4/accounts", Func: ar.CreateAccount},
		{Method: http.MethodGet, Path: "/v4/accounts", Func: ar.ListAccount},
		{Method: http.MethodGet, Path: "/v4/accounts/:name", Func: ar.GetAccount},
		{Method: http.MethodDelete, Path: "/v4/accounts/:name", Func: ar.DeleteAccount},
		{Method: http.MethodPut, Path: "/v4/accounts/:name", Func: ar.UpdateAccount},
		{Method: http.MethodPost, Path: "/v4/accounts/:name/password", Func: ar.ChangePassword},
	}
}

func (ar *AuthResource) CreateAccount(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbac.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	err = rbacsvc.CreateAccount(req.Context(), a)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		writeErrsvcOrInternalErr(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) DeleteAccount(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get(":name")
	err := rbacsvc.DeleteAccount(req.Context(), name)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		writeErrsvcOrInternalErr(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) UpdateAccount(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get(":name")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbac.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}

	err = rbacsvc.UpdateAccount(req.Context(), name, a)
	if err != nil {
		log.Error(errorsEx.MsgOperateAccountFailed, err)
		writeErrsvcOrInternalErr(w, err)
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
	resp := &rbac.AccountResponse{
		Total:    n,
		Accounts: as,
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (ar *AuthResource) GetAccount(w http.ResponseWriter, r *http.Request) {
	a, err := rbacsvc.GetAccount(r.Context(), r.URL.Query().Get(":name"))
	if err != nil {
		log.Error(errorsEx.MsgGetAccountFailed, err)
		writeErrsvcOrInternalErr(w, err)
		return
	}
	a.Password = ""
	rest.WriteResponse(w, r, nil, a)
}

func (ar *AuthResource) ChangePassword(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbac.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	a.Name = req.URL.Query().Get(":name")

	err = rbacsvc.ChangePassword(req.Context(), a)
	if err != nil {
		writeErrsvcOrInternalErr(w, err)
		return
	}
	rest.WriteSuccess(w, req)
}

func (ar *AuthResource) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbac.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	if a.TokenExpirationTime == "" {
		a.TokenExpirationTime = "30m"
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
		writeErrsvcOrInternalErr(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &rbac.Token{TokenStr: t})
}

func MakeBanKey(name, ip string) string {
	return name + "::" + ip
}

func writeErrsvcOrInternalErr(w http.ResponseWriter, err error) {
	e, ok := err.(*errsvc.Error)
	if ok {
		rest.WriteErrsvcError(w, e)
		return
	}
	rest.WriteError(w, discovery.ErrInternal, err.Error())
}

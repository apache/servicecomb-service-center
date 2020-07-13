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
	"context"
	"encoding/json"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/server/service"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/go-chassis/go-chassis/security/authr"
	"io/ioutil"
	"net/http"
)

type AuthResource struct {
}

//URLPatterns define htp pattern
func (r *AuthResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/token", Func: r.Login},
		{Method: http.MethodPost, Path: "/v4/account", Func: r.CreateAccount},
		{Method: http.MethodPost, Path: "/v4/account/:name/password", Func: r.ChangePassword},
	}
}
func (r *AuthResource) CreateAccount(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	a := &rbacframe.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		controller.WriteError(w, scerror.ErrInvalidParams, errorsEx.ErrMsgJSON)
		return
	}
	err = service.ValidateCreateAccount(a)
	if err != nil {
		controller.WriteError(w, scerror.ErrInvalidParams, err.Error())
		return
	}
	err = dao.CreateAccount(context.TODO(), a)
	if err != nil {
		if err == dao.ErrDuplicated {
			controller.WriteError(w, scerror.ErrConflictAccount, "")
			return
		}
		log.Error(errorsEx.ErrMsgCreateAccount, err)
		controller.WriteError(w, scerror.ErrInternal, errorsEx.ErrMsgCreateAccount)
		return
	}
}
func (r *AuthResource) ChangePassword(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	a := &rbacframe.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		controller.WriteError(w, scerror.ErrInvalidParams, errorsEx.ErrMsgJSON)
		return
	}
	a.Name = req.URL.Query().Get(":name")
	err = service.ValidateChangePWD(a)
	if err != nil {
		controller.WriteError(w, scerror.ErrInvalidParams, err.Error())
		return
	}
	changer, err := rbacframe.AccountFromContext(req.Context())
	if err != nil {
		controller.WriteError(w, scerror.ErrInternal, "can not parse account info")
		return
	}
	err = rbac.ChangePassword(context.TODO(), changer.Role, changer.Name, a)
	if err != nil {
		if err == rbac.ErrSamePassword ||
			err == rbac.ErrWrongPassword || err == rbac.ErrEmptyCurrentPassword ||
			err == rbac.ErrNoPermChangeAccount {
			controller.WriteError(w, scerror.ErrInvalidParams, err.Error())
			return
		}
		log.Error("change password failed", err)
		controller.WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
}

func (r *AuthResource) Login(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	a := &rbacframe.Account{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		controller.WriteError(w, scerror.ErrInvalidParams, err.Error())
		return
	}
	t, err := authr.Login(context.TODO(), a.Name, a.Password)
	if err != nil {
		if err == rbac.ErrUnauthorized {
			log.Error("not authorized", err)
			controller.WriteError(w, scerror.ErrUnauthorized, err.Error())
			return
		}
		log.Error("can not sign token", err)
		controller.WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	to := &rbacframe.Token{TokenStr: t}
	b, err := json.Marshal(to)
	if err != nil {
		log.Error("json err", err)
		controller.WriteError(w, scerror.ErrInvalidParams, err.Error())
		return
	}
	controller.WriteJSON(w, b)
}

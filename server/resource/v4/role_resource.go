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
	"github.com/apache/servicecomb-service-center/datasource"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/go-chassis/cari/discovery"
	"io/ioutil"
	"net/http"
)

var ErrConflictRole int32 = 409002

type RoleResource struct {
}

//URLPatterns define htp pattern
func (r *RoleResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/role", Func: r.GetRolePermission},
		{Method: http.MethodPost, Path: "/v4/role", Func: r.CreateRolePermission},
		{Method: http.MethodPut, Path: "/v4/role/:roleName", Func: r.UpdateRolePermission},
		{Method: http.MethodGet, Path: "/v4/role/:roleName", Func: r.GetRole},
		{Method: http.MethodDelete, Path: "/v4/role/:roleName", Func: r.DeleteRole},
	}
}

func (r *RoleResource) GetRolePermission(w http.ResponseWriter, req *http.Request) {
	rs, _, err := dao.ListRole(context.TODO())
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
		return
	}
	resp := &rbacframe.RoleResponse{
		Roles: rs,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
	controller.WriteJSON(w, b)
}

func (r *RoleResource) CreateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacframe.Role{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		controller.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	// TODO: validate role
	err = dao.CreateRole(context.TODO(), a)
	if err != nil {
		if err == datasource.ErrRoleDuplicated {
			controller.WriteError(w, ErrConflictRole, "")
			return
		}
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgOperateRoleFailed)
		return
	}
}

func (r *RoleResource) UpdateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	a := &rbacframe.Role{}
	if err = json.Unmarshal(body, a); err != nil {
		log.Error("json err", err)
		controller.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	// TODO: validate role
	err = dao.EditRole(context.TODO(), a)
	if err != nil {
		if err == datasource.ErrRoleDuplicated {
			controller.WriteError(w, ErrConflictRole, "")
			return
		}
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgOperateRoleFailed)
		return
	}
}

func (r *RoleResource) GetRole(w http.ResponseWriter, req *http.Request) {
	a, err := dao.GetRole(context.TODO(), req.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
	}
	v, err := json.Marshal(a)
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
	controller.WriteJSON(w, v)
}

func (r *RoleResource) DeleteRole(w http.ResponseWriter, req *http.Request) {
	_, err := dao.DeleteRole(context.TODO(), req.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
}

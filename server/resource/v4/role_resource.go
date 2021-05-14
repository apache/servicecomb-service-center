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
	"io/ioutil"
	"net/http"

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/datasource"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/go-chassis/cari/discovery"
)

var ErrConflictRole int32 = 409002

type RoleResource struct {
}

//URLPatterns define http pattern
func (r *RoleResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/role", Func: r.GetRolePermission},
		{Method: http.MethodPost, Path: "/v4/role", Func: r.CreateRolePermission},
		{Method: http.MethodPut, Path: "/v4/role/:roleName", Func: r.UpdateRolePermission},
		{Method: http.MethodGet, Path: "/v4/role/:roleName", Func: r.GetRole},
		{Method: http.MethodDelete, Path: "/v4/role/:roleName", Func: r.DeleteRole},
	}
}

//GetRolePermission list all roles and there's permissions
func (r *RoleResource) GetRolePermission(w http.ResponseWriter, req *http.Request) {
	rs, _, err := dao.ListRole(context.TODO())
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
		return
	}
	resp := &rbac.RoleResponse{
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

//roleParse parse the role info from the request body
func (r *RoleResource) roleParse(body []byte) (*rbac.Role, error) {
	role := &rbac.Role{}
	err := json.Unmarshal(body, role)
	if err != nil {
		log.Error("json err", err)
		return nil, err
	}
	// TODO: validate role
	return role, nil
}

//CreateRolePermission create new role and assign permissions
func (r *RoleResource) CreateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := r.roleParse(body)
	if err != nil {
		controller.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	err = dao.CreateRole(context.TODO(), role)
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

//UpdateRolePermission update role permissions
func (r *RoleResource) UpdateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := r.roleParse(body)
	if err != nil {
		controller.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	err = dao.EditRole(context.TODO(), role)
	if err != nil {
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgOperateRoleFailed)
		return
	}
}

//GetRole get the role info according to role name
func (r *RoleResource) GetRole(w http.ResponseWriter, req *http.Request) {
	role, err := dao.GetRole(context.TODO(), req.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
	}
	v, err := json.Marshal(role)
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
	controller.WriteJSON(w, v)
}

//DeleteRole delete the role info by role name
func (r *RoleResource) DeleteRole(w http.ResponseWriter, req *http.Request) {
	_, err := dao.DeleteRole(context.TODO(), req.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
}

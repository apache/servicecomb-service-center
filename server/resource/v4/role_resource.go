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

	"github.com/apache/servicecomb-service-center/datasource"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/rbac"
)

var ErrConflictRole int32 = 409002

type RoleResource struct {
}

//URLPatterns define http pattern
func (rr *RoleResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/roles", Func: rr.GetRolePermission},
		{Method: http.MethodPost, Path: "/v4/roles", Func: rr.CreateRolePermission},
		{Method: http.MethodPut, Path: "/v4/roles/:roleName", Func: rr.UpdateRolePermission},
		{Method: http.MethodGet, Path: "/v4/roles/:roleName", Func: rr.GetRole},
		{Method: http.MethodDelete, Path: "/v4/roles/:roleName", Func: rr.DeleteRole},
	}
}

//GetRolePermission list all roles and there's permissions
func (rr *RoleResource) GetRolePermission(w http.ResponseWriter, req *http.Request) {
	rs, _, err := dao.ListRole(context.TODO())
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
		return
	}
	resp := &rbac.RoleResponse{
		Roles: rs,
	}
	controller.WriteResponse(w, req, nil, resp)
}

//roleParse parse the role info from the request body
func (rr *RoleResource) roleParse(body []byte) (*rbac.Role, error) {
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
func (rr *RoleResource) CreateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := rr.roleParse(body)
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
	controller.WriteSuccess(w, req)
}

//UpdateRolePermission update role permissions
func (rr *RoleResource) UpdateRolePermission(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := rr.roleParse(body)
	if err != nil {
		controller.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	name := req.URL.Query().Get(":roleName")
	err = dao.EditRole(context.TODO(), name, role)
	if err != nil {
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgOperateRoleFailed)
		return
	}
	controller.WriteSuccess(w, req)
}

//GetRole get the role info according to role name
func (rr *RoleResource) GetRole(w http.ResponseWriter, r *http.Request) {
	role, err := dao.GetRole(context.TODO(), r.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
	}
	controller.WriteResponse(w, r, nil, role)
}

//DeleteRole delete the role info by role name
func (rr *RoleResource) DeleteRole(w http.ResponseWriter, req *http.Request) {
	_, err := dao.DeleteRole(context.TODO(), req.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		controller.WriteError(w, discovery.ErrInternal, errorsEx.MsgJSON)
		return
	}
	controller.WriteSuccess(w, req)
}

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

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/rbac"

	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
)

var ErrConflictRole int32 = 409002

type RoleResource struct {
}

// URLPatterns define http pattern
func (rr *RoleResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/roles", Func: rr.ListRoles},
		{Method: http.MethodPost, Path: "/v4/roles", Func: rr.CreateRole},
		{Method: http.MethodPut, Path: "/v4/roles/:roleName", Func: rr.UpdateRole},
		{Method: http.MethodGet, Path: "/v4/roles/:roleName", Func: rr.GetRole},
		{Method: http.MethodDelete, Path: "/v4/roles/:roleName", Func: rr.DeleteRole},
	}
}

// ListRoles list all roles and there's permissions
func (rr *RoleResource) ListRoles(w http.ResponseWriter, req *http.Request) {
	rs, num, err := rbacsvc.ListRole(req.Context())
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		rest.WriteError(w, discovery.ErrInternal, errorsEx.MsgGetRoleFailed)
		return
	}
	resp := &rbac.RoleResponse{
		Total: num,
		Roles: rs,
	}
	rest.WriteResponse(w, req, nil, resp)
}

// roleParse parse the role info from the request body
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

// CreateRole create new role and assign permissions
func (rr *RoleResource) CreateRole(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := rr.roleParse(body)
	if err != nil {
		rest.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}

	err = rbacsvc.CreateRole(req.Context(), role)
	if err != nil {
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, req, nil, nil)
}

// UpdateRole update role permissions
func (rr *RoleResource) UpdateRole(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get(":roleName")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	role, err := rr.roleParse(body)
	if err != nil {
		rest.WriteError(w, discovery.ErrInvalidParams, errorsEx.MsgJSON)
		return
	}
	err = rbacsvc.EditRole(req.Context(), name, role)
	if err != nil {
		log.Error(errorsEx.MsgOperateRoleFailed, err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, req, nil, nil)
}

// GetRole get the role info according to role name
func (rr *RoleResource) GetRole(w http.ResponseWriter, r *http.Request) {
	resp, err := rbacsvc.GetRole(r.Context(), r.URL.Query().Get(":roleName"))
	if err != nil {
		log.Error(errorsEx.MsgGetRoleFailed, err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, r, nil, resp)
}

// DeleteRole delete the role info by role name
func (rr *RoleResource) DeleteRole(w http.ResponseWriter, req *http.Request) {
	n := req.URL.Query().Get(":roleName")

	err := rbacsvc.DeleteRole(req.Context(), n)
	if err != nil {
		log.Error(errorsEx.MsgJSON, err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, req, nil, nil)
}

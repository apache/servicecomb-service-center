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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
)

var errModifySchemaDisabled = errors.New("schema modify is disabled")

type SchemaService struct {
	//
}

func (s *SchemaService) URLPatterns() []rest.Route {
	var r = []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.GetSchemas},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.DeleteSchemas},
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/schemas", Func: s.ModifySchemas},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/schemas", Func: s.GetAllSchemas},
	}

	if !config.GetRegistry().SchemaDisable {
		r = append(r, rest.Route{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.ModifySchema})
	} else {
		r = append(r, rest.Route{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.DisableSchema})
	}

	return r
}

func (s *SchemaService) DisableSchema(w http.ResponseWriter, r *http.Request) {
	rest.WriteError(w, pb.ErrForbidden, errModifySchemaDisabled.Error())
}

func (s *SchemaService) GetSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.GetSchemaInfo(r.Context(), request)
	w.Header().Add("X-Schema-Summary", resp.SchemaSummary)
	resp.SchemaSummary = ""
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *SchemaService) ModifySchema(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.ModifySchemaRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	request.ServiceId = query.Get(":serviceId")
	request.SchemaId = query.Get(":schemaId")
	resp, err := core.ServiceAPI.ModifySchema(r.Context(), request)
	if err != nil {
		log.Error("can not update schema", err)
		rest.WriteError(w, pb.ErrInternal, "can not update schema")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *SchemaService) ModifySchemas(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	serviceID := r.URL.Query().Get(":serviceId")
	request := &pb.ModifySchemasRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request.ServiceId = serviceID
	resp, err := core.ServiceAPI.ModifySchemas(r.Context(), request)
	if err != nil {
		log.Error("can not update schema", err)
		rest.WriteError(w, pb.ErrInternal, "can not update schema")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *SchemaService) DeleteSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.DeleteSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.DeleteSchema(r.Context(), request)
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *SchemaService) GetAllSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	withSchema := query.Get("withSchema")
	serviceID := query.Get(":serviceId")
	if withSchema != "0" && withSchema != "1" && strings.TrimSpace(withSchema) != "" {
		rest.WriteError(w, pb.ErrInvalidParams, "parameter withSchema must be 1 or 0")
		return
	}
	request := &pb.GetAllSchemaRequest{
		ServiceId:  serviceID,
		WithSchema: withSchema == "1",
	}
	resp, _ := core.ServiceAPI.GetAllSchemaInfo(r.Context(), request)
	rest.WriteResponse(w, r, resp.Response, resp)
}

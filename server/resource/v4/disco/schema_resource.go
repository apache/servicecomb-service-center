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

package disco

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

var errModifySchemaDisabled = errors.New("schema modify is disabled")

type SchemaResource struct {
	//
}

func (s *SchemaResource) URLPatterns() []rest.Route {
	var r = []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.GetSchema},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.DeleteSchema},
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/schemas", Func: s.PutSchemas},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/schemas", Func: s.ListSchema},
	}

	if !config.GetRegistry().SchemaDisable {
		r = append(r, rest.Route{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.PutSchema})
	} else {
		r = append(r, rest.Route{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", Func: s.DisableSchema})
	}

	return r
}

func (s *SchemaResource) DisableSchema(w http.ResponseWriter, r *http.Request) {
	rest.WriteError(w, pb.ErrForbidden, errModifySchemaDisabled.Error())
}

func (s *SchemaResource) GetSchema(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	schema, err := discosvc.GetSchema(r.Context(), request)
	if err != nil {
		log.Error("get schema failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	w.Header().Add("X-Schema-Summary", schema.Summary)
	schema.SchemaId = ""
	schema.Summary = ""
	rest.WriteResponse(w, r, nil, schema)
}

func (s *SchemaResource) PutSchema(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
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
	svcErr := discosvc.PutSchema(r.Context(), request)
	if svcErr != nil {
		log.Error("put schema failed", svcErr)
		rest.WriteServiceError(w, svcErr)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *SchemaResource) PutSchemas(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
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
	svcErr := discosvc.PutSchemas(r.Context(), request)
	if svcErr != nil {
		log.Error("put all schemas failed", svcErr)
		rest.WriteServiceError(w, svcErr)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *SchemaResource) DeleteSchema(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.DeleteSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	err := discosvc.DeleteSchema(r.Context(), request)
	if err != nil {
		log.Error("delete schema failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *SchemaResource) ListSchema(w http.ResponseWriter, r *http.Request) {
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
	schemas, err := discosvc.ListSchema(r.Context(), request)
	if err != nil {
		log.Error("list schema failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &pb.GetAllSchemaResponse{
		Schemas: schemas,
	})
}

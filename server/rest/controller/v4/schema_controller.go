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
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/rest/controller"
	"io/ioutil"
	"net/http"
	"strings"
)

type SchemaService struct {
	//
}

func (this *SchemaService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", this.GetSchemas},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", this.ModifySchema},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId", this.DeleteSchemas},
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/microservices/:serviceId/schemas", this.ModifySchemas},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/schemas", this.GetAllSchemas},
	}
}

func (this *SchemaService) GetSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.GetSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.GetSchemaInfo(r.Context(), request)
	w.Header().Add("X-Schema-Summary", resp.SchemaSummary)
	resp.SchemaSummary = ""
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *SchemaService) ModifySchema(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.ModifySchemaRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "Invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	request.ServiceId = query.Get(":serviceId")
	request.SchemaId = query.Get(":schemaId")
	resp, err := core.ServiceAPI.ModifySchema(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) ModifySchemas(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	serviceId := r.URL.Query().Get(":serviceId")
	request := &pb.ModifySchemasRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Errorf(err, "Invalid json: %s", util.BytesToStringWithNoCopy(message))
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request.ServiceId = serviceId
	resp, err := core.ServiceAPI.ModifySchemas(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) DeleteSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &pb.DeleteSchemaRequest{
		ServiceId: query.Get(":serviceId"),
		SchemaId:  query.Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.DeleteSchema(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) GetAllSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	withSchema := query.Get("withSchema")
	serviceId := query.Get(":serviceId")
	if withSchema != "0" && withSchema != "1" && strings.TrimSpace(withSchema) != "" {
		controller.WriteError(w, scerr.ErrInvalidParams, "parameter withSchema must be 1 or 0")
		return
	}
	request := &pb.GetAllSchemaRequest{
		ServiceId:  serviceId,
		WithSchema: withSchema == "1",
	}
	resp, _ := core.ServiceAPI.GetAllSchemaInfo(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

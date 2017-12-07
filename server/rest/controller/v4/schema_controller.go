//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package v4

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"io/ioutil"
	"net/http"
	"strings"
)

type SchemaService struct {
	//
}

func (this *SchemaService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/microservices/:serviceId/schemas/:schemaId", this.GetSchemas},
		{rest.HTTP_METHOD_PUT, "/v4/:domain/registry/microservices/:serviceId/schemas/:schemaId", this.ModifySchema},
		{rest.HTTP_METHOD_DELETE, "/v4/:domain/registry/microservices/:serviceId/schemas/:schemaId", this.DeleteSchemas},
		{rest.HTTP_METHOD_POST, "/v4/:domain/registry/microservices/:serviceId/schemas", this.ModifySchemas},
		{rest.HTTP_METHOD_GET, "/v4/:domain/registry/microservices/:serviceId/schemas", this.GetAllSchemas},
	}
}

func (this *SchemaService) GetSchemas(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetSchemaRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
		SchemaId:  r.URL.Query().Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.GetSchemaInfo(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, respInternal, resp)
}

func (this *SchemaService) ModifySchema(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.ModifySchemaRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request.ServiceId = r.URL.Query().Get(":serviceId")
	request.SchemaId = r.URL.Query().Get(":schemaId")
	resp, err := core.ServiceAPI.ModifySchema(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) ModifySchemas(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.Logger().Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	serviceId := r.URL.Query().Get(":serviceId")
	request := &pb.ModifySchemasRequest{}
	err = json.Unmarshal(message, request)
	if err != nil {
		util.Logger().Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request.ServiceId = serviceId
	resp, err := core.ServiceAPI.ModifySchemas(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) DeleteSchemas(w http.ResponseWriter, r *http.Request) {
	request := &pb.DeleteSchemaRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
		SchemaId:  r.URL.Query().Get(":schemaId"),
	}
	resp, _ := core.ServiceAPI.DeleteSchema(r.Context(), request)
	controller.WriteResponse(w, resp.Response, nil)
}

func (this *SchemaService) GetAllSchemas(w http.ResponseWriter, r *http.Request) {
	withSchema := r.URL.Query().Get("withSchema")
	serviceId := r.URL.Query().Get(":serviceId")
	util.Logger().Warnf(nil, "Query Service %s schema, withSchema is %s.", serviceId, withSchema)
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

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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
)

var trueOrFalse = map[string]bool{"true": true, "false": false, "1": true, "0": false}

type ServiceResource struct {
	//
}

func (s *ServiceResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/existence", Func: s.ResourceExist},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices", Func: s.ListService},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.GetService},
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices", Func: s.RegisterService},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/properties", Func: s.PutServiceProperties},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId", Func: s.UnregisterService},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices", Func: s.UnregisterManyService},
		// tags
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/tags", Func: s.PutManyTags},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/tags/:key", Func: s.PutTag},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/tags", Func: s.ListTag},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/tags/:key", Func: s.DeleteManyTags},
	}
}

func (s *ServiceResource) RegisterService(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	var request pb.CreateServiceRequest
	err = json.Unmarshal(message, &request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	resp, err := discosvc.RegisterService(r.Context(), &request)
	if err != nil {
		log.Error("create service failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *ServiceResource) PutServiceProperties(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request := &pb.UpdateServicePropsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	err = discosvc.PutServiceProperties(r.Context(), request)
	if err != nil {
		log.Error("can not update service", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *ServiceResource) UnregisterService(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	serviceID := query.Get(":serviceId")
	force := query.Get("force")

	b, ok := trueOrFalse[force]
	if force != "" && !ok {
		rest.WriteError(w, pb.ErrInvalidParams, "parameter force must be false or true")
		return
	}

	request := &pb.DeleteServiceRequest{
		ServiceId: serviceID,
		Force:     b,
	}
	err := discosvc.UnregisterService(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("delete service[%s] failed", serviceID), err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *ServiceResource) ListService(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServicesRequest{
		WithShared: util.StringTRUE(r.URL.Query().Get("withShared")),
	}
	resp, err := discosvc.ListService(r.Context(), request)
	if err != nil {
		log.Error("list services failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *ServiceResource) ResourceExist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	checkType := query.Get("type")
	request := &pb.GetExistenceRequest{
		Type:        checkType,
		Environment: query.Get("env"),
		AppId:       query.Get("appId"),
		ServiceName: query.Get("serviceName"),
		Version:     query.Get("version"),
		ServiceId:   query.Get("serviceId"),
		SchemaId:    query.Get("schemaId"),
	}

	resp, err := resourceExist(ctx, w, request)
	if err != nil {
		log.Error("check resource existence failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func resourceExist(ctx context.Context, w http.ResponseWriter, request *pb.GetExistenceRequest) (*pb.GetExistenceResponse, error) {
	switch request.Type {
	case datasource.ExistTypeMicroservice:
		serviceID, err := discosvc.ExistService(ctx, request)
		if err != nil {
			return nil, err
		}
		return &pb.GetExistenceResponse{ServiceId: serviceID}, nil
	case datasource.ExistTypeSchema:
		schema, err := discosvc.ExistSchema(ctx, &pb.GetSchemaRequest{
			ServiceId: request.ServiceId,
			SchemaId:  request.SchemaId,
		})
		if err != nil {
			return nil, err
		}
		w.Header().Add("X-Schema-Summary", schema.Summary)
		return &pb.GetExistenceResponse{
			ServiceId: request.ServiceId,
			SchemaId:  schema.SchemaId,
		}, nil
	default:
		log.Warn(fmt.Sprintf("unexpected type '%s' for existence query.", request.Type))
		return nil, pb.NewError(pb.ErrInvalidParams, "Only micro-service and schema can be used as type.")
	}
}

func (s *ServiceResource) GetService(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetServiceRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	}
	service, err := discosvc.GetService(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed", request.ServiceId), err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &pb.GetServiceResponse{Service: service})
}

func (s *ServiceResource) UnregisterManyService(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	request := &pb.DelServicesRequest{}

	err = json.Unmarshal(message, request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	resp, err := discosvc.UnregisterManyService(r.Context(), request)
	if err != nil {
		log.Error("delete services failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *ServiceResource) PutManyTags(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	var tags map[string]map[string]string
	err = json.Unmarshal(message, &tags)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}

	err = discosvc.PutManyTags(r.Context(), &pb.AddServiceTagsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
		Tags:      tags["tags"],
	})
	if err != nil {
		log.Error("can not add tag", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *ServiceResource) PutTag(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	err := discosvc.PutTag(r.Context(), &pb.UpdateServiceTagRequest{
		ServiceId: query.Get(":serviceId"),
		Key:       query.Get(":key"),
		Value:     query.Get("value"),
	})
	if err != nil {
		log.Error("can not update tag", err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, r, nil, nil)
}

func (s *ServiceResource) ListTag(w http.ResponseWriter, r *http.Request) {
	resp, err := discosvc.ListTag(r.Context(), &pb.GetServiceTagsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	})
	if err != nil {
		log.Error("can not list tag", err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, r, nil, resp)
}

func (s *ServiceResource) DeleteManyTags(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	keys := query.Get(":key")
	ids := strings.Split(keys, ",")

	err := discosvc.DeleteManyTags(r.Context(), &pb.DeleteServiceTagsRequest{
		ServiceId: query.Get(":serviceId"),
		Keys:      ids,
	})
	if err != nil {
		log.Error("can not delete many tag", err)
		rest.WriteServiceError(w, err)
		return
	}

	rest.WriteResponse(w, r, nil, nil)
}

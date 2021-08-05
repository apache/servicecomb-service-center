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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
)

type TagService struct {
	//
}

func (s *TagService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodPost, Path: "/v4/:project/registry/microservices/:serviceId/tags", Func: s.AddTags},
		{Method: http.MethodPut, Path: "/v4/:project/registry/microservices/:serviceId/tags/:key", Func: s.UpdateTag},
		{Method: http.MethodGet, Path: "/v4/:project/registry/microservices/:serviceId/tags", Func: s.GetTags},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/microservices/:serviceId/tags/:key", Func: s.DeleteTags},
	}
}

func (s *TagService) AddTags(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
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

	resp, err := core.ServiceAPI.AddTags(r.Context(), &pb.AddServiceTagsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
		Tags:      tags["tags"],
	})
	if err != nil {
		log.Error("can not add tag", err)
		rest.WriteError(w, pb.ErrInternal, "can not add tag")
		return
	}
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *TagService) UpdateTag(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	resp, _ := core.ServiceAPI.UpdateTag(r.Context(), &pb.UpdateServiceTagRequest{
		ServiceId: query.Get(":serviceId"),
		Key:       query.Get(":key"),
		Value:     query.Get("value"),
	})
	rest.WriteResponse(w, r, resp.Response, nil)
}

func (s *TagService) GetTags(w http.ResponseWriter, r *http.Request) {
	resp, _ := core.ServiceAPI.GetTags(r.Context(), &pb.GetServiceTagsRequest{
		ServiceId: r.URL.Query().Get(":serviceId"),
	})
	rest.WriteResponse(w, r, resp.Response, resp)
}

func (s *TagService) DeleteTags(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	keys := query.Get(":key")
	ids := strings.Split(keys, ",")

	resp, _ := core.ServiceAPI.DeleteTags(r.Context(), &pb.DeleteServiceTagsRequest{
		ServiceId: query.Get(":serviceId"),
		Keys:      ids,
	})
	rest.WriteResponse(w, r, resp.Response, nil)
}

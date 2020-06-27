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
package controller

import (
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/scerror"
	"net/http"
	"strconv"
)

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := scerror.NewError(code, detail)
	w.Header().Set(rest.HEADER_RESPONSE_STATUS, strconv.Itoa(err.StatusCode()))
	w.Header().Set(rest.HEADER_CONTENT_TYPE, rest.CONTENT_TYPE_JSON)
	w.WriteHeader(err.StatusCode())
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(err.Marshal()))

	if err.InternalError() {
		alarm.Raise(alarm.IdInternalError, alarm.AdditionalContext(detail))
	}
}

func WriteResponse(w http.ResponseWriter, resp *pb.Response, obj interface{}) {
	if resp != nil && resp.GetCode() != pb.Response_SUCCESS {
		WriteError(w, resp.GetCode(), resp.GetMessage())
		return
	}

	if obj == nil {
		w.Header().Set(rest.HEADER_RESPONSE_STATUS, strconv.Itoa(http.StatusOK))
		w.Header().Set(rest.HEADER_CONTENT_TYPE, rest.CONTENT_TYPE_TEXT)
		w.WriteHeader(http.StatusOK)
		return
	}

	objJson, err := json.Marshal(obj)
	if err != nil {
		WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	w.Header().Set(rest.HEADER_RESPONSE_STATUS, strconv.Itoa(http.StatusOK))
	w.Header().Set(rest.HEADER_CONTENT_TYPE, rest.CONTENT_TYPE_JSON)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(objJson))
}

func WriteJsonIfSuccess(w http.ResponseWriter, resp *pb.Response, json []byte) {
	if resp.GetCode() == pb.Response_SUCCESS {
		WriteJson(w, json)
		return
	}
	WriteError(w, resp.GetCode(), resp.GetMessage())
}

//WriteJson simply write json
func WriteJson(w http.ResponseWriter, json []byte) {
	w.Header().Set(rest.HEADER_RESPONSE_STATUS, strconv.Itoa(http.StatusOK))
	w.Header().Set(rest.HEADER_CONTENT_TYPE, rest.CONTENT_TYPE_JSON)
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

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
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/scerror"
	"net/http"
	"strconv"
)

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := scerror.NewError(code, detail)
	w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(err.StatusCode()))
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
	w.WriteHeader(err.StatusCode())
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(err.Marshal()))

	if err.InternalError() {
		err := alarm.Raise(alarm.IDInternalError, alarm.AdditionalContext(detail))
		if err != nil {
			log.Error("", err)
		}
	}
}

func WriteResponse(w http.ResponseWriter, resp *pb.Response, obj interface{}) {
	if resp != nil && resp.GetCode() != proto.ResponseSuccess {
		WriteError(w, resp.GetCode(), resp.GetMessage())
		return
	}

	if obj == nil {
		w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(http.StatusOK))
		w.Header().Set(rest.HeaderContentType, rest.ContentTypeText)
		w.WriteHeader(http.StatusOK)
		return
	}

	b, err := json.Marshal(obj)
	if err != nil {
		WriteError(w, scerror.ErrInternal, err.Error())
		return
	}
	w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(http.StatusOK))
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(b))
}

func WriteJSONIfSuccess(w http.ResponseWriter, resp *pb.Response, json []byte) {
	if resp.GetCode() == proto.ResponseSuccess {
		WriteJSON(w, json)
		return
	}
	WriteError(w, resp.GetCode(), resp.GetMessage())
}

//WriteJSON simply write json
func WriteJSON(w http.ResponseWriter, json []byte) {
	w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(http.StatusOK))
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(json)
	if err != nil {
		log.Error("", err)
	}
}

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
	"net/http"
	"strconv"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	"github.com/go-chassis/cari/discovery"
)

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := discovery.NewError(code, detail)
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

// WriteResponse writes http response
// If the resp is nil or represents success, response status is http.StatusOK,
// response content is obj.
// If the resp represents fail, response status is from the code in the
// resp, response content is from the message in the resp.
func WriteResponse(w http.ResponseWriter, r *http.Request, resp *discovery.Response, obj interface{}) {
	if resp != nil && resp.GetCode() != discovery.ResponseSuccess {
		WriteError(w, resp.GetCode(), resp.GetMessage())
		return
	}

	if obj == nil {
		w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(http.StatusOK))
		w.Header().Set(rest.HeaderContentType, rest.ContentTypeText)
		w.WriteHeader(http.StatusOK)
		return
	}

	util.SetRequestContext(r, rest.CtxResponseObject, obj)

	var (
		data []byte
		err  error
	)
	switch body := obj.(type) {
	case []byte:
		data = body
	default:
		data, err = json.Marshal(body)
		if err != nil {
			WriteError(w, discovery.ErrInternal, err.Error())
			return
		}
	}
	w.Header().Set(rest.HeaderResponseStatus, strconv.Itoa(http.StatusOK))
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintln(w, util.BytesToStringWithNoCopy(data))
	if err != nil {
		log.Error("write response failed", err)
	}
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

func WriteSuccess(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, r, nil, nil)
}

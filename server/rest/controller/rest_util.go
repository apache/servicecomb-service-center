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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/error"
	"net/http"
)

const (
	contentTypeJson = "application/json; charset=UTF-8"
	contentTypeText = "text/plain; charset=UTF-8"
)

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := error.NewError(code, detail)
	w.Header().Add("X-Response-Status", fmt.Sprint(err.StatusCode()))
	w.Header().Set("Content-Type", contentTypeJson)
	w.WriteHeader(err.StatusCode())
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(err.Marshal()))
}

func WriteJsonObject(w http.ResponseWriter, obj interface{}) {
	if obj == nil {
		w.Header().Add("X-Response-Status", fmt.Sprint(http.StatusOK))
		w.Header().Set("Content-Type", contentTypeText)
		w.WriteHeader(http.StatusOK)
		return
	}

	objJson, err := json.Marshal(obj)
	if err != nil {
		WriteError(w, error.ErrInternal, err.Error())
		return
	}
	w.Header().Add("X-Response-Status", fmt.Sprint(http.StatusOK))
	w.Header().Set("Content-Type", contentTypeJson)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, util.BytesToStringWithNoCopy(objJson))
}

func WriteResponse(w http.ResponseWriter, resp *pb.Response, obj interface{}) {
	if resp.GetCode() == pb.Response_SUCCESS {
		WriteJsonObject(w, obj)
		return
	}

	WriteError(w, resp.GetCode(), resp.GetMessage())
}

func WriteBytes(w http.ResponseWriter, resp *pb.Response, json []byte) {
	if resp.GetCode() == pb.Response_SUCCESS {
		w.Header().Add("X-Response-Status", fmt.Sprint(http.StatusOK))
		w.Header().Set("Content-Type", contentTypeJson)
		w.WriteHeader(http.StatusOK)
		w.Write(json)
		return
	}
	WriteError(w, resp.GetCode(), resp.GetMessage())
}

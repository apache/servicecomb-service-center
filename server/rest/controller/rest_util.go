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
	"github.com/ServiceComb/service-center/pkg/util"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/error"
	"net/http"
)

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := error.NewError(code, detail)
	err.HttpWrite(w)
}

func WriteJsonObject(w http.ResponseWriter, obj interface{}) {
	if obj == nil {
		w.Header().Add("X-Response-Status", fmt.Sprint(http.StatusOK))
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	objJson, err := json.Marshal(obj)
	if err != nil {
		WriteError(w, error.ErrInternal, err.Error())
		return
	}
	w.Header().Add("X-Response-Status", fmt.Sprint(http.StatusOK))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
		return
	}
	WriteError(w, resp.GetCode(), resp.GetMessage())
}


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
package rest

import (
	"encoding/json"
	"fmt"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/util"
	"net/http"
	"strconv"
)

var ServiceAPI pb.ServiceCtrlServer
var InstanceAPI pb.SerivceInstanceCtrlServerEx
var GovernServiceAPI pb.GovernServiceCtrlServerEx

func WriteJsonObject(status int, obj interface{}, w http.ResponseWriter) {
	serviceJSON, err := json.Marshal(obj)
	if err != nil {
		util.Logger().Error("marshal response error", err)
		WriteText(http.StatusInternalServerError, fmt.Sprintf("marshal response error, %s", err.Error()), w)
		return
	}
	WriteJson(status, serviceJSON, w)
}

func WriteJson(status int, json []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Add("x-response-status", strconv.Itoa(status))
	w.WriteHeader(status)
	w.Write(json)
}

func WriteText(status int, text string, w http.ResponseWriter) {
	w.Header().Add("x-response-status", strconv.Itoa(status))
	w.WriteHeader(status)
	w.Write(util.StringToBytesWithNoCopy(text))
}

func WriteTextResponse(resp *pb.Response, err error, textIfSuccess string, w http.ResponseWriter) {
	if err != nil {
		WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	if resp.Code != pb.Response_SUCCESS {
		WriteText(http.StatusBadRequest, resp.Message, w)
		return
	}
	WriteText(http.StatusOK, textIfSuccess, w)
}

func WriteJsonResponse(resp *pb.Response, obj interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	if resp.Code != pb.Response_SUCCESS {
		WriteText(http.StatusBadRequest, resp.Message, w)
		return
	}
	objJson, err := json.Marshal(obj)
	if err != nil {
		WriteText(http.StatusInternalServerError, err.Error(), w)
		return
	}
	WriteJson(http.StatusOK, objJson, w)
}

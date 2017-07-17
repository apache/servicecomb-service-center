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
	"github.com/astaxie/beego"
	"github.com/servicecomb/service-center/util"
	"github.com/servicecomb/service-center/util/rest"
	"net/http"
)

type Version struct {
	Api_version string `json:"api_version"`
	Build_tag   string `json:"build_tag"`
}

type Result struct {
	Info   string `json:"info" description:"return info"`
	Status int    `json:"status" description:"http return code"`
}

type MainService struct {
	//
}

func (this *MainService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/version", this.GetVersion},
		{rest.HTTP_METHOD_GET, "/health", this.CluterHealth},
	}
}

func (this *MainService) CluterHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := InstanceAPI.CluterHealth(r.Context())
	if err != nil {
		util.LOGGER.Error("health check failed", err)
		WriteText(http.StatusInternalServerError, "health check failed", w)
		return
	}

	respInternal := resp.Response
	resp.Response = nil
	WriteJsonResponse(respInternal, resp, err, w)
}

func (this *MainService) GetVersion(w http.ResponseWriter, r *http.Request) {
	api_version := beego.AppConfig.String("version")
	build_tag := beego.AppConfig.String("build_tag")
	version := Version{api_version, build_tag}
	versionJSON, _ := json.Marshal(version)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write([]byte(versionJSON))
}

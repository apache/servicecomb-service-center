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
package version

import "github.com/astaxie/beego"

const VERSION = "0.1.1"
const API_VERSION = "3.0.0"

type VersionSet struct {
	Version    string `json:"version"`
	ApiVersion string `json:"apiVersion"`
	BuildTag   string `json:"buildTag"`
}

var version VersionSet

func init() {
	version.Version = VERSION
	version.ApiVersion = API_VERSION
	version.BuildTag = beego.AppConfig.String("build_tag")
}

func Ver() VersionSet {
	return version
}

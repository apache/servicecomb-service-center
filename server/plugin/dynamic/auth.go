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
package dynamic

import (
	"github.com/ServiceComb/service-center/pkg/plugin"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/infra/auth"
	"github.com/astaxie/beego"
	"net/http"
)

var authLib auth.Auth

func init() {
	name := beego.AppConfig.String("auth_plugin")
	if pluginBuilder, ok := auth.AuthPlugins[name]; ok {
		util.Logger().Warnf(nil, "static load plugin '%s' successfully.", name)
		authLib = pluginBuilder()
		return
	}
}

func buildinAuthFunc(r *http.Request) error {
	if authLib == nil {
		return nil
	}
	return authLib.Identify(r)
}

func findAuthFunc(funcName string) func(*http.Request) error {
	ff, err := plugin.FindFunc("auth", funcName)
	if err != nil {
		return buildinAuthFunc
	}
	f, ok := ff.(func(*http.Request) error)
	if !ok {
		util.Logger().Warnf(nil, "unexpected function '%s' format found in plugin 'auth'.", funcName)
		return buildinAuthFunc
	}
	return f
}

func Identify(r *http.Request) error {
	f := findAuthFunc("Identify")
	return f(r)
}

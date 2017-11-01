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
package auth

import (
	"github.com/ServiceComb/service-center/pkg/chain"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/infra/auth"
	"github.com/astaxie/beego"
	"net/http"
)

var plugin auth.Auth

func init() {
	name := beego.AppConfig.String("auth_plugin")
	if pluginBuilder, ok := auth.AuthPlugins[name]; ok {
		util.Logger().Warnf(nil, "service center is in '%s' mode", name)
		plugin = pluginBuilder()
		return
	}
	util.Logger().Warnf(nil, "service center is in 'noAuth' mode")
}

type AuthRequest struct {
}

func (h *AuthRequest) Handle(i *chain.Invocation) {
	if plugin == nil {
		i.Next()
		return
	}
	err := plugin.Identify(i.Context().Value(rest.CTX_REQUEST).(*http.Request))
	if err == nil {
		i.Next()
		return
	}

	w := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter)
	http.Error(w, "Request Unauthorized", http.StatusUnauthorized)
	i.Fail(nil)
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &AuthRequest{})
}

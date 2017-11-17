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
	scerr "github.com/ServiceComb/service-center/server/error"
	"github.com/ServiceComb/service-center/server/plugin/dynamic"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"net/http"
)

type AuthRequest struct {
}

func (h *AuthRequest) Handle(i *chain.Invocation) {
	r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	err := dynamic.Identify(r)
	if err == nil {
		i.Next()
		return
	}

	util.Logger().Errorf(err, "authenticate request failed, %s %s", r.Method, r.RequestURI)

	w := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter)
	controller.WriteError(w, scerr.ErrUnauthorized, err.Error())

	i.Fail(nil)
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &AuthRequest{})
}

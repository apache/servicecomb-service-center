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
package context

import (
	"github.com/ServiceComb/service-center/pkg/chain"
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"net/http"
)

type ContextHandler struct {
}

func (c *ContextHandler) Handle(i *chain.Invocation) {
	pattern := i.Context().Value(roa.CTX_MATCH_PATTERN).(string)
	if IsSkip(pattern) {
		i.Next()
		return
	}

	var (
		err error
		v3  v3Context
		v4  v4Context
	)

	r := i.Context().Value(roa.CTX_REQUEST).(*http.Request)
	switch {
	case v3.IsMatch(r):
		err = v3.Do(r)
	case v4.IsMatch(r):
		err = v4.Do(r)
	}

	if err != nil {
		i.Fail(err)
		return
	}

	i.WithContext("x-remote-ip", util.GetRealIP(r))

	i.Next()
}

func IsSkip(url string) bool {
	l, vl, hl := len(url), len("/version"), len("/health")
	if l > vl && url[l-vl:] == "/version" {
		return true
	}
	if l > hl && url[l-hl:] == "/health" {
		return true
	}
	return false
}

func RegisterHandlers() {
	chain.RegisterHandler(roa.SERVER_CHAIN_NAME, &ContextHandler{})
}

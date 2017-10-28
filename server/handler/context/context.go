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
	"net/http"
)

type ContextHandler struct {
}

func (c *ContextHandler) Handle(i *chain.Invocation) {
	var (
		err error
		v3  v3Context
		v4  v4Context
	)

	ctx := i.HandlerContext()
	r := ctx[roa.CTX_REQUEST].(*http.Request)

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
	i.Next()
}

func RegisterHandlers() {
	chain.RegisterHandler(roa.SERVER_CHAIN_NAME, &ContextHandler{})
}

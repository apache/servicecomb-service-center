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
package perf

import (
	"github.com/ServiceComb/service-center/pkg/chain"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	mr "github.com/ServiceComb/service-center/server/rest"
	"net/http"
	"time"
)

type LatencyStat struct {
}

func (l *LatencyStat) Handle(i *chain.Invocation) {
	i.WithContext("x-start-timestamp", time.Now())

	cb := i.Func
	i.Invoke(func(ret chain.Result) {
		cb(ret)

		r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
		start := i.Context().Value("x-start-timestamp").(time.Time)

		util.LogNilOrWarnf(start, "%s %s", r.Method, r.RequestURI)

		if r.Method != "WEBSOCKET" {
			mr.ReportRequestCompleted(i, start)
		}
	})
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &LatencyStat{})
}

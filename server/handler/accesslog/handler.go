// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build go1.9

package accesslog

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	svr "github.com/apache/servicecomb-service-center/server/rest"
)

// Handler implements chain.Handler
// Handler records access log
type Handler struct {
	logger        *log.Logger
	whiteListAPIs map[string]struct{} // not record access log
}

// Handle handles the request
func (l *Handler) Handle(i *chain.Invocation) {
	matchPattern := i.Context().Value(rest.CTX_MATCH_PATTERN).(string)
	if _, ok := l.whiteListAPIs[matchPattern]; ok {
		i.Next()
		return
	}
	startTimeStr := "unknown"
	start, ok := i.Context().Value(svr.CTX_START_TIMESTAMP).(time.Time)
	if ok {
		startTimeStr = start.Format("2006-01-02T15:04:05.000Z07:00")
	}
	r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	w := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter)
	i.Next(chain.WithAsyncFunc(func(_ chain.Result) {
		delayByMillisecond := "unknown"
		if ok {
			delayByMillisecond = fmt.Sprintf("%d", time.Since(start)/time.Millisecond)
		}
		statusCode := w.Header().Get(rest.HEADER_RESPONSE_STATUS)
		// format:  remoteIp requestReceiveTime "method requestUri proto" statusCode requestBodySize delay(ms)
		// example: 127.0.0.1 2006-01-02T15:04:05.000Z07:00 "GET /v4/default/registry/microservices HTTP/1.1" 200 0 0
		l.logger.Infof("%s %s \"%s %s %s\" %s %d %s",
			util.GetIPFromContext(i.Context()),
			startTimeStr,
			r.Method,
			r.RequestURI,
			r.Proto,
			statusCode,
			r.ContentLength,
			delayByMillisecond)
	}))
}

// NewAccessLogHandler creates a Handler
func NewAccessLogHandler(l *log.Logger, m map[string]struct{}) *Handler {
	return &Handler{
		logger:        l,
		whiteListAPIs: m}
}

// RegisterHandlers registers an access log handler to the handler chain
func RegisterHandlers() {
	if !core.ServerInfo.Config.EnableAccessLog {
		return
	}
	logger := log.NewLogger(log.Config{
		LoggerFile:     os.ExpandEnv(core.ServerInfo.Config.AccessLogFile),
		LogFormatText:  true,
		LogRotateSize:  int(core.ServerInfo.Config.LogRotateSize),
		LogBackupCount: int(core.ServerInfo.Config.LogBackupCount),
	})
	whiteListAPIs := make(map[string]struct{}, 0)
	// no access log for heartbeat
	whiteListAPIs["/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat"] = struct{}{}
	h := NewAccessLogHandler(logger, whiteListAPIs)
	chain.RegisterHandler(rest.ServerChainName, h)
}

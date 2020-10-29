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
	"github.com/apache/servicecomb-service-center/server/config"
	"net/http"
	"os"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	svr "github.com/apache/servicecomb-service-center/server/rest"
)

// Handler implements chain.Handler
// Handler records access log.
// Make sure to complete the initialization before handling the request.
type Handler struct {
	logger        *log.Logger
	whiteListAPIs map[string]struct{} // not record access log
}

// AddWhiteListAPIs adds APIs to white list, where the APIs will be ignored
// in access log.
// Not safe for concurrent use.
func (h *Handler) AddWhiteListAPIs(apis ...string) {
	for _, api := range apis {
		h.whiteListAPIs[api] = struct{}{}
	}
}

// ShouldIgnoreAPI judges whether the API should be ignored in access log.
func (h *Handler) ShouldIgnoreAPI(api string) bool {
	_, ok := h.whiteListAPIs[api]
	return ok
}

// Handle handles the request
func (h *Handler) Handle(i *chain.Invocation) {
	matchPattern := i.Context().Value(rest.CtxMatchPattern).(string)
	if h.ShouldIgnoreAPI(matchPattern) {
		i.Next()
		return
	}
	startTimeStr := "unknown"
	start, ok := i.Context().Value(svr.CtxStartTimestamp).(time.Time)
	if ok {
		startTimeStr = start.Format("2006-01-02T15:04:05.000Z07:00")
	}
	r := i.Context().Value(rest.CtxRequest).(*http.Request)
	w := i.Context().Value(rest.CtxResponse).(http.ResponseWriter)
	i.Next(chain.WithAsyncFunc(func(_ chain.Result) {
		delayByMillisecond := "unknown"
		if ok {
			delayByMillisecond = fmt.Sprintf("%d", time.Since(start)/time.Millisecond)
		}
		statusCode := w.Header().Get(rest.HeaderResponseStatus)
		// format:  remoteIp requestReceiveTime "method requestUri proto" statusCode requestBodySize delay(ms)
		// example: 127.0.0.1 2006-01-02T15:04:05.000Z07:00 "GET /v4/default/registry/microservices HTTP/1.1" 200 0 0
		h.logger.Infof("%s %s \"%s %s %s\" %s %d %s",
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
func NewAccessLogHandler(l *log.Logger) *Handler {
	return &Handler{
		logger:        l,
		whiteListAPIs: make(map[string]struct{})}
}

// RegisterHandlers registers an access log handler to the handler chain
func RegisterHandlers() {
	if !config.GetLog().EnableAccessLog {
		return
	}
	logger := log.NewLogger(log.Config{
		LoggerFile:     os.ExpandEnv(config.GetLog().AccessLogFile),
		LogFormatText:  true,
		LogRotateSize:  int(config.GetLog().LogRotateSize),
		LogBackupCount: int(config.GetLog().LogBackupCount),
		NoCaller:       true,
		NoTime:         true,
		NoLevel:        true,
	})
	h := NewAccessLogHandler(logger)
	// no access log for heartbeat
	h.AddWhiteListAPIs(
		"/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat",
		"/v4/:project/registry/heartbeats",
		"/registry/v3/microservices/:serviceId/instances/:instanceId/heartbeat",
		"/registry/v3/heartbeats",
		"")
	chain.RegisterHandler(rest.ServerChainName, h)
}

/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package exception

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
)

var whitelists = make(map[string]struct{})

// Handler provide a common response writer to handle exceptions
type Handler struct {
}

func (h *Handler) Handle(i *chain.Invocation) {
	w, r, apiPath := i.Context().Value(rest.CtxResponse).(http.ResponseWriter),
		i.Context().Value(rest.CtxRequest).(*http.Request),
		i.Context().Value(rest.CtxMatchPattern).(string)

	i.WithContext(rest.CtxResponseStatus, http.StatusOK)

	if InWhitelist(r.Method, apiPath) {
		i.Next()
		return
	}

	asyncWriter := NewWriter(w)
	i.WithContext(rest.CtxResponse, asyncWriter)
	i.Next(chain.WithFunc(func(ret chain.Result) {
		if !ret.OK {
			i.WithContext(rest.CtxResponseStatus, h.responseError(w, ret.Err))
			return
		}

		i.WithContext(rest.CtxResponseStatus, asyncWriter.StatusCode)
		if err := asyncWriter.Flush(); err != nil {
			log.Error("response writer flush failed", err)
		}
		h.alarmIfInternalError(asyncWriter.StatusCode, util.BytesToStringWithNoCopy(asyncWriter.Body))
	}))
}

func (h *Handler) responseError(w http.ResponseWriter, e error) (statusCode int) {
	statusCode = http.StatusBadRequest
	contentType := rest.ContentTypeText
	body := []byte("Unknown error")
	defer func() {
		w.Header().Set(rest.HeaderContentType, contentType)
		w.WriteHeader(statusCode)
		if _, writeErr := w.Write(body); writeErr != nil {
			log.Error("write response failed", writeErr)
		}
		h.alarmIfInternalError(statusCode, util.BytesToStringWithNoCopy(body))
	}()

	if e == nil {
		log.Warn("callback result is failure but no error")
		return
	}

	body = util.StringToBytesWithNoCopy(e.Error())
	if err, ok := e.(*errsvc.Error); ok {
		statusCode = err.StatusCode()
		contentType = rest.ContentTypeJSON
		body, _ = json.Marshal(err)
	}
	return
}

func (h *Handler) alarmIfInternalError(statusCode int, errMsg string) {
	if statusCode < http.StatusInternalServerError {
		return
	}
	err := alarm.Raise(alarm.IDInternalError, alarm.AdditionalContext(errMsg))
	if err != nil {
		log.Error("raise alarm failed", err)
	}
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.ServerChainName, &Handler{})
}

func RegisterWhitelist(method, apiPath string) {
	whitelists[method+" "+apiPath] = struct{}{}
	log.Debug(fmt.Sprintf("skip handle api [%s]%s exceptions", method, apiPath))
}

func InWhitelist(method, apiPath string) bool {
	_, ok := whitelists[method+" "+apiPath]
	return ok
}

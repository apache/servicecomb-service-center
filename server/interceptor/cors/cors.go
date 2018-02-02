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

 /*
 Copyright (c) 2014 Olivier Poitrey <rs@dailymotion.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is furnished
to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
  */
package cors

import (
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var cors *CORS

func init() {
	cors = New()
}

type CORS struct {
	allowOrigin      string
	allowMethods     map[string]struct{}
	allowHeaders     map[string]struct{}
	allowCredentials bool
	exposeHeaders    string
	maxAge           int
	userHandler      http.Handler
}

func New() *CORS {
	c := new(CORS)
	c.allowOrigin = "*"
	c.allowCredentials = false
	c.allowHeaders = map[string]struct{}{"origin": {}, "content-type": {}, "x-domain-name": {}, "x-consumerid": {}}
	c.allowMethods = map[string]struct{}{"GET": {}, "POST": {}, "PUT": {}, "DELETE": {}, "UPDATE": {}}
	c.maxAge = 1500
	return c
}

func (cors *CORS) AllowMethods() []string {
	return util.MapToList(cors.allowMethods)
}

func (cors *CORS) AllowHeaders() []string {
	return util.MapToList(cors.allowHeaders)
}

func (cors *CORS) handlePreflightRequest(w http.ResponseWriter, r *http.Request) {
	acrm := r.Header.Get("Access-Control-Request-Method")
	if acrm == "" {
		cors.invalid(w, r)
		util.Logger().Warnf(nil, "header 'Access-Control-Request-Method' is empty")
		return
	}
	methods := strings.Split(strings.TrimSpace(acrm), ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		if _, ok := cors.allowMethods[m]; !ok {
			cors.invalid(w, r)
			util.Logger().Warnf(nil, "only supported methods: %v", util.MapToList(cors.allowMethods))
			return
		}
	}
	acrh := r.Header.Get("Access-Control-Request-Headers")
	if acrh != "" {
		headers := strings.Split(strings.TrimSpace(acrh), ",")
		for _, h := range headers {
			h = strings.ToLower(strings.TrimSpace(h))
			if _, ok := cors.allowHeaders[h]; !ok {
				cors.invalid(w, r)
				util.Logger().Warnf(nil, "only supported headers: %v", util.MapToList(cors.allowHeaders))
				return
			}
		}
	}

	w.Header().Add("Access-Control-Allow-Methods", util.StringJoin(cors.AllowMethods(), ","))
	w.Header().Add("Access-Control-Allow-Headers", util.StringJoin(cors.AllowHeaders(), ","))
	w.Header().Add("Access-Control-Max-Age", strconv.Itoa(cors.maxAge))
	cors.addAllowOriginHeader(w, r)
	cors.addAllowCookiesHeader(w, r)
	return
}

func (cors *CORS) invalid(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, "CORS Request Invalid")
	return
}

func (cors *CORS) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	if cors.exposeHeaders != "" {
		w.Header().Add("Access-Control-Expose-Headers", cors.exposeHeaders)
	}
	cors.addAllowOriginHeader(w, r)
	cors.addAllowCookiesHeader(w, r)
	return
}

func (cors *CORS) addAllowOriginHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", cors.allowOrigin)
	return
}

func (cors *CORS) addAllowCookiesHeader(w http.ResponseWriter, r *http.Request) {
	if cors.allowCredentials {
		w.Header().Add("Access-Control-Allow-Credentials", "true")
	}
}

func SetAllowMethods(methods []string) {
	cors.allowMethods = util.ListToMap(methods)
}

func SetAllowHeaders(headers []string) {
	cors.allowHeaders = util.ListToMap(headers)
}

func Intercept(w http.ResponseWriter, r *http.Request) (err error) {
	if origin := r.Header.Get("Origin"); origin == "" {
	} else if r.Method != "OPTIONS" {
		cors.handleActualRequest(w, r)
	} else if acrm := r.Header.Get("Access-Control-Request-Method"); acrm == "" {
		cors.handleActualRequest(w, r)
	} else {
		util.Logger().Debugf("identify the current request is a CORS")
		cors.handlePreflightRequest(w, r)
		err = errors.New("Handle preflight request")
	}
	return
}

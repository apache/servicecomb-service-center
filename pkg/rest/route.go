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
package rest

import (
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/chain"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"net/http"
	"net/url"
	"strings"
)

type URLPattern struct {
	Method string
	Path   string
}

type urlPatternHandler struct {
	Name string
	Path string
	http.Handler
}

type Route struct {
	// Method is one of the following: GET,PUT,POST,DELETE
	Method string
	// Path contains a path pattern
	Path string
	// rest callback function for the specified Method and Path
	Func func(w http.ResponseWriter, r *http.Request)
}

// HTTP request multiplexer
// Attention:
//   1. not thread-safe, must be initialized completely before serve http request
//   2. redirect not supported
type ROAServerHandler struct {
	handlers map[string][]*urlPatternHandler
}

func NewROAServerHander() *ROAServerHandler {
	return &ROAServerHandler{
		handlers: make(map[string][]*urlPatternHandler),
	}
}

func (this *ROAServerHandler) addRoute(route *Route) (err error) {
	method := strings.ToUpper(route.Method)
	if !isValidMethod(method) || !strings.HasPrefix(route.Path, "/") || route.Func == nil {
		message := fmt.Sprintf("Invalid route parameters(method: %s, path: %s)", method, route.Path)
		log.Errorf(nil, message)
		return errors.New(message)
	}

	this.handlers[method] = append(this.handlers[method], &urlPatternHandler{
		util.FormatFuncName(util.FuncName(route.Func)), route.Path, http.HandlerFunc(route.Func)})
	log.Infof("register route %s(%s)", route.Path, method)

	return nil
}

func (this *ROAServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ph := range this.handlers[r.Method] {
		if params, ok := ph.try(r.URL.Path); ok {
			if len(params) > 0 {
				r.URL.RawQuery = params + r.URL.RawQuery
			}

			this.serve(ph, w, r)
			return
		}
	}

	allowed := make([]string, 0, len(this.handlers))
	for method, handlers := range this.handlers {
		if method == r.Method {
			continue
		}

		for _, ph := range handlers {
			if _, ok := ph.try(r.URL.Path); ok {
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Add(HEADER_ALLOW, util.StringJoin(allowed, ", "))
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (this *ROAServerHandler) serve(ph *urlPatternHandler, w http.ResponseWriter, r *http.Request) {
	ctx := util.NewStringContext(r.Context())
	if ctx != r.Context() {
		nr := r.WithContext(ctx)
		*r = *nr
	}

	inv := chain.NewInvocation(ctx, chain.NewChain(SERVER_CHAIN_NAME, chain.Handlers(SERVER_CHAIN_NAME)))
	inv.WithContext(CTX_RESPONSE, w).
		WithContext(CTX_REQUEST, r).
		WithContext(CTX_MATCH_PATTERN, ph.Path).
		WithContext(CTX_MATCH_FUNC, ph.Name).
		Invoke(
			func(ret chain.Result) {
				defer func() {
					err := ret.Err
					itf := recover()
					if itf != nil {
						log.LogPanic(itf)

						err = errorsEx.RaiseError(itf)
					}
					if _, ok := err.(errorsEx.InternalError); ok {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}()
				if ret.OK {
					ph.ServeHTTP(w, r)
				}
			})
}

func (this *urlPatternHandler) try(path string) (p string, _ bool) {
	var i, j int
	l, sl := len(this.Path), len(path)
	for i < sl {
		switch {
		case j >= l:
			if this.Path != "/" && l > 0 && this.Path[l-1] == '/' {
				return p, true
			}
			return "", false
		case this.Path[j] == ':':
			var val string
			var nextc byte
			o := j
			_, nextc, j = match(this.Path, isAlnum, 0, j+1)
			val, _, i = match(path, matchParticial, nextc, i)

			p += url.QueryEscape(this.Path[o:j]) + "=" + url.QueryEscape(val) + "&"
		case path[i] == this.Path[j]:
			i++
			j++
		default:
			return "", false
		}
	}
	if j != l {
		return "", false
	}
	return p, true
}

func match(s string, f func(c byte) bool, exclude byte, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) && s[j] != exclude {
		j++
	}

	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

func matchParticial(c byte) bool {
	return c != '/'
}

func isAlpha(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

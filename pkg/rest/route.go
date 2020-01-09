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
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// URLPattern defines an uri pattern
type URLPattern struct {
	Method string
	Path   string
}

type urlPatternHandler struct {
	Name string
	Path string
	http.Handler
}

// Route is a http route
type Route struct {
	// Method is one of the following: GET,PUT,POST,DELETE
	Method string
	// Path contains a path pattern
	Path string
	// rest callback function for the specified Method and Path
	Func func(w http.ResponseWriter, r *http.Request)
}

// ROAServantService defines a group of Routes
type ROAServantService interface {
	URLPatterns() []Route
}

// ROAServerHandler is a HTTP request multiplexer
// Attention:
//   1. not thread-safe, must be initialized completely before serve http request
//   2. redirect not supported
type ROAServerHandler struct {
	handlers  map[string][]*urlPatternHandler
	chainName string
}

// NewROAServerHander news an ROAServerHandler
func NewROAServerHander() *ROAServerHandler {
	return &ROAServerHandler{
		handlers: make(map[string][]*urlPatternHandler),
	}
}

// RegisterServant registers a ROAServantService
// servant must be an pointer to service object
func (roa *ROAServerHandler) RegisterServant(servant interface{}) {
	val := reflect.ValueOf(servant)
	ind := reflect.Indirect(val)
	typ := ind.Type()
	name := util.FileLastName(typ.PkgPath() + "." + typ.Name())
	if val.Kind() != reflect.Ptr {
		log.Errorf(nil, "<rest.RegisterServant> cannot use non-ptr servant struct `%s`", name)
		return
	}

	urlPatternFunc := val.MethodByName("URLPatterns")
	if !urlPatternFunc.IsValid() {
		log.Errorf(nil, "<rest.RegisterServant> no 'URLPatterns' function in servant struct `%s`", name)
		return
	}

	vals := urlPatternFunc.Call([]reflect.Value{})
	if len(vals) <= 0 {
		log.Errorf(nil, "<rest.RegisterServant> call 'URLPatterns' function failed in servant struct `%s`", name)
		return
	}

	val0 := vals[0]
	if !val.CanInterface() {
		log.Errorf(nil, "<rest.RegisterServant> result of 'URLPatterns' function not interface type in servant struct `%s`", name)
		return
	}

	if routes, ok := val0.Interface().([]Route); ok {
		log.Infof("register servant %s", name)
		for _, route := range routes {
			err := roa.addRoute(&route)
			if err != nil {
				log.Errorf(err, "register route failed.")
			}
		}
	} else {
		log.Errorf(nil, "<rest.RegisterServant> result of 'URLPatterns' function not []*Route type in servant struct `%s`", name)
	}
}

func (roa *ROAServerHandler) setChainName(name string) {
	roa.chainName = name
}

func (roa *ROAServerHandler) addRoute(route *Route) (err error) {
	method := strings.ToUpper(route.Method)
	if !isValidMethod(method) || !strings.HasPrefix(route.Path, "/") || route.Func == nil {
		message := fmt.Sprintf("Invalid route parameters(method: %s, path: %s)", method, route.Path)
		log.Errorf(nil, message)
		return errors.New(message)
	}

	roa.handlers[method] = append(roa.handlers[method], &urlPatternHandler{
		util.FormatFuncName(util.FuncName(route.Func)), route.Path, http.HandlerFunc(route.Func)})
	log.Infof("register route %s(%s)", route.Path, method)

	return nil
}

// ServeHTTP implements http.Handler
func (roa *ROAServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ph := range roa.handlers[r.Method] {
		if params, ok := ph.try(r.URL.Path); ok {
			if len(params) > 0 {
				r.URL.RawQuery = params + r.URL.RawQuery
			}

			roa.serve(ph, w, r)
			return
		}
	}

	allowed := make([]string, 0, len(roa.handlers))
	for method, handlers := range roa.handlers {
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

func (roa *ROAServerHandler) serve(ph *urlPatternHandler, w http.ResponseWriter, r *http.Request) {
	ctx := util.NewStringContext(r.Context())
	if ctx != r.Context() {
		nr := r.WithContext(ctx)
		*r = *nr
	}

	inv := chain.NewInvocation(ctx, chain.NewChain(roa.chainName, chain.Handlers(roa.chainName)))
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

func (roa *urlPatternHandler) try(path string) (p string, _ bool) {
	var i, j int
	l, sl := len(roa.Path), len(path)
	for i < sl {
		switch {
		case j >= l:
			if roa.Path != "/" && l > 0 && roa.Path[l-1] == '/' {
				return p, true
			}
			return "", false
		case roa.Path[j] == ':':
			var val string
			var nextc byte
			o := j
			_, nextc, j = match(roa.Path, isAlnum, 0, j+1)
			val, _, i = match(path, matchParticial, nextc, i)

			p += url.QueryEscape(roa.Path[o:j]) + "=" + url.QueryEscape(val) + "&"
		case path[i] == roa.Path[j]:
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

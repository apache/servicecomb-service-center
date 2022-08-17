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
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var doNothingFunc = func(_ chain.Result) {}

// Router is a HTTP request multiplexer
// Attention:
//  1. not thread-safe, must be initialized completely before serve http request
//  2. redirect not supported
type Router struct {
	handlers  map[string][]*urlPatternHandler
	chainName string
}

// RegisterServant registers a RouteGroup
// servant must be an pointer to service object
func (router *Router) RegisterServant(servant RouteGroup) {
	log.Info(fmt.Sprintf("register servant %s", util.Reflect(servant).Name()))
	for _, route := range servant.URLPatterns() {
		err := router.addRoute(&route)
		if err != nil {
			log.Error("register route failed.", err)
		}
	}
}

func (router *Router) setChainName(name string) {
	router.chainName = name
}

func (router *Router) addRoute(route *Route) (err error) {
	method := strings.ToUpper(route.Method)
	if !isValidMethod(method) || !strings.HasPrefix(route.Path, "/") || route.Func == nil {
		message := fmt.Sprintf("Invalid route parameters(method: %s, path: %s)", method, route.Path)
		log.Error(message, nil)
		return errors.New(message)
	}

	router.handlers[method] = append(router.handlers[method], &urlPatternHandler{
		util.FormatFuncName(util.FuncName(route.Func)), route.Path, http.HandlerFunc(route.Func)})
	log.Info(fmt.Sprintf("register route %s(%s)", route.Path, method))

	return nil
}

// ServeHTTP implements http.Handler
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ph := range router.handlers[r.Method] {
		if params, ok := ph.try(r.URL.Path); ok {
			if len(params) > 0 {
				r.URL.RawQuery = params + r.URL.RawQuery
			}

			router.serve(ph, w, r)
			return
		}
	}

	allowed := make([]string, 0, len(router.handlers))
	for method, handlers := range router.handlers {
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

	w.Header().Add(HeaderAllow, util.StringJoin(allowed, ", "))
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (router *Router) serve(ph *urlPatternHandler, w http.ResponseWriter, r *http.Request) {
	ctx := util.NewStringContext(r.Context())
	if ctx != r.Context() {
		nr := r.WithContext(ctx)
		*r = *nr
	}

	inv := chain.NewInvocation(ctx, chain.NewChain(router.chainName, chain.Handlers(router.chainName)))
	inv.WithContext(CtxResponse, w).
		WithContext(CtxRequest, r).
		WithContext(CtxMatchPattern, ph.Path).
		WithContext(CtxMatchFunc, ph.Name).
		WithContext(CtxRouteHandler, ph).
		Invoke(doNothingFunc)
}

// NewRouter news a Router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string][]*urlPatternHandler),
	}
}

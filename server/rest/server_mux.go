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
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// The rest middleware design:
//   http requests -> DefaultServerMux [ -> ROA router -> handler chain ] -> services
// Usages:
//   1. if use rest.RegisterServant:
//      to register in ROA and requests will go through the handler chain
//   2. if use RegisterServerHandleFunc or RegisterServerHandler:
//      to register in ServeMux directly

const defaultServeMux = "default"

var (
	DefaultServerMux = http.NewServeMux()
	serveMuxMap      = map[string]*http.ServeMux{
		defaultServeMux: DefaultServerMux,
	}
)

func RegisterServeMux(name string) {
	serveMuxMap[name] = http.NewServeMux()
}

func RegisterServeMuxHandleFunc(name, pattern string, f http.HandlerFunc) {
	serveMuxMap[name].HandleFunc(pattern, f)

	log.Infof("register serve mux '%s' http handle function %s(), pattern %s",
		name, util.FuncName(f), pattern)
}

func RegisterServeMuxHandler(name, pattern string, h http.Handler) {
	serveMuxMap[name].Handle(pattern, h)

	t := util.Reflect(h).Type
	log.Infof("register serve mux '%s' http handler %s/%s, pattern %s",
		name, t.PkgPath(), t.Name(), pattern)
}

func RegisterServerHandleFunc(pattern string, f http.HandlerFunc) {
	RegisterServeMuxHandleFunc(defaultServeMux, pattern, f)
}

func RegisterServerHandler(pattern string, h http.Handler) {
	RegisterServeMuxHandler(defaultServeMux, pattern, h)
}

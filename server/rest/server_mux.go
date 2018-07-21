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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"net/http"
)

// The rest middleware design:
//   http requests -> DefaultServerMux [ -> ROA router -> handler chain ] -> services
// Usages:
//   1. if use rest.RegisterServant:
//      to register in ROA and requests will go through the handler chain
//   2. if use RegisterServerHandleFunc or RegisterServerHandler:
//      to register in ServeMux directly
var DefaultServerMux = http.NewServeMux()

func RegisterServerHandleFunc(pattern string, f http.HandlerFunc) {
	DefaultServerMux.HandleFunc(pattern, f)

	util.Logger().Infof("register server http handle function %s(), pattern %s", util.FuncName(f), pattern)
}

func RegisterServerHandler(pattern string, h http.Handler) {
	DefaultServerMux.Handle(pattern, h)

	t := util.Reflect(h).Type
	util.Logger().Infof("register server http handler %s/%s, pattern %s", t.PkgPath(), t.Name(), pattern)
}

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
)

var router *Router // router is the default handler

func init() {
	router = NewRouter()
	router.setChainName(ServerChainName)
}

// RegisterServant registers a RouteGroup into router
// servant must be an pointer to service object
func RegisterServant(group RouteGroup) {
	router.RegisterServant(group)
}

// GetRouter return the router fo REST service
func GetRouter() http.Handler {
	return router
}

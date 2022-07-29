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

package v3

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	v4 "github.com/apache/servicecomb-service-center/server/resource/disco"
)

type MicroServiceService struct {
	v4.ServiceResource
}

func (s *MicroServiceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/registry/v3/existence", Func: s.ResourceExist},
		{Method: http.MethodGet, Path: "/registry/v3/microservices", Func: s.ListService},
		{Method: http.MethodGet, Path: "/registry/v3/microservices/:serviceId", Func: s.GetService},
		{Method: http.MethodPost, Path: "/registry/v3/microservices", Func: s.RegisterService},
		{Method: http.MethodPut, Path: "/registry/v3/microservices/:serviceId/properties", Func: s.PutServiceProperties},
		{Method: http.MethodDelete, Path: "/registry/v3/microservices/:serviceId", Func: s.UnregisterService},
		{Method: http.MethodDelete, Path: "/registry/v3/microservices", Func: s.UnregisterManyService},
		// tags
		{Method: http.MethodPost, Path: "/registry/v3/microservices/:serviceId/tags", Func: s.PutManyTags},
		{Method: http.MethodPut, Path: "/registry/v3/microservices/:serviceId/tags/:key", Func: s.PutTag},
		{Method: http.MethodGet, Path: "/registry/v3/microservices/:serviceId/tags", Func: s.ListTag},
		{Method: http.MethodDelete, Path: "/registry/v3/microservices/:serviceId/tags/:key", Func: s.DeleteManyTags},
	}
}

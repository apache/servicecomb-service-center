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
	v4 "github.com/apache/servicecomb-service-center/server/resource/v4/disco"
)

type MicroServiceInstanceService struct {
	v4.InstanceResource
}

func (s *MicroServiceInstanceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/registry/v3/instances", Func: s.FindInstances},
		{Method: http.MethodGet, Path: "/registry/v3/microservices/:serviceId/instances", Func: s.ListInstance},
		{Method: http.MethodGet, Path: "/registry/v3/microservices/:serviceId/instances/:instanceId", Func: s.GetInstance},
		{Method: http.MethodPost, Path: "/registry/v3/microservices/:serviceId/instances", Func: s.LegacyRegisterInstance},
		{Method: http.MethodDelete, Path: "/registry/v3/microservices/:serviceId/instances/:instanceId", Func: s.UnregisterInstance},
		{Method: http.MethodPut, Path: "/registry/v3/microservices/:serviceId/instances/:instanceId/properties", Func: s.PutInstanceProperties},
		{Method: http.MethodPut, Path: "/registry/v3/microservices/:serviceId/instances/:instanceId/status", Func: s.PutInstanceStatus},
		{Method: http.MethodPut, Path: "/registry/v3/microservices/:serviceId/instances/:instanceId/heartbeat", Func: s.SendHeartbeat},
		{Method: http.MethodPut, Path: "/registry/v3/heartbeats", Func: s.SendManyHeartbeat},
	}
}

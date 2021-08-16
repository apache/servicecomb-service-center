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
	"github.com/apache/servicecomb-service-center/pkg/rest"
	v4 "github.com/apache/servicecomb-service-center/server/rest/controller/v4"
)

type MicroServiceInstanceService struct {
	v4.MicroServiceInstanceService
}

func (this *MicroServiceInstanceService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTPMethodGet, "/registry/v3/instances", this.FindInstances},
		{rest.HTTPMethodGet, "/registry/v3/microservices/:serviceId/instances", this.GetInstances},
		{rest.HTTPMethodGet, "/registry/v3/microservices/:serviceId/instances/:instanceId", this.GetOneInstance},
		{rest.HTTPMethodPost, "/registry/v3/microservices/:serviceId/instances", this.RegisterInstance},
		{rest.HTTPMethodDelete, "/registry/v3/microservices/:serviceId/instances/:instanceId", this.UnregisterInstance},
		{rest.HTTPMethodPut, "/registry/v3/microservices/:serviceId/instances/:instanceId/properties", this.UpdateMetadata},
		{rest.HTTPMethodPut, "/registry/v3/microservices/:serviceId/instances/:instanceId/status", this.UpdateStatus},
		{rest.HTTPMethodPut, "/registry/v3/microservices/:serviceId/instances/:instanceId/heartbeat", this.Heartbeat},
		{rest.HTTPMethodPut, "/registry/v3/heartbeats", this.HeartbeatSet},
	}
}

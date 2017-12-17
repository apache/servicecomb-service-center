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
package govern

import (
	"github.com/ServiceComb/service-center/pkg/rest"
)

// GovernService 治理相关接口服务
type GovernServiceControllerV3 struct {
	GovernServiceControllerV4
}

// URLPatterns 路由
func (governService *GovernServiceControllerV3) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/registry/v3/govern/service/:serviceId", governService.GetServiceDetail},
		{rest.HTTP_METHOD_GET, "/registry/v3/govern/relation", governService.GetGraph},
		{rest.HTTP_METHOD_GET, "/registry/v3/govern/services", governService.GetAllServicesInfo},
	}
}

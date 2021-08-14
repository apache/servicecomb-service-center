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

type RuleService struct {
	v4.RuleService
}

func (this *RuleService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTPMethodPost, "/registry/v3/microservices/:serviceId/rules", this.AddRule},
		{rest.HTTPMethodGet, "/registry/v3/microservices/:serviceId/rules", this.GetRules},
		{rest.HTTPMethodPut, "/registry/v3/microservices/:serviceId/rules/:rule_id", this.UpdateRule},
		{rest.HTTPMethodDelete, "/registry/v3/microservices/:serviceId/rules/:rule_id", this.DeleteRule},
	}
}

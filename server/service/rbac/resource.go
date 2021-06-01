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

package rbac

import (
	"github.com/go-chassis/cari/rbac"
)

const (
	ResourceAccount = "account"
	ResourceRole    = "role"
	ResourceService = "service"
	ResourceGovern  = "governance"
	ResourceSchema  = "service/schema"
	ResourceOps     = "ops"
)

var (
	APITokenGranter = "/v4/token"

	APIAccountList = "/v4/accounts"

	APIRoleList = "/v4/roles"

	APIAccountPassword = "/v4/accounts/:name/password"

	APIOps = "/v4/:project/admin"

	APIGov = "/v1/:project/gov/"

	APILegacyGov = "/v4/:project/govern"

	APIServiceInfo       = "/v4/:project/registry/microservices/:serviceId"
	APIServicesList      = "/v4/:project/registry/microservices"
	APIServiceProperties = "/v4/:project/registry/microservices/:serviceId/properties"
	APIServiceExistence  = "/v4/:project/registry/existence"

	APIProConDependency = "/v4/:project/registry/microservices/:providerId/consumers"
	APIConProDependency = "/v4/:project/registry/microservices/:consumerId/providers"

	APIHeartbeats          = "/v4/:project/registry/heartbeats"
	APIInstanceWatcher     = "/v4/:project/registry/microservices/:serviceId/watcher"
	APIInstanceListWatcher = "/v4/:project/registry/microservices/:serviceId/listwatcher"

	APIServiceTag    = "/v4/:project/registry/microservices/:serviceId/tags"
	APIServiceTagKey = "/v4/:project/registry/microservices/:serviceId/tags/:key"

	APIServiceRule     = "/v4/:project/registry/microservices/:serviceId/rules"
	APIServiceRuleList = "/v4/:project/registry/microservices/:serviceId/rules/rule_id"

	APIServiceSchema = "/v4/:project/registry/microservices/:serviceId/schemas"
)

func InitResourceMap() {
	rbac.PartialMapResource(APIAccountList, ResourceAccount)

	rbac.PartialMapResource(APIRoleList, ResourceRole)

	rbac.PartialMapResource(APIGov, ResourceGovern)

	rbac.PartialMapResource(APIServiceSchema, ResourceSchema)

	rbac.PartialMapResource(APIOps, ResourceOps)

	rbac.PartialMapResource("instances", ResourceService)
	rbac.PartialMapResource(APILegacyGov, ResourceService)

	rbac.MapResource(APIServiceInfo, ResourceService)
	rbac.MapResource(APIServicesList, ResourceService)
	rbac.MapResource(APIServiceProperties, ResourceService)
	rbac.MapResource(APIServiceExistence, ResourceService)
	rbac.MapResource(APIProConDependency, ResourceService)
	rbac.MapResource(APIConProDependency, ResourceService)
	rbac.MapResource(APIHeartbeats, ResourceService)
	rbac.MapResource(APIInstanceWatcher, ResourceService)
	rbac.MapResource(APIInstanceListWatcher, ResourceService)
	rbac.MapResource(APIServiceRuleList, ResourceService)
	rbac.MapResource(APIServiceRule, ResourceService)
	rbac.MapResource(APIServiceTag, ResourceService)
	rbac.MapResource(APIServiceTagKey, ResourceService)
}

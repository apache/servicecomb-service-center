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

import "github.com/apache/servicecomb-service-center/pkg/rbacframe"

const (
	ResourceAccount    = "account"
	ResourceRole       = "role"
	ResourceService    = "service"
	ResourceInstance   = "instance"
	ResourceDep        = "dependencies"
	ResourceRule       = "rule"
	ResourceTag        = "tag"
	ResourceGovern     = "governance"
	ResourceSchema     = "schema"
	ResourceAdminister = "administer"
)

var (
	APIHealth       = "/v4/:project/registry/health"
	APIVersion      = "/v4/:project/registry/version"
	APITokenGranter = "/v4/token"

	APIAccountList  = "/v4/account"
	APIUserAccount  = "/v4/account/:name"
	APIUserPassword = "/v4/account/:name/password"

	APIRoleList = "/v4/role"
	APIRoleInfo = "/v4/role/:roleName"

	APIServiceInfo       = "/v4/:project/registry/microservices/:serviceId"
	APIServicesList      = "/v4/:project/registry/microservices"
	APIServiceProperties = "/v4/:project/registry/microservices/:serviceId/properties"
	APIServiceExistence  = "/v4/:project/registry/existence"

	APIProConDependency = "/v4/:project/registry/microservices/:providerId/consumers"
	APIConProDependency = "/v4/:project/registry/microservices/:consumerId/providers"

	APIInstancesList       = "/v4/:project/registry/microservices/:serviceId/instances"
	APIInstanceInfo        = "/v4/:project/registry/microservices/:serviceId/instances/:instanceId"
	APIInstanceProperties  = "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/properties"
	APIInstanceStatus      = "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/status"
	APIInstanceHeartbeats  = "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat"
	APIHeartbeats          = "/v4/:project/registry/heartbeats"
	APIFindInstances       = "/v4/:project/registry/instances"
	APIInstanceAction      = "/v4/:project/registry/instances/action"
	APIInstanceWatcher     = "/v4/:project/registry/microservices/:serviceId/watcher"
	APIInstanceListWatcher = "/v4/:project/registry/microservices/:serviceId/listwatcher"

	APIServiceTag    = "/v4/:project/registry/microservices/:serviceId/tags"
	APIServiceTagKey = "/v4/:project/registry/microservices/:serviceId/tags/:key"

	APIServiceRule     = "/v4/:project/registry/microservices/:serviceId/rules"
	APIServiceRuleList = "/v4/:project/registry/microservices/:serviceId/rules/rule_id"

	APIServiceSchemaInfo = "/v4/:project/registry/microservices/:serviceId/schemas/:schemaId"
	APIServiceSchema     = "/v4/:project/registry/microservices/:serviceId/schemas"

	APIGovernService         = "/v4/:project/govern/microservices/:serviceId"
	APIGovernServiceInfo     = "/v4/:project/govern/microservices"
	APIGovernServiceRelation = "/v4/:project/govern/relations"
	APIGovernApps            = "/v4/:project/govern/relations"

	APIDump     = "/v4/:project/admin/dump"
	APIClusters = "/v4/:project/admin/clusters"
	APIAlarms   = "/v4/:project/admin/alarms"
)

func initResourceMap() {
	rbacframe.MapResource(APIAccountList, ResourceAccount)
	rbacframe.MapResource(APIUserAccount, ResourceAccount)
	rbacframe.MapResource(APIUserPassword, ResourceAccount)

	rbacframe.MapResource(APIRoleList, ResourceRole)
	rbacframe.MapResource(APIRoleInfo, ResourceRole)

	rbacframe.MapResource(APIServiceInfo, ResourceService)
	rbacframe.MapResource(APIServicesList, ResourceService)
	rbacframe.MapResource(APIServiceProperties, ResourceService)
	rbacframe.MapResource(APIServiceExistence, ResourceService)

	rbacframe.MapResource(APIServiceSchemaInfo, ResourceSchema)
	rbacframe.MapResource(APIServiceSchema, ResourceSchema)

	rbacframe.MapResource(APIProConDependency, ResourceDep)
	rbacframe.MapResource(APIConProDependency, ResourceDep)

	rbacframe.MapResource(APIInstancesList, ResourceInstance)
	rbacframe.MapResource(APIInstanceInfo, ResourceInstance)
	rbacframe.MapResource(APIInstanceProperties, ResourceInstance)
	rbacframe.MapResource(APIInstanceStatus, ResourceInstance)
	rbacframe.MapResource(APIInstanceHeartbeats, ResourceInstance)
	rbacframe.MapResource(APIHeartbeats, ResourceInstance)
	rbacframe.MapResource(APIFindInstances, ResourceInstance)
	rbacframe.MapResource(APIInstanceAction, ResourceInstance)
	rbacframe.MapResource(APIInstanceWatcher, ResourceInstance)
	rbacframe.MapResource(APIInstanceListWatcher, ResourceInstance)

	rbacframe.MapResource(APIServiceRuleList, ResourceRule)
	rbacframe.MapResource(APIServiceRule, ResourceRule)

	rbacframe.MapResource(APIServiceTag, ResourceTag)
	rbacframe.MapResource(APIServiceTagKey, ResourceTag)

	rbacframe.MapResource(APIGovernService, ResourceGovern)
	rbacframe.MapResource(APIGovernServiceInfo, ResourceGovern)
	rbacframe.MapResource(APIGovernServiceRelation, ResourceGovern)
	rbacframe.MapResource(APIGovernApps, ResourceGovern)

	rbacframe.MapResource(APIDump, ResourceAdminister)
	rbacframe.MapResource(APIClusters, ResourceAdminister)
	rbacframe.MapResource(APIAlarms, ResourceAdminister)
}

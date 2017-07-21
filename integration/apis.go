//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package integrationtest

// Basic API
var HEALTH = "/health"
var VERSION = "/version"

// Micro-Service API's
var CHECKEXISTENCE = "/registry/v3/existence"
var GETALLSERVICE = "/registry/v3/microservices"
var GETSERVICEBYID = "/registry/v3/microservices/:serviceId"
var REGISTERMICROSERVICE = "/registry/v3/microservices"
var UPDATEMICROSERVICE = "/registry/v3/microservices/:serviceId/properties"
var UNREGISTERMICROSERVICE = "/registry/v3/microservices/:serviceId"
var GETSCHEMABYID = "/registry/v3/microservices/:serviceId/schemas/:schemaId"
var UPDATESCHEMA = "/registry/v3/microservices/:serviceId/schemas/:schemaId"
var DELETESCHEMA = "/registry/v3/microservices/:serviceId/schemas/:schemaId"
var CREATEDEPENDENCIES = "/registry/v3/dependencies"
var GETCONPRODEPENDENCY = "/registry/v3/microservices/:consumerId/providers"
var GETPROCONDEPENDENCY = "/registry/v3/microservices/:providerId/consumers"

// Instance API's
var FINDINSTANCE = "/registry/v3/instances"
var GETINSTANCE = "/registry/v3/microservices/:serviceId/instances"
var GETINSTANCEBYINSTANCEID = "/registry/v3/microservices/:serviceId/instances/:instanceId"
var REGISTERINSTANCE = "/registry/v3/microservices/:serviceId/instances"
var UNREGISTERINSTANCE = "/registry/v3/microservices/:serviceId/instances/:instanceId"
var UPDATEINSTANCEMETADATA = "/registry/v3/microservices/:serviceId/instances/:instanceId/properties"
var UPDATEINSTANCESTATUS = "/registry/v3/microservices/:serviceId/instances/:instanceId/status"
var INSTANCEHEARTBEAT = "/registry/v3/microservices/:serviceId/instances/:instanceId/heartbeat"
var INSTANCEWATCHER = "/registry/v3/microservices/:serviceId/watcher"
var INSTANCELISTWATCHER = "/registry/v3/microservices/:serviceId/listwatcher"

//Governance API's
var GETGOVERNANCESERVICEDETAILS = "/registry/v3/govern/service/:serviceId"
var GETRELATIONGRAPH = "/registry/v3/govern/relation"
var GETALLSERVICEGOVERNANCEINFO = "/registry/v3/govern/services"

//Rules API's
var ADDRULE = "/registry/v3/microservices/:serviceId/rules"
var GETRULES = "/registry/v3/microservices/:serviceId/rules"
var UPDATERULES = "/registry/v3/microservices/:serviceId/rules/:rule_id"
var DELETERULES = "/registry/v3/microservices/:serviceId/rules/:rule_id"

//Tag API's
var ADDTAGE = "/registry/v3/microservices/:serviceId/tags"
var UPDATETAG = "/registry/v3/microservices/:serviceId/tags/:key"
var GETTAGS = "/registry/v3/microservices/:serviceId/tags"
var DELETETAG = "/registry/v3/microservices/:serviceId/tags/:key"

// HTTP METHODS
var GET = "GET"
var POST = "POST"
var UPDATE = "PUT"
var DELETE = "DELETE"

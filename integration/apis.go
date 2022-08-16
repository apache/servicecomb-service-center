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
package integrationtest

// Basic API
var HEALTH = "/v4/default/registry/health"
var VERSION = "/v4/default/registry/version"

// Micro-Service API's
var CHECKEXISTENCE = "/v4/default/registry/existence"
var GETALLSERVICE = "/v4/default/registry/microservices"
var GETSERVICEBYID = "/v4/default/registry/microservices/:serviceId"
var REGISTERMICROSERVICE = "/v4/default/registry/microservices"
var UPDATEMICROSERVICE = "/v4/default/registry/microservices/:serviceId/properties"
var UNREGISTERMICROSERVICE = "/v4/default/registry/microservices/:serviceId?force=true"
var GETSCHEMABYID = "/v4/default/registry/microservices/:serviceId/schemas/:schemaId"
var UPDATESCHEMA = "/v4/default/registry/microservices/:serviceId/schemas/:schemaId"
var GETSCHEMAS = "/v4/default/registry/microservices/:serviceId/schemas"
var UPDATESCHEMAS = "/v4/default/registry/microservices/:serviceId/schemas"
var DELETESCHEMA = "/v4/default/registry/microservices/:serviceId/schemas/:schemaId"
var CREATEDEPENDENCIES = "/v4/default/registry/dependencies"
var GETCONPRODEPENDENCY = "/v4/default/registry/microservices/:consumerId/providers"
var GETPROCONDEPENDENCY = "/v4/default/registry/microservices/:providerId/consumers"

// Instance API's
var FINDINSTANCE = "/v4/default/registry/instances"
var INSTANCEACTION = "/v4/default/registry/instances/action"
var GETINSTANCE = "/v4/default/registry/microservices/:serviceId/instances"
var GETINSTANCEBYINSTANCEID = "/v4/default/registry/microservices/:serviceId/instances/:instanceId"
var REGISTERINSTANCE = "/v4/default/registry/microservices/:serviceId/instances"
var UNREGISTERINSTANCE = "/v4/default/registry/microservices/:serviceId/instances/:instanceId"
var UPDATEINSTANCEMETADATA = "/v4/default/registry/microservices/:serviceId/instances/:instanceId/properties"
var UPDATEINSTANCESTATUS = "/v4/default/registry/microservices/:serviceId/instances/:instanceId/status"
var INSTANCEHEARTBEAT = "/v4/default/registry/microservices/:serviceId/instances/:instanceId/heartbeat"
var INSTANCEWATCHER = "/v4/default/registry/microservices/:serviceId/watcher"

// Governance API's
var GETGOVERNANCESERVICEDETAILS = "/v4/default/govern/microservices/:serviceId"
var GETRELATIONGRAPH = "/v4/default/govern/relations"
var GETALLSERVICEGOVERNANCEINFO = "/v4/default/govern/microservices"
var GETALLAPPS = "/v4/default/govern/apps"

// Tag API's
var ADDTAGE = "/v4/default/registry/microservices/:serviceId/tags"
var UPDATETAG = "/v4/default/registry/microservices/:serviceId/tags/:key"
var GETTAGS = "/v4/default/registry/microservices/:serviceId/tags"
var DELETETAG = "/v4/default/registry/microservices/:serviceId/tags/:key"

// Admin API's
var DUMP = "/v4/default/admin/dump"

// HTTP METHODS
var GET = "GET"
var POST = "POST"
var UPDATE = "PUT"
var DELETE = "DELETE"

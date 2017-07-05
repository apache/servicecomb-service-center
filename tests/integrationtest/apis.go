package integrationtest

// Basic API
var HEALTH = "/health"
var VERSION = "/version"

// Micro-Service Api's
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

// Instance APi's
var FINDINSTANCE = "/registry/v3/instances"
var GETINSTANCE = "/registry/v3/microservices/:serviceId/instances"
var GETINSTANCEBYINSTANCEID = "/registry/v3/microservices/:serviceId/instances/:instanceId"
var REGISTERINSTANCE = "/registry/v3/microservices/:serviceId/instances"
var UNREGISTERINSTANCE = "/registry/v3/microservices/:serviceId/instances/:instanceId"
var UPDATEINSTANCEMETADATA = "/registry/v3/microservices/:serviceId/instances/:instanceId/properties"
var UPDATEINSTANCESTATUS = "/registry/v3/microservices/:serviceId/instances/:instanceId/status"
var INSTANCEHEARTBEAT = "/registry/v3/microservices/:serviceId/instances/:instanceId/heartbeat"

// HTTP METHODS
var GET = "GET"
var POST = "POST"
var UPDATE = "PUT"
var DELETE = "DELETE"

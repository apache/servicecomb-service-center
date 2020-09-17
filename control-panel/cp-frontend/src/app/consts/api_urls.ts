export enum apiUrls {
  GET_SERVICES = '/v4/:appId/govern/microservices',
  GET_SERVICE_SCHEMAS = '/v4/:appId/registry/microservices/:serviceId/schemas',
  GET_SERVICE_INSTANCES = '/v4/:appId/registry/microservices/:serviceId/instances',
  GET_SERVICE = '/v4/:appId/govern/microservices/:serviceId'
}

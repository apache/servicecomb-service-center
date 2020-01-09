# ToDo List of Service-center Syncer

[中文版](./TODO-ZH.md)

## Functionality

- Improve Service-center plugin to perform lossless data synchronization between Service-center
- Incremental data synchronization
- Support SpringCloud Eureka
- Support Istio
- Support K8S etcd
- Support Consule
- Support data synchronization cross datacenters

## Management

- Lifecycle management of syncer cluster
- Microservices blacklist of cross service-center
- Microservices whitelist of cross service-center-center
- Management of on-demand service-center synchronization
- Support read configuration from config file

## Reliable

- Lifecycle management for syncer's multi-instances in a Data-center
- Store syncer data to etcd

## Security

- Communication encryption
- Authority management
- Authentication management

## Deployment

- Docker deployment
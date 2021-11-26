# ServiceComb gRPC API
This go module contains all grpc service definition of service center 

### APIs
#### sync
service center metadata sync APIs, used in service center peer clusters data sync
#### auth
service center itself act as an auth server 
which maintain account, role, perms data. 
it exposes API for internal services to check token has perms to access resource
### Development Guide
#### To generate grpc code
```shell
cd api/sync/v1
protoc --go_out=. --go-grpc_out=. event_service.proto
```
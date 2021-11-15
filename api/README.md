# ServiceComb gRPC API
This go module contains all grpc service definition of service center 

### APIs
#### sync
service center metadata sync APIs, used in service center peer clusters data sync
### Development Guide
#### To generate grpc code
```shell
cd api/sync/v1
protoc --go_out=. --go-grpc_out=. event_service.proto
```
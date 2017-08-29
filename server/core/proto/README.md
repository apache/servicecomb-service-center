# service-center support for grpc
1. Download 'protoc' compiler, https://github.com/google/protobuf/releases
1. Install the go-grpc plugin, go get -u github.com/golang/protobuf/protoc-gen-go
1. Compile the service.proto file, protoc --go_out=plugins=grpc:. service.proto
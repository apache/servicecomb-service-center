# service-center support for grpc
1. Download 'protoc' compiler, https://github.com/google/protobuf/releases
1. git clone -b v1.0.0 https://github.com/golang/protobuf $GOPATH/src/github.com/golang/protobuf
1. Install the go-grpc plugin, go get github.com/golang/protobuf/protoc-gen-go
1. Compile the service.proto file, protoc --go_out=plugins=grpc:. services.proto
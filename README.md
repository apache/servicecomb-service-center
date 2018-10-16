# Apache-Incubator-ServiceComb-Service-Center 
[![Build Status](https://www.travis-ci.org/apache/incubator-servicecomb-service-center.svg?branch=master)](https://www.travis-ci.org/apache/incubator-servicecomb-service-center)  [![Coverage Status](https://coveralls.io/repos/github/apache/incubator-servicecomb-service-center/badge.svg?branch=master)](https://coveralls.io/github/apache/incubator-servicecomb-service-center?branch=master)  [![Go Report Card](https://goreportcard.com/badge/github.com/apache/incubator-servicecomb-service-center)](https://goreportcard.com/report/github.com/apache/incubator-servicecomb-service-center) [![GoDoc](https://godoc.org/github.com/apache/incubator-servicecomb-service-center?status.svg)](https://godoc.org/github.com/apache/incubator-servicecomb-service-center)  [![HitCount](http://hits.dwyl.io/apache/incubator-servicecomb-service-center.svg)](http://hits.dwyl.io/apache/incubator-servicecomb-service-center) [![Gitter](https://img.shields.io/badge/ServiceComb-Gitter-ff69b4.svg)](https://gitter.im/ServiceCombUsers/Lobby)  

Apache ServiceComb (incubating) Service-Center is a Restful based service-registry that provides micro-services discovery and micro-service management. It is based on Open API format and provides features like service-discovery, fault-tolerance, dynamic routing, notify subscription and scalable by design. It has high performance cache design and separate entity management for micro-services and their instances. It provides out of box support for metrics and tracing. It has a web portal to manage the micro-services.  

## Features
 - **[`Open API`](/server/core/swagger/v4.yaml)**: API doc(Open API format) management for microservice
 - **Metadata**: Metadata management for both microservice and microservice instance
 - **Dependency**: Microservice dependency management
 - **Seperated**: Seperated microservice and microservice instance entity management
 - **Domains**: Logical multiple domains management
 - **Security**: White and back list configuration for service discovery
 - **Discovery**: Support query instance by criteria 
 - **Subscribe**: Use web socket to notify client about instance change events
 - **[`Portal`](/frontend)**: Awesome web portal
 - **Fault tolerance**: Multiple fault tolerance mechanism and design in the architecture
 - **Performance**: Performance/Caching design
 - **[`Metrics`](/docs/integration-grafana.md)**: Able to expose Prometheus metric API automatically
 - **[`Tracing`](/docs/tracing.md)**: Able to report tracing data to Zipkin server
 - **[`Multi Datacenter`](/docs/multidcs.md)**: Additional layer of abstraction to clusters deployed in multiple datacenters
 - **[`Dynmaic Plug-in`](/docs/plugin.md)**: Able to load custom authentication, tls and other dynamic libraries
 - **[`CLI`](/scctl/pkg/plugin/README.md)**: Easy to control service center
 
## Documentation

Project documentation is available on the [ServiceComb website][servicecomb-website]. You can also find full document [`here`](/docs/README.md).

[servicecomb-website]: http://servicecomb.incubator.apache.org/

## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for Linux, Windows and Docker.

[github-release]: http://servicecomb.incubator.apache.org/release/

### Running Service Center using the Release

You can download our latest release from [ServiceComb Website][github-release].When you get these release, you can execute the start script to run Service Center.

Windows(apache-incubator-servicecomb-service-center-XXX-windows-amd64.zip):
```
start-service-center.bat
```

Linux(apache-incubator-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
```sh
./start-service-center.sh
```
Docker:
```sh
docker pull servicecomb/service-center
docker run -d -p 30100:30100 servicecomb/service-center
```

Note: The Releases of Service-Center uses emebeded etcd, if you want to use the seperate instance of etcd then you can deploy the etcd seperately and configure the etcd ip over here.
```
vi conf/app.conf

## Edit this file
# registry address
# 1. if registry_plugin equals to 'embeded_etcd'
# manager_name = "sc-0"
# manager_addr = "http://127.0.0.1:2380"
# manager_cluster = "sc-0=http://127.0.0.1:2380"
# 2. if registry_plugin equals to 'etcd'
# manager_cluster = "127.0.0.1:2379"
manager_cluster = "127.0.0.1:2379"
```

By default the SC comes up on 127.0.0.1:30100, however you can change the configuration of these address over here.

```
vi conf/app.conf

httpaddr = 127.0.0.1
httpport = 30100
```

### Building & Running Service-Center from source

Download the Code
```sh
git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
```

Dependencies

By default, we use glide to manage dependencies. If the go version greater then `go1.11`, 
you can download dependencies directly using command `go mod`. Please follow below steps to 
download all the dependency.

```sh
# greater then go1.11
GO111MODULE=on go mod download
GO111MODULE=on go mod vendor

# lower than go1.11
curl https://glide.sh/get | sh
glide install
```

Build the Service-Center

```sh
go build -o service-center
```

First, you need to run a etcd(version: 3.x) as a database service and then modify the etcd IP and port in the Service Center configuration file (./etc/conf/app.conf : manager_cluster).

```sh
wget https://github.com/coreos/etcd/releases/download/v3.1.8/etcd-v3.1.8-linux-amd64.tar.gz
tar -xvf etcd-v3.1.8-linux-amd64.tar.gz
cd etcd-v3.1.8-linux-amd64
./etcd

cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cp -r ./etc/conf .
./service-center
```
This will bring up Service Center listening on ip/port 127.0.0.1:30100 for service communication.If you want to change the listening ip/port, you can modify it in the Service Center configuration file (./conf/app.conf : httpaddr,httpport).

[github-release]: https://github.com/servicecomb/service-center/releases/

### Running Frontend using the Release

You can download our latest release from ServiceComb Website and then untar it and run start-frontend.sh/start-frontend.bat.
This will bring up the Service-Center UI on [http://127.0.0.1:30103](http://127.0.0.1:30103).

Windows(apache-incubator-servicecomb-service-center-XXX-windows-amd64.zip):
```
start-frontend.bat
```

Linux(apache-incubator-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
```sh
./start-frontend.sh
```

Note: By default frontend runs on 127.0.0.1, if you want to change this then you can change it in `conf/app.conf`. 
```
frontend_host_ip=127.0.0.1
frontend_host_port=30103
```

You can follow the guide over [here](frontend/Readme.md#running-ui-from-source-code) to run the Frontend from source.

## Get The Latest Release

[Download Service Center](http://servicecomb.incubator.apache.org/release/service-center-downloads/)

## Contact

Bugs: [issues](https://issues.apache.org/jira/browse/SCB)

## Contributing

See [Contribution guide](/docs/contribution.md) for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.

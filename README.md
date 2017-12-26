# service-center 
[![Build Status](https://www.travis-ci.org/apache/incubator-servicecomb-service-center.svg?branch=master)](https://www.travis-ci.org/apache/incubator-servicecomb-service-center)  [![Coverage Status](https://coveralls.io/repos/github/apache/incubator-servicecomb-service-center/badge.svg?branch=master)](https://coveralls.io/github/apache/incubator-servicecomb-service-center?branch=master)  [![Go Report Card](https://goreportcard.com/badge/github.com/apache/incubator-servicecomb-service-center)](https://goreportcard.com/report/github.com/apache/incubator-servicecomb-service-center) [![GoDoc](https://godoc.org/github.com/apache/incubator-servicecomb-service-center?status.svg)](https://godoc.org/github.com/apache/incubator-servicecomb-service-center)  

Apache ServiceComb (incubating) service-center allows services to register their instance information and to discover providers of a given service. 
## Features
 -  Seperated microservice and microservice instance entity management
 -  White and back list configuration for service discovery
 -  Use web socket to notify client about instance change events
 -  Support query instance by criteria
 -  Metadata management for both microservice and microservice instance 
 -  API doc(Open API format) management for microservice
 -  Microservice dependency management
 -  Awesome  [web portal](/frontend)
 -	 Multiple fault tolerance mechanism and design in the architecture
 -	 Performance/Caching design
 - **Metrics**: Able to expose Prometheus metric API automatically

## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for Linux, Windows and Docker. Instructions for using these binaries are on the [GitHub releases page][github-release].

[github-release]: https://github.com/servicecomb/service-center/releases/

### Building and Running Service Center

You don't need to build from source to use Service Center (binaries on the [GitHub releases page][github-release]).When you get these binaries, you can execute the start script to run Service Center.

Windows(service-center-xxx-windows-amd64.zip):
```
start.bat
```

Linux(service-center-xxx-linux-amd64.tar.gz):
```sh
./start.sh
```
Docker:
```sh
docker pull servicecomb/service-center
docker run -d -p 30100:30100 servicecomb/service-center
```


##### If you want to try out the latest and greatest, Service Center can be easily built. 

Download the Code
```sh
git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
```

Dependencies

We use gvt for dependency management, please follow below steps to download all the dependency.
```sh
go get github.com/FiloSottile/gvt
gvt restore
```
If you face any issue in downloading the dependency because of insecure connection then you can use ```gvt restore -precaire```

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

## Documentation

Project documentation is available on the [ServiceComb website][servicecomb-website]. You can also find some development guide [here](/docs)

[servicecomb-website]: http://servicecomb.io/
      
## Contact

Bugs: [issues](https://github.com/servicecomb/service-center/issues)

## Contributing

See [Contribution guide](/docs/contribution.md) for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.

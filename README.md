# service-center [![Build Status](https://travis-ci.org/ServiceComb/service-center.svg?branch=master)](https://travis-ci.org/ServiceComb/service-center)

A standalone service center to allow services to register their instance information and to discover providers of a given service

## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for OSX, Linux, Windows, and Docker. Instructions for using these binaries are on the [GitHub releases page][github-release].

[github-release]: https://github.com/servicecomb/service-center/releases/

### Building and Running Service Center

You don't need to build from source to use Service Center (binaries on the [GitHub releases page][github-release]), but if you want to try out the latest and greatest, Service Center can be easily built. 

First, you need to run a etcd(version: 3.x) as a database service and then modify the etcd IP and port in the Service Center configuration file (./conf/app.conf : manager_cluster).

```sh
wget https://github.com/coreos/etcd/releases/download/v3.1.8/etcd-v3.1.8-linux-amd64.tar.gz
tar -xvf etcd-v3.1.8-linux-amd64.tar.gz
cd etcd-v3.1.8-linux-amd64
./etcd

go get github.com/ServiceComb/service-center
cd $GOPATH/src/github.com/ServiceComb/service-center
go build
cp -r ./etc/conf .
./service-center
```
This will bring up Service Center listening on port 30100 for service communication.

[github-release]: https://github.com/servicecomb/service-center/releases/

## Documentation

Project documentation is available on the ServiceComb website.

## Automated Testing

      
## Contact

Mailing list: 

Planning/Roadmap: milestones, roadmap

Bugs: issues

## Contributing

See CONTRIBUTING for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.

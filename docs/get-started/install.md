## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for Linux, Windows and Docker.

[github-release]: http://servicecomb.apache.org/release/

### Running Service Center using the Release

You can download our latest release from [ServiceComb Website][github-release].When you get these release, you can execute the start script to run Service Center.

Windows(apache-servicecomb-service-center-XXX-windows-amd64.zip):
```
start-service-center.bat
```

Linux(apache-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
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

Requirements

+ [Go](https://golang.org) version 1.8+ is required to build the latest version of Service-Center.

Download the Code
```sh
git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
cd $GOPATH/src/github.com/apache/servicecomb-service-center
```

Dependencies

you can download dependencies directly using command `go mod`. Please follow below steps to
download all the dependency.

```sh
# greater then go1.11
GO111MODULE=on go mod download
GO111MODULE=on go mod vendor
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

cd $GOPATH/src/github.com/apache/servicecomb-service-center
cp -r ./etc/conf .
./service-center
```
This will bring up Service Center listening on ip/port 127.0.0.1:30100 for service communication.If you want to change the listening ip/port, you can modify it in the Service Center configuration file (./conf/app.conf : httpaddr,httpport).

[github-release]: https://github.com/servicecomb/service-center/releases/

### Running Frontend using the Release

You can download our latest release from ServiceComb Website and then untar it and run start-frontend.sh/start-frontend.bat.
This will bring up the Service-Center UI on [http://127.0.0.1:30103](http://127.0.0.1:30103).

Windows(apache-servicecomb-service-center-XXX-windows-amd64.zip):
```
start-frontend.bat
```

Linux(apache-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
```sh
./start-frontend.sh
```

Note: By default frontend runs on 127.0.0.1, if you want to change this then you can change it in `conf/app.conf`.
```
frontend_host_ip=127.0.0.1
frontend_host_port=30103
```

You can follow the guide over [here](https://github.com/apache/servicecomb-service-center/blob/master/frontend/Readme.md#running-ui-from-source-code) to run the Frontend from source.
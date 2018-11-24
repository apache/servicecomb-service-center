## scctl

`scctl` enables user to view the list of MicroServices registered in service center(version 1.1.0+).
You can view all the commands from [here](/scctl/pkg/plugin/README.md)

### QuickStart Guide

##### Install
Easiest way to get started with `scctl` is to download the release 
from [`here`](http://servicecomb.apache.org/release/)
and then untar/unzip it based on your OS.

##### Check the version
Windows(apache-servicecomb-service-center-XXX-windows-amd64.zip):
```
scctl.exe version
```

Linux(apache-servicecomb-service-center-XXXX-linux-amd64.tar.gz):
```sh
./scctl version
```

Note: If you already bootstrap SC and listen on `127.0.0.1:30100`, this
command will also print the SC version.

### Running scctl from source code

Requirements

+ [Go](https://golang.org) version 1.8+ is required to build the latest version of scctl.

However if you want to try our latest code then you can follow the below steps
```
#Make sure your GOPATH is set correctly and download all the vendors of SC
git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
cd $GOPATH/src/github.com/apache/servicecomb-service-center

cd scctl

go build

```
Windows:
```
scctl.exe version
```

Linux:
```sh
./scctl version
```

#!/bin/bash

set -e

SCRIPT_DIR=$(cd $(dirname $0); pwd)
BASE_DIR=${SCRIPT_DIR}/../../../

mkdir -p $SCRIPT_DIR/service-center

cd $BASE_DIR

# make CGO_ENABLED=0 since busybox will not execute if it is dynamically linked
export CGO_ENABLED=0
export GOOS="linux"
export GOARCH="amd64"

# buils service-center
go build -o $SCRIPT_DIR/service-center/service-center

# give permission
chmod 755 $SCRIPT_DIR/service-center/service-center

#go to the script directory
cd $SCRIPT_DIR

# copy the conf folder to build-image/service-center 
cp -r $BASE_DIR/etc/conf service-center
sed -i "" "s|^registry_plugin.*=.*$|registry_plugin = embeded_etcd|g" service-center/conf/app.conf
sed -i "" "s|^httpaddr.*=.*$|httpaddr = 0.0.0.0|g" service-center/conf/app.conf
sed -i "" "s|\(^manager_cluster.*=.*$\)|# \1|g" service-center/conf/app.conf

# store the etcd image name  and version
BASE_IMAGE=quay.io/coreos/etcd
BASE_IMAGE_VERSION=v3.1.9

cp Dockerfile.tmpl Dockerfile

sed -i "" "s|{{.BASE_IMAGE}}|${BASE_IMAGE}|g" Dockerfile
sed -i "" "s|{{.BASE_IMAGE_VERSION}}|${BASE_IMAGE_VERSION}|g" Dockerfile

docker build -t developement/servicecomb/service-center:latest .
docker save developement/servicecomb/service-center:latest |gzip >service-center-dev.tgz

# remove the service-center directory from the build-image path
rm -rf service-center Dockerfile 

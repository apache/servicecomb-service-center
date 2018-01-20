#!/bin/bash

set -e

umask 027

SCRIPT_DIR=$(cd $(dirname $0); pwd)
BASE_DIR=${SCRIPT_DIR}/../../..

mkdir -p $SCRIPT_DIR/frontend

cd $BASE_DIR/frontend

# make CGO_ENABLED=0 since busybox will not execute if it is dynamically linked
export CGO_ENABLED=0
export GOOS="linux"
export GOARCH="amd64"

# buils scfrontend
go build -o $SCRIPT_DIR/frontend/scfrontend

#go to the script directory
#cd $SCRIPT_DIR

# copy the app conf folders to build-frontend-image/frontend 
cp -r app conf $BASE_DIR/scripts/frontend/start_linux.sh $SCRIPT_DIR/frontend

chmod 755 $SCRIPT_DIR/frontend/start_linux.sh $SCRIPT_DIR/frontend/scfrontend

sed -i "s|FRONTEND_HOST_IP=127.0.0.1|FRONTEND_HOST_IP=0.0.0.0|g" $SCRIPT_DIR/frontend/conf/app.conf

#go to the script directory
cd $SCRIPT_DIR
tar cf frontend.tar.gz frontend
cp Dockerfile.tmpl Dockerfile

docker build -t servicecomb/scfrontend .
docker save servicecomb/scfrontend:latest |gzip >scfrontend-dev.tgz

# remove the frontend directory from the build-frontend-image path
rm -rf frontend Dockerfile frontend.tar.gz

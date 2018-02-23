# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#!/bin/bash

set -e

umask 027

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

#go to the script directory
cd $SCRIPT_DIR

# copy the conf folder to build-image/service-center 
cp -r $BASE_DIR/etc/conf start.sh service-center

chmod 500 service-center/start.sh service-center/service-center

sed -i "s|^registry_plugin.*=.*$|registry_plugin = embeded_etcd|g" service-center/conf/app.conf
sed -i "s|^httpaddr.*=.*$|httpaddr = 0.0.0.0|g" service-center/conf/app.conf
sed -i "s|\(^manager_cluster.*=.*$\)|# \1|g" service-center/conf/app.conf

# store the etcd image name  and version
BASE_IMAGE=quay.io/coreos/etcd
BASE_IMAGE_VERSION=v3.1.9

cp Dockerfile.tmpl Dockerfile

sed -i "s|{{.BASE_IMAGE}}|${BASE_IMAGE}|g" Dockerfile
sed -i "s|{{.BASE_IMAGE_VERSION}}|${BASE_IMAGE_VERSION}|g" Dockerfile

docker build -t developement/servicecomb/service-center:latest .
docker save developement/servicecomb/service-center:latest |gzip >service-center-dev.tgz

# remove the service-center directory from the build-image path
rm -rf service-center Dockerfile

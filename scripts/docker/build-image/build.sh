#!/usr/bin/env bash

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

set -e

umask 027

SCRIPT_DIR=$(cd $(dirname $0); pwd)

BASE_DIR=${SCRIPT_DIR}/../../../

# build all
export RELEASE=${1:-"latest"}

source ${SCRIPT_DIR}/../../build/tools.sh

PACKAGE_DIR=$SCRIPT_DIR/../$PACKAGE_PREFIX-$RELEASE-linux-amd64
if [ ! -d $PACKAGE_DIR ]; then
    docker_builder_pattern $BASE_DIR $SCRIPT_DIR/../
fi

# create image base dir
mkdir -p $SCRIPT_DIR/service-center

# go to the script directory
cd $SCRIPT_DIR

# copy the conf folder to build-image/service-center
cp -rp $PACKAGE_DIR/conf $PACKAGE_DIR/service-center start.sh service-center

chmod 500 service-center/start.sh service-center/service-center

# store the base image name and version
BASE_IMAGE=${BASE_IMAGE:-"alpine"}

BASE_IMAGE_VERSION=${BASE_IMAGE_VERSION:-"latest"}

cp Dockerfile.tmpl Dockerfile

sed -i "s|{{.BASE_IMAGE}}|${BASE_IMAGE}|g" Dockerfile
sed -i "s|{{.BASE_IMAGE_VERSION}}|${BASE_IMAGE_VERSION}|g" Dockerfile

docker build --no-cache -t servicecomb/service-center:$RELEASE .
docker save servicecomb/service-center:$RELEASE |gzip >service-center-dev.tgz

# remove the service-center directory from the build-image path
rm -rf service-center Dockerfile

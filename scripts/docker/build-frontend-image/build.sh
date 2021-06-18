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
export BUILD="ALL"

source ${SCRIPT_DIR}/../../build/tools.sh

PACKAGE_DIR=$SCRIPT_DIR/../$PACKAGE_PREFIX-$RELEASE-linux-amd64
if [ ! -d $PACKAGE_DIR ]; then
    docker_builder_pattern $BASE_DIR $SCRIPT_DIR/../
fi

# create image base dir
mkdir -p $SCRIPT_DIR/frontend

# go to the script directory
cd $SCRIPT_DIR

# copy the conf folder to build-frontend-image/frontend
cp -rp $PACKAGE_DIR/app $PACKAGE_DIR/conf $PACKAGE_DIR/frontend start.sh frontend
rm -rf frontend/app/bower_components

chmod 500 frontend/start.sh frontend/frontend

# store the base image name and version
BASE_IMAGE=${BASE_IMAGE:-"alpine"}

BASE_IMAGE_VERSION=${BASE_IMAGE_VERSION:-"latest"}

cp Dockerfile.tmpl Dockerfile

sed -i "s|{{.BASE_IMAGE}}|${BASE_IMAGE}|g" Dockerfile
sed -i "s|{{.BASE_IMAGE_VERSION}}|${BASE_IMAGE_VERSION}|g" Dockerfile

docker build --no-cache -t servicecomb/scfrontend:$RELEASE .
docker save servicecomb/scfrontend:$RELEASE |gzip >scfrontend-dev.tgz

# remove the frontend directory from the build-frontend-image path
rm -rf frontend Dockerfile

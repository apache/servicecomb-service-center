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

sed -i "s|frontend_host_ip=127.0.0.1|frontend_host_ip=0.0.0.0|g" $SCRIPT_DIR/frontend/conf/app.conf

#go to the script directory
cd $SCRIPT_DIR
tar cf frontend.tar.gz frontend
cp Dockerfile.tmpl Dockerfile

docker build -t servicecomb/scfrontend .
docker save servicecomb/scfrontend:latest |gzip >scfrontend-dev.tgz

# remove the frontend directory from the build-frontend-image path
rm -rf frontend Dockerfile frontend.tar.gz

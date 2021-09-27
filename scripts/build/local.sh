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
export GOPROXY=${GOPROXY:-"https://goproxy.cn,direct"}
export GOOS=${1:-"linux"}
export GOARCH=${4:-"amd64"}
export CGO_ENABLED=${CGO_ENABLED:-0} # prevent to compile cgo file
export GO_EXTLINK_ENABLED=${GO_EXTLINK_ENABLED:-0} # do not use host linker
export GO_LDFLAGS=${GO_LDFLAGS:-" -s -w"}

RELEASE=${2:-"0.0.1"}

PACKAGE=${3:-"${RELEASE}"}

PACKAGE_PREFIX=apache-servicecomb-service-center

script_path=$(cd "$(dirname "$0")"; pwd)
fail() {
	echo "$1"
	exit 1
}


build_service_center() {
    local app=$PACKAGE_PREFIX-$PACKAGE-$GOOS-$GOARCH

    set +e
    rm -rf $app
    rm -rf $app.tar.gz

    set -e
    mkdir -p $app

    ## Build the Service-Center releases
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    local ldflags="${GO_LDFLAGS} -X 'github.com/apache/servicecomb-service-center/version.BuildTag=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    ldflags="${ldflags} -X 'github.com/apache/servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
    local BINARY_NAME=$app/service-center
    if [ "$GOOS" == "windows" ]; then
        BINARY_NAME=${BINARY_NAME}.exe
    fi
    go build --ldflags "${ldflags}" -o $BINARY_NAME github.com/apache/servicecomb-service-center/cmd/scserver
}


build_scctl() {
    local app=../$PACKAGE_PREFIX-$PACKAGE-$GOOS-$GOARCH
    ## Build the scctl releases
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    local ldflags="${GO_LDFLAGS} -X 'github.com/apache/servicecomb-service-center/scctl/pkg/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    ldflags="${ldflags} -X 'github.com/apache/servicecomb-service-center/scctl/pkg/version.VERSION=$BUILD_NUMBER'"

    local BINARY_NAME=$app/scctl
    if [ "$GOOS" == "windows" ]; then
        BINARY_NAME=${BINARY_NAME}.exe
    fi
    go build --ldflags "${ldflags}" -o $BINARY_NAME github.com/apache/servicecomb-service-center/cmd/scctl
}

build_syncer() {
    local app=../$PACKAGE_PREFIX-$PACKAGE-$GOOS-$GOARCH
    ## Build the syncer releases
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    local ldflags="${GO_LDFLAGS} -X 'github.com/apache/servicecomb-service-center/syncer/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    ldflags="${ldflags} -X 'github.com/apache/servicecomb-service-center/syncer/version.VERSION=$BUILD_NUMBER'"

    local BINARY_NAME=$app/syncer
    if [ "$GOOS" == "windows" ]; then
        BINARY_NAME=${BINARY_NAME}.exe
    fi
    go build --ldflags "${ldflags}" -o $BINARY_NAME github.com/apache/servicecomb-service-center/cmd/syncer
}



## Prepare the Configuration and Make package
package() {
    local app=$PACKAGE_PREFIX-$PACKAGE-$GOOS-$GOARCH

    cp -r etc/conf $app/
    sed -i 's/^manager_cluster.*=.*/manager_name = \"sr-0\"\nmanager_addr = \"http:\/\/127.0.0.1:2380\"\nmanager_cluster = \"sr-0=http:\/\/127.0.0.1:2380\"/g' $app/conf/app.conf
    sed -i 's/^registry_plugin.*=.*/registry_plugin = embeded_etcd/g' $app/conf/app.conf

    ## Copy the Service-Center Releases
    cp -r scripts/release/LICENSE $app/
    cp -r scripts/release/licenses $app/
    cp -r scripts/release/NOTICE $app/
    cp -r README.md $app/

    ## Copy the frontend releases
    cp -r frontend/app $app/

    ## Copy Start Scripts
    cp -r scripts/release/start_scripts/$GOOS/* $app/
    if [ "$GOOS" != "windows" ]; then
        chmod +x $app/*.sh
    fi

    ## Archive the release
    tar -czvf $app.tar.gz $app
}

docker_builder_pattern() {
    local dockerfile_dir=${1:-"."}
    local output=${2:-"."}
    local builder_name=servicecomb/service-center:build
    local builder_path=/go/src/github.com/apache/servicecomb-service-center
    local app=$PACKAGE_PREFIX-$PACKAGE-linux-amd64

    set +e

    docker rmi $builder_name

    set -e

    cd $dockerfile_dir
    docker build -t $builder_name . -f Dockerfile.build
    docker create --name builder $builder_name
    docker cp builder:$builder_path/$app $output
    docker rm -f builder
}
personal_build() {

    build_service_center

    build_scctl

    build_syncer

    package
}

personal_build


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

fail() {
	echo "$1"
	exit 1
}

install_glide() {
    set +e
    GLIDE=$(which glide)
    if [ "$?" == "1" ]; then
        set -e
        curl https://glide.sh/get | sh
    else
        set -e
    fi
}

install_bower() {
    set +e
    BOWER=$(which bower)
    if [ "$?" == "1" ]; then
        set -e

        curl -sL https://deb.nodesource.com/setup_8.x | bash -
        apt-get install -y nodejs

        npm install -g bower
    else
        set -e
    fi
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
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"

    local BINARY_NAME=$app/service-center
    if [ "$GOOS" == "windows" ]; then
        BINARY_NAME=${BINARY_NAME}.exe
    fi
    go build --ldflags "${GO_LDFLAGS}" -o $BINARY_NAME
}

build_frontend() {
    ## Build Frontend Release
    cd frontend
    local BINARY_NAME=$PACKAGE_PREFIX-$PACKAGE-$GOOS-$GOARCH/frontend
    if [ "$GOOS" == "windows" ]; then
        BINARY_NAME=${BINARY_NAME}.exe
    fi
    go build -o ../$BINARY_NAME
    cd ..
}

sc_deps() {
    install_glide

    glide install
}

frontend_deps() {
    install_bower

    ## Download the frontend dependencies using bower
    cd frontend/app
    bower install --allow-root
    cd ../..
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
    cp -r DISCLAIMER $app/
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
    local builder_path=/go/src/github.com/apache/incubator-servicecomb-service-center
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
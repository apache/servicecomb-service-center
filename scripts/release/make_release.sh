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

## Get the Release Number
if [ $2 == "" ]; then
    echo "Invalid version number....exiting...."
    exit 1
else
    RELEASE=$2
fi

#Package prefix for the release directory
PACKAGE_PREFIX=apache-incubator-servicecomb-service-center

## Get the PACKAGE NUMBER
if [ $3 == "" ]; then
    PACKAGE=RELEASE
else
    PACKAGE=$3
fi

## Get the OS Version
case $1 in
 linux )
    OSNAME=linux ;;

 windows )
    OSNAME=windows ;;

 all )
    OSNAME=all ;;

 * )
    echo "Wrong OS Version....exiting....."
    exit 1

esac

## Prepare the Configuration
prepare_conf() {
    set +e
    rm -rf tmp

    set -e
    mkdir tmp
    cp -r etc/conf tmp/
    sed -i 's/# manager_name = \"sc-0\"/manager_name = \"sr-0\"/g' tmp/conf/app.conf
    sed -i 's/# manager_addr = \"http:\/\/127.0.0.1:2380\"/manager_addr = \"http:\/\/127.0.0.1:2380\"/g' tmp/conf/app.conf
    sed -i 's/# manager_cluster = \"sc-0=http:\/\/127.0.0.1:2380\"/manager_cluster = \"sr-0=http:\/\/127.0.0.1:2380\"/g' tmp/conf/app.conf
    sed -i 's/manager_cluster = \"127.0.0.1:2379\"/# manager_cluster = \"127.0.0.1:2379\"/g' tmp/conf/app.conf
    #sed -i s@"manager_cluster.*=.*$"@"manager_name = \"sr-0\"\nmanager_addr = \"http://127.0.0.1:2380\"\nmanager_cluster = \"sr-0=http://127.0.0.1:2380\""@g tmp/conf/app.conf
    sed -i 's/registry_plugin = etcd/registry_plugin = embeded_etcd/g' tmp/conf/app.conf
}

# Build Linux Release
build_linux(){
    if [ $RELEASE == "" ] ; then
         echo "Error in Making Linux Release.....Release Number not specified"
    fi
    if [ $PACKAGE = "" ]; then
        echo "Error in Making Linux Release.....Package Number not specified"
    fi

    set +e
    rm -rf $PACKAGE_PREFIX-$PACKAGE-linux-amd64
    rm -rf $PACKAGE_PREFIX-$PACKAGE-linux-amd64.tar.gz

    set -e
    mkdir -p $PACKAGE_PREFIX-$PACKAGE-linux-amd64

    ## Build the Service-Center releases
    export GOOS=linux
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
    go build --ldflags "${GO_LDFLAGS}" -o apache-incubator-servicecomb-service-center
    cp -r apache-incubator-servicecomb-service-center $PACKAGE_PREFIX-$PACKAGE-linux-amd64

    ## Build Frontend Release
    cd frontend
    go build -o apache-incubator-serviceomb-frontend
    cp -r apache-incubator-serviceomb-frontend ../$PACKAGE_PREFIX-$PACKAGE-linux-amd64
    cd ..

    prepare_conf

    ## Copy the Service-Center Releases
    cp -r tmp/conf $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    cp -r scripts/release/LICENSE $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    cp -r scripts/release/licenses $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    cp -r scripts/release/NOTICE $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    cp -r DISCLAIMER $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    cp -r README.md $PACKAGE_PREFIX-$PACKAGE-linux-amd64/

    ## Copy the frontend releases
    cp -r frontend/app $PACKAGE_PREFIX-$PACKAGE-linux-amd64/

    ## Copy Start Scripts
    cp -r scripts/release/start_scripts/linux/* $PACKAGE_PREFIX-$PACKAGE-linux-amd64/
    chmod +x $PACKAGE_PREFIX-$PACKAGE-linux-amd64/*.sh

    ## Archive the release
    tar -czvf $PACKAGE_PREFIX-$PACKAGE-linux-amd64.tar.gz $PACKAGE_PREFIX-$PACKAGE-linux-amd64

}

# Build Windows Release
build_windows(){
    if [ $RELEASE == "" ] ; then
         echo "Error in Making Windows Release.....Release Number not specified"
    fi
    if [ $PACKAGE = "" ]; then
        echo "Error in Making Windows Release.....Package Number not specified"
    fi

    set +e
    rm -rf $PACKAGE_PREFIX-$PACKAGE-windows-amd64
    rm -rf $PACKAGE_PREFIX-$PACKAGE-windows-amd64.zip

    set -e
    mkdir -p $PACKAGE_PREFIX-$PACKAGE-windows-amd64

    ## Build Service-Center Release
    export GOOS=windows
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
    go build --ldflags "${GO_LDFLAGS}" -o apache-incubator-servicecomb-service-center.exe
    cp -r apache-incubator-servicecomb-service-center.exe $PACKAGE_PREFIX-$PACKAGE-windows-amd64

    ## Build Frontend release
    cd frontend
    go build -o apache-incubator-serviceomb-frontend.exe
    cp -r apache-incubator-serviceomb-frontend.exe ../$PACKAGE_PREFIX-$PACKAGE-windows-amd64
    cd ..

    prepare_conf

    ## Copy the service-center releases
    cp -r tmp/conf $PACKAGE_PREFIX-$PACKAGE-windows-amd64/
    cp -r scripts/release/LICENSE $PACKAGE_PREFIX-$PACKAGE-windows-amd64/
    cp -r scripts/release/licenses $PACKAGE_PREFIX-$PACKAGE-windows-amd64/
    cp -r scripts/release/NOTICE $PACKAGE_PREFIX-$PACKAGE-windows-amd64/
    cp -r DISCLAIMER $PACKAGE_PREFIX-$PACKAGE-windows-amd64/
    cp -r README.md $PACKAGE_PREFIX-$PACKAGE-windows-amd64/

    ## Copy the Frontend releases
    cp -r frontend/app $PACKAGE_PREFIX-$PACKAGE-windows-amd64/

    ## Copy start scripts
    cp -r scripts/release/start_scripts/windows/* $PACKAGE_PREFIX-$PACKAGE-windows-amd64/

    ## Archive the Release
    tar -czvf $PACKAGE_PREFIX-$PACKAGE-windows-amd64.tar.gz $PACKAGE_PREFIX-$PACKAGE-windows-amd64
}

## Compile the binary
case $OSNAME in
 linux )
    build_linux ;;

 windows )
    build_windows ;;

 all )
    build_linux
    build_windows ;;

esac

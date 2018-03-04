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
    rm -rf apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64
    rm -rf apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64.tar.gz

    set -e
    mkdir -p apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64

    export GOOS=linux
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
    go build --ldflags "${GO_LDFLAGS}" -o apache-incubator-servicecomb-service-center
    cp -r apache-incubator-servicecomb-service-center apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64
    prepare_conf
    cp -r tmp/conf apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    echo "./apache-incubator-servicecomb-service-center > start-sc.log 2>&1 &" >> apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/start.sh
    echo "kill -9 \$(ps aux | grep 'apache-incubator-servicecomb-service-center' | awk '{print \$2}')" >> apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/stop.sh
    chmod +x apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/start.sh
    chmod +x apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/stop.sh
    cp -r scripts/release/LICENSE apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    cp -r scripts/release/licenses apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    cp -r scripts/release/NOTICE apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    cp -r DISCLAIMER apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    cp -r README.md apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64/
    tar -czvf apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64.tar.gz apache-incubator-servicecomb-service-center-$PACKAGE-linux-amd64

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
    rm -rf apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64
    rm -rf apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64.zip

    set -e
    mkdir -p apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64
    export GOOS=windows
    export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
    export BUILD_NUMBER=$RELEASE
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
    GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
    go build --ldflags "${GO_LDFLAGS}" -o apache-incubator-servicecomb-service-center.exe
    cp -r apache-incubator-servicecomb-service-center.exe apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64
    prepare_conf
    cp -r tmp/conf apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    echo "apache-incubator-servicecomb-service-center.exe" >> apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/start.bat
    cp -r scripts/release/LICENSE apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    cp -r scripts/release/licenses apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    cp -r scripts/release/NOTICE apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    cp -r DISCLAIMER apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    cp -r README.md apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64/
    tar -czvf apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64.tar.gz apache-incubator-servicecomb-service-center-$PACKAGE-windows-amd64
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

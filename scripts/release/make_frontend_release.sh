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
#!/usr/bin/env bash
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
    cp -r frontend/conf tmp/
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
    rm -rf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64
    rm -rf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64.tar.gz

    set -e
    mkdir -p apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64

    export GOOS=linux
    cd frontend
    go build -o apache-incubator-serviceomb-frontend
    cp -r apache-incubator-serviceomb-frontend ../apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64
    cd ..
    prepare_conf
    cp -r tmp/conf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    cp -r frontend/app apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    echo "./apache-incubator-serviceomb-frontend > start-sc-frontend.log 2>&1 &" >> apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/start.sh
    echo "kill -9 \$(ps aux | grep 'apache-incubator-serviceomb-frontend' | awk '{print \$2}')" >> apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/stop.sh
    chmod +x apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/start.sh
    chmod +x apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/stop.sh
    cp -r LICENSE apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    cp -r NOTICE apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    cp -r DISCLAIMER apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    cp -r frontend/Readme.md apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64/
    tar -czvf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64.tar.gz apache-incubator-servicecomb-frontend-service-center-$PACKAGE-linux-amd64

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
    rm -rf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64
    rm -rf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64.zip

    set -e
    mkdir -p apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64
    export GOOS=windows
    cd frontend
    go build -o apache-incubator-serviceomb-frontend.exe
    cp -r apache-incubator-serviceomb-frontend.exe ../apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64
    cd ..
    prepare_conf
    cp -r tmp/conf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    cp -r frontend/app apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    cp -r LICENSE apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    cp -r NOTICE apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    cp -r DISCLAIMER apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    cp -r frontend/Readme.md apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/
    echo "apache-incubator-serviceomb-frontend.exe" >> apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64/start.bat
    tar -czvf apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64.tar.gz apache-incubator-servicecomb-frontend-service-center-$PACKAGE-windows-amd64
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

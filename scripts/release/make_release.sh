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

## Get the Release Number
if [ $2 == "" ]; then
    echo "Invalid version number....exiting...."
    exit 1
else
    RELEASE=$2
fi

#Package prefix for the release directory
PACKAGE_PREFIX=apache-servicecomb-incubating-service-center

## Get the PACKAGE NUMBER
if [ "X"$3 == "X" ]; then
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

 mac )
    OSNAME=mac ;;

 all )
    OSNAME=all ;;

 * )
    echo "Wrong OS Version....exiting....."
    exit 1

esac

## Get the arch type
export GOARCH=${4:-"amd64"}
export CGO_ENABLED=${CGO_ENABLED:-0} # prevent to compile cgo file
export GO_EXTLINK_ENABLED=${GO_EXTLINK_ENABLED:-0} # do not use host linker
export GO_LDFLAGS=${GO_LDFLAGS:-"-linkmode 'external' -extldflags '-static' -s -w"}

root_path=$(cd "$(dirname "$0")"; pwd)

source ${root_path}/../build/tools.sh

build() {
    frontend_deps

    build_service_center

    build_frontend

    build_scctl

    package
}

# Build Linux Release
build_linux(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Linux Release.....Release Number not specified"
    fi
    if [ "X"$PACKAGE == "X" ]; then
        echo "Error in Making Linux Release.....Package Number not specified"
    fi

    export GOOS=linux

    build
}

# Build Windows Release
build_windows(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Windows Release.....Release Number not specified"
    fi
    if [ "X"$PACKAGE == "X" ]; then
        echo "Error in Making Windows Release.....Package Number not specified"
    fi

    export GOOS=windows

    build
}

# Build Mac Release
build_mac(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Mac Release.....Release Number not specified"
    fi
    if [ "X"$PACKAGE == "X" ]; then
        echo "Error in Making Mac Release.....Package Number not specified"
    fi

    export GOOS=darwin

    build
}

## Compile the binary
case $OSNAME in
 linux )
    build_linux ;;

 windows )
    build_windows ;;

 mac )
    build_mac ;;

 all )
    build_linux
    build_windows
    build_mac ;;

esac

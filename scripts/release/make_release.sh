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

## Get the Release Number
if [[ $2 == "" ]]; then
    echo "Invalid version number....exiting...."
    exit 1
else
    export RELEASE=$2
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
## build all components
export BUILD="ALL"

script_path=$(cd "$(dirname "$0")"; pwd)

source ${script_path}/../build/tools.sh

# Build Linux Release
build_linux(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Linux Release.....Release Number not specified"
    fi

    export GOOS=linux

    build
}

# Build Windows Release
build_windows(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Windows Release.....Release Number not specified"
    fi

    export GOOS=windows

    build
}

# Build Mac Release
build_mac(){
    if [ "X"$RELEASE == "X" ] ; then
         echo "Error in Making Mac Release.....Release Number not specified"
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

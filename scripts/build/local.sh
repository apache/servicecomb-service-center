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

export GOOS=${1:-"linux"}
export GOARCH=${4:-"amd64"}
export CGO_ENABLED=${CGO_ENABLED:-0} # prevent to compile cgo file
export GO_EXTLINK_ENABLED=${GO_EXTLINK_ENABLED:-0} # do not use host linker
export GO_LDFLAGS=${GO_LDFLAGS:-"-linkmode 'external' -extldflags '-static' -s -w"}

RELEASE=${2:-"0.0.1"}

PACKAGE=${3:-"${RELEASE}"}

PACKAGE_PREFIX=apache-servicecomb-incubating-service-center

script_path=$(cd "$(dirname "$0")"; pwd)

source ${script_path}/tools.sh

personal_build() {
    source ${script_path}/deps.sh

    build_service_center

    build_frontend

    build_scctl

    package
}

personal_build

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
CURRENT_PATH=$(cd $(dirname "$0");pwd)
ROOT_PATH=$(dirname "$CURRENT_PATH")
export GOPROXY=https://goproxy.cn,direct
export COVERAGE_PATH=$(pwd)

skip_dir=${1:-'vendor'}

run_test() {
  cd "$1"
  go test -cover -covermode atomic -coverprofile coverage.out
  if [ -f coverage.out ]; then
      sed '1d;$d' coverage.out >> "$ROOT_PATH"/coverage.txt
  fi
}

for d in $(go list -test -f '{{.Dir}}' all | grep servicecomb-service-center | grep -v "eventbase\\|integration\\|test\\|vendor\\|$skip_dir" | sort | uniq); do
  run_test "$d"
done


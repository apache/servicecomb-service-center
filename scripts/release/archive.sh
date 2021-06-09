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

NAME=${1}
VERSION=${2}
USER=${3}

sign() {
  local name=${1}
  local user=${2}
  gpg -u ${user} --armor --output ${name}.tar.gz.asc --detach-sign ${name}.tar.gz
  gpg --verify ${name}.tar.gz.asc ${name}.tar.gz
  sha512sum ${name}.tar.gz >${name}.tar.gz.sha512
  sha512sum --check ${name}.tar.gz.sha512
}

archive() {
  local name=${1}
  local version=${2}
  git archive --format=tar v${version} --prefix=${name}-src/ | gzip >${name}-src.tar.gz
}

# archive source

main() {
  local release=${NAME}-${VERSION}
  local user=${USER}

  echo "archive source ${release} ..."
  archive "${release}" "${VERSION}"

  for pkg in $(ls ${release}-*.tar.gz); do
    echo "sign package ${pkg} by ${user} ..."
    sign "${pkg%.tar.gz}" "${user}"
  done
}

main

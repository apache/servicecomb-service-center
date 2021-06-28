#!/bin/sh

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

cd /opt/service-center

export SERVER_HOST="$(hostname)"
export LOG_FILE=${LOG_FILE:-''}
export LOG_LEVEL=${LOG_LEVEL:-'DEBUG'}
export gov_kie_type=kie
export gov_kie_endpoint=http://kie:30110
if [ -z "${BACKEND_ADDRESS}" ]; then
  export REGISTRY_KIND=${REGISTRY_KIND:-'embedded_etcd'}
else
  export REGISTRY_KIND=${REGISTRY_KIND:-'etcd'}
  export REGISTRY_ETCD_CLUSTER_ENDPOINTS=${BACKEND_ADDRESS}
fi

./service-center
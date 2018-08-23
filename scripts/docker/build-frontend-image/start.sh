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

cd /opt/frontend

sed -i "s/^frontend_host_ip.*=.*$/frontend_host_ip = $(hostname)/g" conf/app.conf

sc_ip_port=${SC_ADDRESS#*//}
sc_ip=${sc_ip_port%:*}
sc_port=${sc_ip_port#*:}

if [ ! -z "${sc_ip}" ]; then
    sed -i "s|^httpaddr.*=.*$|httpaddr = ${sc_ip}|g" conf/app.conf
fi
if [ "X"${sc_port} != "X"${sc_ip} ]; then
    sed -i "s|^httpport.*=.*$|httpport = ${sc_port}|g" conf/app.conf
fi

./frontend
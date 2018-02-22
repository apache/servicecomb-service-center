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
#!/bin/bash

cd /opt/frontend/
cd conf/
. app.conf
echo $SC_HOST_IP
echo $SC_HOST_PORT
echo $SC_HOST_MODE
cd ../
sed -i '/ip/c\ip:"'$SC_HOST_MODE'://'$SC_HOST_IP'",' ./app/apiList/apiList.js
sed -i '/port/c\port:"'$SC_HOST_PORT'"' ./app/apiList/apiList.js
./scfrontend > start-sc-frontend.log 2>&1

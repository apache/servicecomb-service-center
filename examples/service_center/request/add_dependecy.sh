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
PUT /v4/default/registry/dependencies HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: e924c839-df99-5526-5c30-f7ca9a671316

{
    "dependencies": [
        {
            "consumer": {
                "appId": "TestApp8",
                "serviceName": "Test8",
                "version": "1.0.0"
            },
            "providers": [
                {
                     "serviceName": "*",
                     "appId": "TestApp3",
                     "version":"1.0.0"
                }
            ]
        }
    ]
}

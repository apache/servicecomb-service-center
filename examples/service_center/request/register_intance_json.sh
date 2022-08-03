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

POST /v4/default/registry/microservices/2/instances HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
X-Project-Name: default
Cache-Control: no-cache
Postman-Token: bf33f47f-acfe-76fe-8c53-d1f79a46b246

{
	"instance": 
	{
	    "endpoints": [
			"grpc://127.0.0.1:99841"
		],
		"virtualAddress": "xxx.xxx.xxx.local:8080"
		"hostName":"ase",
		"status":"UP",
		"environment":"production",
		"properties": {
			"_TAGS": "A, B",
			"attr1": "a",
			"nodeIP": "one"
		},
		"healthCheck": {
			"mode": "push",
			"interval": 30,
			"times": 2
		}
	}
}

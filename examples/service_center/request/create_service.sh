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
POST /v4/default/registry/microservices HTTP/1.1
Host: localhost:30100
Content-Type: application/json
x-domain-name: default
Cache-Control: no-cache
Postman-Token: 8378df39-dfac-4f9d-205b-981e367c6b05

{
	"service":
	{
		"serviceName": "Test10",
		"appId": "TestApp10",
		"version":"11.0.0",
		"asdasddescription":"examplfsdsfe",
		"level": "FRONT",
		"schemas": [
			"TestService1212"
		],
		"status": "DOWN",
		"properties": {
			"attr2": "a"
		},
		"description":"中文asdfadads。”"
	}
}

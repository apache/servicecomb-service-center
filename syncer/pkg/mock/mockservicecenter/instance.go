/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mockservicecenter

import (
	"net/http"
)

func (m *mockServer) DiscoveryInstances(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{
  "instances": [
    {
      "instanceId": "7a6be9f861a811e9b3f6fa163eca30e0",
      "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
      "endpoints": [
        "rest://192.168.88.75:30100/"
      ],
      "hostName": "chenzhu",
      "status": "UP",
      "healthCheck": {
        "mode": "push",
        "interval": 30,
        "times": 3
      },
      "timestamp": "1555571184",
      "modTimestamp": "1555571184",
      "version": "0.0.1"
    },
    {
      "instanceId": "8e0fe4b961a811e981a6fa163e86b81a",
      "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
      "endpoints": [
        "rest://192.168.88.109:30100/"
      ],
      "hostName": "sunlisen",
      "status": "UP",
      "healthCheck": {
        "mode": "push",
        "interval": 30,
        "times": 3
      },
      "timestamp": "1555571221",
      "modTimestamp": "1555571221",
      "version": "0.0.1"
    }
  ]
}`))
}

func (m *mockServer) RegisterInstance(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{"instanceId": "8e0fe4b961a811e981a6fa163e86b81a"}`))
}

func (m *mockServer) UnregisterInstance(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{"instanceId": "8e0fe4b961a811e981a6fa163e86b81a"}`))
}

func (m *mockServer) Heartbeat(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{"instanceId": "8e0fe4b961a811e981a6fa163e86b81a"}`))
}

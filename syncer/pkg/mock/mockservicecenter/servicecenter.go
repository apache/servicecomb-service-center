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
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/apache/servicecomb-service-center/server/interceptor"

	"github.com/apache/servicecomb-service-center/pkg/rest"
)

const (
	WantStatus  = "X-Want-Status"
	WantContent = "X-Want-Content"
)

func NewMockServer() *httptest.Server {
	ms := &mockServer{}
	interceptor.RegisterInterceptFunc(ms.WantHandler)
	rest.RegisterServant(ms)
	return httptest.NewServer(rest.GetRouter())
}

type mockServer struct{}

func (m *mockServer) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/admin/dump", m.GetAll},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/existence", m.ServiceExistence},
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/microservices", m.CreateService},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices/:serviceId", m.DeleteService},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/instances", m.DiscoveryInstances},
		{rest.HTTP_METHOD_POST, "/v4/:project/registry/microservices/:serviceId/instances", m.RegisterInstance},
		{rest.HTTP_METHOD_DELETE, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId", m.UnregisterInstance},
		{rest.HTTP_METHOD_PUT, "/v4/:project/registry/microservices/:serviceId/instances/:instanceId/heartbeat", m.Heartbeat},
	}
}

func (m *mockServer) WantHandler(rw http.ResponseWriter, req *http.Request) error {
	statusCode, err := strconv.Atoi(req.Header.Get(WantStatus))
	if err != nil || statusCode == 0 {
		statusCode = http.StatusOK
	}
	rw.WriteHeader(statusCode)
	data := req.Header.Get(WantContent)
	if data == "" {
		return nil
	}
	rw.Write([]byte(data))
	return errors.New(data)
}

func (m *mockServer) GetAll(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{
  "cache": {
    "services": [
      {
        "key": "/cse-sr/ms/files/default/default/4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
        "rev": 5,
        "cluster": "sr-0",
        "value": {
          "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd",
          "appId": "default",
          "serviceName": "SERVICECENTER",
          "version": "0.0.1",
          "level": "BACK",
          "schemas": [
            "servicecenter.grpc.api.ServiceCtrl",
            "servicecenter.grpc.api.ServiceInstanceCtrl"
          ],
          "status": "UP",
          "properties": {
            "allowCrossApp": "true"
          },
          "timestamp": "1555571184",
          "alias": "SERVICECENTER",
          "modTimestamp": "1555571184",
          "environment": "production"
        }
      }
    ],
    "serviceIndexes": [
      {
        "key": "/cse-sr/ms/indexes/default/default/production/default/SERVICECENTER/0.0.1",
        "rev": 5,
        "cluster": "sr-0",
        "value": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd"
      }
    ],
    "serviceAliases": [
      {
        "key": "/cse-sr/ms/alias/default/default/production/default/SERVICECENTER/0.0.1",
        "rev": 5,
        "cluster": "sr-0",
        "value": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd"
      }
    ],
    "instances": [
      {
        "key": "/cse-sr/inst/files/default/default/4042a6a3e5a2893698ae363ea99a69eb63fc51cd/7a6be9f861a811e9b3f6fa163eca30e0",
        "rev": 8,
        "cluster": "sr-0",
        "value": {
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
        }
      },
      {
        "key": "/cse-sr/inst/files/default/default/4042a6a3e5a2893698ae363ea99a69eb63fc51cd/8e0fe4b961a811e981a6fa163e86b81a",
        "rev": 9,
        "cluster": "sr-0",
        "value": {
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
      }
    ]
  }
}`))
}

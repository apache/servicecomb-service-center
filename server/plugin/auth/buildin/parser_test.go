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

package buildin_test

import (
	_ "github.com/apache/servicecomb-service-center/test"

	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestGetAPIParseFunc(t *testing.T) {
	rbac.InitResourceMap()

	var serviceIDA, serviceIDB string

	response, _ := datasource.GetMetadataManager().RegisterService(context.Background(), &discovery.CreateServiceRequest{
		Service: &discovery.MicroService{
			AppId:       "TestGetAPIParseFunc",
			ServiceName: "A",
		},
	})
	serviceIDA = response.ServiceId
	response, _ = datasource.GetMetadataManager().RegisterService(context.Background(), &discovery.CreateServiceRequest{
		Service: &discovery.MicroService{
			AppId:       "TestGetAPIParseFunc",
			ServiceName: "B",
		},
	})
	serviceIDB = response.ServiceId

	newRequest := func(method, url string, body string) *http.Request {
		request, _ := http.NewRequest(method, url, strings.NewReader(body))
		util.SetRequestContext(request, rest.CtxMatchPattern, url)
		return request
	}
	tests := []struct {
		name         string
		request      *http.Request
		wantResource *auth.ResourceScope
		wantErr      bool
	}{
		{
			"get all services api should return no labels",
			newRequest(http.MethodGet, "/v4/:project/registry/microservices", ""),
			&auth.ResourceScope{
				Type: "service",
				Verb: "get",
			},
			false,
		},
		{
			"create services api without body should return err",
			newRequest(http.MethodPost, "/v4/:project/registry/microservices", "{}"),
			nil,
			true,
		},
		{
			"create services api with serviceName A should return A",
			newRequest(http.MethodPost, "/v4/:project/registry/microservices",
				"{\n  \"service\": {\n    \"serviceName\": \"A\"\n  }\n}"),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "",
						"serviceName": "A",
					},
				},
				Verb: "create",
			},
			false,
		},
		{
			"delete 2 services api should return 2 labels",
			newRequest(http.MethodDelete, "/v4/:project/registry/microservices",
				"{\n  \"serviceIds\": [\""+serviceIDA+"\", \""+serviceIDB+"\"]\n}"),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "TestGetAPIParseFunc",
						"serviceName": "A",
					},
					{
						"environment": "",
						"appId":       "TestGetAPIParseFunc",
						"serviceName": "B",
					},
				},
				Verb: "delete",
			},
			false,
		},
		{
			"get service sub resource api should return service labels",
			newRequest(http.MethodGet, "/v4/:project/registry/microservices/:serviceId/instances?:serviceId="+serviceIDA, ""),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "TestGetAPIParseFunc",
						"serviceName": "A",
					},
				},
				Verb: "get",
			},
			false,
		},
		{
			"discovery A instances api should return service A labels",
			newRequest(http.MethodGet, "/v4/:project/registry/instances?serviceName=A", ""),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "",
						"serviceName": "A",
					},
				},
				Verb: "get",
			},
			false,
		},
		{
			"govern query A api should return service A labels",
			newRequest(http.MethodGet, "/v4/:project/govern/microservices?serviceName=A", ""),
			&auth.ResourceScope{
				Type: "service",
				Verb: "get",
			},
			false,
		},
		{
			"govern query all api should return nil labels",
			newRequest(http.MethodGet, "/v4/:project/govern/microservices", ""),
			&auth.ResourceScope{
				Type: "service",
				Verb: "get",
			},
			false,
		},
		{
			"batch discovery A instances should return A labels",
			newRequest(http.MethodPost, "/v4/:project/registry/instances/action",
				"{\n  \"consumerId\": \"\",\n  \"instances\": [\n    {\n      \"instance\": {\n        \"serviceId\": \""+serviceIDA+"\",\n        \"instanceId\": \"\"\n      },\n      \"rev\": \"\"\n    }\n  ]\n}"),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "TestGetAPIParseFunc",
						"serviceName": "A",
					},
				},
				Verb: "get",
			},
			false,
		},
		{
			"batch heartbeat A instances should return A labels",
			newRequest(http.MethodPost, "/v4/:project/registry/heartbeats",
				"{\n  \"instances\": [\n    {\n      \"serviceId\": \""+serviceIDA+"\",\n      \"instanceId\": \"\"\n    }\n  ]\n}"),
			&auth.ResourceScope{
				Type: "service",
				Labels: []map[string]string{
					{
						"environment": "",
						"appId":       "TestGetAPIParseFunc",
						"serviceName": "A",
					},
				},
				Verb: "update",
			},
			false,
		},
		// not registered api
		{
			"govern statistics api should apply all",
			newRequest(http.MethodGet, "/v4/:project/govern/statistics", ""),
			&auth.ResourceScope{
				Type: "service",
				Verb: "get",
			},
			false,
		},
		{
			"govern apps api should apply all",
			newRequest(http.MethodGet, "/v4/:project/govern/apps", ""),
			&auth.ResourceScope{
				Type: "service",
				Verb: "get",
			},
			false,
		},
		{
			"account api should apply all",
			newRequest(http.MethodGet, "/v4/accounts", ""),
			&auth.ResourceScope{
				Type: "account",
				Verb: "get",
			},
			false,
		},
		{
			"role api should apply all",
			newRequest(http.MethodGet, "/v4/roles", ""),
			&auth.ResourceScope{
				Type: "role",
				Verb: "get",
			},
			false,
		},
		{
			"role api should apply all",
			newRequest(http.MethodGet, "/v1/:project/gov/", ""),
			&auth.ResourceScope{
				Type: "governance",
				Verb: "get",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := buildin.GetAPIParseFunc(tt.request.URL.Path)(tt.request)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantResource == nil, resource == nil)
			if tt.wantResource == nil {
				return
			}
			assert.Equal(t, tt.wantResource.Type, resource.Type)
			assert.Equal(t, tt.wantResource.Verb, resource.Verb)
			assert.Equal(t, tt.wantResource.Labels == nil, resource.Labels == nil)
			for _, labels := range tt.wantResource.Labels {
				checkLabels(t, labels, resource)
			}
		})
	}
}

func checkLabels(t *testing.T, labels map[string]string, resource *auth.ResourceScope) {
	var (
		targetLabelsLength int
		found              bool
	)
	for k, v := range labels {
		for _, targetLabels := range resource.Labels {
			targetLabelsLength = len(targetLabels)
			found = v == targetLabels[k]
			if found {
				break
			}
		}
		assert.True(t, found, "target label key value not matched %s=%s", k, v)
	}
	assert.Equal(t, len(labels), targetLabelsLength)
}

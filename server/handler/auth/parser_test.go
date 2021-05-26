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

package auth_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestGetAPIParseFunc(t *testing.T) {
	var serviceIDA, serviceIDB string

	response, _ := datasource.Instance().RegisterService(context.Background(), &discovery.CreateServiceRequest{
		Service: &discovery.MicroService{
			AppId:       "TestGetAPIParseFunc",
			ServiceName: "A",
		},
	})
	serviceIDA = response.ServiceId
	response, _ = datasource.Instance().RegisterService(context.Background(), &discovery.CreateServiceRequest{
		Service: &discovery.MicroService{
			AppId:       "TestGetAPIParseFunc",
			ServiceName: "B",
		},
	})
	serviceIDB = response.ServiceId

	t.Run("get all services api should return no labels", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices", nil)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/registry/microservices")(request)
		assert.NoError(t, err)
		assert.Nil(t, labels)
	})

	t.Run("create services api without body should return err", func(t *testing.T) {
		reader := strings.NewReader("{}")

		request, err := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices", reader)
		assert.NoError(t, err)

		_, err = auth.GetAPIParseFunc("/v4/:project/registry/microservices")(request)
		assert.Error(t, err)
	})

	t.Run("create services api with serviceName A should return A", func(t *testing.T) {
		reader := strings.NewReader("{\n  \"service\": {\n    \"serviceName\": \"A\"\n  }\n}")

		request, err := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices", reader)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/registry/microservices")(request)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(labels))
		assert.Equal(t, "A", labels[0]["serviceName"])
	})

	t.Run("delete 2 services api should return 2 labels", func(t *testing.T) {
		reader := strings.NewReader("{\n  \"serviceIds\": [\"" + serviceIDA + "\", \"" + serviceIDB + "\"]\n}")

		request, err := http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices", reader)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/registry/microservices")(request)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(labels))
		assert.Equal(t, "A", labels[0]["serviceName"])
	})

	t.Run("get service sub resource api should return service labels", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceIDA+"/instances?:serviceId="+serviceIDA, nil)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/registry/microservices/:serviceId/instances")(request)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(labels))
		assert.Equal(t, "A", labels[0]["serviceName"])
	})

	t.Run("discovery A instances api should return service A labels", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v4/default/registry/instances?serviceName=A", nil)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/registry/instances")(request)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(labels))
		assert.Equal(t, "A", labels[0]["serviceName"])
	})

	t.Run("govern query A api should return service A labels", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?serviceName=A", nil)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/govern/microservices")(request)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(labels))
		assert.Equal(t, "A", labels[0]["serviceName"])
	})

	t.Run("govern query all api should return nil labels", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices", nil)
		assert.NoError(t, err)

		labels, err := auth.GetAPIParseFunc("/v4/:project/govern/microservices")(request)
		assert.NoError(t, err)
		assert.Nil(t, labels)
	})
}

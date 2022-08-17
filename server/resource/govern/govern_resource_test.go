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

package govern_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/resource/govern"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func init() {
	rest.RegisterServant(&govern.Resource{})
}

func TestResource_GetServiceDetail(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "get_service_detail",
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("query service detail, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices/"+serviceID+"?noCache=true", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServiceDetailResponse
		body, _ := io.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, serviceID, resp.Service.MicroService.ServiceId)
	})

	t.Run("query not exist service detail, should fail", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices/not-exist", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestResource_ListServiceDetail(t *testing.T) {
	ctx := context.TODO()

	const serviceName = "list_service_detail"
	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: serviceName,
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("list service detail, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?noCache=true&options=all&serviceName="+serviceName, nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServicesInfoResponse
		body, _ := io.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, serviceID, resp.AllServicesDetail[0].MicroService.ServiceId)
		assert.NotEqual(t, int64(0), resp.Statistics.Services.Count)
		assert.NotEqual(t, int64(0), resp.Statistics.Apps.Count)
	})

	t.Run("list not exist service detail, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?options=all&serviceName=not-exist", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServicesInfoResponse
		body, _ := io.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(resp.AllServicesDetail))
		assert.NotEqual(t, int64(0), resp.Statistics.Services.Count)
		assert.NotEqual(t, int64(0), resp.Statistics.Apps.Count)
	})
}

func TestResource_ListApp(t *testing.T) {
	ctx := context.TODO()

	const serviceName = "list_app"
	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		AppId:       "list_app_app",
		ServiceName: serviceName,
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("list app, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/apps?noCache=true", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetAppsResponse
		body, _ := io.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.AppIds, "list_app_app")
	})
}

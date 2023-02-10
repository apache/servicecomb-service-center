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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/resource/v4/govern"
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
		Schemas:     []string{"test"},
		Properties:  map[string]string{"test": "list"},
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
		detail := resp.Service
		assert.Equal(t, serviceID, detail.MicroService.ServiceId)
		assert.NotEmpty(t, detail.MicroService.Schemas)
		assert.NotEmpty(t, detail.MicroService.Properties)
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
		Schemas:     []string{"test"},
		Properties:  map[string]string{"test": "list"},
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
		detail := resp.AllServicesDetail[0]
		assert.Equal(t, serviceID, detail.MicroService.ServiceId)
		assert.NotEmpty(t, detail.MicroService.Properties)
		assert.Empty(t, detail.MicroService.Schemas)
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

	t.Run("list service detail with properties filter, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?noCache=true&options=all&serviceName="+serviceName+"&property=test:list", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServicesInfoResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, serviceID, resp.AllServicesDetail[0].MicroService.ServiceId)
		assert.NotEqual(t, int64(0), resp.Statistics.Services.Count)
		assert.NotEqual(t, int64(0), resp.Statistics.Apps.Count)
	})

	t.Run("list service detail with only properties filter, should ok", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?noCache=true&options=all&property=test:list", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServicesInfoResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, serviceID, resp.AllServicesDetail[0].MicroService.ServiceId)
		assert.NotEqual(t, int64(0), resp.Statistics.Services.Count)
		assert.NotEqual(t, int64(0), resp.Statistics.Apps.Count)
	})

	t.Run("list service detail with not exist properties, should return empty", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/govern/microservices?noCache=true&options=all&serviceName="+serviceName+"&property=none:none", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetServicesInfoResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(resp.AllServicesDetail))
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

func TestParseProperties(t *testing.T) {
	type args struct {
		query url.Values
		key   string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"invalid queries", args{nil, "any"}, map[string]string{}},
		{"invalid queries", args{url.Values{}, "any"}, map[string]string{}},
		{"invalid queries", args{url.Values{"a": {"b:a"}}, "other"}, map[string]string{}},
		{"valid queries", args{url.Values{"a": {""}}, "a"}, map[string]string{"": ""}},
		{"valid queries", args{url.Values{"a": {"b"}}, "a"}, map[string]string{"b": ""}},
		{"valid queries", args{url.Values{"a": {"b:"}}, "a"}, map[string]string{"b": ""}},
		{"valid queries", args{url.Values{"a": {":"}}, "a"}, map[string]string{"": ""}},
		{"valid queries", args{url.Values{"a": {":a"}}, "a"}, map[string]string{"": "a"}},
		{"valid queries", args{url.Values{"a": {"b:a"}}, "a"}, map[string]string{"b": "a"}},
		{"valid queries", args{url.Values{"a": {"b:a", "c:d"}}, "a"}, map[string]string{"b": "a", "c": "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := govern.ParseProperties(tt.args.query, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProperties() = %v, want %v", got, tt.want)
			}
		})
	}
}

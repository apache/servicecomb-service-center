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
	"context"
	"net/http"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func init() {
	rbacsvc.InitResourceMap()
}

func TestByServiceKey(t *testing.T) {
	t.Run("discover nothing should return empty scope", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, buildin.APIDiscovery, nil)
		req = req.WithContext(context.WithValue(req.Context(), rest.CtxMatchPattern, buildin.APIDiscovery))
		resp, err := buildin.ByServiceKey(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "service", resp.Type)
		assert.Equal(t, "get", resp.Verb)
		assert.NotEmpty(t, resp.Labels)
		labels := resp.Labels[0]
		assert.Equal(t, "", labels["environment"])
		assert.Equal(t, "", labels["appId"])
		assert.Equal(t, "", labels["serviceName"])
	})

	t.Run("discover provider 'test' should return 'test' scope", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, buildin.APIDiscovery+"?appId=default&serviceName=test", nil)

		service, err := discosvc.RegisterService(req.Context(), &pb.CreateServiceRequest{Service: &pb.MicroService{
			ServiceName: "consumer",
		}})
		assert.NoError(t, err)
		defer discosvc.UnregisterService(req.Context(), &pb.DeleteServiceRequest{ServiceId: service.ServiceId})

		req.Header.Set("X-ConsumerId", service.ServiceId)
		req = req.WithContext(context.WithValue(req.Context(), rest.CtxMatchPattern, buildin.APIDiscovery))
		resp, err := buildin.ByServiceKey(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "service", resp.Type)
		assert.Equal(t, "get", resp.Verb)
		assert.NotEmpty(t, resp.Labels)
		labels := resp.Labels[0]
		assert.Equal(t, "", labels["environment"])
		assert.Equal(t, "default", labels["appId"])
		assert.Equal(t, "test", labels["serviceName"])
	})

	t.Run("discover provider 'test' in development env should return 'test' scope", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, buildin.APIDiscovery+"?appId=default&serviceName=test", nil)

		service, err := discosvc.RegisterService(req.Context(), &pb.CreateServiceRequest{Service: &pb.MicroService{
			Environment: "development",
			ServiceName: "consumer",
		}})
		assert.NoError(t, err)
		defer discosvc.UnregisterService(req.Context(), &pb.DeleteServiceRequest{ServiceId: service.ServiceId})

		req.Header.Set("X-ConsumerId", service.ServiceId)
		req = req.WithContext(context.WithValue(req.Context(), rest.CtxMatchPattern, buildin.APIDiscovery))
		resp, err := buildin.ByServiceKey(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "service", resp.Type)
		assert.Equal(t, "get", resp.Verb)
		assert.NotEmpty(t, resp.Labels)
		labels := resp.Labels[0]
		assert.Equal(t, "development", labels["environment"])
		assert.Equal(t, "default", labels["appId"])
		assert.Equal(t, "test", labels["serviceName"])
	})

	t.Run("discover provider 'test' with query env should return 'test' scope", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, buildin.APIDiscovery+"?appId=default&serviceName=test&env=testing", nil)

		service, err := discosvc.RegisterService(req.Context(), &pb.CreateServiceRequest{Service: &pb.MicroService{
			Environment: "development",
			ServiceName: "consumer",
		}})
		assert.NoError(t, err)
		defer discosvc.UnregisterService(req.Context(), &pb.DeleteServiceRequest{ServiceId: service.ServiceId})

		req.Header.Set("X-ConsumerId", service.ServiceId)
		req = req.WithContext(context.WithValue(req.Context(), rest.CtxMatchPattern, buildin.APIDiscovery))
		resp, err := buildin.ByServiceKey(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "service", resp.Type)
		assert.Equal(t, "get", resp.Verb)
		assert.NotEmpty(t, resp.Labels)
		labels := resp.Labels[0]
		assert.Equal(t, "testing", labels["environment"])
		assert.Equal(t, "default", labels["appId"])
		assert.Equal(t, "test", labels["serviceName"])
	})
}

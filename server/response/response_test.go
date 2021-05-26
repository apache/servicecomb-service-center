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

package response_test

import (
	"github.com/apache/servicecomb-service-center/server/response"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMicroServiceInfoListFilter(t *testing.T) {
	t.Run("obj is invalid, should do nothing", func(t *testing.T) {
		other := &discovery.MicroService{}
		assert.Equal(t, other, response.MicroServiceInfoListFilter(other, nil))
	})

	t.Run("labels is empty, should return empty resources", func(t *testing.T) {
		rs := response.MicroServiceInfoListFilter(&discovery.GetServicesInfoResponse{
			AllServicesDetail: []*discovery.ServiceDetail{
				{MicroService: &discovery.MicroService{ServiceName: "A"}}, {MicroService: &discovery.MicroService{ServiceName: "B"}},
			},
		}, nil)
		assert.Equal(t, 0, len(rs.(*discovery.GetServicesInfoResponse).AllServicesDetail))
	})

	t.Run("not match name, should return empty resources", func(t *testing.T) {
		rs := response.MicroServiceInfoListFilter(&discovery.GetServicesInfoResponse{
			AllServicesDetail: []*discovery.ServiceDetail{
				{MicroService: &discovery.MicroService{ServiceName: "A"}}, {MicroService: &discovery.MicroService{ServiceName: "B"}},
			},
		}, []map[string]string{{"serviceName": "NONE"}})
		assert.Equal(t, 0, len(rs.(*discovery.GetServicesInfoResponse).AllServicesDetail))
	})

	t.Run("match A, should return A resources", func(t *testing.T) {
		rs := response.MicroServiceInfoListFilter(&discovery.GetServicesInfoResponse{
			AllServicesDetail: []*discovery.ServiceDetail{
				{MicroService: &discovery.MicroService{ServiceName: "A"}}, {MicroService: &discovery.MicroService{ServiceName: "B"}},
			},
		}, []map[string]string{{"serviceName": "A"}})
		mss := rs.(*discovery.GetServicesInfoResponse).AllServicesDetail
		assert.Equal(t, 1, len(mss))
		assert.Equal(t, "A", mss[0].MicroService.ServiceName)
	})

	t.Run("not match name & appId, should return empty resources", func(t *testing.T) {
		rs := response.MicroServiceInfoListFilter(&discovery.GetServicesInfoResponse{
			AllServicesDetail: []*discovery.ServiceDetail{
				{MicroService: &discovery.MicroService{ServiceName: "A"}}, {MicroService: &discovery.MicroService{ServiceName: "B"}},
			},
		}, []map[string]string{{"serviceName": "A", "appId": "A"}})
		mss := rs.(*discovery.GetServicesInfoResponse).AllServicesDetail
		assert.Equal(t, 0, len(mss))
	})

	t.Run("match name & appId, should return empty resources", func(t *testing.T) {
		rs := response.MicroServiceInfoListFilter(&discovery.GetServicesInfoResponse{
			AllServicesDetail: []*discovery.ServiceDetail{
				{MicroService: &discovery.MicroService{AppId: "A", ServiceName: "A"}}, {MicroService: &discovery.MicroService{ServiceName: "B"}},
			},
		}, []map[string]string{{"serviceName": "A", "appId": "A"}})
		mss := rs.(*discovery.GetServicesInfoResponse).AllServicesDetail
		assert.Equal(t, 1, len(mss))
	})
}

func TestMicroserviceListFilter(t *testing.T) {
	t.Run("obj is invalid, should do nothing", func(t *testing.T) {
		other := &discovery.MicroService{}
		assert.Equal(t, other, response.MicroserviceListFilter(other, nil))
	})

	t.Run("labels is empty, should return empty resources", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{ServiceName: "A"}, {ServiceName: "B"},
			},
		}, nil)
		assert.Equal(t, 0, len(rs.(*discovery.GetServicesResponse).Services))
	})

	t.Run("not match name, should return empty resources", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{ServiceName: "A"}, {ServiceName: "B"},
			},
		}, []map[string]string{{"serviceName": "NONE"}})
		assert.Equal(t, 0, len(rs.(*discovery.GetServicesResponse).Services))
	})

	t.Run("match A, should return A resources", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{ServiceName: "A"}, {ServiceName: "B"},
			},
		}, []map[string]string{{"serviceName": "A"}})
		mss := rs.(*discovery.GetServicesResponse).Services
		assert.Equal(t, 1, len(mss))
		assert.Equal(t, "A", mss[0].ServiceName)
	})

	t.Run("not match name & appId, should return empty resources", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{ServiceName: "A"}, {ServiceName: "B"},
			},
		}, []map[string]string{{"serviceName": "A", "appId": "A"}})
		mss := rs.(*discovery.GetServicesResponse).Services
		assert.Equal(t, 0, len(mss))
	})

	t.Run("match name & appId, should return A", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{AppId: "A", ServiceName: "A"}, {ServiceName: "B"},
			},
		}, []map[string]string{{"serviceName": "A", "appId": "A"}})
		mss := rs.(*discovery.GetServicesResponse).Services
		assert.Equal(t, 1, len(mss))
		assert.Equal(t, "A", mss[0].ServiceName)
	})

	t.Run("wildcard match name, should return empty resources", func(t *testing.T) {
		rs := response.MicroserviceListFilter(&discovery.GetServicesResponse{
			Services: []*discovery.MicroService{
				{ServiceName: "TestA"}, {ServiceName: "DevB"},
			},
		}, []map[string]string{{"serviceName": "Test*"}})
		mss := rs.(*discovery.GetServicesResponse).Services
		assert.Equal(t, 1, len(mss))
		assert.Equal(t, "TestA", mss[0].ServiceName)
	})
}

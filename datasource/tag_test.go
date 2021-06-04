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

package datasource_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestTags_Add(t *testing.T) {
	var (
		serviceId1 string
	)

	// create service
	t.Run("create service", func(t *testing.T) {
		svc1 := &pb.MicroService{
			AppId:       "create_tag_group_ms",
			ServiceName: "create_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc1,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId1 = resp.ServiceId
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exist")
		resp, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: "noServiceTest",
			Tags: map[string]string{
				"a": "test",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		log.Info("tag quota is equal to the default value and should be paas")
		defaultQuota := quota.DefaultTagQuota
		tags := make(map[string]string, defaultQuota)
		for i := 0; i < defaultQuota; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		resp, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags:      tags,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})
}

func TestTags_Get(t *testing.T) {
	var serviceId string
	t.Run("create service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "get_tag_group_ms",
			ServiceName: "get_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		log.Info("add tags should be passed")
		respAddTags, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exists")
		resp, err := datasource.GetMetadataManager().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "noThisService",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("service's id is empty")
		resp, err = datasource.GetMetadataManager().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: "",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("service's id is invalid")
		resp, err = datasource.GetMetadataManager().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: strings.Repeat("x", 65),
		})
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "test", resp.Tags["a"])
	})
}

func TestTag_Update(t *testing.T) {
	var serviceId string
	t.Run("add service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "update_tag_group_ms",
			ServiceName: "update_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		log.Info("add tags")
		respAddTags, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTags.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {

		log.Info("service does not exists")
		resp, err := datasource.GetMetadataManager().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: "noneservice",
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("tag key does not exist")
		resp, err = datasource.GetMetadataManager().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "notexisttag",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())

		log.Info("tag key is invalid")
		resp, err = datasource.GetMetadataManager().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       strings.Repeat("x", 65),
			Value:     "v",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().UpdateTag(getContext(), &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
	})

	t.Run("find instance, contain tag", func(t *testing.T) {
		log.Info("create consumer")
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group_ms",
				ServiceName: "find_inst_tag_consumer_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		consumerId := resp.ServiceId

		log.Info("create provider")
		resp, err = datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group_ms",
				ServiceName: "find_inst_tag_provider_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		providerId := resp.ServiceId

		log.Info("tag the provider")
		addTagsResp, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: providerId,
			Tags:      map[string]string{"filter_tag": "filter"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, addTagsResp.Response.GetCode())

		log.Info("add instance to provider")
		instanceResp, err := datasource.GetMetadataManager().RegisterInstance(getContext(), &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: providerId,
				Endpoints: []string{
					"findInstanceForTagFilter:127.0.0.1:8080",
				},
				HostName: "UT-HOST",
				Status:   pb.MSI_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, instanceResp.Response.GetCode())

		log.Info("find instance")
		findResp, err := datasource.GetMetadataManager().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			VersionRule:       "1.0.0+",
			Tags:              []string{"not-exist-tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, findResp.Response.GetCode())
		assert.Equal(t, 0, len(findResp.Instances))

		findResp, err = datasource.GetMetadataManager().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			VersionRule:       "1.0.0+",
			Tags:              []string{"filter_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, findResp.Response.GetCode())
		assert.Equal(t, instanceResp.InstanceId, findResp.Instances[0].InstanceId)

		// no add rules

		log.Info("add tags")
		addTagsResp, err = datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: consumerId,
			Tags:      map[string]string{"consumer_tag": "filter"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, addTagsResp.Response.GetCode())

		findResp, err = datasource.GetMetadataManager().FindInstances(getContext(), &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			VersionRule:       "1.0.0+",
			Tags:              []string{"filter_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, findResp.Response.GetCode())
	})
}

func TestTags_Delete(t *testing.T) {
	var serviceId string
	t.Run("create service and add tags", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_tag_group_ms",
				ServiceName: "delete_tag_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		serviceId = resp.ServiceId

		respAddTages, err := datasource.GetMetadataManager().AddTags(getContext(), &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, respAddTages.Response.GetCode())
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exits")
		resp, err := datasource.GetMetadataManager().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: "noneservice",
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrServiceNotExists, resp.Response.GetCode())

		log.Info("tag key does not exits")
		resp, err = datasource.GetMetadataManager().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"c"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ErrTagNotExists, resp.Response.GetCode())
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().DeleteTags(getContext(), &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		respGetTags, err := datasource.GetMetadataManager().GetTags(getContext(), &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "", respGetTags.Tags["a"])
	})
}

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
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestTags_Add(t *testing.T) {
	var (
		serviceId1 string
	)
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId1, Force: true})

	// create service
	t.Run("create service", func(t *testing.T) {
		svc1 := &pb.MicroService{
			AppId:       "create_tag_group_ms",
			ServiceName: "create_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: svc1,
		})
		assert.NoError(t, err)
		assert.NotEqual(t, "", resp.ServiceId)
		serviceId1 = resp.ServiceId
	})

	t.Run("the request is invalid", func(t *testing.T) {
		err := datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: "noServiceTest",
			Tags: map[string]string{
				"a": "test",
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("the request is valid", func(t *testing.T) {
		log.Info("tag quota is equal to the default value and should be paas")
		defaultQuota := int(quotasvc.TagQuota())
		tags := make(map[string]string, defaultQuota)
		for i := 0; i < defaultQuota; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		err := datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags:      tags,
		})
		assert.NoError(t, err)
	})
}

func TestTags_Get(t *testing.T) {
	var serviceId string
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("create service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "get_tag_group_ms",
			ServiceName: "get_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		serviceId = resp.ServiceId

		log.Info("add tags should be passed")
		err = datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exists")
		_, err := datasource.GetMetadataManager().ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: "noThisService",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		log.Info("service's id is empty")
		_, err = datasource.GetMetadataManager().ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		log.Info("service's id is invalid")
		_, err = datasource.GetMetadataManager().ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: strings.Repeat("x", 65),
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)
	})

	t.Run("the request is valid", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())
		assert.Equal(t, "test", resp.Tags["a"])
	})
}

func TestTag_Update(t *testing.T) {
	var serviceId string
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("add service and add tags", func(t *testing.T) {
		svc := &pb.MicroService{
			AppId:       "update_tag_group_ms",
			ServiceName: "update_tag_service_ms",
			Version:     "1.0.0",
			Level:       "FRONT",
			Status:      pb.MS_UP,
		}
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: svc,
		})
		assert.NoError(t, err)
		serviceId = resp.ServiceId

		log.Info("add tags")
		err = datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exists")
		err := datasource.GetMetadataManager().PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: "noneservice",
			Key:       "a",
			Value:     "update",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		log.Info("tag key does not exist")
		err = datasource.GetMetadataManager().PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "notexisttag",
			Value:     "update",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrTagNotExists, testErr.Code)

		log.Info("tag key is invalid")
		err = datasource.GetMetadataManager().PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       strings.Repeat("x", 65),
			Value:     "v",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrTagNotExists, testErr.Code)
	})

	t.Run("the request is valid", func(t *testing.T) {
		err := datasource.GetMetadataManager().PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
	})

	t.Run("find instance, contain tag", func(t *testing.T) {
		log.Info("create consumer")
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group_ms",
				ServiceName: "find_inst_tag_consumer_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		consumerId := resp.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: consumerId, Force: true})

		log.Info("create provider")
		resp, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group_ms",
				ServiceName: "find_inst_tag_provider_ms",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		providerId := resp.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: providerId, Force: true})

		log.Info("tag the provider")
		err = datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: providerId,
			Tags:      map[string]string{"filter_tag": "filter"},
		})
		assert.NoError(t, err)

		log.Info("add instance to provider")
		instanceResp, err := datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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

		log.Info("find instance")
		findResp, err := datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			Tags:              []string{"not-exist-tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(findResp.Instances))

		findResp, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			Tags:              []string{"filter_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, instanceResp.InstanceId, findResp.Instances[0].InstanceId)

		// no add rules

		log.Info("add tags")
		err = datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: consumerId,
			Tags:      map[string]string{"consumer_tag": "filter"},
		})
		assert.NoError(t, err)

		findResp, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group_ms",
			ServiceName:       "find_inst_tag_provider_ms",
			Tags:              []string{"filter_tag"},
		})
		assert.NoError(t, err)
	})
}

func TestTags_Delete(t *testing.T) {
	var serviceId string
	ctx := getContext()
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("create service and add tags", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_tag_group_ms",
				ServiceName: "delete_tag_service_ms",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId = resp.ServiceId

		err = datasource.GetMetadataManager().PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("the request is invalid", func(t *testing.T) {
		log.Info("service does not exits")
		err := datasource.GetMetadataManager().DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: "noneservice",
			Keys:      []string{"a", "b"},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		log.Info("tag key does not exits")
		err = datasource.GetMetadataManager().DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"c"},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrTagNotExists, testErr.Code)
	})

	t.Run("the request is valid", func(t *testing.T) {
		err := datasource.GetMetadataManager().DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)

		respGetTags, err := datasource.GetMetadataManager().ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "", respGetTags.Tags["a"])
	})
}

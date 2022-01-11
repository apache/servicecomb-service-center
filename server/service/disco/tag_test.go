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
package disco_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/server/service/disco"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	pb "github.com/go-chassis/cari/discovery"
)

var (
	TooLongTag = strings.Repeat("x", 65)
)

func TestPutManyTags(t *testing.T) {
	var (
		serviceId1 string
		serviceId2 string
	)
	ctx := getContext()
	max := int(quotasvc.TagQuota())
	defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{serviceId1, serviceId2}, Force: true})

	t.Run("prepare date, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_tag_group",
				ServiceName: "create_tag_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId1 = respCreateService.ServiceId

		respCreateService, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "create_tag_group",
				ServiceName: "create_tag_service",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId2 = respCreateService.ServiceId
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: "",
			Tags: map[string]string{
				"a": "test",
			},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: "noServiceTest",
			Tags: map[string]string{
				"a": "test",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags: map[string]string{
				"": "value",
			},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		err := disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags: map[string]string{
				"a": "test",
			},
		})
		assert.NoError(t, err)

		getServiceTagsResponse, err := disco.ListTag(ctx, &pb.GetServiceTagsRequest{ServiceId: serviceId1})
		assert.NoError(t, err)
		assert.Equal(t, "test", getServiceTagsResponse.Tags["a"])

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId1,
			Tags:      map[string]string{},
		})
		assert.NoError(t, err)

		getServiceTagsResponse, err = disco.ListTag(ctx, &pb.GetServiceTagsRequest{ServiceId: serviceId1})
		assert.NoError(t, err)
		assert.Empty(t, getServiceTagsResponse.Tags)
	})

	t.Run("when create tag out of gauge, should be failed", func(t *testing.T) {
		size := max + 1
		tags := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		err := disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId2,
			Tags:      tags,
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		tags = make(map[string]string, max)
		for i := 0; i < max; i++ {
			s := "tag" + strconv.Itoa(i)
			tags[s] = s
		}
		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId2,
			Tags:      tags,
		})
		assert.NoError(t, err)
	})
}

func TestListTag(t *testing.T) {
	var (
		serviceId string
	)
	ctx := getContext()
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("prepare date, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "get_tag_group",
				ServiceName: "get_tag_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		_, err := disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: "noThisService",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		_, err = disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: "",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		_, err = disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: TooLongServiceID,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		resp, err := disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "test", resp.Tags["a"])
	})
}

func TestPutTag(t *testing.T) {
	var (
		serviceId string
	)
	ctx := getContext()
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "update_tag_group",
				ServiceName: "update_tag_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: "",
			Key:       "a",
			Value:     "update",
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: "noneservice",
			Key:       "a",
			Value:     "update",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "",
			Value:     "update",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "notexisttag",
			Value:     "update",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrTagNotExists, testErr.Code)

		err = disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       TooLongTag,
			Value:     "v",
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		err := disco.PutTag(ctx, &pb.UpdateServiceTagRequest{
			ServiceId: serviceId,
			Key:       "a",
			Value:     "update",
		})
		assert.NoError(t, err)
	})

	t.Run("find instance, contain tag, should passed", func(t *testing.T) {
		resp, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group",
				ServiceName: "find_inst_tag_consumer",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		consumerId := resp.ServiceId

		resp, err = disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "find_inst_tag_group",
				ServiceName: "find_inst_tag_provider",
				Version:     "1.0.1",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		providerId := resp.ServiceId

		defer disco.UnregisterManyService(ctx, &pb.DelServicesRequest{ServiceIds: []string{
			consumerId, providerId,
		}, Force: true})

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: providerId,
			Tags:      map[string]string{"filter_tag": "filter"},
		})
		assert.NoError(t, err)

		instanceResp, err := disco.RegisterInstance(ctx, &pb.RegisterInstanceRequest{
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

		findResp, err := disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group",
			ServiceName:       "find_inst_tag_provider",
			Tags:              []string{"not-exist-tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(findResp.Instances))

		findResp, err = disco.FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumerId,
			AppId:             "find_inst_tag_group",
			ServiceName:       "find_inst_tag_provider",
			Tags:              []string{"filter_tag"},
		})
		assert.NoError(t, err)
		assert.Equal(t, instanceResp.InstanceId, findResp.Instances[0].InstanceId)
	})
}

func TestDeleteManyTags(t *testing.T) {
	var (
		serviceId string
	)
	ctx := getContext()
	max := int(quotasvc.TagQuota())
	defer disco.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceId, Force: true})

	t.Run("prepare data, should be passed", func(t *testing.T) {
		respCreateService, err := disco.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       "delete_tag_group",
				ServiceName: "delete_tag_service",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      pb.MS_UP,
			},
		})
		assert.NoError(t, err)
		serviceId = respCreateService.ServiceId

		err = disco.PutManyTags(ctx, &pb.AddServiceTagsRequest{
			ServiceId: serviceId,
			Tags: map[string]string{
				"a": "test",
				"b": "b",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("when request is invalid, should be failed", func(t *testing.T) {
		err := disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: "",
			Keys:      []string{"a", "b"},
		})
		testErr := err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: "noneservice",
			Keys:      []string{"a", "b"},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrServiceNotExists, testErr.Code)

		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"c"},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrTagNotExists, testErr.Code)

		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{""},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{TooLongTag},
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)

		var arr []string
		for i := 0; i < int(max)+1; i++ {
			arr = append(arr, strconv.Itoa(i))
		}
		err = disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      arr,
		})
		testErr = err.(*errsvc.Error)
		assert.Error(t, testErr)
		assert.Equal(t, pb.ErrInvalidParams, testErr.Code)
	})

	t.Run("when request is valid, should be passed", func(t *testing.T) {
		err := disco.DeleteManyTags(ctx, &pb.DeleteServiceTagsRequest{
			ServiceId: serviceId,
			Keys:      []string{"a", "b"},
		})
		assert.NoError(t, err)

		resp, err := disco.ListTag(ctx, &pb.GetServiceTagsRequest{
			ServiceId: serviceId,
		})
		assert.NoError(t, err)
		assert.Equal(t, "", resp.Tags["a"])
	})
}

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

package util_test

// initialize
import (
	"context"
	"errors"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	proto "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/etcdadpt"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func getContextWith(domain string, project string) context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), domain, project))
}

func TestFindServiceIds(t *testing.T) {
	ctx := getContextWith("default", "default")

	t.Run("no service, should return empty", func(t *testing.T) {
		ids, exist, err := serviceUtil.FindServiceIds(ctx, &proto.MicroServiceKey{}, false)
		assert.NoError(t, err)
		assert.False(t, exist)
		assert.Empty(t, ids)

		ids, exist, err = serviceUtil.FindServiceIds(ctx, &proto.MicroServiceKey{}, true)
		assert.NoError(t, err)
		assert.False(t, exist)
		assert.Empty(t, ids)
	})

	t.Run("exist service, should return empty", func(t *testing.T) {
		resp, err := datasource.GetMetadataManager().RegisterService(ctx, &proto.CreateServiceRequest{
			Service: &proto.MicroService{
				Alias:       "test_find_alias_ids",
				ServiceName: "test_find_service_ids",
				Version:     "2.0",
			},
		})
		assert.NoError(t, err)
		serviceID1 := resp.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &proto.DeleteServiceRequest{ServiceId: serviceID1})

		resp, err = datasource.GetMetadataManager().RegisterService(ctx, &proto.CreateServiceRequest{
			Service: &proto.MicroService{
				Alias:       "test_find_alias_ids",
				ServiceName: "test_find_service_ids",
				Version:     "3.0",
			},
		})
		assert.NoError(t, err)
		serviceID2 := resp.ServiceId
		defer datasource.GetMetadataManager().UnregisterService(ctx, &proto.DeleteServiceRequest{ServiceId: serviceID2})

		ids, exist, err := serviceUtil.FindServiceIds(ctx, &proto.MicroServiceKey{
			Tenant:      "default/default",
			ServiceName: "test_find_service_ids",
			Version:     "1.0",
		}, false)
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, 2, len(ids))

		ids, exist, err = serviceUtil.FindServiceIds(ctx, &proto.MicroServiceKey{
			Tenant:      "default/default",
			ServiceName: "test_find_service_ids",
			Version:     "1.0",
		}, true)
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Empty(t, ids)

		ids, exist, err = serviceUtil.FindServiceIds(ctx, &proto.MicroServiceKey{
			Tenant:      "default/default",
			ServiceName: "test_find_service_ids",
			Version:     "2.0",
		}, true)
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, []string{serviceID1}, ids)
	})
}

func TestGetService(t *testing.T) {
	var err error
	t.Run("when there is no such a service in db", func(t *testing.T) {
		_, err = serviceUtil.GetService(context.Background(), "", "")
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("TestGetService failed")
		}
		assert.Equal(t, datasource.ErrNoData, err)
	})

	_, err = serviceUtil.GetServicesByDomainProject(context.Background(), "")
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetAllServiceUtil(context.Background())
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	t.Run("when there is no such a service in db", func(t *testing.T) {
		_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 0)
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("TestGetService failed")
		}
		assert.Equal(t, datasource.ErrNoData, err)
	})

	// _, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 1)
	// if err != nil {
	//	t.Fatalf("TestGetService failed")
	// }
}

func TestServiceExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestServiceExist failed")
		}
	}()
	serviceUtil.ServiceExist(util.WithCacheOnly(context.Background()), "", "")
}

func TestFromContext(t *testing.T) {
	ctx := util.WithNoCache(context.Background())
	opts := serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op := etcdadpt.OptionsToOp(opts...)
	if op.Mode != etcdadpt.ModeNoCache {
		t.Fatalf("TestFromContext failed")
	}

	ctx = util.WithCacheOnly(context.Background())
	opts = serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op = etcdadpt.OptionsToOp(opts...)
	if op.Mode != etcdadpt.ModeCache {
		t.Fatalf("TestFromContext failed")
	}
}

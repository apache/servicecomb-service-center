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

package etcd_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/little-cui/etcdadpt"
	"github.com/stretchr/testify/assert"
)

const domainProject = "default/default"

func TestGetOldServiceIDs(t *testing.T) {
	const (
		reserve       = 3
		etcdKeyPrefix = "/cse-sr/ms/indexes/default/default//default/"
	)

	type args struct {
		indexesResp *kvstore.Response
	}
	tests := []struct {
		name string
		args args
		want []*etcd.RotateServiceIDKey
	}{
		{"input empty should return empty", args{indexesResp: &kvstore.Response{}}, nil},
		{"less then reserve version count should return empty", args{indexesResp: &kvstore.Response{
			Kvs: []*kvstore.KeyValue{
				{Key: []byte(etcdKeyPrefix + "svc1/1.0"), Value: "svc1-1.0"},
				{Key: []byte(etcdKeyPrefix + "svc1/1.2"), Value: "svc1-1.2"},
				{Key: []byte(etcdKeyPrefix + "svc1/2.0"), Value: "svc1-2.0"},
				{Key: []byte(etcdKeyPrefix + "svc2/1.0"), Value: "svc2-1.0"},
			}, Count: 4,
		}}, nil},
		{"large then reserve version count should return 2 version", args{indexesResp: &kvstore.Response{
			Kvs: []*kvstore.KeyValue{
				{Key: []byte(etcdKeyPrefix + "svc1/1.0"), Value: "svc1-1.0"},
				{Key: []byte(etcdKeyPrefix + "svc1/1.2"), Value: "svc1-1.2"},
				{Key: []byte(etcdKeyPrefix + "svc1/1.3"), Value: "svc1-1.3"},
				{Key: []byte(etcdKeyPrefix + "svc1/2.0"), Value: "svc1-2.0"},
				{Key: []byte(etcdKeyPrefix + "svc1/3.0"), Value: "svc1-3.0"},
				{Key: []byte(etcdKeyPrefix + "svc2/1.0"), Value: "svc2-1.0"},
			}, Count: 6,
		}}, []*etcd.RotateServiceIDKey{
			{domainProject, "svc1-1.2"},
			{domainProject, "svc1-1.0"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := etcd.GetRetireServiceIDs(tt.args.indexesResp, reserve); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRetireServiceIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterInUsed(t *testing.T) {
	const (
		unusedServiceID   = "filterUnused"
		notExistServiceID = "not-exist"
		inusedServiceID   = "filterInused"
	)
	ctx := getContext()

	_, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceId:   unusedServiceID,
			ServiceName: unusedServiceID,
		},
	})
	assert.NoError(t, err)
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: unusedServiceID, Force: true})

	_, err = datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			ServiceId:   inusedServiceID,
			ServiceName: inusedServiceID,
		},
	})
	assert.NoError(t, err)
	defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: inusedServiceID, Force: true})

	_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
		Instance: &pb.MicroServiceInstance{
			ServiceId: inusedServiceID,
			HostName:  inusedServiceID,
		},
	})
	assert.NoError(t, err)

	type args struct {
		serviceIDKeys []*etcd.RotateServiceIDKey
	}
	tests := []struct {
		name string
		args args
		want []*etcd.RotateServiceIDKey
	}{
		{"input empty should return empty", args{serviceIDKeys: []*etcd.RotateServiceIDKey{}}, []*etcd.RotateServiceIDKey{}},
		{"input only one unused should return it", args{serviceIDKeys: []*etcd.RotateServiceIDKey{
			{DomainProject: domainProject, ServiceID: notExistServiceID},
			{DomainProject: domainProject, ServiceID: inusedServiceID},
			{DomainProject: domainProject, ServiceID: unusedServiceID},
		}}, []*etcd.RotateServiceIDKey{
			{DomainProject: domainProject, ServiceID: notExistServiceID},
			{DomainProject: domainProject, ServiceID: unusedServiceID},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := etcd.FilterNoInstance(ctx, tt.args.serviceIDKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterNoInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnregisterManyService(t *testing.T) {
	const serviceIDPrefix = "TestUnregisterManyService"
	ctx := getContext()

	t.Run("delete many should ok", func(t *testing.T) {
		const serviceVersionCount = 10
		var serviceIDs []*etcd.RotateServiceIDKey
		for i := 0; i < serviceVersionCount; i++ {
			serviceID := serviceIDPrefix + fmt.Sprintf("%v", i)
			_, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceId:   serviceID,
					ServiceName: serviceID,
				},
			})
			assert.NoError(t, err)
			serviceIDs = append(serviceIDs, &etcd.RotateServiceIDKey{DomainProject: domainProject, ServiceID: serviceID})
		}

		deleted := etcd.UnregisterManyService(ctx, serviceIDs)
		assert.Equal(t, int64(serviceVersionCount), deleted)

		_, i, err := etcdadpt.List(ctx, path.GenerateServiceKey(domainProject, serviceIDPrefix))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), i)
	})

	t.Run("delete inused without instance, should ok", func(t *testing.T) {
		service, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   serviceIDPrefix + "1",
				ServiceName: serviceIDPrefix + "1",
			},
		})
		assert.NoError(t, err)
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: service.ServiceId, Force: true})

		consumer, err := datasource.GetMetadataManager().RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceId:   serviceIDPrefix + "2",
				ServiceName: serviceIDPrefix + "2",
			},
		})
		assert.NoError(t, err)
		defer datasource.GetMetadataManager().UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: consumer.ServiceId, Force: true})

		_, err = datasource.GetMetadataManager().FindInstances(ctx, &pb.FindInstancesRequest{
			ConsumerServiceId: consumer.ServiceId,
			ServiceName:       serviceIDPrefix + "1",
		})
		assert.NoError(t, err)

		var serviceIDs []*etcd.RotateServiceIDKey
		serviceIDs = append(serviceIDs,
			&etcd.RotateServiceIDKey{DomainProject: domainProject, ServiceID: service.ServiceId},
			&etcd.RotateServiceIDKey{DomainProject: domainProject, ServiceID: consumer.ServiceId},
		)
		_, err = datasource.GetMetadataManager().RegisterInstance(ctx, &pb.RegisterInstanceRequest{
			Instance: &pb.MicroServiceInstance{
				ServiceId: service.ServiceId,
				HostName:  service.ServiceId,
			},
		})
		assert.NoError(t, err)

		deleted := etcd.UnregisterManyService(ctx, serviceIDs)
		assert.Equal(t, int64(2), deleted)

		_, i, err := etcdadpt.List(ctx, path.GenerateServiceKey(domainProject, serviceIDPrefix))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), i)
	})
}

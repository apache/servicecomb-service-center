//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package service_test

import (
	_ "github.com/ServiceComb/service-center/server/plugin/infra/registry/embededetcd"
	_ "github.com/ServiceComb/service-center/server/plugin/infra/registry/etcd"
)

import (
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/etcdsync"
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/ServiceComb/service-center/server/service"
	"testing"
	"time"
)

func init() {
	etcdsync.IsDebug = true
}

func TestServiceController_CreateDependenciesForMircServices(t *testing.T) {
	tryTimes := 3
	testCount := 10
	serviceResource, _, _ := service.AssembleResources()
	for i := 0; i < testCount; i++ {
		_, err := serviceResource.Create(getContext(), &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: fmt.Sprintf("service%d", i),
				AppId:       "test_deps",
				Version:     "1.0.0",
				Level:       "FRONT",
				Status:      "UP",
			},
		})
		if err != nil {
			util.Logger().Error(err.Error(), err)
			return
		}
	}

	for i := 0; i < testCount; i++ {
		go func(i int) {
			serviceName := fmt.Sprintf("service%d", i)
			_, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
				Dependencies: []*pb.MircroServiceDependency{
					{
						Consumer: &pb.DependencyMircroService{
							AppId:       "test_deps",
							ServiceName: serviceName,
							Version:     "1.0.0",
						},
						Providers: []*pb.DependencyMircroService{
							&pb.DependencyMircroService{
								AppId:       "test_deps",
								ServiceName: "service0",
								Version:     "1.0.0",
							},
						},
					},
				},
			})
			if err != nil {
				util.Logger().Errorf(err, "CreateDependenciesForMircServices %s failed.", serviceName)
				return
			}
		}(i)
	}
	for {
		time.Sleep(5 * time.Second)
		key := apt.GenerateProviderDependencyRuleKey("default/default", &pb.MicroServiceKey{
			AppId:       "test_deps",
			ServiceName: "service0",
			Version:     "1.0.0",
		})
		resp, err := backend.Registry().Do(getContext(),
			registry.GET, registry.WithStrKey(key))
		if err != nil {
			util.Logger().Errorf(err, "%s failed.", key)
			return
		}
		if len(resp.Kvs) == 0 {
			util.Logger().Warnf(nil, "%s: 0.", key)
			continue
		}
		d := &pb.MicroServiceDependency{}
		err = json.Unmarshal(resp.Kvs[0].Value, d)
		if err != nil {
			util.Logger().Errorf(err, "%s failed.", key)
			return
		}
		fmt.Println(key, ":", len(d.Dependency))
		if len(d.Dependency) != testCount {
			tryTimes--
			if tryTimes < 0 {
				t.Error("Time out to wait ")
			}
			continue
		}
		return
	}
}

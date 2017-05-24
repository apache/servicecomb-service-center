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
	_ "github.com/servicecomb/service-center/server/core/registry/embededetcd"
	_ "github.com/servicecomb/service-center/server/core/registry/etcd"
)

import (
	"encoding/json"
	"fmt"
	"github.com/servicecomb/service-center/etcdsync"
	apt "github.com/servicecomb/service-center/server/core"
	pb "github.com/servicecomb/service-center/server/core/proto"
	"github.com/servicecomb/service-center/server/core/registry"
	"github.com/servicecomb/service-center/server/service"
	"github.com/servicecomb/service-center/util"
	"testing"
	"time"
)

func init() {
	etcdsync.IsDebug = true
}

func TestServiceController_CreateDependenciesForMircServices(t *testing.T) {
	tryTimes := 3
	testCount := 10
	serviceResource, _, _ := service.AssembleResources(nil)
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
			util.LOGGER.Error(err.Error(), err)
			return
		}
	}

	for i := 0; i < testCount; i++ {
		go func(i int) {
			serviceName := fmt.Sprintf("service%d", i)
			_, err := serviceResource.CreateDependenciesForMircServices(getContext(), &pb.CreateDependenciesRequest{
				Dependencies: []*pb.MircroServiceDependency{
					&pb.MircroServiceDependency{
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
				util.LOGGER.Errorf(err, "CreateDependenciesForMircServices %s failed.", serviceName)
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
		resp, err := registry.GetRegisterCenter().Do(getContext(), &registry.PluginOp{
			Action: registry.GET,
			Key:    []byte(key),
		})
		if err != nil {
			util.LOGGER.Errorf(err, "%s failed.", key)
			return
		}
		if len(resp.Kvs) == 0 {
			util.LOGGER.Warnf(nil, "%s: 0.", key)
			continue
		}
		d := &service.MircroServiceDependency{}
		err = json.Unmarshal(resp.Kvs[0].Value, d)
		if err != nil {
			util.LOGGER.Errorf(err, "%s failed.", key)
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

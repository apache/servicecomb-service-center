// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"encoding/json"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

func DeleteDependencyForDeleteService(domainProject string, serviceID string, service *pb.MicroServiceKey) (registry.PluginOp, error) {
	key := apt.GenerateConsumerDependencyQueueKey(domainProject, serviceID, apt.DepsQueueUUID)
	conDep := new(pb.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*pb.MicroServiceKey{}
	conDep.Override = true
	data, err := json.Marshal(conDep)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

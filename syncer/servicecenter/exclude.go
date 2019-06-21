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
package servicecenter

import (
	"errors"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// exclude instances from other servicecenter, exclude the expired instances in the maps
func (s *servicecenter) exclude(data *pb.SyncData, mapping pb.SyncMapping) (*pb.SyncData, pb.SyncMapping) {
	services := make([]*pb.SyncService, 0, 10)
	instances := make([]*pb.SyncInstance, 0, 10)
	maps := make(pb.SyncMapping, 0, len(mapping))
	for _, inst := range data.Instances {
		if index := mapping.CurrentIndex(inst.InstanceId); index != -1 {
			// exclude the expired instances in the maps
			maps = append(maps, mapping[index])
			continue
		}
		svc := searchService(inst, data.Services)
		if svc == nil {
			err := errors.New("service does not exist")
			log.Errorf(err, "servicecenter.exclude, serviceID = %s, instanceId = %s", inst.ServiceId, inst.InstanceId)
			continue
		}
		// exclude instances from other servicecenter
		instances = append(instances, inst)
		services = append(services, svc)
	}
	data.Services = services
	data.Instances = instances
	return data, maps
}

// searchService search service by instance
func searchService(instance *pb.SyncInstance, services []*pb.SyncService) *pb.SyncService {
	for _, svc := range services {
		if instance.ServiceId == svc.ServiceId {
			return svc
		}
	}
	return nil
}

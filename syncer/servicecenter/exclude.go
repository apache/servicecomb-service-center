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
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// exclude instances from other servicecenter, exclude the expired instances in the maps
func (s *servicecenter) exclude(data *pb.SyncData, mapping pb.SyncMapping) (*pb.SyncData, pb.SyncMapping) {
	services := make([]*pb.SyncService, 0, 10)
	maps := make(pb.SyncMapping, 0, len(mapping))
	for _, svc := range data.Services {

		nis := make([]*scpb.MicroServiceInstance, 0, len(svc.Instances))
		for _, inst := range svc.Instances {
			if index := mapping.CurrentIndex(inst.InstanceId); index != -1 {
				// exclude the expired instances in the maps
				maps = append(maps, mapping[index])
				continue
			}
			// exclude instances from other servicecenter
			nis = append(nis, inst)
		}

		svc.Instances = nis
		services = append(services, svc)
	}
	data.Services = services

	return data, maps
}

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
	"strings"

	"github.com/apache/servicecomb-service-center/server/admin/model"
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// transform servicecenter service cache to SyncData
func transform(cache *model.Cache) (data *pb.SyncData) {
	data = &pb.SyncData{Services: make([]*pb.SyncService, 0, 10)}
	for _, svc := range cache.Microservices {
		instances := instancesFromService(svc.Value, cache.Instances)
		data.Services = append(data.Services, &pb.SyncService{
			DomainProject: strings.Join(strings.Split(svc.Key, "/")[4:6], "/"),
			Service:       svc.Value,
			Instances:     instances,
		})
	}
	return
}

//  instancesFromService Extract instance information from the service cache of the servicecenter
func instancesFromService(service *scpb.MicroService, instances []*model.Instance) []*scpb.MicroServiceInstance {
	instList := make([]*scpb.MicroServiceInstance, 0, 10)
	for _, inst := range instances {
		if inst.Value.ServiceId == service.ServiceId {
			instList = append(instList, inst.Value)
		}
	}
	return instList
}

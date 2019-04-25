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
package repository

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type Adaptor interface {
	New(endpoints []string) (Repository, error)
}

type Repository interface {
	GetAll(ctx context.Context) (*pb.SyncData, error)
	CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
	DeleteService(ctx context.Context, domainProject, serviceId string) error
	ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
	RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error)
	UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) error
	DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error)
	Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) error
}

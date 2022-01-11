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

package proto

import (
	"context"

	"github.com/go-chassis/cari/discovery"
	"google.golang.org/grpc"
)

type ServiceInstanceCtrlServer interface {
	Register(context.Context, *discovery.RegisterInstanceRequest) (*discovery.RegisterInstanceResponse, error)
	Unregister(context.Context, *discovery.UnregisterInstanceRequest) (*discovery.UnregisterInstanceResponse, error)
	Heartbeat(context.Context, *discovery.HeartbeatRequest) (*discovery.HeartbeatResponse, error)
	Find(context.Context, *discovery.FindInstancesRequest) (*discovery.FindInstancesResponse, error)
	GetInstances(context.Context, *discovery.GetInstancesRequest) (*discovery.GetInstancesResponse, error)
	GetOneInstance(context.Context, *discovery.GetOneInstanceRequest) (*discovery.GetOneInstanceResponse, error)
	UpdateStatus(context.Context, *discovery.UpdateInstanceStatusRequest) (*discovery.UpdateInstanceStatusResponse, error)
	UpdateInstanceProperties(context.Context, *discovery.UpdateInstancePropsRequest) (*discovery.UpdateInstancePropsResponse, error)
	Watch(*discovery.WatchInstanceRequest, ServiceInstanceCtrlWatchServer) error
	HeartbeatSet(context.Context, *discovery.HeartbeatSetRequest) (*discovery.HeartbeatSetResponse, error)
}
type ServiceInstanceCtrlWatchServer interface {
	Send(*discovery.WatchInstanceResponse) error
	grpc.ServerStream
}
type GovernServiceCtrlServer interface {
	GetServiceDetail(context.Context, *discovery.GetServiceRequest) (*discovery.GetServiceDetailResponse, error)
	GetServicesInfo(context.Context, *discovery.GetServicesInfoRequest) (*discovery.GetServicesInfoResponse, error)
	GetApplications(context.Context, *discovery.GetAppsRequest) (*discovery.GetAppsResponse, error)
	GetServicesStatistics(context.Context, *discovery.GetServicesRequest) (*discovery.GetServicesInfoStatisticsResponse, error)
}

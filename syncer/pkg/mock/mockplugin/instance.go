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

package mockplugin

import (
	"context"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

var (
	registerInstance   func(ctx context.Context, domainProject, serviceId string, instance *pb.SyncInstance) (string, error)
	unregisterInstance func(ctx context.Context, domainProject, serviceId, instanceId string) error
	heartbeat          func(ctx context.Context, domainProject, serviceId, instanceId string) error
)

func SetRegisterInstance(handler func(ctx context.Context, domainProject, serviceId string, instance *pb.SyncInstance) (string, error)) {
	registerInstance = handler
}

func SetUnregisterInstance(handler func(ctx context.Context, domainProject, serviceId, instanceId string) error) {
	unregisterInstance = handler
}

func SetHeartbeat(handler func(ctx context.Context, domainProject, serviceId, instanceId string) error) {
	heartbeat = handler
}

func (c *mockPlugin) RegisterInstance(ctx context.Context, domainProject, serviceID string, instance *pb.SyncInstance) (string, error) {
	if registerInstance != nil {
		return registerInstance(ctx, domainProject, serviceID, instance)
	}
	return "4d41a637471f11e9888cfa163eca30e0", nil
}

func (c *mockPlugin) UnregisterInstance(ctx context.Context, domainProject, serviceID, instanceID string) error {
	if unregisterInstance != nil {
		return unregisterInstance(ctx, domainProject, serviceID, instanceID)
	}
	return nil
}

func (c *mockPlugin) Heartbeat(ctx context.Context, domainProject, serviceID, instanceID string) error {
	if heartbeat != nil {
		return heartbeat(ctx, domainProject, serviceID, instanceID)
	}
	return nil
}

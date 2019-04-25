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
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error) {
	instanceID, err := c.cli.RegisterInstance(ctx, domainProject, serviceId, instance)
	if err != nil {
		return "", err
	}
	return instanceID, nil
}

func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) error {
	err := c.cli.UnregisterInstance(ctx, domainProject, serviceId, instanceId)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error) {
	instances, err := c.cli.DiscoveryInstances(ctx, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) error {
	err := c.cli.Heartbeat(ctx, domainProject, serviceId, instanceId)
	if err != nil {
		return err
	}
	return nil
}

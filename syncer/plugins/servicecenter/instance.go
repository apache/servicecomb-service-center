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

	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// RegisterInstance register instance to servicecenter
func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceID string, syncInstance *pb.SyncInstance) (string, error) {
	instance := toInstance(syncInstance)
	instance.InstanceId = ""
	instance.ServiceId = serviceID
	domain, project := util.FromDomainProject(domainProject)
	instanceID, err := c.cli.RegisterInstance(ctx, domain, project, serviceID, instance)
	if err != nil {
		return "", err
	}
	return instanceID, nil
}

// UnregisterInstance unregister instance from servicecenter
func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceID, instanceID string) error {
	domain, project := util.FromDomainProject(domainProject)
	err := c.cli.UnregisterInstance(ctx, domain, project, serviceID, instanceID)
	if err != nil {
		return err
	}
	return nil
}

// Heartbeat sends heartbeat to servicecenter
func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceID, instanceID string) error {
	domain, project := util.FromDomainProject(domainProject)
	err := c.cli.Heartbeat(ctx, domain, project, serviceID, instanceID)
	if err != nil {
		return err
	}
	return nil
}

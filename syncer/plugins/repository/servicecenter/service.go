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

func (c *Client) CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	serviceID, err := c.cli.CreateService(ctx, domainProject, service)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}

func (c *Client) DeleteService(ctx context.Context, domainProject, serviceId string) error {
	err := c.cli.DeleteService(ctx, domainProject, serviceId)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	serviceID, err := c.cli.ServiceExistence(ctx, domainProject, service.AppId, service.ServiceName, service.Version, service.Environment)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}

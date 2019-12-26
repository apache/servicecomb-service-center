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

	"github.com/apache/servicecomb-service-center/pkg/log"
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/gogo/protobuf/proto"
)

// CreateService creates the service of servicecenter
func (c *Client) CreateService(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	service := toService(syncService)
	service.ServiceId = ""
	serviceID, err := c.cli.CreateService(ctx, domainProject, service)
	if err != nil {
		log.Debugf("create service err %v", err)
		return "", err
	}

	matches := pb.Expansions(syncService.Expansions).Find(expansionSchema, map[string]string{})
	if len(matches) > 0 {
		schemas := make([]*scpb.Schema, 0, len(matches))
		for _, expansion := range matches {
			schema := &scpb.Schema{}
			err1 := proto.Unmarshal(expansion.Bytes, schema)
			if err1 != nil {
				log.Errorf(err1, "proto unmarshal %s service schema, serviceID = %s, kind = %v, content = %v failed",
					PluginName, serviceID, expansion.Kind, expansion.Bytes)
			}
			schemas = append(schemas, schema)
		}
		err2 := c.CreateSchemas(ctx, domainProject, serviceID, schemas)
		if err2 != nil {
			log.Errorf(err2, "create service schemas failed, serviceID = %s", serviceID)
		}
	}

	return serviceID, nil
}

// DeleteService deletes service from servicecenter
func (c *Client) DeleteService(ctx context.Context, domainProject, serviceId string) error {
	err := c.cli.DeleteService(ctx, domainProject, serviceId)
	if err != nil {
		return err
	}
	return nil
}

// ServiceExistence Checkes service exists in servicecenter
func (c *Client) ServiceExistence(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	service := toService(syncService)
	serviceID, err := c.cli.ServiceExistence(ctx, domainProject, service.AppId, service.ServiceName, service.Version, service.Environment)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}

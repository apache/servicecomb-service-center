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
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	pbsc "github.com/apache/servicecomb-service-center/syncer/proto/sc"
	scpb "github.com/go-chassis/cari/discovery"
	"google.golang.org/protobuf/proto"
)

// CreateService creates the service of servicecenter
func (c *Client) CreateService(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	service := toService(syncService)
	service.ServiceId = ""
	domain, project := util.FromDomainProject(domainProject)
	serviceID, err := c.cli.CreateService(ctx, domain, project, service)
	if err != nil {
		log.Debug(fmt.Sprintf("create service err %v", err))
		return "", err
	}

	matches := pb.Expansions(syncService.Expansions).Find(expansionSchema, map[string]string{})
	if len(matches) > 0 {
		schemas := make([]*scpb.Schema, 0, len(matches))
		for _, expansion := range matches {
			schema := &pbsc.Schema{}
			err1 := proto.Unmarshal(expansion.Bytes, schema)
			if err1 != nil {
				log.Error(fmt.Sprintf("proto unmarshal %s service schema, serviceID = %s, kind = %v, content = %v failed",
					PluginName, serviceID, expansion.Kind, expansion.Bytes), err1)
			}
			schematised := &scpb.Schema{
				SchemaId: schema.SchemaId,
				Summary:  schema.Summary,
				Schema:   schema.Schema,
			}
			schemas = append(schemas, schematised)
		}
		err := c.CreateSchemas(ctx, domain, project, serviceID, schemas)
		if err != nil {
			log.Error(fmt.Sprintf("create service schemas failed, serviceID = %s", serviceID), err)
		}
	}

	return serviceID, nil
}

// DeleteService deletes service from servicecenter
func (c *Client) DeleteService(ctx context.Context, domainProject, serviceID string) error {
	domain, project := util.FromDomainProject(domainProject)
	err := c.cli.DeleteService(ctx, domain, project, serviceID)
	if err != nil {
		return err
	}
	return nil
}

// ServiceExistence Checkes service exists in servicecenter
func (c *Client) ServiceExistence(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	service := toService(syncService)
	domain, project := util.FromDomainProject(domainProject)
	serviceID, err := c.cli.ServiceExistence(ctx, domain, project, service.AppId, service.ServiceName, service.Version, service.Environment)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}

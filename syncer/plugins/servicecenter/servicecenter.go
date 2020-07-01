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
	"github.com/apache/servicecomb-service-center/pkg/client/sc"

	"github.com/apache/servicecomb-service-center/pkg/log"
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const PluginName = "servicecenter"

func init() {
	// Register self as a repository plugin
	plugins.RegisterPlugin(&plugins.Plugin{
		Kind: plugins.PluginServicecenter,
		Name: PluginName,
		New:  New,
	})
}

type adaptor struct{}

func New() plugins.PluginInstance {
	return &adaptor{}
}

// New repository with endpoints
func (*adaptor) New(opts ...plugins.SCConfigOption) (plugins.Servicecenter, error) {
	cli, err := client.NewSCClient(plugins.ToSCConfig(opts...))
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

type Client struct {
	cli *sc.SCClient
}

// GetAll get and transform servicecenter data to SyncData
func (c *Client) GetAll(ctx context.Context) (*pb.SyncData, error) {
	cache, err := c.cli.GetScCache(ctx)
	if err != nil {
		return nil, err
	}

	schemas := make([]*scpb.Schema, 0, len(cache.Microservices))
	for _, service := range cache.Microservices {
		if len(service.Value.Schemas) == 0 {
			continue
		}

		domainProject := getDomainProjectFromServiceKey(service.Key)
		if domainProject == "" {
			continue
		}

		ss, err := c.cli.GetSchemasByServiceId(ctx, domainProject, service.Value.ServiceId)
		if err != nil {
			log.Warnf("get schemas by serviceId failed: %s", err)
			continue
		}
		schemas = append(schemas, ss...)
	}
	return toSyncData(cache, schemas), nil
}

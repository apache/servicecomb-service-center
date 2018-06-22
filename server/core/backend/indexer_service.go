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
package backend

import (
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
)

type serviceSelector struct {
	*pb.MicroServiceKey
	serviceId string
}

func (cfg *serviceSelector) ServiceId(id string) serviceSelector {
	cfg.serviceId = id
	return *cfg
}

func (cfg *serviceSelector) AppId(id string) serviceSelector {
	cfg.MicroServiceKey.AppId = id
	return *cfg
}

func (cfg *serviceSelector) ServiceName(name string) serviceSelector {
	cfg.MicroServiceKey.ServiceName = name
	return *cfg
}

func (cfg *serviceSelector) VersionRule(v string) serviceSelector {
	cfg.MicroServiceKey.Version = v
	return *cfg
}

func ServiceSelector(domainProject string) serviceSelector {
	return serviceSelector{MicroServiceKey: &pb.MicroServiceKey{
		Tenant: domainProject,
	}}
}

type ServiceIndexer struct {
	*Indexer
}

func (i *ServiceIndexer) GetAll(ctx context.Context, cfg serviceSelector) ([]*pb.MicroService, error) {
	if len(cfg.serviceId) > 0 {
		i.Indexer.Search(ctx)
	}
	return nil, nil
}

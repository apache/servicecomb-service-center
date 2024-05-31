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

package mongo

import (
	"context"

	"github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

func (ds *MetadataManager) CountService(ctx context.Context, request *discovery.GetServiceCountRequest) (
	*discovery.GetServiceCountResponse, error) {
	options := []util.Option{util.NotGlobal(), util.Domain(request.Domain)}
	if request.Project != "" {
		options = append(options, util.Project(request.Project))
	}
	count, err := dao.CountService(ctx, util.NewFilter(options...))
	if err != nil {
		return nil, err
	}
	return &discovery.GetServiceCountResponse{
		Count: count,
	}, nil
}

func (ds *MetadataManager) CountInstance(ctx context.Context, request *discovery.GetServiceCountRequest) (
	*discovery.GetServiceCountResponse, error) {
	inFilter, err := ds.getNotGlobalServiceFilter(ctx)
	if err != nil {
		return nil, err
	}
	options := []util.Option{util.Domain(request.Domain)}
	if request.Project != "" {
		options = append(options, util.Project(request.Project))
	}
	options = append(options, util.InstanceServiceID(inFilter))
	count, err := dao.CountInstance(ctx, util.NewFilter(options...))
	if err != nil {
		return nil, err
	}
	return &discovery.GetServiceCountResponse{
		Count: count,
	}, nil
}

func (ds *MetadataManager) getNotGlobalServiceFilter(ctx context.Context) (bson.M, error) {
	serviceIDs := make([]string, 0)
	services, err := dao.GetServices(ctx, util.NewFilter(util.Global()))
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return util.NewFilter(util.NotIn(serviceIDs)), nil
}

func (ds *MetadataManager) CountEnvironment(ctx context.Context, request *ev.GetEnvironmentCountRequest) (*ev.GetEnvironmentCountResponse, error) {
	return nil, nil
}

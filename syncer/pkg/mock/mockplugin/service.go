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
	createServiceHandler func(ctx context.Context, domainProject string, service *pb.SyncService) (string, error)
	deleteService        func(ctx context.Context, domainProject, serviceId string) error
	serviceExistence     func(ctx context.Context, domainProject string, service *pb.SyncService) (string, error)
)

func SetCreateService(handler func(ctx context.Context, domainProject string, service *pb.SyncService) (string, error)) {
	createServiceHandler = handler
}

func SetDeleteService(handler func(ctx context.Context, domainProject, serviceId string) error) {
	deleteService = handler
}

func SetServiceExistence(handler func(ctx context.Context, domainProject string, source *pb.SyncService) (string, error)) {
	serviceExistence = handler
}

func (c *mockPlugin) CreateService(ctx context.Context, domainProject string, service *pb.SyncService) (string, error) {
	if createServiceHandler != nil {
		return createServiceHandler(ctx, domainProject, service)
	}
	return "5db1b794aa6f8a875d6e68110260b5491ee7e223", nil
}

func (c *mockPlugin) DeleteService(ctx context.Context, domainProject, serviceId string) error {
	if createServiceHandler != nil {
		return deleteService(ctx, domainProject, serviceId)
	}
	return nil
}

func (c *mockPlugin) ServiceExistence(ctx context.Context, domainProject string, service *pb.SyncService) (string, error) {
	if serviceExistence != nil {
		return serviceExistence(ctx, domainProject, service)
	}
	return "5db1b794aa6f8a875d6e68110260b5491ee7e223", nil
}

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

package eureka

import (
	"context"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

// CreateService Eureka's application is created with instance and does not need to be processed here.
func (c *Client) CreateService(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	return syncService.Name, nil
}

// DeleteService Eureka's application is created with instance and does not need to be processed here.
func (c *Client) DeleteService(context.Context, string, string) error {
	return nil
}

// ServiceExistence Eureka's application is created with instance and does not need to be processed here.
func (c *Client) ServiceExistence(ctx context.Context, domainProject string, syncService *pb.SyncService) (string, error) {
	return syncService.Name, nil
}

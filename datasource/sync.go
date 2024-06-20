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

package datasource

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	svc "github.com/apache/servicecomb-service-center/syncer/whitelist/service"
)

var ErrSyncAllKeyExists = errors.New("sync all key already exists")

type SyncManager interface {
	SyncAll(ctx context.Context) error
}

// EnableSync implement:
// 1.Check if the service exists
// 2.Check if the service is in the whitelist
func EnableSync(ctx context.Context, serviceID string) (bool, error) {
	remoteIP := util.GetIPFromContext(ctx)
	microservice, err := GetMetadataManager().GetService(ctx, &pb.GetServiceRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed, operator: %s", serviceID, remoteIP), err)
		return false, err
	}
	// Whether the service is in the whitelist. If the service is in the whitelist and needs to be synced.
	if !svc.IsExistInWhiteList(microservice.ServiceName) {
		return false, nil
	}
	return true, nil
}

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

package util

import (
	"context"
	"errors"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	apt "github.com/apache/servicecomb-service-center/server/core"
	scerr "github.com/apache/servicecomb-service-center/pkg/registry"
)

func HeartbeatUtil(ctx context.Context, domainProject string, serviceID string, instanceID string) (leaseID int64, ttl int64, _ *scerr.Error) {
	leaseID, err := GetLeaseID(ctx, domainProject, serviceID, instanceID)
	if err != nil {
		return leaseID, ttl, scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	ttl, err = KeepAliveLease(ctx, domainProject, serviceID, instanceID, leaseID)
	if err != nil {
		return leaseID, ttl, scerr.NewError(scerr.ErrInstanceNotExists, err.Error())
	}
	return leaseID, ttl, nil
}

func KeepAliveLease(ctx context.Context, domainProject, serviceID, instanceID string, leaseID int64) (ttl int64, err error) {
	if leaseID == -1 {
		return ttl, errors.New("leaseId not exist, instance not exist")
	}
	ttl, err = client.KeepAlive(ctx,
		client.WithStrKey(apt.GenerateInstanceLeaseKey(domainProject, serviceID, instanceID)),
		client.WithLease(leaseID))
	if err != nil {
		return ttl, err
	}
	return ttl, nil
}

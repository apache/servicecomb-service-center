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

package checker

import (
	"context"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/service/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func init() {
	heartbeat.Install("checker", NewHeartBeatChecker)
}

type HeartBeatChecker struct {
}

func NewHeartBeatChecker(opts heartbeat.Options) (heartbeat.HealthCheck, error) {
	return &HeartBeatChecker{}, nil
}

func (h *HeartBeatChecker) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := updateInstanceRefreshTime(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrInstanceNotExists, err.Error())),
		}
		return resp, err
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess,
			"Update service instance heartbeat successfully."),
	}, nil
}

func (h *HeartBeatChecker) CheckInstance(ctx context.Context, instance *pb.MicroServiceInstance) error {
	// do nothing
	return nil
}

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

package rpc

import (
	"context"
	"fmt"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	HealthStatusConnected = "CONNECTED"
	HealthStatusAbnormal  = "ABNORMAL"
	HealthStatusClose     = "CLOSE"
)

type Server struct {
	v1sync.UnimplementedEventServiceServer
}

func (s *Server) Sync(ctx context.Context, events *v1sync.EventList) (*v1sync.Results, error) {
	log.Info(fmt.Sprintf("Received: %v", events.Events[0].Action))
	return &v1sync.Results{}, nil
}

func (s *Server) Health(ctx context.Context, request *v1sync.HealthRequest) (*v1sync.HealthReply, error) {
	resp := &v1sync.HealthReply{
		Status:         HealthStatusConnected,
		LocalTimestamp: time.Now().UnixNano(),
	}
	// TODO enable to close syncer
	syncerEnabled := config.GetBool("sync.enableOnStart", false)
	if !syncerEnabled {
		resp.Status = HealthStatusClose
		return resp, nil
	}
	return resp, nil
}

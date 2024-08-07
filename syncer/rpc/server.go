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

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"
)

const (
	HealthStatusConnected = "CONNECTED"
	HealthStatusAbnormal  = "ABNORMAL"
	HealthStatusClose     = "CLOSE"
	HealthStatusAuthFail  = "AuthFail"

	RbacAllowedAccountName = "sync-user"
	RbacAllowedRoleName    = "sync-admin"
)

func NewServer() *Server {
	return &Server{
		replicator: replicator.Manager(),
	}
}

type Server struct {
	v1sync.UnimplementedEventServiceServer

	replicator replicator.Replicator
}

func (s *Server) Sync(ctx context.Context, events *v1sync.EventList) (*v1sync.Results, error) {
	err := auth(ctx)
	if err != nil {
		log.Error("auth failed", err)
		return generateFailedResults(events, err)
	}

	log.Info(fmt.Sprintf("start sync: %s", events.Flag()))

	res := s.replicator.Persist(ctx, events)

	return s.toResults(res), nil
}

func generateFailedResults(events *v1sync.EventList, err error) (*v1sync.Results, error) {
	if events == nil || len(events.Events) == 0 {
		return &v1sync.Results{Results: map[string]*v1sync.Result{}}, nil
	}
	rsts := make(map[string]*v1sync.Result, len(events.Events))
	for _, evt := range events.Events {
		rsts[evt.Id] = &v1sync.Result{
			Code:    resource.Fail,
			Message: err.Error(),
		}
	}
	return &v1sync.Results{Results: rsts}, nil
}

func (s *Server) toResults(results []*resource.Result) *v1sync.Results {
	syncResult := make(map[string]*v1sync.Result, len(results))
	for _, r := range results {
		syncResult[r.EventID] = &v1sync.Result{
			Code:    r.Status,
			Message: r.Message,
		}
	}
	return &v1sync.Results{
		Results: syncResult,
	}
}

func (s *Server) Health(ctx context.Context, _ *v1sync.HealthRequest) (*v1sync.HealthReply, error) {
	resp := &v1sync.HealthReply{
		Status:         HealthStatusConnected,
		LocalTimestamp: time.Now().UnixNano(),
	}
	err := auth(ctx)
	if err != nil {
		resp.Status = HealthStatusAuthFail
		log.Error("auth failed", err)
		return resp, nil
	}

	// TODO enable to close syncer
	if !config.GetConfig().Sync.EnableOnStart {
		resp.Status = HealthStatusClose
		log.Error("unexpected health check when syncer is closed", nil)
		return resp, nil
	}
	return resp, nil
}

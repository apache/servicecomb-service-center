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

package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/connection"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/event"
)

const GRPC = "gRPC"

func Handle(watcher *event.InstanceSubscriber, stream proto.ServiceInstanceCtrl_WatchServer) (err error) {
	timer := time.NewTimer(connection.HeartbeatInterval)
	defer timer.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return
		case <-timer.C:
			timer.Reset(connection.HeartbeatInterval)
		case job := <-watcher.Job:
			if job == nil {
				err = errors.New("channel is closed")
				log.Errorf(err, "watcher caught an exception, subject: %s, group: %s",
					watcher.Subject(), watcher.Group())
				return
			}
			if job.Response == nil {
				continue
			}
			resp := job.Response
			log.Infof("event is coming in, watcher, subject: %s, group: %s",
				watcher.Subject(), watcher.Group())

			err = stream.Send(resp)
			connection.ReportPublishCompleted(job, err)
			if err != nil {
				log.Errorf(err, "send message error, subject: %s, group: %s",
					watcher.Subject(), watcher.Group())
				watcher.SetError(err)
				return
			}
			util.ResetTimer(timer, connection.HeartbeatInterval)
		}
	}
}

func Watch(ctx context.Context, serviceID string, stream proto.ServiceInstanceCtrl_WatchServer) (err error) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	watcher := event.NewInstanceSubscriber(serviceID, domainProject)
	err = event.Center().AddSubscriber(watcher)
	if err != nil {
		return
	}
	connection.ReportSubscriber(domain, GRPC, 1)
	err = Handle(watcher, stream)
	connection.ReportSubscriber(domain, GRPC, -1)
	return
}

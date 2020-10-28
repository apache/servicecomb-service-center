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

package notify

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/proto"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"time"
)

func HandleWatchJob(watcher *InstanceEventListWatcher, stream proto.ServiceInstanceCtrlWatchServer) (err error) {
	timer := time.NewTimer(HeartbeatInterval)
	defer timer.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return
		case <-timer.C:
			timer.Reset(HeartbeatInterval)
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
			ReportPublishCompleted(job, err)
			if err != nil {
				log.Errorf(err, "send message error, subject: %s, group: %s",
					watcher.Subject(), watcher.Group())
				watcher.SetError(err)
				return
			}
			util.ResetTimer(timer, HeartbeatInterval)
		}
	}
}

func DoStreamListAndWatch(ctx context.Context, serviceID string, f func() ([]*pb.WatchInstanceResponse, int64), stream proto.ServiceInstanceCtrlWatchServer) (err error) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	watcher := NewInstanceEventListWatcher(serviceID, apt.GetInstanceRootKey(domainProject)+"/", f)
	err = GetNotifyCenter().AddSubscriber(watcher)
	if err != nil {
		return
	}
	ReportSubscriber(domain, GRPC, 1)
	err = HandleWatchJob(watcher, stream)
	ReportSubscriber(domain, GRPC, -1)
	return
}

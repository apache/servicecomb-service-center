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
package notification

import (
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"time"
)

func HandleWatchJob(watcher *ListWatcher, stream pb.ServiceInstanceCtrl_WatchServer) (err error) {
	timer := time.NewTimer(DEFAULT_HEARTBEAT_INTERVAL)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			timer.Reset(DEFAULT_HEARTBEAT_INTERVAL)

			// TODO grpc 长连接心跳？
		case job := <-watcher.Job:
			if job == nil {
				err = errors.New("channel is closed")
				log.Errorf(err, "watcher caught an exception, subject: %s, group: %s",
					watcher.Subject(), watcher.Group())
				return
			}
			resp := job.Response
			log.Infof("event is coming in, watcher, subject: %s, group: %s",
				watcher.Subject(), watcher.Group())

			err = stream.Send(resp)
			if err != nil {
				log.Errorf(err, "send message error, subject: %s, group: %s",
					watcher.Subject(), watcher.Group())
				watcher.SetError(err)
				return
			}

			util.ResetTimer(timer, DEFAULT_HEARTBEAT_INTERVAL)
		}
	}
}

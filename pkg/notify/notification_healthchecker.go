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

import "github.com/apache/servicecomb-service-center/pkg/log"

const (
	ServerCheckerName  = "__HealthChecker__"
	ServerCheckSubject = "__NotifyServerHealthCheck__"
)

//Notifier 健康检查
type ServiceHealthChecker struct {
	Subscriber
}

type ServiceHealthCheckJob struct {
	Event
	ErrorSubscriber Subscriber
}

func (s *ServiceHealthChecker) OnMessage(job Event) {
	j := job.(*ServiceHealthCheckJob)
	err := j.ErrorSubscriber.Err()

	if j.ErrorSubscriber.Type() == NOTIFTY {
		log.Errorf(nil, "remove %s watcher failed, here cause a dead lock, subject: %s, group: %s",
			j.ErrorSubscriber.Type(), j.ErrorSubscriber.Subject(), j.ErrorSubscriber.Group())
		return
	}

	log.Debugf("notification service remove %s watcher, error: %v, subject: %s, group: %s",
		j.ErrorSubscriber.Type(), err, j.ErrorSubscriber.Subject(), j.ErrorSubscriber.Group())
	s.Service().RemoveSubscriber(j.ErrorSubscriber)
}

func NewNotifyServiceHealthChecker() *ServiceHealthChecker {
	return &ServiceHealthChecker{
		Subscriber: NewSubscriber(NOTIFTY, ServerCheckSubject, ServerCheckerName),
	}
}

func NewNotifyServiceHealthCheckJob(s Subscriber) *ServiceHealthCheckJob {
	return &ServiceHealthCheckJob{
		Event:           NewEvent(NOTIFTY, ServerCheckSubject, ServerCheckerName),
		ErrorSubscriber: s,
	}
}

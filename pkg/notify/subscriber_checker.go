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
	"github.com/apache/servicecomb-service-center/pkg/log"
)

const (
	groupCheck   = "__HealthChecker__"
	subjectCheck = "__SubscriberHealthCheck__"
)

var TypeCheck = RegisterType("CHECKER", DefaultQueueSize)

//Notifier 健康检查
type SubscriberChecker struct {
	Subscriber
}

type ErrEvent struct {
	Event
	Subscriber Subscriber
}

func (s *SubscriberChecker) OnMessage(evt Event) {
	j := evt.(*ErrEvent)
	err := j.Subscriber.Err()

	if j.Subscriber.Type() == TypeCheck {
		log.Errorf(nil, "remove %s subscriber failed, here cause a dead lock, subject: %s, group: %s",
			j.Subscriber.Type(), j.Subscriber.Subject(), j.Subscriber.Group())
		return
	}

	log.Debugf("notification service remove %s subscriber, error: %v, subject: %s, group: %s",
		j.Subscriber.Type(), err, j.Subscriber.Subject(), j.Subscriber.Group())
	s.Service().RemoveSubscriber(j.Subscriber)
}

func NewSubscriberChecker() *SubscriberChecker {
	return &SubscriberChecker{
		Subscriber: NewSubscriber(TypeCheck, subjectCheck, groupCheck),
	}
}

func NewErrEvent(by Subscriber) *ErrEvent {
	return &ErrEvent{
		Event:      NewEvent(TypeCheck, subjectCheck, groupCheck),
		Subscriber: by,
	}
}

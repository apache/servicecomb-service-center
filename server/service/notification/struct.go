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
	"fmt"
	"strconv"
	"time"
)

const (
	DEFAULT_MAX_QUEUE = 1000
	DEFAULT_TIMEOUT   = 30 * time.Second

	NOTIFTY NotifyType = iota
	INSTANCE
	typeEnd
)

type NotifyType int

func (nt NotifyType) String() string {
	if int(nt) < len(notifyTypeNames) {
		return notifyTypeNames[nt]
	}
	return "NotifyType" + strconv.Itoa(int(nt))
}

type NotifyServiceConfig struct {
	AddTimeout    time.Duration
	NotifyTimeout time.Duration
	MaxQueue      int64
}

func (nsc NotifyServiceConfig) String() string {
	return fmt.Sprintf("{acceptQueue: %d, accept: %s, notify: %s}",
		nsc.MaxQueue, nsc.AddTimeout, nsc.NotifyTimeout)
}

type Subscriber interface {
	Err() error
	SetError(err error)
	Id() string
	Subject() string
	Type() NotifyType
	Service() *NotifyService
	SetService(*NotifyService)
	OnAccept()
	OnMessage(job NotifyJob)
	Close()
}

type NotifyJob interface {
	SubscriberId() string
	Subject() string
	Type() NotifyType
}

type BaseSubscriber struct {
	id      string
	subject string
	nType   NotifyType
	service *NotifyService
	err     error
}

func (s *BaseSubscriber) Err() error {
	return s.err
}

func (s *BaseSubscriber) SetError(err error) {
	s.err = err
	// 触发清理job
	s.Service().AddJob(NewNotifyServiceHealthCheckJob(s))
}

func (s *BaseSubscriber) Id() string {
	return s.id
}

func (s *BaseSubscriber) Subject() string {
	return s.subject
}

func (s *BaseSubscriber) Type() NotifyType {
	return s.nType
}

func (s *BaseSubscriber) Service() *NotifyService {
	return s.service
}

func (s *BaseSubscriber) SetService(svc *NotifyService) {
	s.service = svc
}

func (s *BaseSubscriber) OnAccept() {
}

func (s *BaseSubscriber) OnMessage(job NotifyJob) {
	s.SetError(errors.New("do not call base notifier OnMessage method"))
}

func (s *BaseSubscriber) Close() {

}

type BaseNotifyJob struct {
	subscriberId string
	subject      string
	nType        NotifyType
}

func (s *BaseNotifyJob) SubscriberId() string {
	return s.subscriberId
}

func (s *BaseNotifyJob) Subject() string {
	return s.subject
}

func (s *BaseNotifyJob) Type() NotifyType {
	return s.nType
}

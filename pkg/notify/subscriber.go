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
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Subscriber interface {
	Id() string
	Subject() string
	Group() string
	Type() Type
	Service() *NotifyService
	SetService(*NotifyService)

	Err() error
	SetError(err error)

	Close()
	OnAccept()
	// The event bus will callback this function, so it must be non-blocked.
	OnMessage(Event)
}

type baseSubscriber struct {
	nType   Type
	id      string
	subject string
	group   string
	service *NotifyService
	err     error
}

func (s *baseSubscriber) Id() string                    { return s.id }
func (s *baseSubscriber) Subject() string               { return s.subject }
func (s *baseSubscriber) Group() string                 { return s.group }
func (s *baseSubscriber) Type() Type                    { return s.nType }
func (s *baseSubscriber) Service() *NotifyService       { return s.service }
func (s *baseSubscriber) SetService(svc *NotifyService) { s.service = svc }
func (s *baseSubscriber) Err() error                    { return s.err }
func (s *baseSubscriber) SetError(err error)            { s.err = err }
func (s *baseSubscriber) Close()                        {}
func (s *baseSubscriber) OnAccept()                     {}
func (s *baseSubscriber) OnMessage(job Event) {
	s.SetError(errors.New("do not call base notifier OnMessage method"))
}

func NewSubscriber(nType Type, subject, group string) Subscriber {
	return &baseSubscriber{
		id:      util.GenerateUuid(),
		group:   group,
		subject: subject,
		nType:   nType,
	}
}

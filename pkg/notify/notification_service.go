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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"sync"
)

type NotifyService struct {
	processors map[Type]*Processor
	closeMux   sync.RWMutex
	isClose    bool
}

func (s *NotifyService) init() {
	for _, t := range Types() {
		s.processors[t] = NewProcessor(t.String(), t.QueueSize())
	}
}

func (s *NotifyService) Start() {
	if !s.Closed() {
		log.Warnf("notify service is already running")
		return
	}
	s.closeMux.Lock()
	s.isClose = false
	s.closeMux.Unlock()

	// 错误subscriber清理
	s.AddSubscriber(NewNotifyServiceHealthChecker())

	s.startProcessors()

	log.Debugf("notify service is started")
}

func (s *NotifyService) AddSubscriber(n Subscriber) error {
	if n == nil {
		err := errors.New("required Subscriber")
		log.Errorf(err, "add subscriber failed")
		return err
	}

	p, ok := s.processors[n.Type()]
	if !ok {
		err := errors.New("unknown subscribe type")
		log.Errorf(err, "add %s subscriber[%s/%s] failed", n.Type(), n.Subject(), n.Group())
		return err
	}
	n.SetService(s)
	n.OnAccept()

	p.AddSubscriber(n)
	return nil
}

func (s *NotifyService) RemoveSubscriber(n Subscriber) {
	p, ok := s.processors[n.Type()]
	if !ok {
		return
	}

	p.Remove(n)
	n.Close()
}

func (s *NotifyService) startProcessors() {
	for _, p := range s.processors {
		p.Run()
	}
}

func (s *NotifyService) stopProcessors() {
	for _, p := range s.processors {
		p.Clear()
		p.Stop()
	}
}

//通知内容塞到队列里
func (s *NotifyService) Publish(job Event) error {
	if s.Closed() {
		return errors.New("add notify job failed for server shutdown")
	}

	p, ok := s.processors[job.Type()]
	if !ok {
		return errors.New("Unknown job type")
	}
	p.Accept(job)
	return nil
}

func (s *NotifyService) Closed() (b bool) {
	s.closeMux.RLock()
	b = s.isClose
	s.closeMux.RUnlock()
	return
}

func (s *NotifyService) Stop() {
	if s.Closed() {
		return
	}
	s.closeMux.Lock()
	s.isClose = true
	s.closeMux.Unlock()

	s.stopProcessors()

	log.Debug("notify service stopped")
}

func NewNotifyService() *NotifyService {
	ns := &NotifyService{
		processors: make(map[Type]*Processor),
		isClose:    true,
	}
	ns.init()
	return ns
}

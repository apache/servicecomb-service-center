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
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type Service struct {
	processors map[Type]*Processor
	mux        sync.RWMutex
	isClose    bool
}

func (s *Service) newProcessor(t Type) *Processor {
	s.mux.RLock()
	p, ok := s.processors[t]
	if ok {
		s.mux.RUnlock()
		return p
	}
	s.mux.RUnlock()

	s.mux.Lock()
	p, ok = s.processors[t]
	if ok {
		s.mux.Unlock()
		return p
	}
	p = NewProcessor(t.String(), t.QueueSize())
	s.processors[t] = p
	s.mux.Unlock()

	p.Run()
	return p
}

func (s *Service) Start() {
	if !s.Closed() {
		log.Warnf("notify service is already running")
		return
	}
	s.mux.Lock()
	s.isClose = false
	s.mux.Unlock()

	// 错误subscriber清理
	err := s.AddSubscriber(NewSubscriberChecker())
	if err != nil {
		log.Error("", err)
	}

	log.Debugf("notify service is started")
}

func (s *Service) AddSubscriber(n Subscriber) error {
	if n == nil {
		err := errors.New("required Subscriber")
		log.Errorf(err, "add subscriber failed")
		return err
	}

	if !n.Type().IsValid() {
		err := errors.New("unknown subscribe type")
		log.Errorf(err, "add %s subscriber[%s/%s] failed", n.Type(), n.Subject(), n.Group())
		return err
	}

	p := s.newProcessor(n.Type())
	n.SetService(s)
	n.OnAccept()

	p.AddSubscriber(n)
	return nil
}

func (s *Service) RemoveSubscriber(n Subscriber) {
	s.mux.RLock()
	p, ok := s.processors[n.Type()]
	if !ok {
		s.mux.RUnlock()
		return
	}
	s.mux.RUnlock()

	p.Remove(n)
	n.Close()
}

func (s *Service) stopProcessors() {
	s.mux.RLock()
	for _, p := range s.processors {
		p.Clear()
		p.Stop()
	}
	s.mux.RUnlock()
}

//通知内容塞到队列里
func (s *Service) Publish(evt Event) error {
	if s.Closed() {
		return errors.New("add notify event failed for server shutdown")
	}

	s.mux.RLock()
	p, ok := s.processors[evt.Type()]
	if !ok {
		s.mux.RUnlock()
		return errors.New("unknown event type")
	}
	s.mux.RUnlock()
	p.Accept(evt)
	return nil
}

func (s *Service) Closed() (b bool) {
	s.mux.RLock()
	b = s.isClose
	s.mux.RUnlock()
	return
}

func (s *Service) Stop() {
	if s.Closed() {
		return
	}
	s.mux.Lock()
	s.isClose = true
	s.mux.Unlock()

	s.stopProcessors()

	log.Debug("notify service stopped")
}

func NewNotifyService() *Service {
	return &Service{
		processors: make(map[Type]*Processor),
		isClose:    true,
	}
}

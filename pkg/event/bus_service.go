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

package event

import (
	"errors"
	"fmt"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

// BusService is the daemon service to manage multiple type Bus
// and wrap handle methods of Bus
type BusService struct {
	// buses is the map of event handler, key is event source type
	buses   map[Type]*Bus
	mux     sync.RWMutex
	isClose bool
}

func (s *BusService) newBus(t Type) *Bus {
	s.mux.RLock()
	p, ok := s.buses[t]
	if ok {
		s.mux.RUnlock()
		return p
	}
	s.mux.RUnlock()

	s.mux.Lock()
	p, ok = s.buses[t]
	if ok {
		s.mux.Unlock()
		return p
	}
	p = NewBus(t.String(), t.QueueSize())
	s.buses[t] = p
	s.mux.Unlock()

	p.Run()
	return p
}

func (s *BusService) Start() {
	if !s.Closed() {
		log.Warn("notify service is already running")
		return
	}
	s.mux.Lock()
	s.isClose = false
	s.mux.Unlock()

	// 错误subscriber清理
	err := s.AddSubscriber(NewSubscriberHealthChecker())
	if err != nil {
		log.Error("", err)
	}

	log.Debug("notify service is started")
}

func (s *BusService) AddSubscriber(n Subscriber) error {
	if n == nil {
		err := errors.New("required Subscriber")
		log.Error("add subscriber failed", err)
		return err
	}

	if !n.Type().IsValid() {
		err := errors.New("unknown subscribe type")
		log.Error(fmt.Sprintf("add %s subscriber[%s/%s] failed", n.Type(), n.Subject(), n.Group()), err)
		return err
	}

	p := s.newBus(n.Type())
	n.SetBus(s)
	n.OnAccept()

	p.AddSubscriber(n)
	return nil
}

func (s *BusService) RemoveSubscriber(n Subscriber) {
	s.mux.RLock()
	p, ok := s.buses[n.Type()]
	if !ok {
		s.mux.RUnlock()
		return
	}
	s.mux.RUnlock()

	p.RemoveSubscriber(n)
	n.Close()
}

func (s *BusService) closeBuses() {
	s.mux.RLock()
	for _, p := range s.buses {
		p.Clear()
		p.Stop()
	}
	s.mux.RUnlock()
}

// 通知内容塞到队列里
func (s *BusService) Fire(evt Event) error {
	if s.Closed() {
		return errors.New("add notify evt failed for server shutdown")
	}

	s.mux.RLock()
	bus, ok := s.buses[evt.Type()]
	if !ok {
		s.mux.RUnlock()
		return fmt.Errorf("no %s subscriber on this service center", evt.Type())
	}
	s.mux.RUnlock()
	bus.Fire(evt)
	return nil
}

func (s *BusService) Closed() (b bool) {
	s.mux.RLock()
	b = s.isClose
	s.mux.RUnlock()
	return
}

func (s *BusService) Stop() {
	if s.Closed() {
		return
	}
	s.mux.Lock()
	s.isClose = true
	s.mux.Unlock()

	s.closeBuses()

	log.Debug("notify service stopped")
}

func NewBusService() *BusService {
	return &BusService{
		buses:   make(map[Type]*Bus),
		isClose: true,
	}
}

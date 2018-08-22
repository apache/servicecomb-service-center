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
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
	"sync"
)

var notifyService *NotifyService

func init() {
	notifyService = &NotifyService{
		isClose:   true,
		goroutine: gopool.New(context.Background()),
	}
}

type NotifyService struct {
	processors *util.ConcurrentMap
	goroutine  *gopool.Pool
	err        chan error
	closeMux   sync.RWMutex
	isClose    bool
}

func (s *NotifyService) Err() <-chan error {
	return s.err
}

func (s *NotifyService) init() {
	s.processors = util.NewConcurrentMap(int(typeEnd))
	s.err = make(chan error, 1)
	for i := NotifyType(0); i != typeEnd; i++ {
		s.processors.Put(i, NewProcessor(i.String(), i.QueueSize()))
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

	s.init()
	// 错误subscriber清理
	s.AddSubscriber(NewNotifyServiceHealthChecker())

	log.Debugf("notify service is started")

	s.processors.ForEach(func(item util.MapItem) (next bool) {
		s.goroutine.Do(item.Value.(*Processor).Do)
		return true
	})
}

func (s *NotifyService) AddSubscriber(n Subscriber) error {
	if s.Closed() {
		return errors.New("server is shutting down")
	}

	itf, ok := s.processors.Get(n.Type())
	if !ok {
		return errors.New("Unknown subscribe type")
	}
	n.SetService(s)
	n.OnAccept()

	itf.(*Processor).AddSubscriber(n)
	return nil
}

func (s *NotifyService) RemoveSubscriber(n Subscriber) {
	itf, ok := s.processors.Get(n.Type())
	if !ok {
		return
	}

	itf.(*Processor).Remove(n)
	n.Close()
}

func (s *NotifyService) RemoveAllSubscribers() {
	s.processors.ForEach(func(item util.MapItem) (next bool) {
		item.Value.(*Processor).Clear()
		return true
	})
}

//通知内容塞到队列里
func (s *NotifyService) AddJob(job NotifyJob) error {
	if s.Closed() {
		return errors.New("add notify job failed for server shutdown")
	}

	itf, ok := s.processors.Get(job.Type())
	if !ok {
		return errors.New("Unknown job type")
	}
	itf.(*Processor).Accept(job)
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

	s.goroutine.Close(true)

	s.RemoveAllSubscribers()

	close(s.err)

	log.Debug("notify service stopped")
}

func GetNotifyService() *NotifyService {
	return notifyService
}

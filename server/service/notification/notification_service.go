//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package notification

import (
	"container/list"
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"sync"
	"time"
)

var notifyTypeNames = []string{
	NOTIFTY:  "NOTIFTY",
	INSTANCE: "INSTANCE",
}

var notifyService *NotifyService

func init() {
	notifyService = &NotifyService{
		isClose: true,
	}
}

type subscriberIndex map[string]*list.List

type subscriberSubjectIndex map[string]subscriberIndex

type serviceIndex map[NotifyType]subscriberSubjectIndex

type NotifyService struct {
	Config NotifyServiceConfig

	services serviceIndex
	queues   map[NotifyType]chan NotifyJob
	waits    sync.WaitGroup
	mutexes  map[NotifyType]*sync.Mutex
	err      chan error
	closeMux sync.RWMutex
	isClose  bool
}

func (s *NotifyService) Err() <-chan error {
	return s.err
}

func (s *NotifyService) AddSubscriber(n Subscriber) error {
	if s.Closed() {
		return errors.New("server is shutting down")
	}

	s.mutexes[n.Type()].Lock()
	ss, ok := s.services[n.Type()]
	if !ok {
		s.mutexes[n.Type()].Unlock()
		return errors.New("Unknown subscribe type")
	}

	sr, ok := ss[n.Subject()]
	if !ok {
		sr = make(subscriberIndex)
		ss[n.Subject()] = sr
	}

	ns, ok := sr[n.Id()]
	if !ok {
		ns = list.New()
	}
	ns.PushBack(n)
	sr[n.Id()] = ns

	n.SetService(s)
	s.mutexes[n.Type()].Unlock()

	n.OnAccept()
	return nil
}

func (s *NotifyService) RemoveSubscriber(n Subscriber) {
	s.mutexes[n.Type()].Lock()
	defer s.mutexes[n.Type()].Unlock()
	ss, ok := s.services[n.Type()]
	if !ok {
		return
	}

	m, ok := ss[n.Subject()]
	if !ok {
		return
	}

	ns, ok := m[n.Id()]
	if !ok {
		return
	}

	for sr := ns.Front(); sr != nil; sr = sr.Next() {
		if sr.Value == n {
			ns.Remove(sr)
			n.Close()
			break
		}
	}
}

func (s *NotifyService) RemoveAllSubscribers() {
	for t, ss := range s.services {
		s.mutexes[t].Lock()
		for _, subscribers := range ss {
			for _, ns := range subscribers {
				for e, n := ns.Front(), ns.Front(); e != nil; e = n {
					e.Value.(Subscriber).Close()
					n = e.Next()
					ns.Remove(e)
				}
			}
		}
		s.mutexes[t].Unlock()
	}
}

//通知内容塞到队列里
func (s *NotifyService) AddJob(job NotifyJob) error {
	if s.Closed() {
		return errors.New("add notify job failed for server shutdown")
	}

	defer util.RecoverAndReport()
	select {
	case s.queues[job.Type()] <- job:
		return nil
	case <-time.After(s.Config.AddTimeout):
		util.Logger().Errorf(nil, "Add job failed.%s")
		return errors.New("add notify job timeout")
	}
}

func (s *NotifyService) publish2Subscriber(t NotifyType) {
	defer s.waits.Done()
	for job := range s.queues[t] {
		util.Logger().Infof("notification service got a job %s: %s to notify subscriber %s",
			job.Type(), job.Subject(), job.SubscriberId())

		s.mutexes[t].Lock()

		if s.Closed() && len(s.services[t]) == 0 {
			s.mutexes[t].Unlock()
			return
		}

		m, ok := s.services[t][job.Subject()]
		if ok {
			// publish的subject如果带上id，则单播，否则广播
			if len(job.SubscriberId()) != 0 {
				ns, ok := m[job.SubscriberId()]
				if ok {
					for n := ns.Front(); n != nil; n = n.Next() {
						go n.Value.(Subscriber).OnMessage(job)
					}
				}
				s.mutexes[t].Unlock()
				continue
			}
			for key := range m {
				ns := m[key]
				for n := ns.Front(); n != nil; n = n.Next() {
					go n.Value.(Subscriber).OnMessage(job)
				}
			}
		}

		s.mutexes[t].Unlock()
	}
}

func (s *NotifyService) init() {
	if s.Config.AddTimeout <= 0 {
		s.Config.AddTimeout = DEFAULT_TIMEOUT
	}
	if s.Config.NotifyTimeout <= 0 {
		s.Config.NotifyTimeout = DEFAULT_TIMEOUT
	}
	if s.Config.MaxQueue <= 0 || s.Config.MaxQueue > DEFAULT_MAX_QUEUE {
		s.Config.MaxQueue = DEFAULT_MAX_QUEUE
	}

	s.services = make(serviceIndex)
	s.err = make(chan error, 1)
	s.queues = make(map[NotifyType]chan NotifyJob)
	s.mutexes = make(map[NotifyType]*sync.Mutex)
	for i := NotifyType(0); i != typeEnd; i++ {
		s.services[i] = make(subscriberSubjectIndex)
		s.queues[i] = make(chan NotifyJob, s.Config.MaxQueue)
		s.mutexes[i] = &sync.Mutex{}
		s.waits.Add(1)
	}
}

func (s *NotifyService) Start() {
	if !s.Closed() {
		util.Logger().Warnf(nil, "notify service is already running with config %s", s.Config)
		return
	}
	s.closeMux.Lock()
	s.isClose = false
	s.closeMux.Unlock()

	s.init()
	// 错误subscriber清理
	s.AddSubscriber(NewNotifyServiceHealthChecker())

	util.Logger().Infof("notify service is started with config %s", s.Config)

	for i := NotifyType(0); i != typeEnd; i++ {
		go s.publish2Subscriber(i)
	}
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

	for _, c := range s.queues {
		close(c)
	}
	s.waits.Wait()

	s.RemoveAllSubscribers()

	close(s.err)

	util.Logger().Info("notify service stopped.")
}

func GetNotifyService() *NotifyService {
	return notifyService
}

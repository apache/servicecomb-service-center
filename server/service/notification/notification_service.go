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
	"encoding/json"
	"errors"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/service/dependency"
	"github.com/ServiceComb/service-center/server/service/microservice"
	"github.com/ServiceComb/service-center/util"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_QUEUE           = 100
	DEFAULT_TIMEOUT             = 30 * time.Second
	DEFAULT_LISTWATCH_TIMEOUT   = 30 * time.Second
	NOTIFY_SERVER_CHECKER_NAME  = "__HealthChecker__"
	NOTIFY_SERVER_CHECK_SUBJECT = "__NotifyServerHealthCheck__"
)

type NotifyServerConfig struct {
	AddTimeout    time.Duration
	NotifyTimeout time.Duration
	MaxQueue      int64
}

type NotifyService struct {
	Config *NotifyServerConfig

	cacherMap map[string]registry.Cacher
	notifiers map[string]map[string]*list.List
	queue     chan NotifyJob
	err       chan error
	mux       sync.Mutex
	isClose   bool
}

func (s *NotifyService) Err() <-chan error {
	return s.err
}

func (s *NotifyService) ListWatcher(key string) registry.Cacher {
	return s.cacherMap[key]
}

func (s *NotifyService) AddNotifier(n Worker) error {
	if s.isClose {
		return errors.New("server is shutting down")
	}

	s.mux.Lock()
	_, ok := s.notifiers[n.Subject()]
	if !ok {
		s.notifiers[n.Subject()] = map[string]*list.List{}
	}

	m := s.notifiers[n.Subject()]
	ns, ok := m[n.Id()]
	if !ok {
		ns = list.New()
	}
	ns.PushBack(n)
	m[n.Id()] = ns

	n.SetService(s)
	s.mux.Unlock()

	n.OnAccept()
	return nil
}

func (s *NotifyService) RemoveNotifier(r Worker) {
	s.mux.Lock()
	m, ok := s.notifiers[r.Subject()]
	if ok {
		ns, ok := m[r.Id()]
		if ok {
			for n := ns.Front(); n != nil; n = n.Next() {
				if n.Value == r {
					ns.Remove(n)
					r.Close()
					break
				}
			}
		}
	}
	s.mux.Unlock()
}

//通知内容塞到队列里
func (s *NotifyService) AddJob(job NotifyJob) error {
	if s.isClose {
		return errors.New("add notify job failed for server shutdown")
	}
	select {
	case s.queue <- job:
		return nil
	case <-time.After(s.Config.AddTimeout):
		util.LOGGER.Errorf(nil, "Add job failed.%s")
		return errors.New("add notify job timeout")
	}
}

func (s *NotifyService) publish2Subscriber() {
	for job := range s.queue {
		util.LOGGER.Infof("notification server got a job %s %s", job.Subject(), job.Id())

		s.mux.Lock()

		m, ok := s.notifiers[job.Subject()]
		if ok {
			// publish的subject如果带上id，则单播，否则广播
			if len(job.Id()) != 0 {
				ns, ok := m[job.Id()]
				if ok {
					for n := ns.Front(); n != nil; n = n.Next() {
						go n.Value.(Worker).OnMessage(job)
					}
				}
				s.mux.Unlock()
				continue
			}
			for key := range m {
				ns := m[key]
				for n := ns.Front(); n != nil; n = n.Next() {
					go n.Value.(Worker).OnMessage(job)
				}
			}
		}

		s.mux.Unlock()
	}
}

func (s *NotifyService) StartNotifyService() {
	util.LOGGER.Info("starting notify service")

	if s.Config.AddTimeout <= 0 {
		s.Config.AddTimeout = DEFAULT_TIMEOUT
	}
	if s.Config.NotifyTimeout <= 0 {
		s.Config.NotifyTimeout = DEFAULT_TIMEOUT
	}
	if s.Config.MaxQueue <= 0 || s.Config.MaxQueue > DEFAULT_MAX_QUEUE {
		s.Config.MaxQueue = DEFAULT_MAX_QUEUE
	}

	s.cacherMap = make(map[string]registry.Cacher)
	s.notifiers = map[string]map[string]*list.List{}
	s.err = make(chan error, 1)
	s.queue = make(chan NotifyJob, s.Config.MaxQueue)
	s.isClose = false

	go s.publish2Subscriber()

	// 错误notifier清理
	s.AddNotifier(NewNotifyServiceHealthChecker())
	s.WatchTenants()
	util.LOGGER.Info("notify service is ready")
}

func (s *NotifyService) WatchTenants() {
	key := apt.GenerateTenantKey("")

	c := registry.NewKvCacher(&registry.KvCacherConfig{
		Key:     key[:len(key)-1],
		Timeout: DEFAULT_LISTWATCH_TIMEOUT,
		Period:  time.Second,
		OnEvent: func(evt *registry.KvEvent) error {
			kv := evt.KV
			action := evt.Action
			tenant := pb.GetInfoFromTenantChangeEvent(kv)
			if len(tenant) == 0 {
				util.LOGGER.Errorf(nil,
					"unmarshal tenant info failed, key %s [%s] event", string(kv.Key), action)
				return nil
			}
			if action != pb.EVT_CREATE {
				return nil
			}

			util.LOGGER.Warnf(nil, "new tenant %s instances watcher is created", tenant)
			s.WatchInstance(apt.GetInstanceRootKey(tenant))
			return nil
		},
	})
	s.cacherMap[key] = c
	c.Run()
}

//SC 负责监控所有实例变化
func (s *NotifyService) WatchInstance(instanceWatchByTenantKey string) {
	c := registry.NewKvCacher(&registry.KvCacherConfig{
		Key:     instanceWatchByTenantKey,
		Timeout: DEFAULT_LISTWATCH_TIMEOUT,
		Period:  time.Second,
		OnEvent: func(evt *registry.KvEvent) error {
			kv := evt.KV
			action := evt.Action
			providerId, providerInstanceId, tenantProject, data := pb.GetInfoFromInstChangedEvent(kv)
			if data == nil {
				util.LOGGER.Errorf(nil,
					"unmarshal provider service instance file failed, instance %s/%s [%s] event, data is nil",
					providerId, providerInstanceId, action)
				return nil
			}
			util.LOGGER.Warnf(nil, "notification service catch instance %s/%s [%s] event",
				providerId, providerInstanceId, action)

			var instance pb.MicroServiceInstance
			err := json.Unmarshal(data, &instance)
			if err != nil {
				util.LOGGER.Errorf(err, "unmarshal provider service instance %s/%s file failed",
					providerId, providerInstanceId)
				return nil
			}
			// 查询服务版本信息
			ms, err := microservice.GetByIdInCache(tenantProject, providerId)
			if ms == nil {
				util.LOGGER.Errorf(err, "get provider service %s/%s id in cache failed",
					providerId, providerInstanceId)
				return nil
			}

			// 查询所有consumer
			Kvs, err := dependency.GetConsumersInCache(tenantProject, providerId)
			if err != nil {
				util.LOGGER.Errorf(err, "query service %s consumers failed", providerId)
				return nil
			}

			response := &pb.WatchInstanceResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "watch instance successfully"),
				Action:   string(action),
				Key: &pb.MicroServiceKey{
					AppId:       ms.AppId,
					ServiceName: ms.ServiceName,
					Version:     ms.Version,
				},
				Instance: &instance,
			}
			for _, dependence := range Kvs {
				consumer := string(dependence.Key)
				consumer = consumer[strings.LastIndex(consumer, "/")+1:]
				job := NewWatchJob(consumer, apt.GetInstanceRootKey(tenantProject)+"/",
					evt.Revision, response)
				util.LOGGER.Debugf("publish event to notify server, %v", job)

				// TODO add超时怎么处理？
				s.AddJob(job)
			}
			return nil
		},
	})
	s.cacherMap[instanceWatchByTenantKey+"/"] = c
	c.Run()
}

func (s *NotifyService) Close() {
	s.isClose = true

	close(s.queue)

	s.mux.Lock()
	for subject := range s.notifiers {
		for key := range s.notifiers[subject] {
			ns := s.notifiers[subject][key]
			for n := ns.Front(); n != nil; n = n.Next() {
				n.Value.(Worker).Close()
			}
		}
	}
	s.mux.Unlock()

	close(s.err)
}

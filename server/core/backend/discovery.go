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
package backend

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"time"
)

var store = &KvStore{}

func init() {
	store.Initialize()
	registerInnerTypes()
}

type KvStore struct {
	AddOns      map[discovery.Type]AddOn
	adaptors    util.ConcurrentMap
	taskService task.TaskService
	ready       chan struct{}
	goroutine   *gopool.Pool
	isClose     bool
	rev         int64
}

func (s *KvStore) Initialize() {
	s.AddOns = make(map[discovery.Type]AddOn)
	s.taskService = task.NewTaskService()
	s.ready = make(chan struct{})
	s.goroutine = gopool.New(context.Background())
}

func (s *KvStore) OnCacheEvent(evt discovery.KvEvent) {
	if s.rev < evt.Revision {
		s.rev = evt.Revision
	}
}

func (s *KvStore) InjectConfig(cfg *discovery.Config) *discovery.Config {
	return cfg.AppendEventFunc(s.OnCacheEvent)
}

func (s *KvStore) repo() discovery.AdaptorRepository {
	return plugin.Plugins().Discovery()
}

func (s *KvStore) getOrCreateAdaptor(t discovery.Type) discovery.Adaptor {
	v, _ := s.adaptors.Fetch(t, func() (interface{}, error) {
		addOn, ok := s.AddOns[t]
		if ok {
			adaptor := s.repo().New(t, addOn.(AddOn).Config())
			adaptor.Run()
			return adaptor, nil
		}
		log.Warnf("type '%s' not found", t)
		return nil, nil
	})
	return v.(discovery.Adaptor)
}

func (s *KvStore) Run() {
	s.goroutine.Do(s.store)
	s.goroutine.Do(s.autoClearCache)
	s.taskService.Run()
}

func (s *KvStore) store(ctx context.Context) {
	// new all types
	for _, t := range discovery.Types {
		select {
		case <-ctx.Done():
			return
		case <-s.getOrCreateAdaptor(t).Ready():
		}
	}

	util.SafeCloseChan(s.ready)

	log.Debugf("all adaptors are ready")
}

func (s *KvStore) autoClearCache(ctx context.Context) {
	if core.ServerInfo.Config.CacheTTL == 0 {
		return
	}

	log.Infof("start auto clear cache in %v", core.ServerInfo.Config.CacheTTL)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(core.ServerInfo.Config.CacheTTL):
			for _, t := range discovery.Types {
				cache, ok := s.getOrCreateAdaptor(t).Cache().(discovery.Cache)
				if !ok {
					log.Error("the discovery adaptor does not implement the Cache", nil)
					continue
				}
				cache.MarkDirty()
			}
			log.Warnf("caches are marked dirty!")
		}
	}
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	s.adaptors.ForEach(func(item util.MapItem) bool {
		item.Value.(discovery.Adaptor).Stop()
		return true
	})

	s.taskService.Stop()

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	log.Debugf("store daemon stopped")
}

func (s *KvStore) Ready() <-chan struct{} {
	<-s.taskService.Ready()
	return s.ready
}

func (s *KvStore) Install(addOn AddOn) (id discovery.Type, err error) {
	if addOn == nil || len(addOn.Name()) == 0 || addOn.Config() == nil {
		return discovery.TypeError, errors.New("invalid parameter")
	}

	id, err = discovery.RegisterType(addOn.Name())
	if err != nil {
		return
	}

	discovery.EventProxy(id).InjectConfig(addOn.Config())

	s.InjectConfig(addOn.Config())

	s.AddOns[id] = addOn

	log.Infof("install new type %d:%s->%s", id, addOn.Name(), addOn.Config().Key)
	return
}

func (s *KvStore) MustInstall(addOn AddOn) discovery.Type {
	id, err := s.Install(addOn)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *KvStore) Adaptors(id discovery.Type) discovery.Adaptor { return s.getOrCreateAdaptor(id) }
func (s *KvStore) Service() discovery.Adaptor                   { return s.Adaptors(SERVICE) }
func (s *KvStore) SchemaSummary() discovery.Adaptor             { return s.Adaptors(SCHEMA_SUMMARY) }
func (s *KvStore) Instance() discovery.Adaptor                  { return s.Adaptors(INSTANCE) }
func (s *KvStore) Lease() discovery.Adaptor                     { return s.Adaptors(LEASE) }
func (s *KvStore) ServiceIndex() discovery.Adaptor              { return s.Adaptors(SERVICE_INDEX) }
func (s *KvStore) ServiceAlias() discovery.Adaptor              { return s.Adaptors(SERVICE_ALIAS) }
func (s *KvStore) ServiceTag() discovery.Adaptor                { return s.Adaptors(SERVICE_TAG) }
func (s *KvStore) Rule() discovery.Adaptor                      { return s.Adaptors(RULE) }
func (s *KvStore) RuleIndex() discovery.Adaptor                 { return s.Adaptors(RULE_INDEX) }
func (s *KvStore) Schema() discovery.Adaptor                    { return s.Adaptors(SCHEMA) }
func (s *KvStore) DependencyRule() discovery.Adaptor            { return s.Adaptors(DEPENDENCY_RULE) }
func (s *KvStore) DependencyQueue() discovery.Adaptor           { return s.Adaptors(DEPENDENCY_QUEUE) }
func (s *KvStore) Domain() discovery.Adaptor                    { return s.Adaptors(DOMAIN) }
func (s *KvStore) Project() discovery.Adaptor                   { return s.Adaptors(PROJECT) }

// KeepAlive will always return ok when registry is unavailable
// unless the registry response is LeaseNotFound
func (s *KvStore) KeepAlive(ctx context.Context, opts ...registry.PluginOpOption) (int64, error) {
	op := registry.OpPut(opts...)

	t := NewLeaseAsyncTask(op)
	if op.Mode == registry.MODE_NO_CACHE {
		log.Debugf("keep alive lease WitchNoCache, request etcd server, op: %s", op)
		err := t.Do(ctx)
		ttl := t.TTL
		return ttl, err
	}

	err := s.taskService.Add(ctx, t)
	if err != nil {
		return 0, err
	}
	itf, err := s.taskService.LatestHandled(t.Key())
	if err != nil {
		return 0, err
	}
	pt := itf.(*LeaseTask)
	return pt.TTL, pt.Err()
}

func Store() *KvStore {
	return store
}

func Revision() int64 {
	return store.rev
}

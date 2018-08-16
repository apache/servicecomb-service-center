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
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/task"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"golang.org/x/net/context"
	"sync"
)

var store = &KvStore{}

func init() {
	store.Initialize()
	registerInnerTypes()
}

type KvStore struct {
	addOns      *util.ConcurrentMap
	entities    *util.ConcurrentMap
	taskService task.TaskService
	lock        sync.RWMutex
	ready       chan struct{}
	goroutine   *gopool.Pool
	isClose     bool
	rev         int64
}

func (s *KvStore) Initialize() {
	s.addOns = util.NewConcurrentMap(0)
	s.entities = util.NewConcurrentMap(0)
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

func (s *KvStore) repo() discovery.EntityRepository {
	return plugin.Plugins().Discovery()
}

func (s *KvStore) getOrCreateEntity(t discovery.StoreType) discovery.Entity {
	v, _ := s.entities.Fetch(t, func() (interface{}, error) {
		addOn, ok := s.addOns.Get(t)
		if ok {
			return s.repo().NewEntity(t, addOn.(discovery.AddOn).Config()), nil
		}
		log.Warnf("type '%s' not found", t)
		return s.repo().NewEntity(t, nil), nil
	})
	return v.(discovery.Entity)
}

func (s *KvStore) Run() {
	s.goroutine.Do(s.store)
	s.taskService.Run()
}

func (s *KvStore) store(ctx context.Context) {
	// new all types
	for _, t := range discovery.StoreTypes() {
		select {
		case <-ctx.Done():
			return
		case <-s.getOrCreateEntity(t).Ready():
		}
	}

	util.SafeCloseChan(s.ready)

	log.Debugf("all entities are ready")
}

func (s *KvStore) closed() bool {
	return s.isClose
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	s.entities.ForEach(func(item util.MapItem) bool {
		item.Value.(discovery.Entity).Stop()
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

func (s *KvStore) Install(addOn discovery.AddOn) (id discovery.StoreType, err error) {
	if id, err = discovery.Install(addOn); err != nil {
		return
	}
	s.InjectConfig(addOn.Config())

	s.addOns.Put(id, addOn)

	log.Infof("install new type %d:%s->%s", id, addOn.Name(), addOn.Config().Key)
	return
}

func (s *KvStore) MustInstall(addOn discovery.AddOn) discovery.StoreType {
	id, err := s.Install(addOn)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *KvStore) Entities(id discovery.StoreType) discovery.Entity { return s.getOrCreateEntity(id) }
func (s *KvStore) Service() discovery.Entity                        { return s.Entities(SERVICE) }
func (s *KvStore) SchemaSummary() discovery.Entity                  { return s.Entities(SCHEMA_SUMMARY) }
func (s *KvStore) Instance() discovery.Entity                       { return s.Entities(INSTANCE) }
func (s *KvStore) Lease() discovery.Entity                          { return s.Entities(LEASE) }
func (s *KvStore) ServiceIndex() discovery.Entity                   { return s.Entities(SERVICE_INDEX) }
func (s *KvStore) ServiceAlias() discovery.Entity                   { return s.Entities(SERVICE_ALIAS) }
func (s *KvStore) ServiceTag() discovery.Entity                     { return s.Entities(SERVICE_TAG) }
func (s *KvStore) Rule() discovery.Entity                           { return s.Entities(RULE) }
func (s *KvStore) RuleIndex() discovery.Entity                      { return s.Entities(RULE_INDEX) }
func (s *KvStore) Schema() discovery.Entity                         { return s.Entities(SCHEMA) }
func (s *KvStore) Dependency() discovery.Entity                     { return s.Entities(DEPENDENCY) }
func (s *KvStore) DependencyRule() discovery.Entity                 { return s.Entities(DEPENDENCY_RULE) }
func (s *KvStore) DependencyQueue() discovery.Entity                { return s.Entities(DEPENDENCY_QUEUE) }
func (s *KvStore) Domain() discovery.Entity                         { return s.Entities(DOMAIN) }
func (s *KvStore) Project() discovery.Entity                        { return s.Entities(PROJECT) }

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

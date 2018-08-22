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
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/task"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
)

var store = &KvStore{}

func init() {
	store.Initialize()
}

type KvStore struct {
	entities    *util.ConcurrentMap
	taskService task.TaskService
	lock        sync.RWMutex
	ready       chan struct{}
	goroutine   *gopool.Pool
	isClose     bool
	rev         int64
}

func (s *KvStore) Initialize() {
	s.entities = util.NewConcurrentMap(0)
	s.taskService = task.NewTaskService()
	s.ready = make(chan struct{})
	s.goroutine = gopool.New(context.Background())
}

func (s *KvStore) OnCacheEvent(evt KvEvent) {
	if s.rev < evt.Revision {
		s.rev = evt.Revision
	}
}

func (s *KvStore) injectConfig(t StoreType) *Config {
	return TypeConfig[t].AppendEventFunc(s.OnCacheEvent)
}

func (s *KvStore) getOrCreateEntity(t StoreType) Entity {
	v, err := s.entities.Fetch(t, func() (interface{}, error) {
		cfg, ok := TypeConfig[t]
		if !ok {
			// do not new instance
			return nil, ErrNoImpl
		}

		e := NewKvEntity(t.String(), cfg)
		e.Run()
		return e, nil
	})
	if err != nil {
		log.Errorf(err, "can not find entity '%s', new default one", t.String())
		return DefaultKvEntity()

	}
	return v.(Entity)
}

func (s *KvStore) Run() {
	s.goroutine.Do(s.store)
	s.taskService.Run()
}

func (s *KvStore) store(ctx context.Context) {
	for i := range TypeConfig {
		select {
		case <-ctx.Done():
			return
		case <-s.Entities(i).Ready():
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
		item.Value.(Entity).Stop()
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

func (s *KvStore) installType(e Extension) (id StoreType, err error) {
	if e == nil {
		return TypeError, errors.New("invalid parameter")
	}
	for _, n := range TypeNames {
		if n == e.Name() {
			return TypeError, fmt.Errorf("redeclare store type '%s'", n)
		}
	}
	for _, r := range TypeConfig {
		if r.Key == e.Config().Key {
			return TypeError, fmt.Errorf("redeclare store root '%s'", r)
		}
	}

	id = StoreType(len(TypeNames))
	TypeNames = append(TypeNames, e.Name())
	// config
	TypeConfig[id] = e.Config()
	// event proxy
	NewEventProxy(id)

	s.injectConfig(id)
	return
}

func (s *KvStore) Install(e Extension) (id StoreType, err error) {
	if id, err = s.installType(e); err != nil {
		return
	}

	log.Infof("install new store entity %d:%s->%s", id, e.Name(), e.Config().Key)
	return
}

func (s *KvStore) MustInstall(e Extension) StoreType {
	id, err := s.Install(e)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *KvStore) Entities(id StoreType) Entity { return s.getOrCreateEntity(id) }
func (s *KvStore) Service() Entity              { return s.Entities(SERVICE) }
func (s *KvStore) SchemaSummary() Entity        { return s.Entities(SCHEMA_SUMMARY) }
func (s *KvStore) Instance() Entity             { return s.Entities(INSTANCE) }
func (s *KvStore) Lease() Entity                { return s.Entities(LEASE) }
func (s *KvStore) ServiceIndex() Entity         { return s.Entities(SERVICE_INDEX) }
func (s *KvStore) ServiceAlias() Entity         { return s.Entities(SERVICE_ALIAS) }
func (s *KvStore) ServiceTag() Entity           { return s.Entities(SERVICE_TAG) }
func (s *KvStore) Rule() Entity                 { return s.Entities(RULE) }
func (s *KvStore) RuleIndex() Entity            { return s.Entities(RULE_INDEX) }
func (s *KvStore) Schema() Entity               { return s.Entities(SCHEMA) }
func (s *KvStore) Dependency() Entity           { return s.Entities(DEPENDENCY) }
func (s *KvStore) DependencyRule() Entity       { return s.Entities(DEPENDENCY_RULE) }
func (s *KvStore) DependencyQueue() Entity      { return s.Entities(DEPENDENCY_QUEUE) }
func (s *KvStore) Domain() Entity               { return s.Entities(DOMAIN) }
func (s *KvStore) Project() Entity              { return s.Entities(PROJECT) }

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

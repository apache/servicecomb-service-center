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
	"github.com/apache/incubator-servicecomb-service-center/pkg/async"
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
	indexers    *util.ConcurrentMap
	taskService *async.TaskService
	lock        sync.RWMutex
	ready       chan struct{}
	goroutine   *util.GoRoutine
	isClose     bool
	rev         int64
}

func (s *KvStore) Initialize() {
	s.indexers = util.NewConcurrentMap(0)
	s.taskService = async.NewTaskService()
	s.ready = make(chan struct{})
	s.goroutine = util.NewGo(context.Background())
}

func (s *KvStore) OnCacheEvent(evt KvEvent) {
	if s.rev < evt.Revision {
		s.rev = evt.Revision
	}
}

func (s *KvStore) injectConfig(t StoreType) *Config {
	return TypeConfig[t].AppendEventFunc(s.OnCacheEvent).
		AppendEventFunc(EventProxies[t].OnEvent)
}

func (s *KvStore) getOrCreateIndexer(t StoreType) Indexer {
	v, err := s.indexers.Fetch(t, func() (interface{}, error) {
		if _, ok := TypeConfig[t]; !ok {
			return nil, ErrNoImpl
		}

		i := NewIndexer(t.String(), s.injectConfig(t))
		i.Run()
		return i, nil
	})
	if err != nil {
		util.Logger().Errorf(err, "can not find entity '%s', new base indexer for it", t.String())
		return NullIndexer
	}
	return v.(Indexer)
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
		case <-s.Entity(i).Ready():
		}
	}

	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("all indexers are ready")
}

func (s *KvStore) closed() bool {
	return s.isClose
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	s.indexers.ForEach(func(item util.MapItem) bool {
		item.Value.(Indexer).Stop()
		return true
	})

	s.taskService.Stop()

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("store daemon stopped")
}

func (s *KvStore) Ready() <-chan struct{} {
	<-s.taskService.Ready()
	return s.ready
}

func (s *KvStore) installType(e Entity) (id StoreType, err error) {
	if e == nil {
		return NOT_EXIST, errors.New("invalid parameter")
	}
	for _, n := range TypeNames {
		if n == e.Name() {
			return NOT_EXIST, fmt.Errorf("redeclare store type '%s'", n)
		}
	}
	for _, r := range TypeConfig {
		if r.Prefix == e.Config().Prefix {
			return NOT_EXIST, fmt.Errorf("redeclare store root '%s'", r)
		}
	}

	id = StoreType(len(TypeNames))
	TypeNames = append(TypeNames, e.Name())
	TypeConfig[id] = e.Config()
	EventProxies[id] = NewEventProxy()
	return
}

func (s *KvStore) Install(e Entity) (id StoreType, err error) {
	if id, err = s.installType(e); err != nil {
		return
	}

	util.Logger().Infof("install new store entity %d:%s->%s", id, e.Name(), e.Config().Prefix)
	return
}

func (s *KvStore) MustInstall(e Entity) StoreType {
	id, err := s.Install(e)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *KvStore) Entity(id StoreType) Indexer { return s.getOrCreateIndexer(id) }
func (s *KvStore) Service() Indexer            { return s.Entity(SERVICE) }
func (s *KvStore) SchemaSummary() Indexer      { return s.Entity(SCHEMA_SUMMARY) }
func (s *KvStore) Instance() Indexer           { return s.Entity(INSTANCE) }
func (s *KvStore) Lease() Indexer              { return s.Entity(LEASE) }
func (s *KvStore) ServiceIndex() Indexer       { return s.Entity(SERVICE_INDEX) }
func (s *KvStore) ServiceAlias() Indexer       { return s.Entity(SERVICE_ALIAS) }
func (s *KvStore) ServiceTag() Indexer         { return s.Entity(SERVICE_TAG) }
func (s *KvStore) Rule() Indexer               { return s.Entity(RULE) }
func (s *KvStore) RuleIndex() Indexer          { return s.Entity(RULE_INDEX) }
func (s *KvStore) Schema() Indexer             { return s.Entity(SCHEMA) }
func (s *KvStore) Dependency() Indexer         { return s.Entity(DEPENDENCY) }
func (s *KvStore) DependencyRule() Indexer     { return s.Entity(DEPENDENCY_RULE) }
func (s *KvStore) DependencyQueue() Indexer    { return s.Entity(DEPENDENCY_QUEUE) }
func (s *KvStore) Domain() Indexer             { return s.Entity(DOMAIN) }
func (s *KvStore) Project() Indexer            { return s.Entity(PROJECT) }

func (s *KvStore) KeepAlive(ctx context.Context, opts ...registry.PluginOpOption) (int64, error) {
	op := registry.OpPut(opts...)

	t := NewLeaseAsyncTask(op)
	if op.Mode == registry.MODE_NO_CACHE {
		util.Logger().Debugf("keep alive lease WitchNoCache, request etcd server, op: %s", op)
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

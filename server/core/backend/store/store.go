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
package store

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/async"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
)

var store = &KvStore{}

func init() {
	store.Initialize()

	AddEventHandleFunc(LEASE, store.onLeaseEvent)
}

type KvStore struct {
	indexers    map[StoreType]*Indexer
	taskService *async.TaskService
	lock        sync.RWMutex
	ready       chan struct{}
	goroutine   *util.GoRoutine
	isClose     bool
}

func (s *KvStore) Initialize() {
	s.indexers = make(map[StoreType]*Indexer)
	s.taskService = async.NewTaskService()
	s.ready = make(chan struct{})
	s.goroutine = util.NewGo(context.Background())

	for i := StoreType(0); i != typeEnd; i++ {
		store.newNullStore(i)
	}
}

func (s *KvStore) dispatchEvent(t StoreType, evt KvEvent) {
	s.indexers[t].OnCacheEvent(evt)
	EventProxy(t).OnEvent(evt)
}

func (s *KvStore) newStore(t StoreType, opts ...KvCacherCfgOption) {
	s.newIndexBuilder(t, NewKvCacher(t.String(), s.getKvCacherCfgOptions(t)...))
}

func (s *KvStore) newNullStore(t StoreType) {
	s.newIndexBuilder(t, NullCacher)
}

func (s *KvStore) newIndexBuilder(t StoreType, cacher Cacher) {
	indexer := NewCacheIndexer(cacher)
	s.indexers[t] = indexer
	indexer.Run()
}

func (s *KvStore) Run() {
	s.goroutine.Do(s.store)
	s.taskService.Run()
}

func (s *KvStore) getKvCacherCfgOptions(t StoreType) (opts []KvCacherCfgOption) {
	switch t {
	case INSTANCE:
		opts = append(opts, WithDeferHandler(s.SelfPreservationHandler()))
	}
	sz := TypeInitSize[t]
	if sz > 0 {
		opts = append(opts, WithInitSize(sz))
	}
	opts = append(opts,
		WithKey(TypeRoots[t]),
		WithEventFunc(func(evt KvEvent) { s.dispatchEvent(t, evt) }))
	return
}

func (s *KvStore) SelfPreservationHandler() DeferHandler {
	return &InstanceEventDeferHandler{Percent: DEFAULT_SELF_PRESERVATION_PERCENT}
}

func (s *KvStore) store(ctx context.Context) {
	for t := StoreType(0); t != typeEnd; t++ {
		switch t {
		case SCHEMA:
			continue
		default:
			s.newStore(t)
		}
	}

	s.wait(ctx)
}

func (s *KvStore) wait(ctx context.Context) {
	for _, i := range s.indexers {
		select {
		case <-ctx.Done():
			return
		case <-i.Ready():
		}
	}
	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("all indexers are ready")
}

func (s *KvStore) onLeaseEvent(evt KvEvent) {
	if evt.Type != pb.EVT_DELETE {
		return
	}

	key := util.BytesToStringWithNoCopy(evt.Object.(*mvccpb.KeyValue).Key)
	s.taskService.DeferRemove(ToLeaseAsyncTaskKey(key))
}
func (s *KvStore) closed() bool {
	return s.isClose
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	for _, i := range s.indexers {
		i.Stop()
	}

	s.taskService.Stop()

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("store daemon stopped")
}

func (s *KvStore) Ready() <-chan struct{} {
	<-s.taskService.Ready()
	return s.ready
}

func (s *KvStore) Service() *Indexer {
	return s.indexers[SERVICE]
}

func (s *KvStore) SchemaSummary() *Indexer {
	return s.indexers[SCHEMA_SUMMARY]
}

func (s *KvStore) Instance() *Indexer {
	return s.indexers[INSTANCE]
}

func (s *KvStore) Lease() *Indexer {
	return s.indexers[LEASE]
}

func (s *KvStore) ServiceIndex() *Indexer {
	return s.indexers[SERVICE_INDEX]
}

func (s *KvStore) ServiceAlias() *Indexer {
	return s.indexers[SERVICE_ALIAS]
}

func (s *KvStore) ServiceTag() *Indexer {
	return s.indexers[SERVICE_TAG]
}

func (s *KvStore) Rule() *Indexer {
	return s.indexers[RULE]
}

func (s *KvStore) RuleIndex() *Indexer {
	return s.indexers[RULE_INDEX]
}

func (s *KvStore) Schema() *Indexer {
	return s.indexers[SCHEMA]
}

func (s *KvStore) Dependency() *Indexer {
	return s.indexers[DEPENDENCY]
}

func (s *KvStore) DependencyRule() *Indexer {
	return s.indexers[DEPENDENCY_RULE]
}

func (s *KvStore) DependencyQueue() *Indexer {
	return s.indexers[DEPENDENCY_QUEUE]
}

func (s *KvStore) Domain() *Indexer {
	return s.indexers[DOMAIN]
}

func (s *KvStore) Project() *Indexer {
	return s.indexers[PROJECT]
}

func (s *KvStore) Endpoints() *Indexer {
	return s.indexers[ENDPOINTS]
}

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

func Revision() (rev int64) {
	for _, i := range Store().indexers {
		if rev < i.Cache().Version() {
			rev = i.Cache().Version()
		}
	}
	return
}

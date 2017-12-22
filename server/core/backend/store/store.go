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
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"strconv"
	"sync"
)

const (
	DOMAIN StoreType = iota
	PROJECT
	SERVICE
	SERVICE_INDEX
	SERVICE_ALIAS
	SERVICE_TAG
	RULE
	RULE_INDEX
	DEPENDENCY
	DEPENDENCY_RULE
	DEPENDENCY_QUEUE
	SCHEMA // big data should not be stored in memory.
	SCHEMA_SUMMARY
	INSTANCE
	LEASE
	ENDPOINTS
	typeEnd
)

const TIME_FORMAT = "15:04:05.000"

var TypeNames = []string{
	SERVICE:          "SERVICE",
	INSTANCE:         "INSTANCE",
	DOMAIN:           "DOMAIN",
	SCHEMA:           "SCHEMA",
	SCHEMA_SUMMARY:   "SCHEMA_SUMMARY",
	RULE:             "RULE",
	LEASE:            "LEASE",
	SERVICE_INDEX:    "SERVICE_INDEX",
	SERVICE_ALIAS:    "SERVICE_ALIAS",
	SERVICE_TAG:      "SERVICE_TAG",
	RULE_INDEX:       "RULE_INDEX",
	DEPENDENCY:       "DEPENDENCY",
	DEPENDENCY_RULE:  "DEPENDENCY_RULE",
	DEPENDENCY_QUEUE: "DEPENDENCY_QUEUE",
	PROJECT:          "PROJECT",
	ENDPOINTS:        "ENDPOINTS",
}

var TypeRoots = map[StoreType]string{
	SERVICE:          apt.GetServiceRootKey(""),
	INSTANCE:         apt.GetInstanceRootKey(""),
	DOMAIN:           apt.GetDomainRootKey() + "/",
	SCHEMA:           apt.GetServiceSchemaRootKey(""),
	SCHEMA_SUMMARY:   apt.GetServiceSchemaSummaryRootKey(""),
	RULE:             apt.GetServiceRuleRootKey(""),
	LEASE:            apt.GetInstanceLeaseRootKey(""),
	SERVICE_INDEX:    apt.GetServiceIndexRootKey(""),
	SERVICE_ALIAS:    apt.GetServiceAliasRootKey(""),
	SERVICE_TAG:      apt.GetServiceTagRootKey(""),
	RULE_INDEX:       apt.GetServiceRuleIndexRootKey(""),
	DEPENDENCY:       apt.GetServiceDependencyRootKey(""),
	DEPENDENCY_RULE:  apt.GetServiceDependencyRuleRootKey(""),
	DEPENDENCY_QUEUE: apt.GetServiceDependencyQueueRootKey(""),
	PROJECT:          apt.GetProjectRootKey(""),
	ENDPOINTS:        apt.GetEndpointsRootKey(""),
}

var store = &KvStore{}

func init() {
	store.Initialize()

	AddEventHandleFunc(LEASE, store.onLeaseEvent)
}

type StoreType int

func (st StoreType) String() string {
	if int(st) < len(TypeNames) {
		return TypeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

type KvStore struct {
	indexers     map[StoreType]*Indexer
	asyncTaskSvc *async.AsyncTaskService
	lock         sync.RWMutex
	ready        chan struct{}
	isClose      bool
}

func (s *KvStore) Initialize() {
	s.indexers = make(map[StoreType]*Indexer)
	s.asyncTaskSvc = async.NewAsyncTaskService()
	s.ready = make(chan struct{})

	for i := StoreType(0); i != typeEnd; i++ {
		store.newNullStore(i)
	}
}

func (s *KvStore) dispatchEvent(t StoreType, evt *KvEvent) {
	s.indexers[t].OnCacheEvent(evt)
	select {
	case <-s.Ready():
	default:
		if evt.Action == pb.EVT_CREATE {
			evt.Action = pb.EVT_INIT
		}
	}
	EventProxy(t).OnEvent(evt)
}

func (s *KvStore) newStore(t StoreType, opts ...KvCacherCfgOption) {
	opts = append(opts,
		WithKey(TypeRoots[t]),
		WithInitSize(s.StoreSize(t)),
		WithEventFunc(func(evt *KvEvent) { s.dispatchEvent(t, evt) }),
	)
	s.newIndexer(t, NewKvCacher(opts...))
}

func (s *KvStore) newNullStore(t StoreType) {
	s.newIndexer(t, NullCacher)
}

func (s *KvStore) newIndexer(t StoreType, cacher Cacher) {
	indexer := NewCacheIndexer(t, cacher)
	s.indexers[t] = indexer
	indexer.Run()
}

func (s *KvStore) Run() {
	go s.store()
	s.asyncTaskSvc.Run()
}

func (s *KvStore) StoreSize(t StoreType) int {
	switch t {
	case DOMAIN:
		return 10
	case INSTANCE, LEASE:
		return 1000
	default:
		return 100
	}
}

func (s *KvStore) SelfPreservationHandler() DeferHandler {
	return &InstanceEventDeferHandler{Percent: DEFAULT_SELF_PRESERVATION_PERCENT}
}

func (s *KvStore) store() {
	for t := StoreType(0); t != typeEnd; t++ {
		switch t {
		case INSTANCE:
			s.newStore(t, WithDeferHandler(s.SelfPreservationHandler()))
		case SCHEMA:
			continue
		default:
			s.newStore(t)
		}
	}
	for _, i := range s.indexers {
		<-i.Ready()
	}
	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("all indexers are ready")
}

func (s *KvStore) onLeaseEvent(evt *KvEvent) {
	if evt.Action != pb.EVT_DELETE {
		return
	}

	key := util.BytesToStringWithNoCopy(evt.KV.Key)
	leaseID := util.BytesToStringWithNoCopy(evt.KV.Value)

	s.asyncTaskSvc.DeferRemove(ToLeaseAsyncTaskKey(key))

	util.Logger().Debugf("push task to async remove queue successfully, key %s %s [%s] event",
		key, leaseID, evt.Action)
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

	s.asyncTaskSvc.Stop()

	util.SafeCloseChan(s.ready)

	util.Logger().Debugf("store daemon stopped.")
}

func (s *KvStore) Ready() <-chan struct{} {
	<-s.asyncTaskSvc.Ready()
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

	err := s.asyncTaskSvc.Add(ctx, t)
	if err != nil {
		return 0, err
	}
	itf, err := s.asyncTaskSvc.LatestHandled(t.Key())
	if err != nil {
		return 0, err
	}
	pt := itf.(*LeaseAsyncTask)
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

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
package store

import (
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strconv"
	"sync"
)

const (
	SERVICE StoreType = iota
	INSTANCE
	DOMAIN
	SCHEMA // big data should not be stored in memory.
	RULE
	LEASE
	SERVICE_INDEX
	SERVICE_ALIAS
	SERVICE_TAG
	RULE_INDEX
	DEPENDENCY
	ENDPOINTS_INDEX
	typeEnd
)

var typeNames = []string{
	SERVICE:         "SERVICE",
	INSTANCE:        "INSTANCE",
	DOMAIN:          "DOMAIN",
	SCHEMA:          "SCHEMA",
	RULE:            "RULE",
	LEASE:           "LEASE",
	SERVICE_INDEX:   "SERVICE_INDEX",
	SERVICE_ALIAS:   "SERVICE_ALIAS",
	SERVICE_TAG:     "SERVICE_TAG",
	RULE_INDEX:      "RULE_INDEX",
	DEPENDENCY:      "DEPENDENCY",
	ENDPOINTS_INDEX: "ENDPOINTS_INDEX",
}

var store *KvStore

func init() {
	store = &KvStore{
		indexers:    make(map[StoreType]Indexer),
		asyncTasker: NewAsyncTasker(),
		ready:       make(chan struct{}),
	}
	for i := StoreType(0); i != typeEnd; i++ {
		store.newNullStore(i)
	}
	AddEventHandleFunc(DOMAIN, store.onDomainEvent)
	AddEventHandleFunc(LEASE, store.onLeaseEvent)
}

type LeaseAsyncTask struct {
	key     string
	LeaseID int64
	TTL     int64
	err     error
}

func (lat *LeaseAsyncTask) Key() string {
	return lat.key
}

func (lat *LeaseAsyncTask) Do(ctx context.Context) error {
	lat.TTL, lat.err = registry.GetRegisterCenter().LeaseRenew(ctx, lat.LeaseID)
	if lat.err != nil {
		util.LOGGER.Errorf(lat.err, "renew lease %d failed, key %s", lat.LeaseID, lat.Key())
	}
	return lat.err
}

func (lat *LeaseAsyncTask) Err() error {
	return lat.err
}

func NewLeaseAsyncTask(op *registry.PluginOp) *LeaseAsyncTask {
	return &LeaseAsyncTask{
		key:     registry.BytesToStringWithNoCopy(op.Key),
		LeaseID: op.Lease,
	}
}

type StoreType int

func (st StoreType) String() string {
	if int(st) < len(typeNames) {
		return typeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

type KvStore struct {
	indexers    map[StoreType]Indexer
	asyncTasker AsyncTasker
	lock        sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (s *KvStore) newStore(t StoreType, prefix string) {
	s.newCacherStore(t, NewCacher(prefix,
		func(evt *KvEvent) {
			EventHandler(t).OnEvent(evt)
		}))
}

func (s *KvStore) newNullStore(t StoreType) {
	s.newCacherStore(t, NullCacher)
}

func (s *KvStore) newCacherStore(t StoreType, cacher Cacher) {
	indexer := NewKvCacheIndexer(t, cacher)
	s.indexers[t] = indexer
	indexer.Run()
}

func (s *KvStore) Run() {
	go s.store()
	s.asyncTasker.Run()
}

func (s *KvStore) store() {
	// TODO should cache data group by domain.
	s.newStore(DOMAIN, apt.GetDomainRootKey())
	s.newStore(SERVICE, apt.GetServiceRootKey(""))
	s.newStore(INSTANCE, apt.GetInstanceRootKey(""))
	s.newStore(LEASE, apt.GetInstanceLeaseRootKey(""))
	s.newStore(SERVICE_INDEX, apt.GetServiceIndexRootKey(""))
	s.newStore(SERVICE_ALIAS, apt.GetServiceAliasRootKey(""))
	s.newStore(ENDPOINTS_INDEX, apt.GetInstancesEndpointsIndexRootKey(""))
	// TODO current key design does not support cache store.
	// s.newStore(DEPENDENCY, apt.GetServiceDependencyRootKey(domain))
	// s.newStore(SERVICE_TAG, apt.GetServiceTagRootKey(domain))
	// s.newStore(RULE, apt.GetServiceRuleRootKey(domain))
	// s.newStore(RULE_INDEX, apt.GetServiceRuleIndexRootKey(domain))
	for _, i := range s.indexers {
		<-i.Ready()
	}
	util.SafeCloseChan(s.ready)

	util.LOGGER.Debugf("all indexers are ready")
}

func (s *KvStore) onDomainEvent(evt *KvEvent) {
	kv := evt.KV
	action := evt.Action
	tenant := pb.GetInfoFromTenantKV(kv)

	if action != pb.EVT_CREATE {
		util.LOGGER.Infof("tenant '%s' is %s", tenant, action)
		return
	}

	if len(tenant) == 0 {
		util.LOGGER.Errorf(nil,
			"unmarshal tenant info failed, key %s [%s] event", string(kv.Key), action)
		return
	}

	util.LOGGER.Infof("new tenant %s is created", tenant)
}

func (s *KvStore) onLeaseEvent(evt *KvEvent) {
	if evt.Action != pb.EVT_DELETE {
		return
	}

	key := registry.BytesToStringWithNoCopy(evt.KV.Key)
	leaseID := registry.BytesToStringWithNoCopy(evt.KV.Value)

	s.removeAsyncTaske(key)

	util.LOGGER.Debugf("push task to async remove queue successfully, key %s %s [%s] event",
		key, leaseID, evt.Action)
}

func (s *KvStore) removeAsyncTaske(key string) {
	s.asyncTasker.RemoveTask(key)
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

	s.asyncTasker.Stop()

	util.SafeCloseChan(s.ready)

	util.LOGGER.Debugf("store daemon stopped.")
}

func (s *KvStore) Ready() <-chan struct{} {
	<-s.asyncTasker.Ready()
	return s.ready
}

func (s *KvStore) Service() Indexer {
	return s.indexers[SERVICE]
}

func (s *KvStore) Instance() Indexer {
	return s.indexers[INSTANCE]
}

func (s *KvStore) Lease() Indexer {
	return s.indexers[LEASE]
}

func (s *KvStore) ServiceIndex() Indexer {
	return s.indexers[SERVICE_INDEX]
}

func (s *KvStore) ServiceAlias() Indexer {
	return s.indexers[SERVICE_ALIAS]
}

func (s *KvStore) ServiceTag() Indexer {
	return s.indexers[SERVICE_TAG]
}

func (s *KvStore) Rule() Indexer {
	return s.indexers[RULE]
}

func (s *KvStore) RuleIndex() Indexer {
	return s.indexers[RULE_INDEX]
}

func (s *KvStore) Schema() Indexer {
	return s.indexers[SCHEMA]
}

func (s *KvStore) Dependency() Indexer {
	return s.indexers[DEPENDENCY]
}

func (s *KvStore) EndpointsIndex() Indexer {
	return s.indexers[ENDPOINTS_INDEX]
}

func (s *KvStore) KeepAlive(ctx context.Context, op *registry.PluginOp) (int64, error) {
	t := NewLeaseAsyncTask(op)
	if op.WithNoCache {
		util.LOGGER.Debugf("keep alive lease WitchNoCache, request etcd server, op: %s", op)
		err := t.Do(ctx)
		ttl := t.TTL
		return ttl, err
	}

	err := s.asyncTasker.AddTask(ctx, t)
	if err != nil {
		return 0, err
	}
	itf, err := s.asyncTasker.LatestHandled(t.Key())
	if err != nil {
		return 0, err
	}
	t = itf.(*LeaseAsyncTask)
	return t.TTL, t.Err()
}

func Store() *KvStore {
	return store
}

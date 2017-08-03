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
	"github.com/ServiceComb/service-center/util"
	"strconv"
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
	typeEnd
)

var typeNames = []string{
	SERVICE:       "SERVICE",
	INSTANCE:      "INSTANCE",
	DOMAIN:        "DOMAIN",
	SCHEMA:        "SCHEMA",
	RULE:          "RULE",
	LEASE:         "LEASE",
	SERVICE_INDEX: "SERVICE_INDEX",
	SERVICE_ALIAS: "SERVICE_ALIAS",
	SERVICE_TAG:   "SERVICE_TAG",
	RULE_INDEX:    "RULE_INDEX",
}

var (
	kvStoreEventFuncMap map[StoreType][]KvEventFunc
	store               *KvStore
)

func init() {
	store = &KvStore{
		cachers:  make(map[StoreType]Cacher),
		indexers: make(map[StoreType]Indexer),
		ready:    make(chan struct{}),
	}
	kvStoreEventFuncMap = make(map[StoreType][]KvEventFunc)
	for i := StoreType(0); i != typeEnd; i++ {
		kvStoreEventFuncMap[i] = make([]KvEventFunc, 0, 5)
		store.newNullStore(i)
	}
	AddKvStoreEventFunc(DOMAIN, store.onDomainEvent)
	AddKvStoreEventFunc(LEASE, store.onLeaseEvent)
}

type StoreType int

func (st StoreType) String() string {
	if int(st) < len(typeNames) {
		return typeNames[st]
	}
	return "TYPE" + strconv.Itoa(int(st))
}

type KvStore struct {
	cachers  map[StoreType]Cacher
	indexers map[StoreType]Indexer
	isClose  bool
	ready    chan struct{}
}

func (s *KvStore) newStore(t StoreType, prefix string) {
	c := NewCacher(prefix, func(evt *KvEvent) { s.onEvent(t, evt) })
	s.cachers[t] = c
	c.Run()

	s.indexers[t] = NewKvCacheIndexer(c.Cache())
}

func (s *KvStore) newNullStore(t StoreType) {
	s.cachers[t] = NullCacher
	s.indexers[t] = NewKvCacheIndexer(NullCacher.Cache())
}

func (s *KvStore) onEvent(t StoreType, evt *KvEvent) {
	fs := kvStoreEventFuncMap[t]
	for _, f := range fs {
		f(evt) // TODO how to be parallel?
	}
}

func (s *KvStore) Run() {
	s.storeDomain()
	<-s.ready
}

func (s *KvStore) storeDomain() {
	key := apt.GenerateTenantKey("")
	s.newStore(DOMAIN, key[:len(key)-1])
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
	s.storeDomainData(tenant)
	return
}

func (s *KvStore) onLeaseEvent(evt *KvEvent) {
	if evt.Action != pb.EVT_DELETE {
		return
	}

	switch evt.Action {
	case pb.EVT_CREATE:
	case pb.EVT_DELETE:
		temp := string(evt.KV.Value)
		leaseID, err := strconv.ParseInt(temp, 10, 64)
		if err != nil {
			util.LOGGER.Errorf(err,
				"remove async lease tasker failed, key %s %s [%s] event",
				string(evt.KV.Key), temp, evt.Action)
		}
		s.Lease().(*KvCacheIndexer).RemoveAsyncLeaseTasker(leaseID)
		util.LOGGER.Debugf("remove async lease tasker successfully, key %s %s [%s] event",
			string(evt.KV.Key), temp, evt.Action)
	}
}

func (s *KvStore) storeDomainData(domain string) {
	s.newStore(SERVICE, apt.GetServiceRootKey(domain))
	s.newStore(INSTANCE, apt.GetInstanceRootKey(domain))
	s.newStore(LEASE, apt.GetInstanceLeaseRootKey(domain))
	s.newStore(SERVICE_INDEX, apt.GetServiceIndexRootKey(domain))
	s.newStore(SERVICE_ALIAS, apt.GetServiceAliasRootKey(domain))
	// TODO current key design does not support cache store.
	// s.newStore(SERVICE_TAG, apt.GetServiceTagRootKey(domain))
	// s.newStore(RULE, apt.GetServiceRuleRootKey(domain))
	// s.newStore(RULE_INDEX, apt.GetServiceRuleIndexRootKey(domain))
}

func (s *KvStore) closed() bool {
	return s.isClose
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	for _, c := range s.cachers {
		c.Stop()
	}
	util.LOGGER.Debugf("store daemon stopped.")
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

func Store() *KvStore {
	return store
}

func AddKvStoreEventFunc(t StoreType, f KvEventFunc) {
	kvStoreEventFuncMap[t] = append(kvStoreEventFuncMap[t], f)
}

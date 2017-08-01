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
	TAG
	SCHEMA
	LEASE
	SERVICE_INDEX
	SERVICE_ALIAS
	typeEnd
)

var typeNames = []string{
	SERVICE:       "SERVICE",
	INSTANCE:      "INSTANCE",
	DOMAIN:        "DOMAIN",
	TAG:           "TAG",
	SCHEMA:        "SCHEMA",
	LEASE:         "LEASE",
	SERVICE_INDEX: "SERVICE_INDEX",
	SERVICE_ALIAS: "SERVICE_ALIAS",
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
	AddKvStoreEventFunc(DOMAIN, store.onDomainCreate)
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

func (s *KvStore) storeService(key string) {
	s.newStore(SERVICE, key)
}

func (s *KvStore) storeInstance(key string) {
	s.newStore(INSTANCE, key)
}

func (s *KvStore) storeLease(key string) {
	s.newStore(LEASE, key)
}

func (s *KvStore) storeServiceIndex(key string) {
	s.newStore(SERVICE_INDEX, key)
}

func (s *KvStore) storeServiceAlias(key string) {
	s.newStore(SERVICE_ALIAS, key)
}

func (s *KvStore) onDomainCreate(evt *KvEvent) {
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
	s.storeService(apt.GetServiceRootKey(tenant))
	s.storeInstance(apt.GetInstanceRootKey(tenant))
	s.storeLease(apt.GetInstanceLeaseRootKey(tenant))
	s.storeServiceIndex(apt.GetServiceIndexRootKey(tenant))
	s.storeServiceAlias(apt.GetServiceAliasRootKey(tenant))
	return
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

func Store() *KvStore {
	return store
}

func AddKvStoreEventFunc(t StoreType, f KvEventFunc) {
	kvStoreEventFuncMap[t] = append(kvStoreEventFuncMap[t], f)
}

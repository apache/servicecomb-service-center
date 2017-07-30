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
	"time"
)

const (
	DEFAULT_LISTWATCH_TIMEOUT           = 30 * time.Second
	SERVICE                   StoreType = iota
	INSTANCE
	DOMAIN
	TAG
	SCHEMA
	LEASE
	INDEX
	typeEnd
)

var typeNames = []string{
	SERVICE:  "SERVICE",
	INSTANCE: "INSTANCE",
	DOMAIN:   "DOMAIN",
	TAG:      "TAG",
	SCHEMA:   "SCHEMA",
	LEASE:    "LEASE",
	INDEX:    "INDEX",
}

var (
	kvStoreEventFuncMap map[StoreType][]KvEventFunc
	store               *KvStore
)

func init() {
	kvStoreEventFuncMap = make(map[StoreType][]KvEventFunc)
	for i := StoreType(0); i != typeEnd; i++ {
		kvStoreEventFuncMap[i] = make([]KvEventFunc, 0, 5)
	}

	store = &KvStore{
		cachers: make(map[StoreType]Cacher),
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
	cachers map[StoreType]Cacher

	isClose bool
}

func (s *KvStore) newCache(prefix string, callback KvEventFunc) Cacher {
	return s.newTimeoutCache(DEFAULT_LISTWATCH_TIMEOUT, prefix, callback)
}

func (s *KvStore) newTimeoutCache(ot time.Duration, prefix string, callback KvEventFunc) Cacher {
	return NewKvCacher(&KvCacherConfig{
		Key:     prefix,
		Timeout: ot,
		Period:  time.Second,
		OnEvent: callback,
	})
}

func (s *KvStore) newStoreCacher(t StoreType, prefix string) Cacher {
	c := s.newCache(prefix, func(evt *KvEvent) { s.onEvent(t, evt) })
	s.cachers[t] = c
	return c
}

func (s *KvStore) onEvent(t StoreType, evt *KvEvent) {
	fs := kvStoreEventFuncMap[t]
	for _, f := range fs {
		f(evt) // TODO how to be parallel?
	}
}

func (s *KvStore) Run() {
	s.storeDomain()
}

func (s *KvStore) storeDomain() {
	key := apt.GenerateTenantKey("")
	s.newStoreCacher(DOMAIN, key[:len(key)-1]).Run()
}

func (s *KvStore) storeService(serviceWatchByTenantKey string) {
	s.newStoreCacher(SERVICE, serviceWatchByTenantKey).Run()
}

func (s *KvStore) storeInstance(instanceWatchByTenantKey string) {
	s.newStoreCacher(INSTANCE, instanceWatchByTenantKey).Run()
}

func (s *KvStore) storeLease(leaseWatchByTenantKey string) {
	s.newStoreCacher(LEASE, leaseWatchByTenantKey).Run()
}

func (s *KvStore) storeServiceIndex(indexWatchByTenantKey string) {
	s.newStoreCacher(INDEX, indexWatchByTenantKey).Run()
}

func (s *KvStore) onDomainCreate(evt *KvEvent) {
	kv := evt.KV
	action := evt.Action

	if action != pb.EVT_CREATE {
		return
	}

	tenant := pb.GetInfoFromTenantKV(kv)
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
	return
}

func (s *KvStore) closed() bool {
	return s.isClose
}

func (s *KvStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true // TODO really stop?
}

func Store() *KvStore {
	return store
}

func AddKvStoreEventFunc(t StoreType, f KvEventFunc) {
	kvStoreEventFuncMap[t] = append(kvStoreEventFuncMap[t], f)
}

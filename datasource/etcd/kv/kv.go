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

package kv

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"time"
)

var store = &KvStore{}

func init() {
	store.Initialize()
	registerInnerTypes()
}

type KvStore struct {
	AddOns    map[sd.Type]AddOn
	adaptors  util.ConcurrentMap
	ready     chan struct{}
	goroutine *gopool.Pool
	isClose   bool
	rev       int64
}

func (s *KvStore) Initialize() {
	s.AddOns = make(map[sd.Type]AddOn)
	s.ready = make(chan struct{})
	s.goroutine = gopool.New(context.Background())
}

func (s *KvStore) OnCacheEvent(evt sd.KvEvent) {
	if s.rev < evt.Revision {
		s.rev = evt.Revision
	}
}

func (s *KvStore) InjectConfig(cfg *sd.Config) *sd.Config {
	return cfg.AppendEventFunc(s.OnCacheEvent)
}

func (s *KvStore) repo() sd.AdaptorRepository {
	return sd.Instance()
}

func (s *KvStore) getOrCreateAdaptor(t sd.Type) sd.Adaptor {
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
	return v.(sd.Adaptor)
}

func (s *KvStore) Run() {
	s.goroutine.Do(s.store)
	s.goroutine.Do(s.autoClearCache)
}

func (s *KvStore) store(ctx context.Context) {
	// new all types
	for _, t := range sd.Types {
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
			for _, t := range sd.Types {
				cache, ok := s.getOrCreateAdaptor(t).Cache().(sd.Cache)
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
		item.Value.(sd.Adaptor).Stop()
		return true
	})

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	log.Debugf("store daemon stopped")
}

func (s *KvStore) Ready() <-chan struct{} {
	return s.ready
}

func (s *KvStore) Install(addOn AddOn) (id sd.Type, err error) {
	if addOn == nil || len(addOn.Name()) == 0 || addOn.Config() == nil {
		return sd.TypeError, errors.New("invalid parameter")
	}

	id, err = sd.RegisterType(addOn.Name())
	if err != nil {
		return
	}

	sd.EventProxy(id).InjectConfig(addOn.Config())

	s.InjectConfig(addOn.Config())

	s.AddOns[id] = addOn

	log.Infof("install new type %d:%s->%s", id, addOn.Name(), addOn.Config().Key)
	return
}

func (s *KvStore) MustInstall(addOn AddOn) sd.Type {
	id, err := s.Install(addOn)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *KvStore) Adaptors(id sd.Type) sd.Adaptor { return s.getOrCreateAdaptor(id) }
func (s *KvStore) Service() sd.Adaptor            { return s.Adaptors(SERVICE) }
func (s *KvStore) SchemaSummary() sd.Adaptor      { return s.Adaptors(SchemaSummary) }
func (s *KvStore) Instance() sd.Adaptor           { return s.Adaptors(INSTANCE) }
func (s *KvStore) Lease() sd.Adaptor              { return s.Adaptors(LEASE) }
func (s *KvStore) ServiceIndex() sd.Adaptor       { return s.Adaptors(ServiceIndex) }
func (s *KvStore) ServiceAlias() sd.Adaptor       { return s.Adaptors(ServiceAlias) }
func (s *KvStore) ServiceTag() sd.Adaptor         { return s.Adaptors(ServiceTag) }
func (s *KvStore) Rule() sd.Adaptor               { return s.Adaptors(RULE) }
func (s *KvStore) RuleIndex() sd.Adaptor          { return s.Adaptors(RuleIndex) }
func (s *KvStore) Schema() sd.Adaptor             { return s.Adaptors(SCHEMA) }
func (s *KvStore) DependencyRule() sd.Adaptor     { return s.Adaptors(DependencyRule) }
func (s *KvStore) DependencyQueue() sd.Adaptor    { return s.Adaptors(DependencyQueue) }
func (s *KvStore) Domain() sd.Adaptor             { return s.Adaptors(DOMAIN) }
func (s *KvStore) Project() sd.Adaptor            { return s.Adaptors(PROJECT) }

func Store() *KvStore {
	return store
}

func Revision() int64 {
	return store.rev
}

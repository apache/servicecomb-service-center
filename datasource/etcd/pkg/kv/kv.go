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

// kv package provides a TypeStore to manage the implementations of sd package, see types.go
package kv

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/datasource/etcd/pkg/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/config"
	"time"
)

var store = &TypeStore{}

func init() {
	store.Initialize()
	registerInnerTypes()
}

type TypeStore struct {
	AddOns    map[sd.Type]AddOn
	adaptors  util.ConcurrentMap
	ready     chan struct{}
	goroutine *gopool.Pool
	isClose   bool
	rev       int64
}

func (s *TypeStore) Initialize() {
	s.AddOns = make(map[sd.Type]AddOn)
	s.ready = make(chan struct{})
	s.goroutine = gopool.New(context.Background())
}

func (s *TypeStore) OnCacheEvent(evt sd.KvEvent) {
	if s.rev < evt.Revision {
		s.rev = evt.Revision
	}
}

func (s *TypeStore) InjectConfig(cfg *sd.Config) *sd.Config {
	return cfg.AppendEventFunc(s.OnCacheEvent)
}

func (s *TypeStore) repo() sd.AdaptorRepository {
	return sd.Instance()
}

func (s *TypeStore) getOrCreateAdaptor(t sd.Type) sd.Adaptor {
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

func (s *TypeStore) Run() {
	s.goroutine.Do(s.store)
	s.goroutine.Do(s.autoClearCache)
}

func (s *TypeStore) store(ctx context.Context) {
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

func (s *TypeStore) autoClearCache(ctx context.Context) {
	if config.ServerInfo.Config.CacheTTL == 0 {
		return
	}

	log.Infof("start auto clear cache in %v", config.ServerInfo.Config.CacheTTL)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(config.ServerInfo.Config.CacheTTL):
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

func (s *TypeStore) Stop() {
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

func (s *TypeStore) Ready() <-chan struct{} {
	return s.ready
}

func (s *TypeStore) Install(addOn AddOn) (id sd.Type, err error) {
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

func (s *TypeStore) MustInstall(addOn AddOn) sd.Type {
	id, err := s.Install(addOn)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *TypeStore) Adaptors(id sd.Type) sd.Adaptor { return s.getOrCreateAdaptor(id) }
func (s *TypeStore) Service() sd.Adaptor            { return s.Adaptors(SERVICE) }
func (s *TypeStore) SchemaSummary() sd.Adaptor      { return s.Adaptors(SchemaSummary) }
func (s *TypeStore) Instance() sd.Adaptor           { return s.Adaptors(INSTANCE) }
func (s *TypeStore) Lease() sd.Adaptor              { return s.Adaptors(LEASE) }
func (s *TypeStore) ServiceIndex() sd.Adaptor       { return s.Adaptors(ServiceIndex) }
func (s *TypeStore) ServiceAlias() sd.Adaptor       { return s.Adaptors(ServiceAlias) }
func (s *TypeStore) ServiceTag() sd.Adaptor         { return s.Adaptors(ServiceTag) }
func (s *TypeStore) Rule() sd.Adaptor               { return s.Adaptors(RULE) }
func (s *TypeStore) RuleIndex() sd.Adaptor          { return s.Adaptors(RuleIndex) }
func (s *TypeStore) Schema() sd.Adaptor             { return s.Adaptors(SCHEMA) }
func (s *TypeStore) DependencyRule() sd.Adaptor     { return s.Adaptors(DependencyRule) }
func (s *TypeStore) DependencyQueue() sd.Adaptor    { return s.Adaptors(DependencyQueue) }
func (s *TypeStore) Domain() sd.Adaptor             { return s.Adaptors(DOMAIN) }
func (s *TypeStore) Project() sd.Adaptor            { return s.Adaptors(PROJECT) }

func Store() *TypeStore {
	return store
}

func Revision() int64 {
	return store.rev
}

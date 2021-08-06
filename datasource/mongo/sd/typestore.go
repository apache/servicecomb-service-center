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

// Package sd provides a TypeStore to manage the implementations of sd package, see types.go
package sd

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/foundation/gopool"
)

var store = &TypeStore{}

func init() {
	store.Initialize()
}

type TypeStore struct {
	caches    util.ConcurrentMap
	ready     chan struct{}
	goroutine *gopool.Pool
	isClose   bool
}

func (s *TypeStore) Initialize() {
	s.ready = make(chan struct{})
	s.goroutine = goutil.New()
}

func (s *TypeStore) Run() {
	s.goroutine.Do(s.store)
	s.goroutine.Do(s.autoClearCache)
}

func (s *TypeStore) store(ctx context.Context) {
	// new all types
	for t := range CacherRegister {
		select {
		case <-ctx.Done():
			return
		case <-s.getOrCreateCache(t).Ready():
		}
	}
	util.SafeCloseChan(s.ready)
	log.Debug("all caches are ready")
}

func (s *TypeStore) autoClearCache(ctx context.Context) {
	ttl := config.GetRegistry().CacheTTL

	if ttl == 0 {
		return
	}

	log.Info(fmt.Sprintf("start auto clear cache in %v", ttl))
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(ttl):
			for t := range CacherRegister {
				cache := s.getOrCreateCache(t).Cache()
				cache.MarkDirty()
			}
			log.Warn("caches are marked dirty!")
		}
	}
}

func (s *TypeStore) getOrCreateCache(t string) *MongoCacher {
	cache, ok := s.caches.Get(t)
	if ok {
		return cache.(*MongoCacher)
	}
	f, ok := CacherRegister[t]
	if !ok {
		log.Fatal(fmt.Sprintf("unexpected type store "+t), nil)
	}
	cacher := f()
	cacher.Run()

	s.caches.Put(t, cacher)
	return cacher
}

func (s *TypeStore) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	log.Debug("store daemon stopped")
}

func (s *TypeStore) Ready() <-chan struct{} {
	return s.ready
}

func (s *TypeStore) TypeCacher(id string) *MongoCacher { return s.getOrCreateCache(id) }
func (s *TypeStore) Service() *MongoCacher             { return s.TypeCacher(service) }
func (s *TypeStore) Instance() *MongoCacher            { return s.TypeCacher(instance) }
func (s *TypeStore) Dep() *MongoCacher                 { return s.TypeCacher(dep) }

func Store() *TypeStore {
	return store
}

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

// Package state provides a State to manage the implementations of sd package, see types.go
package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-chassis/foundation/gopool"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Manager struct {
	Repository Repository
	Rev        int64

	states     map[kvstore.Type]State
	statesLock sync.RWMutex

	goroutine *gopool.Pool

	ready   chan struct{}
	isClose bool
}

func (s *Manager) Initialize() {
	s.states = make(map[kvstore.Type]State)
	s.ready = make(chan struct{})
	s.goroutine = goutil.New()
}

func (s *Manager) OnCacheEvent(evt kvstore.Event) {
	if s.Rev < evt.Revision {
		s.Rev = evt.Revision
	}
}

func (s *Manager) InjectConfig(cfg *kvstore.Options) *kvstore.Options {
	if !Configuration().EnableCache {
		cfg.WithInitSize(0)
	}
	cfg.AppendEventFunc(s.OnCacheEvent)
	return cfg
}

func (s *Manager) repo() Repository {
	return s.Repository
}

func (s *Manager) getOrCreateState(t kvstore.Type) State {
	s.statesLock.RLock()
	v, ok := s.states[t]
	if ok {
		s.statesLock.RUnlock()
		return v
	}
	s.statesLock.RUnlock()

	s.statesLock.Lock()
	p, ok := Plugins()[t]
	if ok {
		cfg := p.Config()
		kvstore.EventProxy(t).InjectConfig(cfg)
		s.InjectConfig(cfg)

		state := s.repo().New(t, cfg)
		state.Run()

		s.states[t] = state
		s.statesLock.Unlock()
		return state
	}
	s.statesLock.Unlock()

	log.Warn(fmt.Sprintf("type '%s' not found", t))
	return nil
}

func (s *Manager) stopStates() {
	s.statesLock.RLock()
	for _, state := range s.states {
		state.Stop()
	}
	s.statesLock.RUnlock()
}

func (s *Manager) Run() {
	s.goroutine.Do(s.store)
}

func (s *Manager) store(ctx context.Context) {
	// new all types
	for _, t := range kvstore.Types {
		state := s.getOrCreateState(t)
		if state == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-state.Ready():
		}
	}

	util.SafeCloseChan(s.ready)

	log.Debug("all states are ready")
}

func (s *Manager) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	s.stopStates()

	s.goroutine.Close(true)

	util.SafeCloseChan(s.ready)

	log.Debug("store daemon stopped")
}

func (s *Manager) Ready() <-chan struct{} {
	return s.ready
}

func (s *Manager) States(id kvstore.Type) State { return s.getOrCreateState(id) }

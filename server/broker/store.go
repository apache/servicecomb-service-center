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
package broker

import (
	"sync"

	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	sstore "github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	"golang.org/x/net/context"
)

const (
	PARTICIPANT sstore.StoreType = iota
	VERSION
	PACT
	PACT_VERSION
	PACT_TAG
	VERIFICATION
	PACT_LATEST
	typeEnd
)

var TypeNames = []string{
	PARTICIPANT:  "PARTICIPANT",
	VERSION:      "VERSION",
	PACT:         "PACT",
	PACT_VERSION: "PACT_VERSION",
	PACT_TAG:     "PACT_TAG",
	VERIFICATION: "VERIFICATION",
	PACT_LATEST:  "PACT_LATEST",
}

var TypeRoots = map[sstore.StoreType]string{
	PARTICIPANT:  GetBrokerParticipantKey(""),
	VERSION:      GetBrokerVersionKey(""),
	PACT:         GetBrokerPactKey(""),
	PACT_VERSION: GetBrokerPactVersionKey(""),
	PACT_TAG:     GetBrokerTagKey(""),
	VERIFICATION: GetBrokerVerificationKey(""),
	PACT_LATEST:  GetBrokerLatestKey(""),
}

var store = &BKvStore{}

func Store() *BKvStore {
	return store
}

func (s *BKvStore) StoreSize(t sstore.StoreType) int {
	return 100
}

func (s *BKvStore) newStore(t sstore.StoreType, opts ...sstore.KvCacherCfgOption) {
	opts = append(opts,
		sstore.WithKey(TypeRoots[t]),
		sstore.WithInitSize(s.StoreSize(t)),
	)
	s.newIndexer(t, sstore.NewKvCacher(opts...))
}

func (s *BKvStore) store(ctx context.Context) {
	for t := sstore.StoreType(0); t != typeEnd; t++ {
		s.newStore(t)
	}
	for _, i := range s.bindexers {
		select {
		case <-ctx.Done():
			return
		case <-i.Ready():
		}
	}
	util.SafeCloseChan(s.bready)

	util.Logger().Debugf("all indexers are ready")
}

func init() {
	store.Initialize()
	store.Run()
	store.Ready()
}

type BKvStore struct {
	*sstore.KvStore
	bindexers map[sstore.StoreType]*sstore.Indexer
	block     sync.RWMutex
	bready    chan struct{}
	bisClose  bool
}

func (s *BKvStore) Initialize() {
	s.KvStore = sstore.Store()
	s.KvStore.Initialize()
	s.bindexers = make(map[sstore.StoreType]*sstore.Indexer)
	s.bready = make(chan struct{})

	for i := sstore.StoreType(0); i != typeEnd; i++ {
		store.newNullStore(i)
	}
}

func (s *BKvStore) newNullStore(t sstore.StoreType) {
	s.newIndexer(t, sstore.NullCacher)
}

func (s *BKvStore) newIndexer(t sstore.StoreType, cacher sstore.Cacher) {
	indexer := sstore.NewCacheIndexer(t, cacher)
	s.bindexers[t] = indexer
	indexer.Run()
}

func (s *BKvStore) Run() {
	util.Go(func(ctx context.Context) {
		s.store(ctx)
		select {
		case <-ctx.Done():
			s.Stop()
		}
	})
}

func (s *BKvStore) Ready() <-chan struct{} {
	return s.bready
}

func (s *BKvStore) Participant() *sstore.Indexer {
	return s.bindexers[PARTICIPANT]
}

func (s *BKvStore) Version() *sstore.Indexer {
	return s.bindexers[VERSION]
}

func (s *BKvStore) Pact() *sstore.Indexer {
	return s.bindexers[PACT]
}

func (s *BKvStore) PactVersion() *sstore.Indexer {
	return s.bindexers[PACT_VERSION]
}

func (s *BKvStore) PactTag() *sstore.Indexer {
	return s.bindexers[PACT_TAG]
}

func (s *BKvStore) Verification() *sstore.Indexer {
	return s.bindexers[VERIFICATION]
}

func (s *BKvStore) PactLatest() *sstore.Indexer {
	return s.bindexers[PACT_LATEST]
}

func (s *BKvStore) Stop() {
	if s.bisClose {
		return
	}
	s.bisClose = true

	for _, i := range s.bindexers {
		i.Stop()
	}

	util.SafeCloseChan(s.bready)

	util.Logger().Debugf("broker store daemon stopped")
}

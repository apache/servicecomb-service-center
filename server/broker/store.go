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
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"
)

var (
	PARTICIPANT  discovery.Type
	VERSION      discovery.Type
	PACT         discovery.Type
	PACT_VERSION discovery.Type
	PACT_TAG     discovery.Type
	VERIFICATION discovery.Type
	PACT_LATEST  discovery.Type
)

var brokerKvStore = &BKvStore{}

func init() {
	PARTICIPANT = backend.Store().MustInstall(discovery.NewAddOn("PARTICIPANT", discovery.Configure().WithPrefix(GetBrokerParticipantKey(""))))
	VERSION = backend.Store().MustInstall(discovery.NewAddOn("VERSION", discovery.Configure().WithPrefix(GetBrokerVersionKey(""))))
	PACT = backend.Store().MustInstall(discovery.NewAddOn("PACT", discovery.Configure().WithPrefix(GetBrokerPactKey(""))))
	PACT_VERSION = backend.Store().MustInstall(discovery.NewAddOn("PACT_VERSION", discovery.Configure().WithPrefix(GetBrokerPactVersionKey(""))))
	PACT_TAG = backend.Store().MustInstall(discovery.NewAddOn("PACT_TAG", discovery.Configure().WithPrefix(GetBrokerTagKey(""))))
	VERIFICATION = backend.Store().MustInstall(discovery.NewAddOn("VERIFICATION", discovery.Configure().WithPrefix(GetBrokerVerificationKey(""))))
	PACT_LATEST = backend.Store().MustInstall(discovery.NewAddOn("PACT_LATEST", discovery.Configure().WithPrefix(GetBrokerLatestKey(""))))
}

type BKvStore struct {
}

func (s *BKvStore) Participant() discovery.Indexer {
	return backend.Store().Adaptors(PARTICIPANT)
}

func (s *BKvStore) Version() discovery.Indexer {
	return backend.Store().Adaptors(VERSION)
}

func (s *BKvStore) Pact() discovery.Indexer {
	return backend.Store().Adaptors(PACT)
}

func (s *BKvStore) PactVersion() discovery.Indexer {
	return backend.Store().Adaptors(PACT_VERSION)
}

func (s *BKvStore) PactTag() discovery.Indexer {
	return backend.Store().Adaptors(PACT_TAG)
}

func (s *BKvStore) Verification() discovery.Indexer {
	return backend.Store().Adaptors(VERIFICATION)
}

func (s *BKvStore) PactLatest() discovery.Indexer {
	return backend.Store().Adaptors(PACT_LATEST)
}

func Store() *BKvStore {
	return brokerKvStore
}

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
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
)

var (
	PARTICIPANT  sd.Type
	VERSION      sd.Type
	PACT         sd.Type
	PactVersion  sd.Type
	PactTag      sd.Type
	VERIFICATION sd.Type
	PactLatest   sd.Type
)

var brokerKvStore = &BKvStore{}

type BKvStore struct {
}

func (s *BKvStore) Participant() sd.Indexer {
	return kv.Store().Adaptors(PARTICIPANT)
}

func (s *BKvStore) Version() sd.Indexer {
	return kv.Store().Adaptors(VERSION)
}

func (s *BKvStore) Pact() sd.Indexer {
	return kv.Store().Adaptors(PACT)
}

func (s *BKvStore) PactVersion() sd.Indexer {
	return kv.Store().Adaptors(PactVersion)
}

func (s *BKvStore) PactTag() sd.Indexer {
	return kv.Store().Adaptors(PactTag)
}

func (s *BKvStore) Verification() sd.Indexer {
	return kv.Store().Adaptors(VERIFICATION)
}

func (s *BKvStore) PactLatest() sd.Indexer {
	return kv.Store().Adaptors(PactLatest)
}

func Store() *BKvStore {
	return brokerKvStore
}

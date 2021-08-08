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

package etcd

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEtcdState(t *testing.T) {
	i := NewEtcdState("a", kvstore.NewOptions().WithInitSize(1))
	_, ok := i.Indexer.(*CacheIndexer)
	assert.True(t, ok)

	i = NewEtcdState("a", kvstore.NewOptions().WithInitSize(0))
	_, ok = i.Indexer.(*Indexer)
	assert.True(t, ok)

	i = NewEtcdState("a", kvstore.NewOptions())
	_, ok = i.Indexer.(*CacheIndexer)
	assert.True(t, ok)
	_, ok = i.Cacher.(*KvCacher)
	assert.True(t, ok)
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(state.Config{})
	i := repo.(*Repository).New(0, kvstore.NewOptions())
	_, ok := i.(*State)
	assert.True(t, ok)
}

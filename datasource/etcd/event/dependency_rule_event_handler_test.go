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
package event

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
)

func TestNewDependencyRuleEventHandler(t *testing.T) {
	consumerId := "1"
	provider := &pb.MicroServiceKey{Tenant: "x/y", Version: "0+"}
	b := cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
	h := NewDependencyRuleEventHandler()
	h.OnEvent(kvstore.Event{Type: pb.EVT_CREATE})
	b = cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if !b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
	h.OnEvent(kvstore.Event{Type: pb.EVT_INIT})
	b = cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if !b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
	h.OnEvent(kvstore.Event{Type: pb.EVT_UPDATE, KV: &kvstore.KeyValue{
		Key: []byte(path.GenerateProviderDependencyRuleKey("x/y", provider))}})
	b = cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
	h.OnEvent(kvstore.Event{Type: pb.EVT_DELETE, KV: &kvstore.KeyValue{
		Key: []byte(path.GenerateProviderDependencyRuleKey("x/y", provider))}})
	b = cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
	h.OnEvent(kvstore.Event{Type: pb.EVT_DELETE, KV: &kvstore.KeyValue{
		Key: []byte(path.GenerateConsumerDependencyRuleKey("x/y", provider))}})
	b = cache.DependencyRule.ExistRule(context.Background(), consumerId, provider)
	if !b {
		t.Fatalf("TestNewDependencyRuleEventHandler failed")
	}
}

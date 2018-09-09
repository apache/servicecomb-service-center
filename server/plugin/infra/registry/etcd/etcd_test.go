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

import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/tracing/buildin"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/security/buildin"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/tls/buildin"
import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	dialTimeout = 500 * time.Millisecond
	endpoint    = "127.0.0.1:2379"
)

func TestEtcdClient(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	select {
	case <-etcd.Ready():
	default:
		t.Fatalf("TestEtcdClient failed")
	}

	// base test
	registry.RegistryConfig().ClusterAddresses = endpoint
	inst := NewRegistry()
	if inst == nil || firstEndpoint != "http://"+endpoint {
		t.Fatalf("TestEtcdClient failed, %#v", firstEndpoint)
	}
	old1 := registry.RegistryConfig().ClusterAddresses
	old2 := registry.RegistryConfig().DialTimeout
	registry.RegistryConfig().ClusterAddresses = "x"
	registry.RegistryConfig().DialTimeout = dialTimeout
	inst = NewRegistry()
	if inst == nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	select {
	case <-inst.(*EtcdClient).Err():
	default:
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	registry.RegistryConfig().ClusterAddresses = old1
	registry.RegistryConfig().DialTimeout = old2

	// case: etcd do
	// put
	resp, err := etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/b"),
		registry.WithStrValue("b"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/b"))
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/a"),
		registry.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/a"),
		registry.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/a" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/a"),
		registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/c"),
		registry.WithStrValue("c"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/d"),
		registry.WithStrValue("d"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/dd"),
		registry.WithStrValue("dd"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || string(resp.Kvs[0].Value) != "d" ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || string(resp.Kvs[1].Value) != "dd" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix(), registry.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix(), registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd")) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" ||
		string(resp.Kvs[1].Key) != "/test_range/c" || string(resp.Kvs[1].Value) != "c" ||
		string(resp.Kvs[2].Key) != "/test_range/d" || string(resp.Kvs[2].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithKeyOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/c" || resp.Kvs[1].Value != nil ||
		string(resp.Kvs[2].Key) != "/test_range/d" || resp.Kvs[2].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithCountOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix paging
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || string(resp.Kvs[0].Value) != "c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" || string(resp.Kvs[1].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(), registry.WithKeyOnly(),
		registry.WithOffset(4), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/dd" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/d"), registry.WithPrefix(), registry.WithKeyOnly(),
		registry.WithOffset(0), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 2 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(), registry.WithCountOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(6), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// if offset < -1, just paging by limit
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(-2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 5 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range paging
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithKeyOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/d" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithCountOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(5), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(4), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcd.Do(context.Background(), registry.DEL,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd")) // [b, d) !!!
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || len(resp.Kvs) != 2 || string(resp.Kvs[1].Key) != "/test_range/dd" {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", resp.Kvs)
	}
	// delete prefix
	resp, err = etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}

	// large data
	var wg sync.WaitGroup
	for i := 0; i < registry.DEFAULT_PAGE_COUNT+1; i++ {
		wg.Add(1)
		v := strconv.Itoa(i)
		go func() {
			defer wg.Done()
			resp, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_page/"+v),
				registry.WithStrValue(v))
			if err != nil || !resp.Succeeded {
				t.Fatalf("TestEtcdClient_Do failed, %#v", err)
			}
		}()
	}
	wg.Wait()
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"),
		registry.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded || resp.Count != registry.DEFAULT_PAGE_COUNT+1 ||
		len(resp.Kvs) != registry.DEFAULT_PAGE_COUNT+1 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"), registry.WithPrefix(), registry.WithDescendOrder())
	if err != nil || !resp.Succeeded || resp.Count != registry.DEFAULT_PAGE_COUNT+1 ||
		len(resp.Kvs) != registry.DEFAULT_PAGE_COUNT+1 ||
		string(resp.Kvs[0].Key) != "/test_page/999" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcd.Do(context.Background(), registry.DEL,
		registry.WithStrKey("/test_page/"),
		registry.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"), registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
}

func TestEtcdClient_Compact(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	err = etcd.Compact(context.Background(), 0)
	if err != nil {
		t.Fatalf("TestEtcdClient_Compact failed")
	}
	err = etcd.Compact(context.Background(), 0)
	if err == nil {
		t.Fatalf("TestEtcdClient_Compact failed")
	}
}

func TestEtcdClient_Txn(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	resp, err := etcd.Txn(context.Background(), nil)
	if err == nil || resp != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	success, err := etcd.PutNoOverride(context.Background(), registry.WithStrKey("/test_txn/a"))
	if err != nil || !success {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	success, err = etcd.PutNoOverride(context.Background(), registry.WithStrKey("/test_txn/a"), registry.WithStrValue("a"))
	if err != nil || success {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	resp, err = etcd.Txn(context.Background(), []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: registry.Put, Key: []byte("/test_txn/b"), Value: []byte("b")},
	})
	if err != nil || resp == nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_txn/"),
		registry.WithPrefix(), registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcd.TxnWithCmp(context.Background(), []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: registry.Put, Key: []byte("/test_txn/b"), Value: []byte("b")},
	}, []registry.CompareOp{
		{[]byte("/test_txn/a"), registry.CMP_VALUE, registry.CMP_EQUAL, "a"},
	}, []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_txn/c"), Value: []byte("c")},
		{Action: registry.Put, Key: []byte("/test_txn/d"), Value: []byte("d")},
	})
	if err != nil || resp == nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	// case: range request
	resp, err = etcd.TxnWithCmp(context.Background(), nil, []registry.CompareOp{
		{[]byte("/test_txn/c"), registry.CMP_VALUE, registry.CMP_EQUAL, "c"},
	}, []registry.PluginOp{
		{Action: registry.Get, Key: []byte("/test_txn/a")},
		{Action: registry.Get, Key: []byte("/test_txn/"), Prefix: true},
	})
	if err != nil || resp == nil || resp.Succeeded || resp.Count != 3 { // a + [a,b]
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	// case: test key not exist
	resp, err = etcd.TxnWithCmp(context.Background(), []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: registry.Put, Key: []byte("/test_txn/b"), Value: []byte("b")},
	}, []registry.CompareOp{
		{[]byte("/test_txn/c"), registry.CMP_VALUE, registry.CMP_EQUAL, "c"},
	}, []registry.PluginOp{
		{Action: registry.Delete, Key: []byte("/test_txn/"), Prefix: true},
	})
	if err != nil || resp == nil || resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	resp, err = etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_txn/"),
		registry.WithPrefix(), registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
}

func TestEtcdClient_LeaseRenew(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	id, err := etcd.LeaseGrant(context.Background(), -1)
	if err != nil || id == 0 {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	id, err = etcd.LeaseGrant(context.Background(), 0)
	if err != nil || id == 0 {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	id, err = etcd.LeaseGrant(context.Background(), 2)
	if err != nil || id == 0 {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	ttl, err := etcd.LeaseRenew(context.Background(), id)
	if err != nil || ttl != 2 {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	err = etcd.LeaseRevoke(context.Background(), id)
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	ttl, err = etcd.LeaseRenew(context.Background(), id)
	if err == nil || ttl != 0 {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
}

func TestEtcdClient_HealthCheck(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:        []string{endpoint},
		DialTimeout:      dialTimeout,
		AutoSyncInterval: time.Millisecond,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	err = etcd.ReOpen()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), dialTimeout)
	err = etcd.SyncMembers(ctx)
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	etcd.Endpoints = []string{"x"}
	err = etcd.ReOpen()
	if err == nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	ctx, _ = context.WithTimeout(context.Background(), dialTimeout)
	err = etcd.SyncMembers(ctx)
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	etcd.Endpoints = []string{endpoint}

	etcd.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				_, err = etcd.Do(context.Background(), registry.GET,
					registry.WithStrKey("/test_health/"))
				if err != nil {
					continue
				}
			case <-ctx.Done():
			}
			break
		}
		if err != nil {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
		cancel()
	}()
	etcd.healthCheckLoop(ctx)
}

func TestEtcdClient_Watch(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcd.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()

	defer func() {
		resp, err := etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_watch/"),
			registry.WithPrefix())
		if err != nil || !resp.Succeeded {
			t.Fatalf("TestEtcdClient_Do failed, %#v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = etcd.Watch(ctx, registry.WithStrKey("/test_watch/a"))
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	ch := make(chan struct{})
	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), registry.WithStrKey("/test_watch/a"),
			registry.WithWatchCallback(func(message string, evt *registry.PluginResponse) error {
				if evt.Count != 1 || len(evt.Kvs) != 1 || evt.Action != registry.Put ||
					string(evt.Kvs[0].Key) != "/test_watch/a" || string(evt.Kvs[0].Value) != "a" {
					t.Fatalf("TestEtcdClient failed, %#v", evt)
				}
				return fmt.Errorf("error")
			}))
		if err == nil || err.Error() != "error" {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
	}()

	<-time.After(500 * time.Millisecond)
	resp, err := etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_watch/a"),
		registry.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), registry.WithStrKey("/test_watch/"),
			registry.WithPrefix(),
			registry.WithWatchCallback(func(message string, evt *registry.PluginResponse) error {
				equalA := evt.Action == registry.Put && string(evt.Kvs[0].Key) == "/test_watch/a" && string(evt.Kvs[0].Value) == "a"
				equalB := evt.Action == registry.Put && string(evt.Kvs[1].Key) == "/test_watch/b" && string(evt.Kvs[0].Value) == "b"
				if evt.Count != 2 || len(evt.Kvs) != 2 || !(equalA || equalB) {
					t.Fatalf("TestEtcdClient failed, %#v", evt)
				}
				return fmt.Errorf("error")
			}))
		if err == nil || err.Error() != "error" {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
	}()

	<-time.After(500 * time.Millisecond)
	resp, err = etcd.Txn(context.Background(), []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_watch/a"), Value: []byte("a")},
		{Action: registry.Put, Key: []byte("/test_watch/b"), Value: []byte("b")},
	})
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	// diff action type will be split
	go func() {
		defer func() { ch <- struct{}{} }()
		var times = 3
		err = etcd.Watch(context.Background(), registry.WithStrKey("/test_watch/"),
			registry.WithPrefix(),
			registry.WithWatchCallback(func(message string, evt *registry.PluginResponse) error {
				equalA := evt.Action == registry.Delete && string(evt.Kvs[0].Key) == "/test_watch/a" && evt.Kvs[0].Value == nil
				equalB := evt.Action == registry.Put && string(evt.Kvs[0].Key) == "/test_watch/b" && string(evt.Kvs[0].Value) == "b"
				equalC := evt.Action == registry.Put && string(evt.Kvs[0].Key) == "/test_watch/c" && string(evt.Kvs[0].Value) == "c"
				if evt.Count != 1 || len(evt.Kvs) != 1 || !(equalA || equalB || equalC) {
					t.Fatalf("TestEtcdClient failed, %#v", evt)
				}
				times--
				if times == 0 {
					return fmt.Errorf("error")
				}
				return nil
			}))
		if err == nil || err.Error() != "error" {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
	}()

	<-time.After(500 * time.Millisecond)
	resp, err = etcd.Txn(context.Background(), []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_watch/c"), Value: []byte("c")},
		{Action: registry.Delete, Key: []byte("/test_watch/a"), Value: []byte("a")},
		{Action: registry.Put, Key: []byte("/test_watch/b"), Value: []byte("b")},
	})
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	// watch with rev
	resp, err = etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_watch/c"),
		registry.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	rev := resp.Revision
	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), registry.WithStrKey("/test_watch/"),
			registry.WithPrefix(),
			registry.WithRev(rev),
			registry.WithWatchCallback(func(message string, evt *registry.PluginResponse) error {
				if evt.Count != 1 || len(evt.Kvs) != 1 || evt.Action != registry.Delete ||
					string(evt.Kvs[0].Key) != "/test_watch/c" || evt.Kvs[0].Value != nil {
					t.Fatalf("TestEtcdClient failed, %#v", evt)
				}
				return fmt.Errorf("error")
			}))
		if err == nil || err.Error() != "error" {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
	}()
	<-ch

	// delete with prevKV
	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), registry.WithStrKey("/test_watch/"),
			registry.WithPrefix(), registry.WithPrevKv(),
			registry.WithWatchCallback(func(message string, evt *registry.PluginResponse) error {
				if len(evt.Kvs) != 1 || evt.Action != registry.Delete ||
					string(evt.Kvs[0].Key) != "/test_watch/b" || string(evt.Kvs[0].Value) != "b" {
					t.Fatalf("TestEtcdClient failed, %#v", evt)
				}
				return fmt.Errorf("error")
			}))
		if err == nil || err.Error() != "error" {
			t.Fatalf("TestEtcdClient failed, %#v", err)
		}
	}()
	<-time.After(500 * time.Millisecond)
	resp, err = etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_watch/b"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch
}

func TestNewRegistry(t *testing.T) {
	etcd := &EtcdClient{
		Endpoints:        []string{endpoint, "0.0.0.0:2379"},
		DialTimeout:      dialTimeout,
		AutoSyncInterval: time.Millisecond,
	}
	err := etcd.Initialize()
	if err == nil {
		// should be err, member list does not contain one of the endpoints.
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
}

type mockTLSHandler struct {
}

func (m *mockTLSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestWithTLS(t *testing.T) {
	sslRoot := "../../../../../examples/service_center/ssl/"
	os.Setenv("SSL_ROOT", sslRoot)

	core.ServerInfo.Config.SslEnabled = true
	registry.RegistryConfig().SslEnabled = true
	defer func() {
		core.ServerInfo.Config.SslEnabled = false
		registry.RegistryConfig().SslEnabled = false
		os.Setenv("SSL_ROOT", "")
	}()

	svr, err := rpc.NewServer("127.0.0.1:0")
	go func() {
		svr.Serve()
	}()
	defer svr.Stop()

	etcd := &EtcdClient{
		DialTimeout: dialTimeout,
		Endpoints:   []string{svr.Listener.Addr().String()},
	}

	err = etcd.Initialize()
	// initialize the etcd client will check member list firstly,
	// so will raise an grpc error but not TLS errors.
	if _, ok := status.FromError(err); !ok {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcd.Close()
}

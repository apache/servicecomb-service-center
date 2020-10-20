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
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	_ "github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/stretchr/testify/assert"
)
import _ "github.com/apache/servicecomb-service-center/server/plugin/security/cipher/buildin"
import _ "github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf/buildin"
import (
	context2 "context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/registry"

	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const (
	dialTimeout = 500 * time.Millisecond
)

var (
	endpoint = etcd.Configuration().ClusterAddresses
)

func TestInitCluster(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		etcd.Configuration().ClusterAddresses = "127.0.0.1:2379"
		etcd.Configuration().InitClusterInfo()
		if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "127.0.0.1:2379" {
			t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
		}
	})
	t.Run("not normal2", func(t *testing.T) {
		etcd.Configuration().ClusterAddresses = "127.0.0.1:2379,127.0.0.2:2379"
		etcd.Configuration().InitClusterInfo()
		if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "127.0.0.1:2379,127.0.0.2:2379" {
			t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
		}
	})

	etcd.Configuration().ClusterName = "sc-0"
	etcd.Configuration().ClusterAddresses = "sc-0=127.0.0.1:2379,127.0.0.2:2379"
	etcd.Configuration().InitClusterInfo()
	if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "127.0.0.1:2379,127.0.0.2:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
	}
	if strings.Join(etcd.Configuration().Clusters["sc-0"], ",") != "127.0.0.1:2379,127.0.0.2:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().Clusters)
	}
	etcd.Configuration().ClusterName = "sc-0"
	etcd.Configuration().ClusterAddresses = "sc-1=127.0.0.1:2379,127.0.0.2:2379,sc-2=127.0.0.3:2379"
	etcd.Configuration().InitClusterInfo()
	if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
	}
	if strings.Join(etcd.Configuration().Clusters["sc-1"], ",") != "127.0.0.1:2379,127.0.0.2:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().Clusters)
	}
	if strings.Join(etcd.Configuration().Clusters["sc-2"], ",") != "127.0.0.3:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().Clusters)
	}
	etcd.Configuration().ClusterName = "sc-0"
	etcd.Configuration().ClusterAddresses = "sc-0=127.0.0.1:2379,sc-1=127.0.0.3:2379,127.0.0.4:2379"
	etcd.Configuration().InitClusterInfo()
	if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "127.0.0.1:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
	}
	if strings.Join(etcd.Configuration().Clusters["sc-1"], ",") != "127.0.0.3:2379,127.0.0.4:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().Clusters)
	}
	etcd.Configuration().ClusterName = "sc-0"
	etcd.Configuration().ManagerAddress = "127.0.0.1:2379,127.0.0.2:2379"
	etcd.Configuration().ClusterAddresses = "sc-0=127.0.0.1:30100,sc-1=127.0.0.2:30100"
	etcd.Configuration().InitClusterInfo()
	if strings.Join(etcd.Configuration().RegistryAddresses(), ",") != "127.0.0.1:2379,127.0.0.2:2379" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().RegistryAddresses())
	}
	if strings.Join(etcd.Configuration().Clusters["sc-1"], ",") != "127.0.0.2:30100" {
		t.Fatalf("TestInitCluster failed, %v", etcd.Configuration().Clusters)
	}
}

func TestEtcdClient(t *testing.T) {
	etcd.Configuration().ClusterName = ""
	etcd.Configuration().ManagerAddress = ""
	etcd.Configuration().ClusterAddresses = endpoint
	etcd.Configuration().InitClusterInfo()

	etcdc := &Client{
		Endpoints:   []string{endpoint},
		DialTimeout: dialTimeout,
	}
	err := etcdc.Initialize()
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	defer etcdc.Close()

	select {
	case <-etcdc.Ready():
	default:
		t.Fatalf("TestEtcdClient failed")
	}

	// base test
	inst := NewRegistry(registry.Options{})
	if inst == nil || strings.Index(endpoint, firstEndpoint) < 0 {
		t.Fatalf("TestEtcdClient failed, %s != %s", firstEndpoint, endpoint)
	}
	old1 := etcd.Configuration().ClusterAddresses
	old2 := etcd.Configuration().DialTimeout
	etcd.Configuration().ClusterAddresses = "x"
	etcd.Configuration().InitClusterInfo()
	etcd.Configuration().DialTimeout = dialTimeout
	inst = NewRegistry(registry.Options{})
	if inst == nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	select {
	case <-inst.(*Client).Err():
	default:
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	etcd.Configuration().ClusterAddresses = old1
	etcd.Configuration().InitClusterInfo()
	etcd.Configuration().DialTimeout = old2

	// case: etcdc do
	// put
	resp, err := etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/b"),
		registry.WithStrValue("b"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/b"))
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/a"),
		registry.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/a"),
		registry.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/a" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/a"),
		registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/c"),
		registry.WithStrValue("c"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/d"),
		registry.WithStrValue("d"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/dd"),
		registry.WithStrValue("dd"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || string(resp.Kvs[0].Value) != "d" ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || string(resp.Kvs[1].Value) != "dd" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix(), registry.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/d"),
		registry.WithPrefix(), registry.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd")) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" ||
		string(resp.Kvs[1].Key) != "/test_range/c" || string(resp.Kvs[1].Value) != "c" ||
		string(resp.Kvs[2].Key) != "/test_range/d" || string(resp.Kvs[2].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithKeyOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/c" || resp.Kvs[1].Value != nil ||
		string(resp.Kvs[2].Key) != "/test_range/d" || resp.Kvs[2].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithCountOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix paging
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || string(resp.Kvs[0].Value) != "c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" || string(resp.Kvs[1].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(), registry.WithKeyOnly(),
		registry.WithOffset(4), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/dd" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/d"), registry.WithPrefix(), registry.WithKeyOnly(),
		registry.WithOffset(0), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 2 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(), registry.WithCountOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(6), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// if offset < -1, just paging by limit
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/"), registry.WithPrefix(),
		registry.WithOffset(-2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 5 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range paging
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithKeyOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/d" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"), registry.WithCountOnly(),
		registry.WithOffset(2), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(5), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_range/a"),
		registry.WithStrEndKey("/test_range/dd"),
		registry.WithOffset(4), registry.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcdc.Do(context.Background(), registry.DEL,
		registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/dd")) // [b, d) !!!
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || len(resp.Kvs) != 2 || string(resp.Kvs[1].Key) != "/test_range/dd" {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", resp.Kvs)
	}
	// delete prefix
	resp, err = etcdc.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/"),
		registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}

	// large data
	var wg sync.WaitGroup
	for i := 0; i < registry.DefaultPageCount+1; i++ {
		wg.Add(1)
		v := strconv.Itoa(i)
		go func() {
			defer wg.Done()
			resp, err = etcdc.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_page/"+v),
				registry.WithStrValue(v))
			if err != nil || !resp.Succeeded {
				t.Fatalf("TestEtcdClient_Do failed, %#v", err)
			}
		}()
	}
	wg.Wait()
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"),
		registry.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded || resp.Count != registry.DefaultPageCount+1 ||
		len(resp.Kvs) != registry.DefaultPageCount+1 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"), registry.WithPrefix(), registry.WithDescendOrder())
	if err != nil || !resp.Succeeded || resp.Count != registry.DefaultPageCount+1 ||
		len(resp.Kvs) != registry.DefaultPageCount+1 ||
		string(resp.Kvs[0].Key) != "/test_page/999" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcdc.Do(context.Background(), registry.DEL,
		registry.WithStrKey("/test_page/"),
		registry.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), registry.GET,
		registry.WithStrKey("/test_page/"), registry.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
}

func TestEtcdClient_Compact(t *testing.T) {
	etcd := &Client{
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
	etcd := &Client{
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
		{[]byte("/test_txn/a"), registry.CmpValue, registry.CmpEqual, "a"},
	}, []registry.PluginOp{
		{Action: registry.Put, Key: []byte("/test_txn/c"), Value: []byte("c")},
		{Action: registry.Put, Key: []byte("/test_txn/d"), Value: []byte("d")},
	})
	if err != nil || resp == nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	// case: range request
	resp, err = etcd.TxnWithCmp(context.Background(), nil, []registry.CompareOp{
		{[]byte("/test_txn/c"), registry.CmpValue, registry.CmpEqual, "c"},
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
		{[]byte("/test_txn/c"), registry.CmpValue, registry.CmpEqual, "c"},
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
	etcd := &Client{
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
	etcdc := &Client{
		Endpoints:        []string{endpoint},
		DialTimeout:      dialTimeout,
		AutoSyncInterval: time.Millisecond,
	}
	err := etcdc.Initialize()
	assert.NoError(t, err)
	defer etcdc.Close()

	err = etcdc.ReOpen()
	assert.NoError(t, err)
	ctx, _ := context.WithTimeout(context.Background(), dialTimeout)
	err = etcdc.SyncMembers(ctx)
	assert.NoError(t, err)
	etcdc.Endpoints = []string{"x"}
	err = etcdc.ReOpen()
	assert.Error(t, err)

	t.Run("before check", func(t *testing.T) {
		ctx, _ = context.WithTimeout(context.Background(), dialTimeout)
		err = etcdc.SyncMembers(ctx)
		assert.NoError(t, err)
	})

	etcdc.Endpoints = []string{endpoint}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	go etcdc.healthCheckLoop(ctx)
	for {
		_, err = etcdc.Do(context.Background(), registry.GET,
			registry.WithStrKey("/test_health/"))
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	assert.NoError(t, err)

}

func TestEtcdClient_Watch(t *testing.T) {
	etcd := &Client{
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

type mockKVForPagine struct {
	rangeCount int
	countResp  *clientv3.GetResponse
	rangeResp1 *clientv3.GetResponse
	rangeResp2 *clientv3.GetResponse
}

func (m *mockKVForPagine) Put(ctx context2.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func (m *mockKVForPagine) Get(ctx context2.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	op := &clientv3.Op{}
	for _, o := range opts {
		o(op)
	}
	if op.IsCountOnly() {
		return m.countResp, nil
	}
	if m.rangeCount == 0 {
		m.rangeCount = 1
		return m.rangeResp1, nil
	}
	return m.rangeResp2, nil
}

func (m *mockKVForPagine) Delete(ctx context2.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, nil
}

func (m *mockKVForPagine) Compact(ctx context2.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (m *mockKVForPagine) Do(ctx context2.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (m *mockKVForPagine) Txn(ctx context2.Context) clientv3.Txn {
	return nil
}

// test scenario: db data decreases during paging.
func TestEtcdClient_paging(t *testing.T) {
	// key range: [startKey, endKey)
	generateGetResp := func(startKey, endKey int) *clientv3.GetResponse {
		resp := &clientv3.GetResponse{
			Count: int64(endKey - startKey),
			Header: &etcdserverpb.ResponseHeader{
				Revision: 0,
			},
			Kvs: make([]*mvccpb.KeyValue, 0),
		}
		if resp.Count <= 0 {
			return resp
		}
		for i := startKey; i < endKey; i++ {
			kvPart := &mvccpb.KeyValue{
				Key:   []byte(fmt.Sprint(i)),
				Value: []byte(""),
			}
			resp.Kvs = append(resp.Kvs, kvPart)
		}
		return resp
	}

	mockKv := &mockKVForPagine{
		rangeCount: 0,
		// if count only, return 4097 kvs
		countResp: generateGetResp(0, 4097),
		// the first paging request, return 4096 kvs
		rangeResp1: generateGetResp(0, 4096),
		// the second paging request, return 0 kv
		// meaning data decreases during paging
		rangeResp2: generateGetResp(0, 0),
	}
	c := Client{
		Client: &clientv3.Client{
			KV: mockKv,
		},
	}

	op := registry.PluginOp{
		Offset: -1,
		Limit:  registry.DefaultPageCount,
	}
	r, err := c.paging(context2.Background(), op)
	if err != nil {
		t.Fatalf("TestEtcdClient_paging failed, %#v", err)
	}
	if len(r.Kvs) <= 0 {
		t.Fatalf("TestEtcdClient_paging failed")
	}
}

func TestNewRegistry(t *testing.T) {
	etcd := &Client{
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

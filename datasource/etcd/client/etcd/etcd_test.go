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

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"

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
	inst := NewRegistry(client.Options{})
	if inst == nil || strings.Index(endpoint, firstEndpoint) < 0 {
		t.Fatalf("TestEtcdClient failed, %s != %s", firstEndpoint, endpoint)
	}
	old1 := etcd.Configuration().ClusterAddresses
	old2 := etcd.Configuration().DialTimeout
	etcd.Configuration().ClusterAddresses = "x"
	etcd.Configuration().InitClusterInfo()
	etcd.Configuration().DialTimeout = dialTimeout
	inst = NewRegistry(client.Options{})
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
	resp, err := etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_range/b"),
		client.WithStrValue("b"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/b"))
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_range/a"),
		client.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/a"),
		client.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/a" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/a"),
		client.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 1 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_range/c"),
		client.WithStrValue("c"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_range/d"),
		client.WithStrValue("d"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_range/dd"),
		client.WithStrValue("dd"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/d"),
		client.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || string(resp.Kvs[0].Value) != "d" ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || string(resp.Kvs[1].Value) != "dd" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/d"),
		client.WithPrefix(), client.WithKeyOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/d"),
		client.WithPrefix(), client.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd")) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || string(resp.Kvs[0].Value) != "b" ||
		string(resp.Kvs[1].Key) != "/test_range/c" || string(resp.Kvs[1].Value) != "c" ||
		string(resp.Kvs[2].Key) != "/test_range/d" || string(resp.Kvs[2].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd"), client.WithKeyOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 ||
		string(resp.Kvs[0].Key) != "/test_range/b" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/c" || resp.Kvs[1].Value != nil ||
		string(resp.Kvs[2].Key) != "/test_range/d" || resp.Kvs[2].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd"), client.WithCountOnly()) // [b, dd) !!!
	if err != nil || !resp.Succeeded || resp.Count != 3 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get prefix paging
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/"), client.WithPrefix(),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || string(resp.Kvs[0].Value) != "c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" || string(resp.Kvs[1].Value) != "d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/"), client.WithPrefix(), client.WithKeyOnly(),
		client.WithOffset(4), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/dd" || resp.Kvs[0].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/d"), client.WithPrefix(), client.WithKeyOnly(),
		client.WithOffset(0), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 2 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/d" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/dd" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/"), client.WithPrefix(), client.WithCountOnly(),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/"), client.WithPrefix(),
		client.WithOffset(6), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// if offset < -1, just paging by limit
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/"), client.WithPrefix(),
		client.WithOffset(-2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 5 || len(resp.Kvs) != 5 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// get range paging
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd"),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 1 ||
		string(resp.Kvs[0].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/a"),
		client.WithStrEndKey("/test_range/dd"),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" ||
		string(resp.Kvs[1].Key) != "/test_range/d" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/a"),
		client.WithStrEndKey("/test_range/dd"), client.WithKeyOnly(),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 2 ||
		string(resp.Kvs[0].Key) != "/test_range/c" || resp.Kvs[0].Value != nil ||
		string(resp.Kvs[1].Key) != "/test_range/d" || resp.Kvs[1].Value != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/a"),
		client.WithStrEndKey("/test_range/dd"), client.WithCountOnly(),
		client.WithOffset(2), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || resp.Kvs != nil {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd"),
		client.WithOffset(5), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 3 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_range/a"),
		client.WithStrEndKey("/test_range/dd"),
		client.WithOffset(4), client.WithLimit(2))
	if err != nil || !resp.Succeeded || resp.Count != 4 || len(resp.Kvs) != 0 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcdc.Do(context.Background(), client.DEL,
		client.WithStrKey("/test_range/b"),
		client.WithStrEndKey("/test_range/dd")) // [b, d) !!!
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/"),
		client.WithPrefix())
	if err != nil || !resp.Succeeded || len(resp.Kvs) != 2 || string(resp.Kvs[1].Key) != "/test_range/dd" {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", resp.Kvs)
	}
	// delete prefix
	resp, err = etcdc.Do(context.Background(), client.DEL, client.WithStrKey("/test_range/"),
		client.WithPrefix())
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET, client.WithStrKey("/test_range/"),
		client.WithPrefix())
	if err != nil || !resp.Succeeded || resp.Count != 0 {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}

	// large data
	var wg sync.WaitGroup
	for i := 0; i < client.DefaultPageCount+1; i++ {
		wg.Add(1)
		v := strconv.Itoa(i)
		go func() {
			defer wg.Done()
			resp, err = etcdc.Do(context.Background(), client.PUT, client.WithStrKey("/test_page/"+v),
				client.WithStrValue(v))
			if err != nil || !resp.Succeeded {
				t.Fatalf("TestEtcdClient_Do failed, %#v", err)
			}
		}()
	}
	wg.Wait()
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_page/"),
		client.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded || resp.Count != client.DefaultPageCount+1 ||
		len(resp.Kvs) != client.DefaultPageCount+1 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_page/"), client.WithPrefix(), client.WithDescendOrder())
	if err != nil || !resp.Succeeded || resp.Count != client.DefaultPageCount+1 ||
		len(resp.Kvs) != client.DefaultPageCount+1 ||
		string(resp.Kvs[0].Key) != "/test_page/999" {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	// delete range
	resp, err = etcdc.Do(context.Background(), client.DEL,
		client.WithStrKey("/test_page/"),
		client.WithStrEndKey("/test_page/9999"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", err)
	}
	resp, err = etcdc.Do(context.Background(), client.GET,
		client.WithStrKey("/test_page/"), client.WithPrefix())
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

	success, err := etcd.PutNoOverride(context.Background(), client.WithStrKey("/test_txn/a"))
	if err != nil || !success {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	success, err = etcd.PutNoOverride(context.Background(), client.WithStrKey("/test_txn/a"), client.WithStrValue("a"))
	if err != nil || success {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	resp, err = etcd.Txn(context.Background(), []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: client.ActionPut, Key: []byte("/test_txn/b"), Value: []byte("b")},
	})
	if err != nil || resp == nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}
	resp, err = etcd.Do(context.Background(), client.GET, client.WithStrKey("/test_txn/"),
		client.WithPrefix(), client.WithCountOnly())
	if err != nil || !resp.Succeeded || resp.Count != 2 {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}

	resp, err = etcd.TxnWithCmp(context.Background(), []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: client.ActionPut, Key: []byte("/test_txn/b"), Value: []byte("b")},
	}, []client.CompareOp{
		{[]byte("/test_txn/a"), client.CmpValue, client.CmpEqual, "a"},
	}, []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_txn/c"), Value: []byte("c")},
		{Action: client.ActionPut, Key: []byte("/test_txn/d"), Value: []byte("d")},
	})
	if err != nil || resp == nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	// case: range request
	resp, err = etcd.TxnWithCmp(context.Background(), nil, []client.CompareOp{
		{[]byte("/test_txn/c"), client.CmpValue, client.CmpEqual, "c"},
	}, []client.PluginOp{
		{Action: client.ActionGet, Key: []byte("/test_txn/a")},
		{Action: client.ActionGet, Key: []byte("/test_txn/"), Prefix: true},
	})
	if err != nil || resp == nil || resp.Succeeded || resp.Count != 3 { // a + [a,b]
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	// case: test key not exist
	resp, err = etcd.TxnWithCmp(context.Background(), []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_txn/a"), Value: []byte("a")},
		{Action: client.ActionPut, Key: []byte("/test_txn/b"), Value: []byte("b")},
	}, []client.CompareOp{
		{[]byte("/test_txn/c"), client.CmpValue, client.CmpEqual, "c"},
	}, []client.PluginOp{
		{Action: client.ActionDelete, Key: []byte("/test_txn/"), Prefix: true},
	})
	if err != nil || resp == nil || resp.Succeeded {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	resp, err = etcd.Do(context.Background(), client.GET, client.WithStrKey("/test_txn/"),
		client.WithPrefix(), client.WithCountOnly())
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
		_, err = etcdc.Do(context.Background(), client.GET,
			client.WithStrKey("/test_health/"))
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
		resp, err := etcd.Do(context.Background(), client.DEL, client.WithStrKey("/test_watch/"),
			client.WithPrefix())
		if err != nil || !resp.Succeeded {
			t.Fatalf("TestEtcdClient_Do failed, %#v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = etcd.Watch(ctx, client.WithStrKey("/test_watch/a"))
	if err != nil {
		t.Fatalf("TestEtcdClient failed, %#v", err)
	}

	ch := make(chan struct{})
	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), client.WithStrKey("/test_watch/a"),
			client.WithWatchCallback(func(message string, evt *client.PluginResponse) error {
				if evt.Count != 1 || len(evt.Kvs) != 1 || evt.Action != client.ActionPut ||
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
	resp, err := etcd.Do(context.Background(), client.PUT, client.WithStrKey("/test_watch/a"),
		client.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), client.WithStrKey("/test_watch/"),
			client.WithPrefix(),
			client.WithWatchCallback(func(message string, evt *client.PluginResponse) error {
				equalA := evt.Action == client.ActionPut && string(evt.Kvs[0].Key) == "/test_watch/a" && string(evt.Kvs[0].Value) == "a"
				equalB := evt.Action == client.ActionPut && string(evt.Kvs[1].Key) == "/test_watch/b" && string(evt.Kvs[0].Value) == "b"
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
	resp, err = etcd.Txn(context.Background(), []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_watch/a"), Value: []byte("a")},
		{Action: client.ActionPut, Key: []byte("/test_watch/b"), Value: []byte("b")},
	})
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	// diff action type will be split
	go func() {
		defer func() { ch <- struct{}{} }()
		var times = 3
		err = etcd.Watch(context.Background(), client.WithStrKey("/test_watch/"),
			client.WithPrefix(),
			client.WithWatchCallback(func(message string, evt *client.PluginResponse) error {
				equalA := evt.Action == client.ActionDelete && string(evt.Kvs[0].Key) == "/test_watch/a" && evt.Kvs[0].Value == nil
				equalB := evt.Action == client.ActionPut && string(evt.Kvs[0].Key) == "/test_watch/b" && string(evt.Kvs[0].Value) == "b"
				equalC := evt.Action == client.ActionPut && string(evt.Kvs[0].Key) == "/test_watch/c" && string(evt.Kvs[0].Value) == "c"
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
	resp, err = etcd.Txn(context.Background(), []client.PluginOp{
		{Action: client.ActionPut, Key: []byte("/test_watch/c"), Value: []byte("c")},
		{Action: client.ActionDelete, Key: []byte("/test_watch/a"), Value: []byte("a")},
		{Action: client.ActionPut, Key: []byte("/test_watch/b"), Value: []byte("b")},
	})
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	<-ch

	// watch with rev
	resp, err = etcd.Do(context.Background(), client.DEL, client.WithStrKey("/test_watch/c"),
		client.WithStrValue("a"))
	if err != nil || !resp.Succeeded {
		t.Fatalf("TestEtcdClient_Do failed, %#v", err)
	}
	rev := resp.Revision
	go func() {
		defer func() { ch <- struct{}{} }()
		err = etcd.Watch(context.Background(), client.WithStrKey("/test_watch/"),
			client.WithPrefix(),
			client.WithRev(rev),
			client.WithWatchCallback(func(message string, evt *client.PluginResponse) error {
				if evt.Count != 1 || len(evt.Kvs) != 1 || evt.Action != client.ActionDelete ||
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
		err = etcd.Watch(context.Background(), client.WithStrKey("/test_watch/"),
			client.WithPrefix(), client.WithPrevKv(),
			client.WithWatchCallback(func(message string, evt *client.PluginResponse) error {
				if len(evt.Kvs) != 1 || evt.Action != client.ActionDelete ||
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
	resp, err = etcd.Do(context.Background(), client.DEL, client.WithStrKey("/test_watch/b"))
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

	op := client.PluginOp{
		Offset: -1,
		Limit:  client.DefaultPageCount,
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

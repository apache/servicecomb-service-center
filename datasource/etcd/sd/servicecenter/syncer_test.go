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

package servicecenter

import (
	"fmt"
	"testing"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func TestClusterIndexer_Sync(t *testing.T) {
	syncer := &Syncer{}
	c := kvstore.NewKvCache("test", kvstore.NewOptions())
	cfg := kvstore.NewOptions()
	sccacher := NewServiceCenterCacher(cfg, c)
	arr := dump.MicroserviceIndexSlice{}

	// case: sync empty data
	cfg.WithEventFunc(func(kvstore.Event) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*dump.KV, dump.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: CREATE
	cfg.WithEventFunc(func(evt kvstore.Event) {
		if evt.Type != pb.EVT_CREATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	arr.SetValue(&dump.KV{Key: "/a", Value: "a", Rev: 1, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*dump.KV, dump.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: UPDATE
	cfg.WithEventFunc(func(evt kvstore.Event) {
		if evt.Type != pb.EVT_UPDATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	arr.SetValue(&dump.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *dump.KV, _ dump.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: UPDATE the same one
	cfg.WithEventFunc(func(evt kvstore.Event) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*dump.KV, dump.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: conflict but not print log
	cfg.WithEventFunc(func(evt kvstore.Event) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	arr.SetValue(&dump.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	arr.SetValue(&dump.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "b"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, syncer.logConflictFunc)

	// case: conflict and print log
	func() {
		defer log.Recover()
		cfg.WithEventFunc(func(evt kvstore.Event) {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		})
		arr = dump.MicroserviceIndexSlice{}
		arr.SetValue(&dump.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
		arr.SetValue(&dump.KV{Key: "/a", Value: "ab", Rev: 2, ClusterName: "b"})
		syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, syncer.logConflictFunc)
		// '/a' is incorrect key and logConflictFunc will be excepted to panic here
		t.Fatalf("TestClusterIndexer_Sync failed")
	}()

	// case: some cluster err and do not overwrite the cache
	cfg.WithEventFunc(func(evt kvstore.Event) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	arr.SetValue(&dump.KV{Key: "/a", Value: "ab", Rev: 3, ClusterName: "b"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, map[string]error{"a": fmt.Errorf("error")}, func(kv *dump.KV, _ dump.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: DELETE but the cluster err
	cfg.WithEventFunc(func(evt kvstore.Event) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	syncer.checkWithConflictHandleFunc(sccacher, &arr, map[string]error{"a": fmt.Errorf("error")}, func(kv *dump.KV, _ dump.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: DELETE
	cfg.WithEventFunc(func(evt kvstore.Event) {
		fmt.Println(evt)
		if evt.Type != pb.EVT_DELETE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
	})
	arr = dump.MicroserviceIndexSlice{}
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *dump.KV, _ dump.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: CREATE again and set cluster to local cluster name
	cfg.WithEventFunc(func(evt kvstore.Event) {
		if evt.Type != pb.EVT_CREATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	configuration := etcd.Configuration()
	arr.SetValue(&dump.KV{Key: "/a", Value: "a", Rev: 1, ClusterName: configuration.ClusterName})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*dump.KV, dump.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: UPDATE but skip local cluster
	cfg.WithEventFunc(func(evt kvstore.Event) {
		if evt.Type != pb.EVT_UPDATE && evt.KV.Value != "aa" {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = dump.MicroserviceIndexSlice{}
	arr.SetValue(&dump.KV{Key: "/a", Value: "x", Rev: 2, ClusterName: configuration.ClusterName})
	arr.SetValue(&dump.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *dump.KV, _ dump.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})
}

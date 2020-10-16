// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicecenter

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	registry2 "github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/pkg/log"
	model2 "github.com/apache/servicecomb-service-center/pkg/model"
	"testing"
)

func TestClusterIndexer_Sync(t *testing.T) {
	syncer := &Syncer{}
	cache := registry2.NewKvCache("test", registry2.Configure())
	cfg := registry2.Configure()
	sccacher := NewServiceCenterCacher(cfg, cache)
	arr := model2.MicroserviceIndexSlice{}

	// case: sync empty data
	cfg.WithEventFunc(func(registry2.KvEvent) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*model2.KV, model2.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: CREATE
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		if evt.Type != registry2.EVT_CREATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "a", Rev: 1, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*model2.KV, model2.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: UPDATE
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		if evt.Type != registry2.EVT_UPDATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *model2.KV, _ model2.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: UPDATE the same one
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*model2.KV, model2.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: conflict but not print log
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	arr.SetValue(&model2.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "b"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, syncer.logConflictFunc)

	// case: conflict and print log
	func() {
		defer log.Recover()
		cfg.WithEventFunc(func(evt registry2.KvEvent) {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		})
		arr = model2.MicroserviceIndexSlice{}
		arr.SetValue(&model2.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
		arr.SetValue(&model2.KV{Key: "/a", Value: "ab", Rev: 2, ClusterName: "b"})
		syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, syncer.logConflictFunc)
		// '/a' is incorrect key and logConflictFunc will be excepted to panic here
		t.Fatalf("TestClusterIndexer_Sync failed")
	}()

	// case: some cluster err and do not overwrite the cache
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "ab", Rev: 3, ClusterName: "b"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, map[string]error{"a": fmt.Errorf("error")}, func(kv *model2.KV, _ model2.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: DELETE but the cluster err
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	syncer.checkWithConflictHandleFunc(sccacher, &arr, map[string]error{"a": fmt.Errorf("error")}, func(kv *model2.KV, _ model2.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: DELETE
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		fmt.Println(evt)
		if evt.Type != registry2.EVT_DELETE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
	})
	arr = model2.MicroserviceIndexSlice{}
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *model2.KV, _ model2.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})

	// case: CREATE again and set cluster to local cluster name
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		if evt.Type != registry2.EVT_CREATE {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "a", Rev: 1, ClusterName: etcd.Configuration().ClusterName})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(*model2.KV, model2.Getter, int) {
		t.Fatalf("TestClusterIndexer_Sync failed")
	})

	// case: UPDATE but skip local cluster
	cfg.WithEventFunc(func(evt registry2.KvEvent) {
		if evt.Type != registry2.EVT_UPDATE && evt.KV.Value != "aa" {
			t.Fatalf("TestClusterIndexer_Sync failed, %v", evt)
		}
		fmt.Println(evt)
	})
	arr = model2.MicroserviceIndexSlice{}
	arr.SetValue(&model2.KV{Key: "/a", Value: "x", Rev: 2, ClusterName: etcd.Configuration().ClusterName})
	arr.SetValue(&model2.KV{Key: "/a", Value: "aa", Rev: 2, ClusterName: "a"})
	syncer.checkWithConflictHandleFunc(sccacher, &arr, nil, func(kv *model2.KV, _ model2.Getter, _ int) {
		t.Fatalf("TestClusterIndexer_Sync failed %v", kv)
	})
}

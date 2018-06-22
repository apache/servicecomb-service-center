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
package backend

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"math/rand"
	"testing"
)

func BenchmarkFilter(b *testing.B) {
	inst := &pb.MicroServiceInstance{
		HealthCheck: &pb.HealthCheck{
			Interval: 4,
			Times:    0,
		},
	}
	v, _ := json.Marshal(inst)

	cacher := &KvCacher{Cfg: DefaultConfig().WithParser(InstanceParser)}

	n := 300 * 1000 // 30w
	cache := NewKvCache(cacher, n)
	items := make([]*mvccpb.KeyValue, 0, n)
	for ; n > 0; n-- {
		k := fmt.Sprintf("/%d", n)
		if n <= 10*1000 {
			// create
			items = append(items, &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: int64(rand.Int()),
			})
		} else if n > 100*1000 && n <= 20*1000 {
			// update
			cache.store[k] = &KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       inst,
				ModRevision: 1,
			}
			items = append(items, &mvccpb.KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       v,
				ModRevision: int64(rand.Int()),
			})
		} else {
			// delete
			cache.store[k] = &KeyValue{
				Key:         util.StringToBytesWithNoCopy(k),
				Value:       inst,
				ModRevision: 1,
			}
		}
	}
	cacher.cache = cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacher.filter(1, items)
	}
	b.ReportAllocs()

	// TODO bad performance!!!
	//20	  82367261 ns/op	37964987 B/op	   80132 allocs/op
}

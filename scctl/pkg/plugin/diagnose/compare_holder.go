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

package diagnose

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/value"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/coreos/etcd/mvcc/mvccpb"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
)

type CompareHolder interface {
	Compare() *CompareResult
}

type DataStore struct {
	Data       []*mvccpb.KeyValue
	DataParser value.Parser
}

func (d *DataStore) ForEach(f func(i int, v *dump.KV) bool) {
	for i, kv := range d.Data {
		obj, _ := d.DataParser.Unmarshal(kv.Value)
		if !f(i, &dump.KV{Key: string(kv.Key), Rev: kv.ModRevision, Value: obj}) {
			return
		}
	}
}

type CompareResult struct {
	Name    string
	Results map[int][]string
}

type abstractCompareHolder struct {
	Cache        dump.Getter
	DataStore    *DataStore
	MismatchFunc func(v *dump.KV) string
}

func (h *abstractCompareHolder) toMap(getter dump.Getter) map[string]*dump.KV {
	m := make(map[string]*dump.KV)
	getter.ForEach(func(i int, v *dump.KV) bool {
		m[v.Key] = v
		return true
	})
	return m
}

func (h *abstractCompareHolder) Compare() *CompareResult {
	result := &CompareResult{
		Results: make(map[int][]string),
	}
	leftCh := make(chan map[string]*dump.KV, 2)
	rightCh := make(chan map[string]*dump.KV, 2)

	var (
		add    []string
		update []string
		del    []string
	)

	gopool.New(gopool.Configure().Workers(5)).
		Do(func(_ context.Context) {
			left := h.toMap(h.Cache)
			leftCh <- left
			leftCh <- left
		}).
		Do(func(_ context.Context) {
			right := h.toMap(h.DataStore)
			rightCh <- right
			rightCh <- right
		}).
		Do(func(_ context.Context) {
			left := <-leftCh
			right := <-rightCh
			// add or update
			for lk, lkv := range left {
				rkv, ok := right[lk]
				if !ok {
					add = append(add, h.MismatchFunc(lkv))
					continue
				}
				if rkv.Rev != lkv.Rev {
					update = append(update, h.MismatchFunc(lkv))
				}
			}
		}).
		Do(func(_ context.Context) {
			left := <-leftCh
			right := <-rightCh
			// delete
			for rk, rkv := range right {
				if _, ok := left[rk]; !ok {
					del = append(del, h.MismatchFunc(rkv))
				}
			}
		}).
		Done()

	if len(add) > 0 {
		result.Results[greater] = add
	}
	if len(update) > 0 {
		result.Results[mismatch] = update
	}
	if len(del) > 0 {
		result.Results[less] = del
	}
	return result
}

type ServiceCompareHolder struct {
	*abstractCompareHolder
	Cache dump.MicroserviceSlice
	Kvs   []*mvccpb.KeyValue
}

func (h *ServiceCompareHolder) Compare() *CompareResult {
	h.abstractCompareHolder = &abstractCompareHolder{
		Cache: &h.Cache, DataStore: &DataStore{Data: h.Kvs, DataParser: value.ServiceParser}, MismatchFunc: h.toName,
	}
	r := h.abstractCompareHolder.Compare()
	r.Name = service
	return r
}
func (h *ServiceCompareHolder) toName(kv *dump.KV) string {
	s, ok := kv.Value.(*pb.MicroService)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s/%s/%s(%s)", s.AppId, s.ServiceName, s.Version, s.ServiceId)
}

type InstanceCompareHolder struct {
	*abstractCompareHolder
	Cache dump.InstanceSlice
	Kvs   []*mvccpb.KeyValue
}

func (h *InstanceCompareHolder) Compare() *CompareResult {
	h.abstractCompareHolder = &abstractCompareHolder{
		Cache: &h.Cache, DataStore: &DataStore{Data: h.Kvs, DataParser: value.InstanceParser}, MismatchFunc: h.toName,
	}
	r := h.abstractCompareHolder.Compare()
	r.Name = instance
	return r
}
func (h *InstanceCompareHolder) toName(kv *dump.KV) string {
	s, ok := kv.Value.(*pb.MicroServiceInstance)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%v(%s/%s)", s.Endpoints, s.ServiceId, s.InstanceId)
}

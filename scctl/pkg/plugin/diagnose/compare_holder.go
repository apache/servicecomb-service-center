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
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type CompareHolder interface {
	Compare() *CompareResult
}

type DataStore struct {
	Data       []*mvccpb.KeyValue
	DataParser pb.Parser
}

func (d *DataStore) ForEach(f func(i int, v *model.KV) bool) {
	for i, kv := range d.Data {
		obj, _ := d.DataParser.Unmarshal(kv.Value)
		if !f(i, &model.KV{Key: string(kv.Key), Rev: kv.ModRevision, Value: obj}) {
			return
		}
	}
}

type CompareResult struct {
	Name    string
	Results map[int][]string
}

type abstractCompareHolder struct {
	Cache        model.Getter
	DataStore    *DataStore
	MismatchFunc func(v *model.KV) string
}

func (h *abstractCompareHolder) toMap(getter model.Getter) map[string]*model.KV {
	m := make(map[string]*model.KV)
	getter.ForEach(func(i int, v *model.KV) bool {
		m[v.Key] = v
		return true
	})
	return m
}

func (h *abstractCompareHolder) Compare() *CompareResult {
	result := &CompareResult{
		Results: make(map[int][]string),
	}
	leftCh := make(chan map[string]*model.KV, 2)
	rightCh := make(chan map[string]*model.KV, 2)

	var (
		add    []string
		update []string
		del    []string
	)

	gopool.New(context.Background(), gopool.Configure().Workers(3)).
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
	Cache model.MicroserviceSlice
	Kvs   []*mvccpb.KeyValue
}

func (h *ServiceCompareHolder) Compare() *CompareResult {
	h.abstractCompareHolder = &abstractCompareHolder{
		Cache: &h.Cache, DataStore: &DataStore{Data: h.Kvs, DataParser: pb.ServiceParser}, MismatchFunc: h.toName,
	}
	r := h.abstractCompareHolder.Compare()
	r.Name = service
	return r
}
func (h *ServiceCompareHolder) toName(kv *model.KV) string {
	s, ok := kv.Value.(*pb.MicroService)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s/%s/%s(%s)", s.AppId, s.ServiceName, s.Version, s.ServiceId)
}

type InstanceCompareHolder struct {
	*abstractCompareHolder
	Cache model.InstanceSlice
	Kvs   []*mvccpb.KeyValue
}

func (h *InstanceCompareHolder) Compare() *CompareResult {
	h.abstractCompareHolder = &abstractCompareHolder{
		Cache: &h.Cache, DataStore: &DataStore{Data: h.Kvs, DataParser: pb.InstanceParser}, MismatchFunc: h.toName,
	}
	r := h.abstractCompareHolder.Compare()
	r.Name = instance
	return r
}
func (h *InstanceCompareHolder) toName(kv *model.KV) string {
	s, ok := kv.Value.(*pb.MicroServiceInstance)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%v(%s/%s)", s.Endpoints, s.ServiceId, s.InstanceId)
}
